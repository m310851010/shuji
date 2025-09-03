package data_import

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"shuji/db"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// 文件类型常量
const (
	TableName1       = "规上企业"
	TableName2       = "其他单位"
	TableName3       = "新上项目"
	TableAttachment2 = "区域综合"
)

// 加密状态常量 - 这些值需要在运行时初始化
var (
	EncryptedZero = "" // 加密后的"0"
	EncryptedOne  = "" // 加密后的"1"
)

// initEncryptedConstants 初始化加密常量（包级别初始化）
func initEncryptedConstants(app App) {
	if EncryptedZero == "" {
		encryptedZero, _ := app.SM4Encrypt("0")
		EncryptedZero = encryptedZero
	}

	if EncryptedOne == "" {
		encryptedOne, _ := app.SM4Encrypt("1")
		EncryptedOne = encryptedOne
	}
}

// getDecryptedStatus 获取解密后的状态值
func (s *DataImportService) getDecryptedStatus(encryptedValue interface{}) string {
	if encryptedValue == nil {
		return ""
	}

	// 直接比较加密值，避免解密操作
	if encryptedValue == EncryptedZero {
		return "0"
	}
	if encryptedValue == EncryptedOne {
		return "1"
	}
	return ""
}

// ValidationError 验证错误结构
type ValidationError struct {
	RowNumber int      `json:"row_number"` // 错误行号
	Message   string   `json:"message"`    // 错误信息
	Cells     []string `json:"cells"`      // 涉及到的单元格位置，如["A1", "B1", "C1"]
}

const (
	TableType1           = "table1"
	TableType2           = "table2"
	TableType3           = "table3"
	TableTypeAttachment2 = "attachment2"
)

// App 应用接口，用于访问数据库和其他功能
type App interface {
	GetDB() *db.Database
	GetAreaConfig() db.QueryResult
	GetAreaStr() string
	GetEnhancedAreaConfig() db.QueryResult
	InsertImportRecord(fileName, fileType, importState, describe string)
	IsEnterpriseListExist() (bool, error)
	GetEnterpriseInfoByCreditCode(creditCode string) db.QueryResult
	CacheFileExists(tableType string, fileName string) db.QueryResult
	CopyFileToCache(tableType string, fileName string) db.QueryResult
	IsEquipmentListExist() (bool, error)
	GetEquipmentByCreditCode(creditCode string) db.QueryResult
	SM4Encrypt(plaintext string) (string, error)
	SM4Decrypt(ciphertext string) (string, error)
	GetCachePath(tableType string) string
	GetCurrentOSUser() string
	GetCtx() context.Context
	GetDBPassword() string
}

// DataImportService 数据导入服务
type DataImportService struct {
	app App
}

// NewDataImportService 创建数据导入服务
func NewDataImportService(app App) *DataImportService {
	// 确保加密常量已初始化
	initEncryptedConstants(app)

	return &DataImportService{
		app: app,
	}
}

// convertToInterfaceSlice 将字符串切片转换为接口切片
func (s *DataImportService) convertToInterfaceSlice(strSlice []string) []interface{} {
	result := make([]interface{}, len(strSlice))
	for i, v := range strSlice {
		result[i] = v
	}
	return result
}

// cleanCellValue 清理单元格值
func (s *DataImportService) cleanCellValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\t", "")
	return value
}

// validateRequiredFieldsOrdered 批量校验必填字段（保证顺序）
func (s *DataImportService) validateRequiredFieldsOrdered(data map[string]interface{}, fieldOrder []string, fields map[string]string, rowNo int) []string {
	errors := []string{}
	for _, fieldName := range fieldOrder {
		if displayName, exists := fields[fieldName]; exists {
			fieldErrors := s.validateRequiredField(data, fieldName, displayName, rowNo)
			errors = append(errors, fieldErrors...)
		}
	}
	return errors
}

// validateRequiredField 校验必填字段的通用函数
func (s *DataImportService) validateRequiredField(data map[string]interface{}, fieldName, fieldDisplayName string, rowNo int) []string {
	errors := []string{}
	if value, ok := data[fieldName].(string); !ok || value == "" {
		errors = append(errors, fmt.Sprintf("第%d行：%s不能为空", rowNo, fieldDisplayName))
	}
	return errors
}

// validateRequiredFields 批量校验必填字段
func (s *DataImportService) validateRequiredFields(data map[string]interface{}, fields map[string]string, rowNo int) []string {
	errors := []string{}
	for fieldName, displayName := range fields {
		fieldErrors := s.validateRequiredField(data, fieldName, displayName, rowNo)
		errors = append(errors, fieldErrors...)
	}
	return errors
}

// validateTable3TimeFields 校验附表3时间字段（拟投产时间和实际投产时间至少选择其一）
func (s *DataImportService) validateTable3TimeFields(data map[string]interface{}, excelRowNum int) []string {
	errors := []string{}

	scheduledTime, _ := data["scheduled_time"].(string)
	actualTime, _ := data["actual_time"].(string)

	// 检查拟投产时间和实际投产时间是否都为空
	if scheduledTime == "" && actualTime == "" {
		errors = append(errors, fmt.Sprintf("第%d行：拟投产时间和实际投产时间至少选择其一填写", excelRowNum))
	}

	return errors
}

// getStringValue 获取字符串值的通用函数
func (s *DataImportService) getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// validateEnterpriseAndCreditCode 企业名称和统一社会信用代码校验（从表1、表2提取的通用逻辑）
func (s *DataImportService) validateEnterpriseAndCreditCode(data map[string]interface{}, unitRowNum, regionRowNum int) []string {

	errors := []string{}
	unitName, _ := data["unit_name"].(string)
	creditCode, _ := data["credit_code"].(string)

	provinceName := s.getStringValue(data["province_name"])
	cityName := s.getStringValue(data["city_name"])
	countryName := s.getStringValue(data["country_name"])

	if unitName != "" && creditCode != "" {

		// 第一步: 调用s.app.IsEnterpriseListExist(), 检查企业清单是否存在, 不存在直接校验通过
		hasEnterpriseList, err := s.app.IsEnterpriseListExist()
		if err != nil {
			errors = append(errors, fmt.Sprintf("第%d行：企业清单检查失败", unitRowNum))
			return errors
		}

		if hasEnterpriseList {
			// 第二步: 如果企业清单存在, 调用s.app.GetEnterpriseNameByCreditCode,检查统一信用代码是否有对应的企业名称, 未查询到企业名称校验失败
			result := s.app.GetEnterpriseInfoByCreditCode(creditCode)
			if !result.Ok || result.Data == nil {
				errors = append(errors, fmt.Sprintf("第%d行：%s企业，统一信用代码%s未在清单表里", unitRowNum, unitName, creditCode))
				return errors
			}

			if provinceName != "" && cityName != "" && countryName != "" {
				enterpriseInfo := result.Data.(map[string]interface{})
				dbUnitName := enterpriseInfo["unit_name"].(string)
				dbProvinceName := enterpriseInfo["province_name"].(string)
				dbCityName := enterpriseInfo["city_name"].(string)
				dbCountryName := enterpriseInfo["country_name"].(string)

				// 如果查询到企业名了，比较企业名称是否相同
				if dbUnitName != unitName {
					errors = append(errors, fmt.Sprintf("第%d行：统一信用代码%s和导入的企业名称不对应", unitRowNum, creditCode))
					return errors
				}

				// 如果查询到企业名了，比较省市县是否相同
				errors = s.checkRegionMatch(provinceName, cityName, countryName, dbProvinceName, dbCityName, dbCountryName, regionRowNum)
			}

		} else if provinceName != "" && cityName != "" && countryName != "" {
			errors = s.validateRegionOnly(data, regionRowNum)
		}
	}

	return errors
}

// validateEquipmentAndCreditCode 企业名称和统一社会信用代码校验（从表1、表2提取的通用逻辑）
func (s *DataImportService) validateEquipmentAndCreditCode(data map[string]interface{}, unitRowNum, regionRowNum int) []string {

	errors := []string{}
	unitName, _ := data["unit_name"].(string)
	creditCode, _ := data["credit_code"].(string)

	provinceName := s.getStringValue(data["province_name"])
	cityName := s.getStringValue(data["city_name"])
	countryName := s.getStringValue(data["country_name"])

	if unitName != "" && creditCode != "" {

		// 第一步: 调用s.app.IsEquipmentListExist(), 检查企业清单是否存在, 不存在直接校验通过
		hasEquipmentList, err := s.app.IsEquipmentListExist()
		if err != nil {
			errors = append(errors, fmt.Sprintf("第%d行：企业清单检查失败", unitRowNum))
			return errors
		}

		if hasEquipmentList {
			// 第二步: 如果企业清单存在, 调用s.app.GetEquipmentNameByCreditCode,检查统一信用代码是否有对应的企业名称, 未查询到企业名称校验失败
			result := s.app.GetEquipmentByCreditCode(creditCode)
			if !result.Ok || result.Data == nil {
				errors = append(errors, fmt.Sprintf("第%d行：%s企业，统一信用代码%s未在清单表里", unitRowNum, unitName, creditCode))
				return errors
			}

			if provinceName != "" && cityName != "" && countryName != "" {
				equipmentInfo := result.Data.(map[string]interface{})
				dbUnitName := equipmentInfo["unit_name"].(string)
				dbProvinceName := equipmentInfo["province_name"].(string)
				dbCityName := equipmentInfo["city_name"].(string)
				dbCountryName := equipmentInfo["country_name"].(string)

				// 如果查询到企业名了，比较企业名称是否相同
				if dbUnitName != unitName {
					errors = append(errors, fmt.Sprintf("第%d行：统一信用代码%s和导入的企业名称不对应", unitRowNum, creditCode))
					return errors
				}

				// 如果查询到企业名了，比较省市县是否相同
				errors = s.checkRegionMatch(provinceName, cityName, countryName, dbProvinceName, dbCityName, dbCountryName, regionRowNum)
			}

		} else if provinceName != "" && cityName != "" && countryName != "" {
			errors = s.validateRegionOnly(data, regionRowNum)
		}
	}

	return errors
}

// checkRegionMatch 检查省市县是否匹配的公共函数
func (s *DataImportService) checkRegionMatch(provinceName, cityName, countryName, expectedProvince, expectedCity, expectedCountry string, excelRowNum int) []string {
	errors := []string{}

	// 1.检查省是否匹配, 失败,返回
	if expectedProvince != "" && provinceName != expectedProvince {
		errors = append(errors, fmt.Sprintf("第%d行：导入的数据对应的区域与当前配置的区域[省（市、区）]不符", excelRowNum))
		return errors
	}
	// 2.检查市, city_name有值时,是否匹配, 无值时返回成功
	if expectedCity != "" && cityName != expectedCity {
		errors = append(errors, fmt.Sprintf("第%d行：导入的数据对应的区域与当前配置的区域[地市（州）]不符", excelRowNum))
		return errors
	}
	// 3.检查县, country_name有值时,是否匹配, 无值时返回成功
	if expectedCountry != "" && countryName != expectedCountry {
		errors = append(errors, fmt.Sprintf("第%d行：导入的数据对应的区域与当前配置的区域[县（区）]不符", excelRowNum))
		return errors
	}

	return errors
}

// validateRegionOnly 仅省市县校验（从表3、附件2提取的通用逻辑）
func (s *DataImportService) validateRegionOnly(data map[string]interface{}, excelRowNum int) []string {

	provinceName, _ := data["province_name"].(string)
	cityName, _ := data["city_name"].(string)
	countryName, _ := data["country_name"].(string)

	// 调用s.app.GetAreaConfig(), 获取province_name, city_name, country_name
	areaResult := s.app.GetAreaConfig()
	row := areaResult.Data.(map[string]interface{})
	expectedProvince := s.getStringValue(row["province_name"])
	expectedCity := s.getStringValue(row["city_name"])
	expectedCountry := s.getStringValue(row["country_name"])

	// 使用公共函数检查省市县是否匹配
	return s.checkRegionMatch(provinceName, cityName, countryName, expectedProvince, expectedCity, expectedCountry, excelRowNum)
}

// GetAreaConfig 获取区域配置 - 直接调用App的方法
func (s *DataImportService) GetAreaConfig() db.QueryResult {
	return s.app.GetAreaConfig()
}

// decryptValue 解密数值
func (s *DataImportService) decryptValue(value interface{}) string {
	if value == nil {
		return ""
	}

	encryptedValue := s.getStringValue(value)
	if encryptedValue == "" {
		return ""
	}

	decryptedValue, err := s.app.SM4Decrypt(encryptedValue)
	if err != nil {
		return ""
	}

	return decryptedValue
}

// 导入状态常量
const (
	ImportStateSuccess = "导入成功"
	ImportStateFailed  = "导入失败"
)

// 表名常量
const (
	TableEnterpriseCoalConsumptionMain    = "enterprise_coal_consumption_main"
	TableEnterpriseCoalConsumptionUsage   = "enterprise_coal_consumption_usage"
	TableEnterpriseCoalConsumptionEquip   = "enterprise_coal_consumption_equip"
	TableCriticalCoalEquipmentConsumption = "critical_coal_equipment_consumption"
	TableFixedAssetsInvestmentProject     = "fixed_assets_investment_project"
	TableCoalConsumptionReport            = "coal_consumption_report"
)

// 字段显示名称映射
var (
	// 附表1必填字段
	Table1RequiredFields = map[string]string{
		"stat_date":                      "年份",
		"unit_name":                      "单位名称",
		"credit_code":                    "统一社会信用代码",
		"trade_a":                        "行业门类",
		"trade_b":                        "行业大类",
		"trade_c":                        "行业中类",
		"province_name":                  "单位所在省/市/区",
		"city_name":                      "单位所在地市",
		"country_name":                   "单位所在区县",
		"tel":                            "联系电话",
		"annual_energy_equivalent_value": "年综合能耗当量值（万吨标准煤，含原料用能）",
		"annual_energy_equivalent_cost":  "年综合能耗等价值（万吨标准煤，含原料用能）",
	}

	// 附表2必填字段
	Table2RequiredFields = map[string]string{
		"coal_type":     "类型",
		"coal_no":       "编号",
		"usage_time":    "累计使用时间",
		"design_life":   "设计年限",
		"capacity_unit": "容量单位",
		"capacity":      "容量",
	}

	// 附表3必填字段
	Table3RequiredFields = map[string]string{
		"project_name":              "项目名称",
		"project_code":              "项目代码",
		"document_number":           "审查意见文号",
		"construction_unit":         "建设单位",
		"main_construction_content": "主要建设内容",
		"province_name":             "项目所在省",
		"city_name":                 "项目所在市",
		"country_name":              "项目所在县",
		"trade_a":                   "所属行业大类",
		"trade_c":                   "所属行业小类",
		"examination_approval_time": "节能审查批复时间",
		"examination_authority":     "节能审查机关",
		"equivalent_value":          "当量值",
		"equivalent_cost":           "等价值",
	}
)

// generateUUID 生成UUID
func (s *DataImportService) generateUUID() string {
	return uuid.New().String()
}

// parseFloat 解析浮点数
func (s *DataImportService) parseFloat(value string) float64 {
	if value == "" {
		return 0
	}

	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return result
}

// isIntegerEqual 使用整数计算判断两个float64是否相等（乘以1000转换为整数计算）
func (s *DataImportService) isIntegerEqual(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA == intB
}

// isIntegerLessThan 使用整数计算判断a是否小于b（乘以1000转换为整数计算）
func (s *DataImportService) isIntegerLessThan(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA < intB
}

// isIntegerGreaterThan 使用整数计算判断a是否大于b（乘以1000转换为整数计算）
func (s *DataImportService) isIntegerGreaterThan(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA > intB
}

// isIntegerLessThanOrEqual 使用整数计算判断a是否小于等于b（乘以1000转换为整数计算）
func (s *DataImportService) isIntegerLessThanOrEqual(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA <= intB
}

// isIntegerGreaterThanOrEqual 使用整数计算判断a是否大于等于b（乘以1000转换为整数计算）
func (s *DataImportService) isIntegerGreaterThanOrEqual(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA >= intB
}

// isIntegerInteger 使用整数计算判断float64是否为整数（乘以1000转换为整数计算）
func (s *DataImportService) isIntegerInteger(value float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intValue := int64(value * 1000)
	// 判断是否能被1000整除
	return intValue%1000 == 0
}

// 精度安全的算术运算函数（基于整数计算）
// addFloat64 精度安全的浮点数加法
func (s *DataImportService) addFloat64(a, b float64) float64 {
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	result := intA + intB
	return float64(result) / 1000
}

// subtractFloat64 精度安全的浮点数减法
func (s *DataImportService) subtractFloat64(a, b float64) float64 {
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	result := intA - intB
	return float64(result) / 1000
}

// multiplyFloat64 精度安全的浮点数乘法
func (s *DataImportService) multiplyFloat64(a, b float64) float64 {
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	result := intA * intB
	return float64(result) / 1000000 // 除以1000000是因为两个数都乘以了1000
}

// divideFloat64 精度安全的浮点数除法
func (s *DataImportService) divideFloat64(a, b float64) float64 {
	if b == 0 {
		return 0 // 避免除零错误
	}
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	result := intA / intB
	return float64(result) // 不需要再除以1000，因为分子分母都乘以了1000
}

// sumFloat64 精度安全的浮点数求和（可变参数）
func (s *DataImportService) sumFloat64(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var total int64
	for _, value := range values {
		total += int64(value * 1000)
	}
	return float64(total) / 1000
}

// GetCellValue 获取单元格值
func (s *DataImportService) GetCellValueByRow(row []string, cellIndex int) string {
	if len(row) > cellIndex {
		return s.cleanCellValue(row[cellIndex])
	}
	return ""
}

func (s *DataImportService) GetCellValueByRows(rows [][]string, rowIndex int, cellIndex int) string {
	if len(rows) > rowIndex {
		if len(rows[rowIndex]) > cellIndex {
			return s.cleanCellValue(rows[rowIndex][cellIndex])
		}
	}
	return ""
}

// getExcelRowNumber 获取记录中的Excel行号
func (s *DataImportService) getExcelRowNumber(data map[string]interface{}) int {
	// 尝试获取记录的行号
	if rowNum, ok := data["_excel_row"].(int); ok {
		return rowNum
	}

	// 如果没有找到行号，返回默认值
	return 1
}

func (s *DataImportService) getTableName(tableType string) string {
	switch tableType {
	case TableType1:
		return TableName1
	case TableType2:
		return TableName2
	case TableType3:
		return TableName3
	case TableTypeAttachment2:
		return TableAttachment2
	}
	return ""
}

// createValidationErrorZip 创建验证错误文件的ZIP包
func (s *DataImportService) createValidationErrorZip(failedFiles []string, tableType, tableName string) error {
	if len(failedFiles) == 0 {
		return nil
	}

	cacheDir := s.app.GetCachePath(tableType)
	// 创建ZIP文件
	zipFileName := fmt.Sprintf("%s模型报告.zip", tableName)
	zipPath := filepath.Join(cacheDir, zipFileName)

	zipFile, err := os.OpenFile(zipPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("创建ZIP文件失败: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加文件到ZIP
	for _, filePath := range failedFiles {
		file, err := os.Open(filePath)
		if err != nil {
			continue // 跳过无法打开的文件
		}
		defer file.Close()

		// 获取文件名
		fileName := filepath.Base(filePath)

		// 创建ZIP条目
		zipEntry, err := zipWriter.Create(fileName)
		if err != nil {
			continue
		}

		// 复制文件内容
		_, err = io.Copy(zipEntry, file)
		if err != nil {
			continue
		}
	}
	return nil
}

// encryptNumericFields 通用数值字段加密函数
func (s *DataImportService) encryptNumericFields(record map[string]interface{}, numericFields []string) map[string]interface{} {
	encrypted := make(map[string]interface{})

	for _, field := range numericFields {
		if value, ok := record[field].(string); ok && value != "" {
			if encryptedValue, err := s.app.SM4Encrypt(value); err == nil {
				encrypted[field] = encryptedValue
			} else {
				encrypted[field] = ""
			}
		} else {
			encrypted[field] = ""
		}
	}

	return encrypted
}

// formatErrorMessages 格式化错误信息，每条错误使用序号标识并换行
func formatErrorMessages(errorMsg string) string {
	if errorMsg == "" {
		return ""
	}

	// 按分号分割错误信息
	errors := strings.Split(errorMsg, ";")

	var formattedErrors []string
	for i, err := range errors {
		err = strings.TrimSpace(err)
		if err != "" {
			// 添加序号并换行
			formattedErrors = append(formattedErrors, fmt.Sprintf("%d. %s", i+1, err))
		}
	}

	// 用换行符连接所有错误信息
	return strings.Join(formattedErrors, "\n")
}

// ExcelFieldMapping 字段名称到Excel位置的映射结构
type ExcelFieldMapping struct {
	TableType string
	FieldName string
	Column    string // Excel列名，如 "A", "B", "C"
	RowOffset int    // 相对于数据行的偏移量
}

// getTable1FieldMapping 获取附表1字段映射
func (s *DataImportService) getTable1FieldMapping() map[string]ExcelFieldMapping {
	return map[string]ExcelFieldMapping{
		// 主表字段映射（企业基本信息 + 综合能源消费情况）
		"annual_energy_equivalent_value": {TableType: "table1", FieldName: "annual_energy_equivalent_value", Column: "A", RowOffset: 0},
		"annual_energy_equivalent_cost":  {TableType: "table1", FieldName: "annual_energy_equivalent_cost", Column: "B", RowOffset: 0},
		"annual_raw_material_energy":     {TableType: "table1", FieldName: "annual_raw_material_energy", Column: "C", RowOffset: 0},
		"annual_total_coal_consumption":  {TableType: "table1", FieldName: "annual_total_coal_consumption", Column: "D", RowOffset: 0},
		"annual_total_coal_products":     {TableType: "table1", FieldName: "annual_total_coal_products", Column: "E", RowOffset: 0},
		"annual_raw_coal":                {TableType: "table1", FieldName: "annual_raw_coal", Column: "F", RowOffset: 0},
		"annual_raw_coal_consumption":    {TableType: "table1", FieldName: "annual_raw_coal_consumption", Column: "G", RowOffset: 0},
		"annual_clean_coal_consumption":  {TableType: "table1", FieldName: "annual_clean_coal_consumption", Column: "H", RowOffset: 0},
		"annual_other_coal_consumption":  {TableType: "table1", FieldName: "annual_other_coal_consumption", Column: "I", RowOffset: 0},
		"annual_coke_consumption":        {TableType: "table1", FieldName: "annual_coke_consumption", Column: "J", RowOffset: 0},

		// 用途表字段映射
		"input_quantity":  {TableType: "table1", FieldName: "input_quantity", Column: "F", RowOffset: 0},
		"output_quantity": {TableType: "table1", FieldName: "output_quantity", Column: "I", RowOffset: 0},

		// 设备表字段映射
		"total_runtime":           {TableType: "table1", FieldName: "total_runtime", Column: "D", RowOffset: 0},
		"design_life":             {TableType: "table1", FieldName: "design_life", Column: "E", RowOffset: 0},
		"energy_efficiency":       {TableType: "table1", FieldName: "energy_efficiency", Column: "F", RowOffset: 0},
		"capacity":                {TableType: "table1", FieldName: "capacity", Column: "H", RowOffset: 0},
		"annual_coal_consumption": {TableType: "table1", FieldName: "annual_coal_consumption", Column: "J", RowOffset: 0},
	}
}

// getTable2FieldMapping 获取附表2字段映射
func (s *DataImportService) getTable2FieldMapping() map[string]ExcelFieldMapping {
	return map[string]ExcelFieldMapping{
		"usage_time":              {TableType: "table2", FieldName: "usage_time", Column: "D", RowOffset: 0},
		"design_life":             {TableType: "table2", FieldName: "design_life", Column: "E", RowOffset: 0},
		"capacity":                {TableType: "table2", FieldName: "capacity", Column: "H", RowOffset: 0},
		"annual_coal_consumption": {TableType: "table2", FieldName: "annual_coal_consumption", Column: "K", RowOffset: 0},
	}
}

// getTable3FieldMapping 获取附表3字段映射
func (s *DataImportService) getTable3FieldMapping() map[string]ExcelFieldMapping {
	return map[string]ExcelFieldMapping{
		"equivalent_value":           {TableType: "table3", FieldName: "equivalent_value", Column: "P", RowOffset: 0},
		"equivalent_cost":            {TableType: "table3", FieldName: "equivalent_cost", Column: "Q", RowOffset: 0},
		"pq_total_coal_consumption":  {TableType: "table3", FieldName: "pq_total_coal_consumption", Column: "R", RowOffset: 0},
		"pq_coal_consumption":        {TableType: "table3", FieldName: "pq_coal_consumption", Column: "S", RowOffset: 0},
		"pq_coke_consumption":        {TableType: "table3", FieldName: "pq_coke_consumption", Column: "T", RowOffset: 0},
		"pq_blue_coke_consumption":   {TableType: "table3", FieldName: "pq_blue_coke_consumption", Column: "U", RowOffset: 0},
		"sce_total_coal_consumption": {TableType: "table3", FieldName: "sce_total_coal_consumption", Column: "V", RowOffset: 0},
		"sce_coal_consumption":       {TableType: "table3", FieldName: "sce_coal_consumption", Column: "W", RowOffset: 0},
		"sce_coke_consumption":       {TableType: "table3", FieldName: "sce_coke_consumption", Column: "X", RowOffset: 0},
		"sce_blue_coke_consumption":  {TableType: "table3", FieldName: "sce_blue_coke_consumption", Column: "Y", RowOffset: 0},
		"substitution_quantity":      {TableType: "table3", FieldName: "substitution_quantity", Column: "AB", RowOffset: 0},
		"pq_annual_coal_quantity":    {TableType: "table3", FieldName: "pq_annual_coal_quantity", Column: "AC", RowOffset: 0},
		"sce_annual_coal_quantity":   {TableType: "table3", FieldName: "sce_annual_coal_quantity", Column: "AD", RowOffset: 0},
	}
}

// getAttachment2FieldMapping 获取附件2字段映射
func (s *DataImportService) getAttachment2FieldMapping() map[string]ExcelFieldMapping {
	return map[string]ExcelFieldMapping{
		"total_coal":       {TableType: "attachment2", FieldName: "total_coal", Column: "E", RowOffset: 0},
		"raw_coal":         {TableType: "attachment2", FieldName: "raw_coal", Column: "F", RowOffset: 0},
		"washed_coal":      {TableType: "attachment2", FieldName: "washed_coal", Column: "G", RowOffset: 0},
		"other_coal":       {TableType: "attachment2", FieldName: "other_coal", Column: "H", RowOffset: 0},
		"power_generation": {TableType: "attachment2", FieldName: "power_generation", Column: "I", RowOffset: 0},
		"heating":          {TableType: "attachment2", FieldName: "heating", Column: "J", RowOffset: 0},
		"coal_washing":     {TableType: "attachment2", FieldName: "coal_washing", Column: "K", RowOffset: 0},
		"coking":           {TableType: "attachment2", FieldName: "coking", Column: "L", RowOffset: 0},
		"oil_refining":     {TableType: "attachment2", FieldName: "oil_refining", Column: "M", RowOffset: 0},
		"gas_production":   {TableType: "attachment2", FieldName: "gas_production", Column: "N", RowOffset: 0},
		"industry":         {TableType: "attachment2", FieldName: "industry", Column: "O", RowOffset: 0},
		"raw_materials":    {TableType: "attachment2", FieldName: "raw_materials", Column: "P", RowOffset: 0},
		"other_uses":       {TableType: "attachment2", FieldName: "other_uses", Column: "Q", RowOffset: 0},
		"coke":             {TableType: "attachment2", FieldName: "coke", Column: "R", RowOffset: 0},
	}
}

// getFieldMapping 根据表格类型获取字段映射
func (s *DataImportService) getFieldMapping(tableType string) map[string]ExcelFieldMapping {
	switch tableType {
	case TableType1:
		return s.getTable1FieldMapping()
	case TableType2:
		return s.getTable2FieldMapping()
	case TableType3:
		return s.getTable3FieldMapping()
	case TableTypeAttachment2:
		return s.getAttachment2FieldMapping()
	default:
		return make(map[string]ExcelFieldMapping)
	}
}

// getCellPosition 根据字段名称和行号获取Excel单元格位置
func (s *DataImportService) getCellPosition(tableType, fieldName string, rowNumber int) string {
	fieldMapping := s.getFieldMapping(tableType)
	if mapping, exists := fieldMapping[fieldName]; exists {
		// 计算实际行号（考虑偏移量）
		actualRow := rowNumber + mapping.RowOffset
		return mapping.Column + fmt.Sprintf("%d", actualRow)
	}
	return ""
}

// highlightCellsInExcel 在Excel中高亮指定的单元格
func (s *DataImportService) highlightCellsInExcel(f *excelize.File, sheetName string, cells []string) error {
	// 创建黄色背景样式
	style, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFF00"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		return err
	}

	// 去重单元格列表
	uniqueCells := make(map[string]bool)
	for _, cell := range cells {
		if cell != "" {
			uniqueCells[cell] = true
		}
	}

	// 应用样式到每个单元格
	for cell := range uniqueCells {
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	return nil
}

// UnprotecFile 解除Excel文件的保护并保存
func (s *DataImportService) UnprotecFile(filePath string) error {
	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer f.Close()

	// 解除所有工作表的保护
	sheets := f.GetSheetList()
	for _, sheetName := range sheets {
		// 首先尝试使用空密码解除工作表保护
		err := f.UnprotectSheet(sheetName)
		if err != nil {
			// 如果空密码失败，尝试使用"shuji"密码
			_ = f.UnprotectSheet(sheetName, "shuji")
		}
	}

	// 保存修改后的文件
	if err := f.Save(); err != nil {
		return fmt.Errorf("保存解除保护后的文件失败: %v", err)
	}

	fmt.Printf("成功解除文件 %s 的保护并保存\n", filePath)
	return nil
}
