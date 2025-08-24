package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"shuji/data_import"

	"github.com/xuri/excelize/v2"
)

// DataCheckResult 数据校验结果
type DataCheckResult struct {
	Ok      bool     `json:"ok"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

// ManualCheckResult 人工校验结果
type ManualCheckResult struct {
	ObjID       string `json:"obj_id"`
	TableName   string `json:"table_name"`
	CheckResult string `json:"check_result"` // "1": 已校核, "2": 校核未通过
	CheckRemark string `json:"check_remark"` // 校核备注
	CheckUser   string `json:"check_user"`
	CheckTime   string `json:"check_time"`
}

// DataCheckValidationError 数据校验错误结构
type DataCheckValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Type    string `json:"type"` // "required", "range", "logic", "format"
}

// 自动校验相关的常量
const (
	// 数值范围常量
	MAX_VALUE_100000     = 100000
	MAX_VALUE_1000000000 = 1000000000
	MAX_VALUE_200000     = 200000
	MIN_VALUE_0          = 0
	MAX_YEARS_50         = 50
)

// ==================== 附表1 模型校验 ====================

// ModelDataCheckTable1 附表1模型校验（自动校验规则）
func (a *App) ModelDataCheckTable1() DataCheckResult {
	result := DataCheckResult{
		Ok:      false,
		Message: "",
		Errors:  []string{},
	}

	var allErrors []string

	// 1. 校验附表1主表数据
	table1MainErrors := a.validateTable1MainDataAuto()
	for _, err := range table1MainErrors {
		allErrors = append(allErrors, err.Message)
	}

	// 2. 校验附表1用途数据
	table1UsageErrors := a.validateTable1UsageDataAuto()
	for _, err := range table1UsageErrors {
		allErrors = append(allErrors, err.Message)
	}

	// 3. 校验附表1设备数据
	table1EquipErrors := a.validateTable1EquipDataAuto()
	for _, err := range table1EquipErrors {
		allErrors = append(allErrors, err.Message)
	}

	if len(allErrors) > 0 {
		result.Ok = false
		result.Message = data_import.TableName1 + "模型校验发现错误"
		result.Errors = allErrors
	} else {
		result.Ok = true
		result.Message = data_import.TableName1 + "模型校验通过"
	}

	return result
}

// validateTable1MainDataAuto 自动校验附表1主表数据
func (a *App) validateTable1MainDataAuto() []DataCheckValidationError {
	var errors []DataCheckValidationError

	// 查询所有附表1主表数据
	result, err := a.db.Query("SELECT annual_energy_equivalent_value, annual_energy_equivalent_cost, annual_raw_material_energy FROM enterprise_coal_consumption_main WHERE is_confirm = 1")
	if err != nil {
		return errors
	}

	// 处理查询结果
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				energyEquivalentValue := getStringValue(row["annual_energy_equivalent_value"])
				energyEquivalentCost := getStringValue(row["annual_energy_equivalent_cost"])
				rawMaterialEnergy := getStringValue(row["annual_raw_material_energy"])

				// 年综合能耗当量值校验
				if energyEquivalentValue != "" {
					if val, err := strconv.ParseFloat(energyEquivalentValue, 64); err == nil {
						if val < MIN_VALUE_0 {
							errors = append(errors, DataCheckValidationError{
								Field:   "年综合能耗当量值",
								Message: "年综合能耗当量值数值错误，年综合能耗当量值应该≥0",
								Type:    "range",
							})
						} else if val > MAX_VALUE_100000 {
							errors = append(errors, DataCheckValidationError{
								Field:   "年综合能耗当量值",
								Message: "年综合能耗当量值数值错误，年综合能耗当量值应该≤100000",
								Type:    "range",
							})
						}
					}
				}

				// 年综合能耗等价值校验
				if energyEquivalentCost != "" {
					if val, err := strconv.ParseFloat(energyEquivalentCost, 64); err == nil {
						if val < MIN_VALUE_0 {
							errors = append(errors, DataCheckValidationError{
								Field:   "年综合能耗等价值",
								Message: "年综合能耗等价值数值错误，年综合能耗等价值应该≥0",
								Type:    "range",
							})
						} else if val > MAX_VALUE_100000 {
							errors = append(errors, DataCheckValidationError{
								Field:   "年综合能耗等价值",
								Message: "年综合能耗等价值数值错误，年综合能耗等价值应该≤100000",
								Type:    "range",
							})
						}
					}
				}

				// 逻辑关系校验：年综合能耗当量值应≥年原料用能消费量
				if energyEquivalentValue != "" && rawMaterialEnergy != "" {
					if energyVal, err1 := strconv.ParseFloat(energyEquivalentValue, 64); err1 == nil {
						if rawMaterialVal, err2 := strconv.ParseFloat(rawMaterialEnergy, 64); err2 == nil {
							if energyVal < rawMaterialVal {
								errors = append(errors, DataCheckValidationError{
									Field:   "年综合能耗当量值",
									Message: "年综合能耗当量值数值错误，年综合能耗当量值应该≥年原料用能消费量",
									Type:    "logic",
								})
							}
						}
					}
				}
			}
		}
	}

	return errors
}

// validateTable1UsageDataAuto 自动校验附表1用途数据
func (a *App) validateTable1UsageDataAuto() []DataCheckValidationError {
	var errors []DataCheckValidationError

	// 查询所有附表1用途数据
	result, err := a.db.Query("SELECT input_quantity FROM enterprise_coal_consumption_usage WHERE is_confirm = 1")
	if err != nil {
		return errors
	}

	// 处理查询结果
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				inputQuantity := getStringValue(row["input_quantity"])

				// 投入量数值校验
				if inputQuantity != "" {
					if val, err := strconv.ParseFloat(inputQuantity, 64); err == nil {
						if val < MIN_VALUE_0 {
							errors = append(errors, DataCheckValidationError{
								Field:   "投入量",
								Message: "投入量数值错误，投入量应该≥0",
								Type:    "range",
							})
						} else if val > MAX_VALUE_100000 {
							errors = append(errors, DataCheckValidationError{
								Field:   "投入量",
								Message: "投入量数值错误，投入量应该≤100000",
								Type:    "range",
							})
						}
					}
				}
			}
		}
	}

	return errors
}

// validateTable1EquipDataAuto 自动校验附表1设备数据
func (a *App) validateTable1EquipDataAuto() []DataCheckValidationError {
	var errors []DataCheckValidationError

	// 查询所有附表1设备数据
	result, err := a.db.Query("SELECT total_runtime, design_life, capacity FROM enterprise_coal_consumption_equip WHERE is_confirm = 1")
	if err != nil {
		return errors
	}

	// 处理查询结果
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				totalRuntime := getStringValue(row["total_runtime"])
				designLife := getStringValue(row["design_life"])
				capacity := getStringValue(row["capacity"])

				// 累计使用时间校验
				if totalRuntime != "" {
					if val, err := strconv.Atoi(totalRuntime); err == nil {
						if val < 0 || val > MAX_YEARS_50 {
							errors = append(errors, DataCheckValidationError{
								Field:   "累计使用时间",
								Message: "累计使用时间数值错误，累计使用时间应该为0-50之间的整数",
								Type:    "range",
							})
						}
					}
				}

				// 设计年限校验
				if designLife != "" {
					if val, err := strconv.Atoi(designLife); err == nil {
						if val < 0 || val > MAX_YEARS_50 {
							errors = append(errors, DataCheckValidationError{
								Field:   "设计年限",
								Message: "设计年限数值错误，设计年限应该为0-50之间的整数",
								Type:    "range",
							})
						}
					}
				}

				// 容量校验
				if capacity != "" {
					if val, err := strconv.Atoi(capacity); err == nil {
						if val < 0 {
							errors = append(errors, DataCheckValidationError{
								Field:   "容量",
								Message: "容量数值错误，容量应该≥0",
								Type:    "range",
							})
						}
					}
				}
			}
		}
	}

	return errors
}

// ==================== 附表2 模型校验 ====================

// ModelDataCheckTable2 附表2模型校验（自动校验规则）
func (a *App) ModelDataCheckTable2() DataCheckResult {
	result := DataCheckResult{
		Ok:      false,
		Message: "",
		Errors:  []string{},
	}

	var allErrors []string

	// 校验附表2数据
	table2Errors := a.validateTable2DataAuto()
	for _, err := range table2Errors {
		allErrors = append(allErrors, err.Message)
	}

	if len(allErrors) > 0 {
		result.Ok = false
		result.Message = data_import.TableName2 + "模型校验发现错误"
		result.Errors = allErrors
	} else {
		result.Ok = true
		result.Message = data_import.TableName2 + "模型校验通过"
	}

	return result
}

// validateTable2DataAuto 自动校验附表2数据
func (a *App) validateTable2DataAuto() []DataCheckValidationError {
	var errors []DataCheckValidationError

	// 查询所有附表2数据
	result, err := a.db.Query("SELECT annual_coal_consumption FROM critical_coal_equipment_consumption WHERE is_confirm = 1")
	if err != nil {
		return errors
	}

	// 处理查询结果
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				annualCoalConsumption := getStringValue(row["annual_coal_consumption"])

				// 年耗煤量数值校验
				if annualCoalConsumption != "" {
					if val, err := strconv.ParseFloat(annualCoalConsumption, 64); err == nil {
						if val < MIN_VALUE_0 {
							errors = append(errors, DataCheckValidationError{
								Field:   "年耗煤量",
								Message: "年耗煤量数值错误，年耗煤量应该≥0",
								Type:    "range",
							})
						} else if val > MAX_VALUE_1000000000 {
							errors = append(errors, DataCheckValidationError{
								Field:   "年耗煤量",
								Message: "年耗煤量数值错误，年耗煤量应该≤1000000000",
								Type:    "range",
							})
						}
					}
				}
			}
		}
	}

	return errors
}

// ==================== 附表3 模型校验 ====================

// ModelDataCheckTable3 附表3模型校验（自动校验规则）
func (a *App) ModelDataCheckTable3() DataCheckResult {
	result := DataCheckResult{
		Ok:      false,
		Message: "",
		Errors:  []string{},
	}

	var allErrors []string

	// 校验附表3数据
	table3Errors := a.validateTable3DataAuto()
	for _, err := range table3Errors {
		allErrors = append(allErrors, err.Message)
	}

	if len(allErrors) > 0 {
		result.Ok = false
		result.Message = data_import.TableName3 + "模型校验发现错误"
		result.Errors = allErrors
	} else {
		result.Ok = true
		result.Message = data_import.TableName3 + "模型校验通过"
	}

	return result
}

// validateTable3DataAuto 自动校验附表3数据
func (a *App) validateTable3DataAuto() []DataCheckValidationError {
	var errors []DataCheckValidationError

	// 查询所有附表3数据
	result, err := a.db.Query("SELECT equivalent_value FROM fixed_assets_investment_project WHERE is_confirm = 1")
	if err != nil {
		return errors
	}

	// 处理查询结果
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				equivalentValue := getStringValue(row["equivalent_value"])

				// 当量值数值校验
				if equivalentValue != "" {
					if val, err := strconv.ParseFloat(equivalentValue, 64); err == nil {
						if val < MIN_VALUE_0 {
							errors = append(errors, DataCheckValidationError{
								Field:   "当量值",
								Message: "当量值数值错误，当量值应该≥0",
								Type:    "range",
							})
						} else if val > MAX_VALUE_100000 {
							errors = append(errors, DataCheckValidationError{
								Field:   "当量值",
								Message: "当量值数值错误，当量值应该≤100000",
								Type:    "range",
							})
						}
					}
				}
			}
		}
	}

	return errors
}

// ==================== 附件2 模型校验 ====================

// ModelDataCheckAttachment2 附件2模型校验（自动校验规则）
func (a *App) ModelDataCheckAttachment2() DataCheckResult {
	result := DataCheckResult{
		Ok:      false,
		Message: "",
		Errors:  []string{},
	}

	var allErrors []string

	// 校验附件2数据
	attachment2Errors := a.validateAttachment2DataAuto()
	for _, err := range attachment2Errors {
		allErrors = append(allErrors, err.Message)
	}

	if len(allErrors) > 0 {
		result.Ok = false
		result.Message = data_import.TableAttachment2 + "模型校验发现错误"
		result.Errors = allErrors
	} else {
		result.Ok = true
		result.Message = data_import.TableAttachment2 + "模型校验通过"
	}

	return result
}

// validateAttachment2DataAuto 自动校验附件2数据
func (a *App) validateAttachment2DataAuto() []DataCheckValidationError {
	var errors []DataCheckValidationError

	// 查询所有附件2数据
	result, err := a.db.Query("SELECT total_coal, raw_coal, washed_coal, other_coal FROM coal_consumption_report WHERE is_confirm = 1")
	if err != nil {
		return errors
	}

	// 处理查询结果
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				totalCoal := getStringValue(row["total_coal"])
				rawCoal := getStringValue(row["raw_coal"])
				washedCoal := getStringValue(row["washed_coal"])
				otherCoal := getStringValue(row["other_coal"])

				// 煤合计数值校验
				if totalCoal != "" {
					if val, err := strconv.ParseFloat(totalCoal, 64); err == nil {
						if val < MIN_VALUE_0 {
							errors = append(errors, DataCheckValidationError{
								Field:   "煤合计",
								Message: "煤合计数值错误，煤合计应该≥0",
								Type:    "range",
							})
						} else if val > MAX_VALUE_200000 {
							errors = append(errors, DataCheckValidationError{
								Field:   "煤合计",
								Message: "煤合计数值错误，煤合计应该≤200000",
								Type:    "range",
							})
						}
					}
				}

				// 逻辑关系校验：煤合计应=原煤+洗精煤+其他
				if totalCoal != "" && rawCoal != "" && washedCoal != "" && otherCoal != "" {
					if totalCoalVal, err1 := strconv.ParseFloat(totalCoal, 64); err1 == nil {
						if rawCoalVal, err2 := strconv.ParseFloat(rawCoal, 64); err2 == nil {
							if washedCoalVal, err3 := strconv.ParseFloat(washedCoal, 64); err3 == nil {
								if otherCoalVal, err4 := strconv.ParseFloat(otherCoal, 64); err4 == nil {
									if totalCoalVal != rawCoalVal+washedCoalVal+otherCoalVal {
										errors = append(errors, DataCheckValidationError{
											Field:   "煤合计",
											Message: "煤合计数值错误，煤合计应该=原煤+洗精煤+其他",
											Type:    "logic",
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return errors
}

// ==================== 企业清单相关校验 ====================

// checkEnterpriseInListForValidation 检查企业是否在企业清单中（用于校验）
func (a *App) checkEnterpriseInListForValidation(creditCode, unitName string) (bool, string, bool) {
	// 首先检查企业名称是否在enterprise_list表中
	result, err := a.db.Query("SELECT credit_code FROM enterprise_list WHERE unit_name = ?", unitName)
	if err != nil {
		return false, fmt.Sprintf("企业名称数值错误，企业名称应该为有效的企业名称"), false
	}

	// 检查企业名称是否在清单中
	if result.Data == nil {
		return false, fmt.Sprintf("企业名称数值错误，企业名称%s未在企业清单里", unitName), false // 不在清单中
	}

	// 提取credit_code
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		return false, "企业名称数值错误，数据格式异常", false
	}

	dbCreditCode, exists := data["credit_code"]
	if !exists {
		return false, "企业名称数值错误，数据库异常", false
	}

	var existingCreditCode string
	if dbCreditCode == nil {
		existingCreditCode = ""
	} else {
		existingCreditCode = fmt.Sprintf("%v", dbCreditCode)
	}

	// 检查统一信用代码是否对应
	if existingCreditCode != creditCode {
		return false, fmt.Sprintf("统一信用代码数值错误，统一信用代码应该为%s", existingCreditCode), true // 在清单中但代码不对应
	}

	return true, "", true // 在清单中且代码对应
}

// validateEnterpriseListRules 校验企业清单相关规则
// validateEnterpriseListRules 校验企业清单相关规则（只校验附表1和附表2中的单位名称）
func (a *App) validateEnterpriseListRules() []DataCheckValidationError {
	var errors []DataCheckValidationError

	// 校验附表1（enterprise_coal_consumption_main）中的单位名称
	table1Result, err := a.db.Query("SELECT unit_name FROM enterprise_coal_consumption_main WHERE is_confirm = 1")
	if err == nil && table1Result.Data != nil {
		if dataList, ok := table1Result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				unitName := getStringValue(row["unit_name"])
				if unitName != "" {
					// 检查单位名称是否在enterprise_list表中
					listResult, listErr := a.db.Query("SELECT unit_name FROM enterprise_list WHERE unit_name = ?", unitName)
					if listErr != nil {
						errors = append(errors, DataCheckValidationError{
							Field:   "企业名称",
							Message: "企业名称数值错误，企业名称应该为有效的企业名称",
							Type:    "info",
						})
						continue
					}

					if listResult.Data == nil {
						// 企业名称未在企业清单里
						errors = append(errors, DataCheckValidationError{
							Field:   "企业名称",
							Message: "企业名称数值错误，企业名称" + unitName + "未在企业清单里",
							Type:    "info",
						})
					}
				}
			}
		}
	}

	// 校验附表2（critical_coal_equipment_consumption）中的单位名称
	table2Result, err := a.db.Query("SELECT unit_name FROM critical_coal_equipment_consumption WHERE is_confirm = 1")
	if err == nil && table2Result.Data != nil {
		if dataList, ok := table2Result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				unitName := getStringValue(row["unit_name"])
				if unitName != "" {
					// 检查单位名称是否在enterprise_list表中
					listResult, listErr := a.db.Query("SELECT unit_name FROM enterprise_list WHERE unit_name = ?", unitName)
					if listErr != nil {
						errors = append(errors, DataCheckValidationError{
							Field:   "企业名称",
							Message: "企业名称数值错误，企业名称应该为有效的企业名称",
							Type:    "info",
						})
						continue
					}

					if listResult.Data == nil {
						// 企业名称未在企业清单里
						errors = append(errors, DataCheckValidationError{
							Field:   "企业名称",
							Message: "企业名称数值错误，企业名称" + unitName + "未在企业清单里",
							Type:    "info",
						})
					}
				}
			}
		}
	}

	return errors
}

// ==================== 完整模型校验 ====================

// ModelDataCheck 完整模型校验（包含所有表的校验）
func (a *App) ModelDataCheck() DataCheckResult {
	result := DataCheckResult{
		Ok:      false,
		Message: "",
		Errors:  []string{},
	}

	var allErrors []string

	// 1. 企业清单相关校验
	enterpriseListErrors := a.validateEnterpriseListRules()
	for _, err := range enterpriseListErrors {
		allErrors = append(allErrors, err.Message)
	}

	// 2. 附表1校验
	table1Result := a.ModelDataCheckTable1()
	allErrors = append(allErrors, table1Result.Errors...)

	// 3. 附表2校验
	table2Result := a.ModelDataCheckTable2()
	allErrors = append(allErrors, table2Result.Errors...)

	// 4. 附表3校验
	table3Result := a.ModelDataCheckTable3()
	allErrors = append(allErrors, table3Result.Errors...)

	// 5. 附件2校验
	attachment2Result := a.ModelDataCheckAttachment2()
	allErrors = append(allErrors, attachment2Result.Errors...)

	// 判断校验结果：只有非info类型的错误才影响校验结果
	var criticalErrors []string
	for _, err := range enterpriseListErrors {
		if err.Type != "info" {
			criticalErrors = append(criticalErrors, err.Message)
		}
	}

	// 合并所有严重错误
	criticalErrors = append(criticalErrors, table1Result.Errors...)
	criticalErrors = append(criticalErrors, table2Result.Errors...)
	criticalErrors = append(criticalErrors, table3Result.Errors...)
	criticalErrors = append(criticalErrors, attachment2Result.Errors...)

	if len(criticalErrors) > 0 {
		result.Ok = false
		result.Message = "模型校验发现错误"
		result.Errors = allErrors // 返回所有错误信息（包括info类型）
	} else {
		result.Ok = true
		result.Message = "模型校验通过"
		result.Errors = allErrors // 即使通过也返回info类型的提示信息
	}

	return result
}

// ==================== 人工校验 ====================

// ManualDataCheckTable1 附表1人工校验
func (a *App) ManualDataCheckTable1(checkResult ManualCheckResult) DataCheckResult {
	result := DataCheckResult{
		Ok:      false,
		Message: "",
		Errors:  []string{},
	}

	// 更新附表1相关表中的is_check字段
	tables := []string{
		"enterprise_coal_consumption_main",
		"enterprise_coal_consumption_usage",
		"enterprise_coal_consumption_equip",
	}

	for _, tableName := range tables {
		_, err := a.db.Exec(`
			UPDATE `+tableName+` 
			SET is_check = ?, check_user = ?, check_time = ? 
			WHERE obj_id = ?
		`, checkResult.CheckResult, checkResult.CheckUser, time.Now(), checkResult.ObjID)

		if err != nil {
			result.Message = data_import.TableName1 + "人工校验失败: " + err.Error()
			return result
		}
	}

	result.Ok = true
	result.Message = data_import.TableName1 + "人工校验成功"
	return result
}

// ManualDataCheckTable2 附表2人工校验
func (a *App) ManualDataCheckTable2(checkResult ManualCheckResult) DataCheckResult {
	result := DataCheckResult{
		Ok:      false,
		Message: "",
		Errors:  []string{},
	}

	// 更新附表2表中的is_check字段
	_, err := a.db.Exec(`
		UPDATE critical_coal_equipment_consumption 
		SET is_check = ?, check_user = ?, check_time = ? 
		WHERE obj_id = ?
	`, checkResult.CheckResult, checkResult.CheckUser, time.Now(), checkResult.ObjID)

	if err != nil {
		result.Message = data_import.TableName2 + "人工校验失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Message = data_import.TableName2 + "人工校验成功"
	return result
}

// ManualDataCheckTable3 附表3人工校验
func (a *App) ManualDataCheckTable3(checkResult ManualCheckResult) DataCheckResult {
	result := DataCheckResult{
		Ok:      false,
		Message: "",
		Errors:  []string{},
	}

	// 更新附表3表中的is_check字段
	_, err := a.db.Exec(`
		UPDATE fixed_assets_investment_project 
		SET is_check = ?, check_user = ?, check_time = ? 
		WHERE obj_id = ?
	`, checkResult.CheckResult, checkResult.CheckUser, time.Now(), checkResult.ObjID)

	if err != nil {
		result.Message = data_import.TableName3 + "人工校验失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Message = data_import.TableName3 + "人工校验成功"
	return result
}

// ManualDataCheckAttachment2 附件2人工校验
func (a *App) ManualDataCheckAttachment2(checkResult ManualCheckResult) DataCheckResult {
	result := DataCheckResult{
		Ok:      false,
		Message: "",
		Errors:  []string{},
	}

	// 更新附件2表中的is_check字段
	_, err := a.db.Exec(`
		UPDATE coal_consumption_report 
		SET is_check = ?, check_user = ?, check_time = ? 
		WHERE obj_id = ?
	`, checkResult.CheckResult, checkResult.CheckUser, time.Now(), checkResult.ObjID)

	if err != nil {
		result.Message = data_import.TableAttachment2 + "人工校验失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Message = data_import.TableAttachment2 + "人工校验成功"
	return result
}

// ManualDataCheck 通用人工校验（根据表名自动选择对应的校验函数）
func (a *App) ManualDataCheck(checkResult ManualCheckResult) DataCheckResult {
	switch checkResult.TableName {
	case "enterprise_coal_consumption_main", "enterprise_coal_consumption_usage", "enterprise_coal_consumption_equip":
		return a.ManualDataCheckTable1(checkResult)
	case "critical_coal_equipment_consumption":
		return a.ManualDataCheckTable2(checkResult)
	case "fixed_assets_investment_project":
		return a.ManualDataCheckTable3(checkResult)
	case "coal_consumption_report":
		return a.ManualDataCheckAttachment2(checkResult)
	default:
		return DataCheckResult{
			Ok:      false,
			Message: "未知的表名: " + checkResult.TableName,
			Errors:  []string{},
		}
	}
}

// ==================== 数据获取函数 ====================

// GetDataCheckListTable1 获取附表1数据校验列表
func (a *App) GetDataCheckListTable1(page int, pageSize int) interface{} {
	offset := (page - 1) * pageSize

	// 获取附表1所有表的总数
	var total int
	result1, _ := a.db.Query("SELECT COUNT(*) FROM enterprise_coal_consumption_main WHERE is_confirm = 1")
	if result1.Data != nil {
		if dataList, ok := result1.Data.([]map[string]interface{}); ok && len(dataList) > 0 {
			if countStr := getStringValue(dataList[0]["COUNT(*)"]); countStr != "" {
				if count, err := strconv.Atoi(countStr); err == nil {
					total = count
				}
			}
		}
	}

	var count2, count3 int
	result2, _ := a.db.Query("SELECT COUNT(*) FROM enterprise_coal_consumption_usage WHERE is_confirm = 1")
	if result2.Data != nil {
		if dataList, ok := result2.Data.([]map[string]interface{}); ok && len(dataList) > 0 {
			if countStr := getStringValue(dataList[0]["COUNT(*)"]); countStr != "" {
				if count, err := strconv.Atoi(countStr); err == nil {
					count2 = count
				}
			}
		}
	}

	result3, _ := a.db.Query("SELECT COUNT(*) FROM enterprise_coal_consumption_equip WHERE is_confirm = 1")
	if result3.Data != nil {
		if dataList, ok := result3.Data.([]map[string]interface{}); ok && len(dataList) > 0 {
			if countStr := getStringValue(dataList[0]["COUNT(*)"]); countStr != "" {
				if count, err := strconv.Atoi(countStr); err == nil {
					count3 = count
				}
			}
		}
	}
	total += count2 + count3

	// 获取附表1主表数据
	mainData := a.getTableCheckDataWithTableName("enterprise_coal_consumption_main", data_import.TableName1+"主表", pageSize, offset)

	// 获取附表1用途数据
	usageData := a.getTableCheckDataWithTableName("enterprise_coal_consumption_usage", data_import.TableName1+"用途", pageSize, offset)

	// 获取附表1设备数据
	equipData := a.getTableCheckDataWithTableName("enterprise_coal_consumption_equip", data_import.TableName1+"设备", pageSize, offset)

	// 合并数据
	var allData []interface{}
	for _, item := range mainData {
		allData = append(allData, item)
	}
	for _, item := range usageData {
		allData = append(allData, item)
	}
	for _, item := range equipData {
		allData = append(allData, item)
	}

	return map[string]interface{}{
		"total": total,
		"data":  allData,
		"page":  page,
		"size":  pageSize,
	}
}

// GetDataCheckListTable2 获取附表2数据校验列表
func (a *App) GetDataCheckListTable2(page int, pageSize int) interface{} {
	offset := (page - 1) * pageSize

	// 获取总数
	var total int
	result, _ := a.db.Query("SELECT COUNT(*) FROM critical_coal_equipment_consumption WHERE is_confirm = 1")
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok && len(dataList) > 0 {
			if countStr := getStringValue(dataList[0]["COUNT(*)"]); countStr != "" {
				if count, err := strconv.Atoi(countStr); err == nil {
					total = count
				}
			}
		}
	}

	// 获取数据列表
	data := a.getTableCheckDataWithTableName("critical_coal_equipment_consumption", data_import.TableName2, pageSize, offset)

	return map[string]interface{}{
		"total": total,
		"data":  data,
		"page":  page,
		"size":  pageSize,
	}
}

// GetDataCheckListTable3 获取附表3数据校验列表
func (a *App) GetDataCheckListTable3(page int, pageSize int) interface{} {
	offset := (page - 1) * pageSize

	// 获取总数
	var total int
	result, _ := a.db.Query("SELECT COUNT(*) FROM fixed_assets_investment_project WHERE is_confirm = 1")
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok && len(dataList) > 0 {
			if countStr := getStringValue(dataList[0]["COUNT(*)"]); countStr != "" {
				if count, err := strconv.Atoi(countStr); err == nil {
					total = count
				}
			}
		}
	}

	// 获取数据列表
	data := a.getTableCheckDataWithTableName("fixed_assets_investment_project", data_import.TableName3, pageSize, offset)

	return map[string]interface{}{
		"total": total,
		"data":  data,
		"page":  page,
		"size":  pageSize,
	}
}

// GetDataCheckListAttachment2 获取附件2数据校验列表
func (a *App) GetDataCheckListAttachment2(page int, pageSize int) interface{} {
	offset := (page - 1) * pageSize

	// 获取总数
	var total int
	result, _ := a.db.Query("SELECT COUNT(*) FROM coal_consumption_report WHERE is_confirm = 1")
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok && len(dataList) > 0 {
			if countStr := getStringValue(dataList[0]["COUNT(*)"]); countStr != "" {
				if count, err := strconv.Atoi(countStr); err == nil {
					total = count
				}
			}
		}
	}

	// 获取数据列表
	data := a.getTableCheckDataWithTableName("coal_consumption_report", data_import.TableAttachment2, pageSize, offset)

	return map[string]interface{}{
		"total": total,
		"data":  data,
		"page":  page,
		"size":  pageSize,
	}
}

// GetDataCheckList 通用获取数据校验列表（根据表名自动选择对应的函数）
func (a *App) GetDataCheckList(tableName string, page int, pageSize int) interface{} {
	switch tableName {
	case "enterprise_coal_consumption_main", "enterprise_coal_consumption_usage", "enterprise_coal_consumption_equip":
		return a.GetDataCheckListTable1(page, pageSize)
	case "critical_coal_equipment_consumption":
		return a.GetDataCheckListTable2(page, pageSize)
	case "fixed_assets_investment_project":
		return a.GetDataCheckListTable3(page, pageSize)
	case "coal_consumption_report":
		return a.GetDataCheckListAttachment2(page, pageSize)
	default:
		return map[string]interface{}{
			"total": 0,
			"data":  []interface{}{},
			"page":  page,
			"size":  pageSize,
		}
	}
}

// GetDataCheckDetail 获取数据详情
func (a *App) GetDataCheckDetail(tableName string, objID string) interface{} {
	// 根据表名构建查询语句
	var query string
	switch tableName {
	case "enterprise_coal_consumption_main":
		query = "SELECT * FROM enterprise_coal_consumption_main WHERE obj_id = ?"
	case "enterprise_coal_consumption_usage":
		query = "SELECT * FROM enterprise_coal_consumption_usage WHERE obj_id = ?"
	case "enterprise_coal_consumption_equip":
		query = "SELECT * FROM enterprise_coal_consumption_equip WHERE obj_id = ?"
	case "critical_coal_equipment_consumption":
		query = "SELECT * FROM critical_coal_equipment_consumption WHERE obj_id = ?"
	case "fixed_assets_investment_project":
		query = "SELECT * FROM fixed_assets_investment_project WHERE obj_id = ?"
	case "coal_consumption_report":
		query = "SELECT * FROM coal_consumption_report WHERE obj_id = ?"
	default:
		return map[string]interface{}{
			"data": []interface{}{},
		}
	}

	result, err := a.db.Query(query, objID)
	if err != nil {
		return map[string]interface{}{
			"data": []interface{}{},
		}
	}

	var data []interface{}
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				item := make(map[string]interface{})
				for col, val := range row {
					if val != nil {
						item[col] = val
					} else {
						item[col] = ""
					}
				}
				data = append(data, item)
			}
		}
	}

	return map[string]interface{}{
		"data": data,
	}
}

// ==================== 结果导出 ====================

// ExportCheckResultsToExcel 导出校验结果到Excel文件（在原文件基础上添加错误标记）
func (a *App) ExportCheckResultsToExcel(originalFilePath string) (string, error) {
	// 检查原文件是否存在
	if _, err := os.Stat(originalFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("原文件不存在: %s", originalFilePath)
	}

	// 生成新文件名
	dir := filepath.Dir(originalFilePath)
	baseName := filepath.Base(originalFilePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	newFileName := fmt.Sprintf("%s_校验结果_%s%s", nameWithoutExt, time.Now().Format("20060102_150405"), ext)
	newFilePath := filepath.Join(dir, newFileName)

	// 复制原文件
	if err := copyFile(originalFilePath, newFilePath); err != nil {
		return "", fmt.Errorf("复制文件失败: %v", err)
	}

	// 打开新文件进行编辑
	file, err := excelize.OpenFile(newFilePath)
	if err != nil {
		return "", fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer file.Close()

	// 执行模型校验并获取结果
	modelCheckResult := a.ModelDataCheck()

	// 如果校验未通过，在Excel中添加错误标记
	if !modelCheckResult.Ok && len(modelCheckResult.Errors) > 0 {
		// 为每个工作表添加错误标记
		for _, sheetName := range file.GetSheetList() {
			// 获取工作表的最大行数
			rows, err := file.GetRows(sheetName)
			if err != nil {
				continue
			}

			// 在最后一列添加错误信息
			errorCol := len(rows[0]) + 1 // 从最后一列开始
			errorColName := string(rune('A' + errorCol - 1))

			// 设置错误列标题
			file.SetCellValue(sheetName, fmt.Sprintf("%s1", errorColName), "校验错误信息")

			// 设置错误列标题的样式（红色）
			style, err := file.NewStyle(&excelize.Style{
				Font: &excelize.Font{
					Color: "FF0000", // 红色
					Bold:  true,
				},
			})
			if err == nil {
				file.SetCellStyle(sheetName, fmt.Sprintf("%s1", errorColName), fmt.Sprintf("%s1", errorColName), style)
			}

			// 为每一行添加对应的错误信息
			for i, errorMsg := range modelCheckResult.Errors {
				if i+2 <= len(rows) { // 确保不超出实际行数
					cell := fmt.Sprintf("%s%d", errorColName, i+2)
					file.SetCellValue(sheetName, cell, errorMsg)

					// 设置错误信息的样式（红色）
					errorStyle, err := file.NewStyle(&excelize.Style{
						Font: &excelize.Font{
							Color: "FF0000", // 红色
						},
					})
					if err == nil {
						file.SetCellStyle(sheetName, cell, cell, errorStyle)
					}
				}
			}
		}
	}

	// 保存文件
	if err := file.Save(); err != nil {
		return "", fmt.Errorf("保存Excel文件失败: %v", err)
	}

	return newFilePath, nil
}

// CheckAndExportToExcel 校验并导出到Excel（返回校验结果和文件路径）
func (a *App) CheckAndExportToExcel(originalFilePath string) (bool, string, error) {
	// 执行模型校验
	modelCheckResult := a.ModelDataCheck()

	// 导出到Excel
	exportedFilePath, err := a.ExportCheckResultsToExcel(originalFilePath)
	if err != nil {
		return false, "", err
	}

	// 返回校验结果、文件路径和错误信息
	return modelCheckResult.Ok, exportedFilePath, nil
}

// ==================== 辅助函数 ====================

// getStringValue 安全获取字符串值
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// getTableCheckData 获取指定表的校验数据
func (a *App) getTableCheckData(tableName string) []map[string]interface{} {
	var data []map[string]interface{}
	result, err := a.db.Query(`
		SELECT obj_id, unit_name, stat_date, is_check, check_user, check_time 
		FROM ` + tableName + ` 
		WHERE is_confirm = 1 
		ORDER BY create_time DESC
	`)
	if err != nil {
		return data
	}

	// 处理查询结果
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				item := map[string]interface{}{
					"obj_id":     getStringValue(row["obj_id"]),
					"unit_name":  getStringValue(row["unit_name"]),
					"stat_date":  getStringValue(row["stat_date"]),
					"is_check":   getStringValue(row["is_check"]),
					"check_user": getStringValue(row["check_user"]),
					"check_time": row["check_time"],
				}
				data = append(data, item)
			}
		}
	}

	return data
}

// getTableCheckDataWithTableName 获取指定表的校验数据（带表名标识）
func (a *App) getTableCheckDataWithTableName(tableName string, tableDisplayName string, pageSize int, offset int) []map[string]interface{} {
	var data []map[string]interface{}
	result, err := a.db.Query(`
		SELECT obj_id, unit_name, stat_date, is_check, check_user, check_time 
		FROM `+tableName+` 
		WHERE is_confirm = 1 
		ORDER BY create_time DESC
		LIMIT ? OFFSET ?
	`, pageSize, offset)
	if err != nil {
		return data
	}

	// 处理查询结果
	if result.Data != nil {
		if dataList, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range dataList {
				item := map[string]interface{}{
					"table_name": tableDisplayName,
					"obj_id":     getStringValue(row["obj_id"]),
					"unit_name":  getStringValue(row["unit_name"]),
					"stat_date":  getStringValue(row["stat_date"]),
					"is_check":   getStringValue(row["is_check"]),
					"check_user": getStringValue(row["check_user"]),
					"check_time": row["check_time"],
				}
				data = append(data, item)
			}
		}
	}

	return data
}
