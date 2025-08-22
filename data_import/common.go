package main

import (
	"fmt"
	"time"
)

// insertImportRecord 插入导入记录 - 公共函数
func (s *App) insertImportRecord(fileName, fileType, importState, describe string) {
	record := &DataImportRecord{
		FileName:    fileName,
		FileType:    fileType,
		ImportTime:  time.Now(),
		ImportState: importState,
		Describe:    describe,
		CreateUser:  GetCurrentOSUser(),
	}

	recordService := NewDataImportRecordService(s.db)
	err := recordService.InsertImportRecord(record)
	if err != nil {
		fmt.Printf("插入导入记录失败: %v", err)
	}
}

// checkTableHasData 检查表是否有数据的通用函数
func (s *App) checkTableHasData(tableName string) bool {
	query := fmt.Sprintf("SELECT COUNT(*) as count FROM %s", tableName)
	result, err := s.db.Query(query)
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
func (s *App) clearTableData(tableName string) error {
	query := fmt.Sprintf("DELETE FROM %s", tableName)
	_, err := s.db.Exec(query)
	return err
}

// validateRequiredField 校验必填字段的通用函数
func (s *App) validateRequiredField(data map[string]interface{}, fieldName, fieldDisplayName string, rowIndex int) []string {
	errors := []string{}
	if value, ok := data[fieldName].(string); !ok || value == "" {
		errors = append(errors, fmt.Sprintf("第%d行：%s不能为空", rowIndex+1, fieldDisplayName))
	}
	return errors
}

// validateRequiredFields 批量校验必填字段
func (s *App) validateRequiredFields(data map[string]interface{}, fields map[string]string, rowIndex int) []string {
	errors := []string{}
	for fieldName, displayName := range fields {
		fieldErrors := s.validateRequiredField(data, fieldName, displayName, rowIndex)
		errors = append(errors, fieldErrors...)
	}
	return errors
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
	TableEnterpriseCoalConsumptionMain   = "enterprise_coal_consumption_main"
	TableEnterpriseCoalConsumptionUsage  = "enterprise_coal_consumption_usage"
	TableEnterpriseCoalConsumptionEquip  = "enterprise_coal_consumption_equip"
	TableCriticalCoalEquipmentConsumption = "critical_coal_equipment_consumption"
	TableFixedAssetsInvestmentProject    = "fixed_assets_investment_project"
	TableCoalConsumptionReport           = "coal_consumption_report"
)

// 字段显示名称映射
var (
	// 附表1必填字段
	Table1RequiredFields = map[string]string{
		"stat_date":    "年份",
		"credit_code":  "统一社会信用代码",
		"unit_name":    "企业名称",
	}

	// 附表2必填字段
	Table2RequiredFields = map[string]string{
		"credit_code": "统一社会信用代码",
		"unit_name":   "单位名称",
		"stat_date":   "数据年份",
		"coal_type":   "耗煤类型",
		"coal_no":     "编号",
	}

	// 附表3必填字段
	Table3RequiredFields = map[string]string{
		"project_name":     "项目名称",
		"province_name":    "项目所在省",
		"city_name":        "项目所在市",
		"country_name":     "项目所在县",
		"approval_number":  "节能审查批复文号",
	}

	// 附件2必填字段
	Attachment2RequiredFields = map[string]string{
		"stat_date":     "数据年份",
		"province_name": "单位省级名称",
		"city_name":     "单位市级名称",
		"country_name":  "单位县级名称",
		"unit_name":     "单位名称",
	}
)
