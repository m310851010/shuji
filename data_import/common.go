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
	RowNumber int    `json:"row_number"`
	Message   string `json:"message"`
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

// parseNumericValue 解析数值
func (s *DataImportService) parseNumericValue(value string) string {
	value = s.cleanCellValue(value)
	if value == "" {
		return "0"
	}

	// 移除常见的非数字字符，保留数字、小数点和负号
	var result strings.Builder
	for _, char := range value {
		if (char >= '0' && char <= '9') || char == '.' || char == '-' || char == ',' {
			result.WriteRune(char)
		}
	}

	cleaned := result.String()
	if cleaned == "" || cleaned == "-" {
		return "0"
	}

	return cleaned
}

// validateRequiredField 校验必填字段的通用函数
func (s *DataImportService) validateRequiredField(data map[string]interface{}, fieldName, fieldDisplayName string, rowIndex int) []string {
	errors := []string{}
	if value, ok := data[fieldName].(string); !ok || value == "" {
		errors = append(errors, fmt.Sprintf("第%d行：%s不能为空", rowIndex+1, fieldDisplayName))
	}
	return errors
}

// validateRequiredFields 批量校验必填字段
func (s *DataImportService) validateRequiredFields(data map[string]interface{}, fields map[string]string, rowIndex int) []string {
	errors := []string{}
	for fieldName, displayName := range fields {
		fieldErrors := s.validateRequiredField(data, fieldName, displayName, rowIndex)
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

	if unitName != "" && creditCode != "" {
		provinceName := s.getStringValue(data["province_name"])
		cityName := s.getStringValue(data["city_name"])
		countryName := s.getStringValue(data["country_name"])

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
		} else {
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
		errors = append(errors, fmt.Sprintf("第%d行：导入的数据单位与当前单位不符", excelRowNum))
		return errors
	}
	// 2.检查市, city_name有值时,是否匹配, 无值时返回成功
	if expectedCity != "" && cityName != expectedCity {
		errors = append(errors, fmt.Sprintf("第%d行：导入的数据单位与当前单位不符", excelRowNum))
		return errors
	}
	// 3.检查县, country_name有值时,是否匹配, 无值时返回成功
	if expectedCountry != "" && countryName != expectedCountry {
		errors = append(errors, fmt.Sprintf("第%d行：导入的数据单位与当前单位不符", excelRowNum))
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
		"unit_name":     "单位名称",
		"credit_code":   "统一社会信用代码",
		"province_name": "单位地址",
		"city_name":     "单位地址",
		"country_name":  "单位地址",
		"trade_a":       "所属行业",
		"trade_b":       "所属行业",
		"trade_c":       "所属行业",
		"stat_date":     "数据年份",
		"coal_type":     "类型",
		"row_no":        "编号",
		"coal_no":       "累计使用时间",
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

	// 附件2必填字段
	Attachment2RequiredFields = map[string]string{
		"stat_date":        "数据年份",
		"province_name":    "省（市、区）",
		"city_name":        "地市（州）",
		"country_name":     "县（区）",
		"total_coal":       "煤合计",
		"raw_coal":         "原煤",
		"washed_coal":      "洗精煤",
		"other_coal":       "其他",
		"power_generation": "火力发电",
		"heating":          "供热",
		"coal_washing":     "煤炭洗选",
		"coking":           "炼焦",
		"oil_refining":     "炼油及煤制油",
		"gas_production":   "制气",
		"industry":         "工业",
		"raw_materials":    "用作原料、材料",
		"other_uses":       "其他用途",
		"coke":             "焦炭消费摸底",
	}
)

// generateUUID 生成UUID
func (s *DataImportService) generateUUID() string {
	return uuid.New().String()
}

// parseFloat 解析浮点数
func (s *DataImportService) parseFloat(value string) (float64, error) {
	// 移除逗号
	value = strings.ReplaceAll(value, ",", "")
	// 移除空格
	value = strings.TrimSpace(value)

	if value == "" {
		return 0, nil
	}

	return strconv.ParseFloat(value, 64)
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
