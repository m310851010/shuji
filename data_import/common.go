package data_import

import (
	"fmt"
	"shuji/db"
	"strings"
)

// App 应用接口，用于访问数据库和其他功能
type App interface {
	GetDB() *db.Database
	GetAreaConfig() db.QueryResult
	InsertImportRecord(fileName, fileType, importState, describe string) db.QueryResult
}

// DataImportService 数据导入服务
type DataImportService struct {
	app App
}


// NewDataImportService 创建数据导入服务
func NewDataImportService(app App) *DataImportService {
	return &DataImportService{app: app}
}

// checkTableHasData 检查表是否有数据的通用函数
func (s *DataImportService) checkTableHasData(tableName string) bool {
	query := fmt.Sprintf("SELECT COUNT(*) as count FROM %s", tableName)
	result, err := s.app.GetDB().Query(query)
	if err != nil {
		return false
	}

	if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
		if count, ok := data[0]["count"].(int64); ok {
			return count > 0
		}
	}
	return false
}

// clearTableData 清空表数据的通用函数
func (s *DataImportService) clearTableData(tableName string) error {
	query := fmt.Sprintf("DELETE FROM %s", tableName)
	_, err := s.app.GetDB().Exec(query)
	return err
}

// 通用Excel解析辅助函数

// findTableStartRow 查找表格开始行
func (s *DataImportService) findTableStartRow(rows [][]string, keywords ...string) int {
	for i, row := range rows {
		if len(row) > 0 {
			firstCell := strings.TrimSpace(row[0])
			for _, keyword := range keywords {
				if strings.Contains(firstCell, keyword) {
					return i
				}
			}
		}
	}
	return -1
}

// findTableEndRow 查找表格结束行
func (s *DataImportService) findTableEndRow(rows [][]string, startRow int, endKeywords ...string) int {
	for i := startRow + 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) > 0 {
			firstCell := strings.TrimSpace(row[0])
			for _, keyword := range endKeywords {
				if strings.Contains(firstCell, keyword) {
					return i
				}
			}
		}
	}
	return len(rows)
}

// isValidDataRow 检查是否为有效数据行
func (s *DataImportService) isValidDataRow(row []string) bool {
	if len(row) == 0 {
		return false
	}

	// 检查第一列是否为空
	firstCell := strings.TrimSpace(row[0])
	if firstCell == "" {
		return false
	}

	// 检查是否为表格标题行
	titleKeywords := []string{"表格", "情况", "汇总", "信息", "消费", "项目", "企业", "单位"}
	for _, keyword := range titleKeywords {
		if strings.Contains(firstCell, keyword) && len(firstCell) > 10 {
			return false
		}
	}

	return true
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

// parseDateValue 解析日期值
func (s *DataImportService) parseDateValue(value string) string {
	value = s.cleanCellValue(value)
	if value == "" {
		return ""
	}

	// 尝试解析常见的日期格式
	dateFormats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006.01.02",
		"2006年01月02日",
		"2006年1月2日",
	}

	for _, format := range dateFormats {
		if len(value) >= len(format) {
			// 简单匹配，实际项目中可能需要更复杂的日期解析
			if strings.Contains(value, "年") && strings.Contains(value, "月") {
				// 处理中文日期格式
				value = strings.ReplaceAll(value, "年", "-")
				value = strings.ReplaceAll(value, "月", "-")
				value = strings.ReplaceAll(value, "日", "")
				return value
			}
		}
	}

	return value
}

// validateRequiredField 校验必填字段的通用函数
func (s *DataImportService) validateRequiredField(data map[string]interface{}, fieldName, fieldDisplayName string, rowIndex int) []string {
	errors := []string{}
	if value, ok := data[fieldName].(string); !ok || value == "" {
		errors = append(errors, fmt.Sprintf("第%d行：%s不能为空，请核对并重新上传数据", rowIndex+1, fieldDisplayName))
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

// getStringValue 获取字符串值的通用函数
func (s *DataImportService) getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// checkEnterpriseInList 检查企业是否在清单中
func (s *DataImportService) checkEnterpriseInList(unitName, creditCode string) (bool, error) {
	// 先通过统一信用代码查询enterprise_list表，获取企业名称
	query := "SELECT unit_name FROM enterprise_list WHERE credit_code = ?"
	result, err := s.app.GetDB().Query(query, creditCode)
	if err != nil {
		return false, err
	}

	// 如果没有找到该统一信用代码对应的企业，就不检查
	if data, ok := result.Data.([]map[string]interface{}); !ok || len(data) == 0 {
		return true, nil // 不在清单中，但不报错
	}

	// 如果找到了，比较企业名称是否相同
	if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
		if dbUnitName, ok := data[0]["unit_name"].(string); ok {
			return dbUnitName == unitName, nil // 相同则校验通过，不同则校验失败
		}
	}

	return false, nil
}

// validateEnterpriseNameCreditCodeCorrespondence 通用的企业名称和统一信用代码对应关系校验函数（单条数据）
func (s *DataImportService) validateEnterpriseNameCreditCodeCorrespondence(unitName, creditCode string, rowIndex int, isEnterpriseListCheck bool) []string {
	errors := []string{}

	if unitName != "" && creditCode != "" {
		// 检查企业名称和统一信用代码是否对应
		corresponds, err := s.checkEnterpriseInList(unitName, creditCode)
		if err != nil {
			if isEnterpriseListCheck {
				errors = append(errors, fmt.Sprintf("第%d行：企业清单检查失败", rowIndex+1))
			} else {
				errors = append(errors, fmt.Sprintf("第%d行：企业名称和统一信用代码对应关系检查失败", rowIndex+1))
			}
			return errors
		}

		if !corresponds {
			if isEnterpriseListCheck {
				errors = append(errors, fmt.Sprintf("第%d行：%s企业，统一信用代码%s未在清单表里", rowIndex+1, unitName, creditCode))
			} else {
				errors = append(errors, fmt.Sprintf("第%d行：统一信用代码%s和上传的企业名称不对应", rowIndex+1, creditCode))
			}
		}
	}

	return errors
}

// GetAreaConfig 获取区域配置 - 直接调用App的方法
func (s *DataImportService) GetAreaConfig() db.QueryResult {
	return s.app.GetAreaConfig()
}

// 文件类型常量
const (
	FileTypeTable1      = "附表1"
	FileTypeTable2      = "附表2"
	FileTypeTable3      = "附表3"
	FileTypeAttachment2 = "附件2"
)

// 导入状态常量
const (
	ImportStateSuccess = "上传成功"
	ImportStateFailed  = "上传失败"
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
		"stat_date":   "年份",
		"credit_code": "统一社会信用代码",
		"unit_name":   "企业名称",
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
		"coal_no":       "编号",
		"usage_time":    "累计使用时间",
		"design_life":   "设计年限",
		"capacity_unit": "容量单位",
		"capacity":      "容量",
	}

	// 附表3必填字段
	Table3RequiredFields = map[string]string{
		"project_name":      "项目名称",
		"project_code":      "项目代码",
		"approval_number":   "审查意见文号",
		"stat_date":         "年份",
		"unit_name":         "建设单位",
		"main_content":      "主要建设内容",
		"province_name":     "项目所在省",
		"city_name":         "项目所在市",
		"country_name":      "项目所在县",
		"industry_type":     "所属行业大类",
		"industry_subtype":  "所属行业小类",
		"approval_time":     "节能审查批复时间",
		"put_into_use_time": "拟投产时间/实际投产时间",
		"approval_org":      "节能审查机关",
		"equivalent_value":  "当量值",
		"equivalent_cost":   "等价值",
	}

	// 附件2必填字段
	Attachment2RequiredFields = map[string]string{
		"stat_date":     "数据年份",
		"province_name": "省（市、区）",
		"city_name":     "地市（州）",
		"country_name":  "县（区）",
		"report_unit":   "制表单位",
	}
)
