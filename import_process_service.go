package main

import (
	"fmt"
	"shuji/db"
)

// 导入进度

// EnterpriseRecordStatus 企业记录状态结构
type EnterpriseRecordStatus struct {
	UnitName   string `json:"unit_name"`   // 企业名称
	CreditCode string `json:"credit_code"` // 企业代码
	// 动态年份字段将在运行时添加
}

// QueryTable1Process 查询附表1的导入进度
func (a *App) QueryTable1Process() db.QueryResult {
	result := db.QueryResult{
		Ok:   false,
		Data: make([]map[string]interface{}, 0),
	}

	// 1. 查询企业清单表作为基准
	enterpriseQuery := `
		SELECT 
			unit_name,
			credit_code
		FROM enterprise_list 
		ORDER BY unit_name
	`
	enterpriseResult, err := a.db.Query(enterpriseQuery)
	if err != nil {
		result.Message = "查询企业清单失败: " + err.Error()
		return result
	}

	result.Ok = true
	if !enterpriseResult.Ok || enterpriseResult.Data == nil {
		return result
	}

	// 2. 查询附表1主表数据，按credit_code和stat_date分组
	table1Query := `
		SELECT 
			credit_code,
			stat_date,
			COUNT(1) as record_count
		FROM enterprise_coal_consumption_main 
		GROUP BY credit_code, stat_date
		ORDER BY credit_code, stat_date
	`
	table1Result, err := a.db.Query(table1Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询附表1数据失败: " + err.Error()
		return result
	}

	// 3. 构建企业记录状态映射
	table1DataMap := make(map[string]map[string]bool) // credit_code -> stat_date -> hasRecord

	// 处理附表1数据
	if table1Result.Ok && table1Result.Data != nil {
		if data, ok := table1Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				creditCode := ""
				if code, ok := row["credit_code"].(string); ok {
					creditCode = code
				}

				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				// 初始化企业数据映射
				if table1DataMap[creditCode] == nil {
					table1DataMap[creditCode] = make(map[string]bool)
				}
				table1DataMap[creditCode][statDate] = true
			}
		}
	}

	// 4. 处理企业清单数据，构建最终结果
	enterpriseList := make([]map[string]interface{}, 0)
	if data, ok := enterpriseResult.Data.([]map[string]interface{}); ok {
		for _, row := range data {
			unitName := ""
			if name, ok := row["unit_name"].(string); ok {
				unitName = name
			}

			creditCode := ""
			if code, ok := row["credit_code"].(string); ok {
				creditCode = code
			}

			// 创建企业记录状态
			enterpriseStatus := map[string]interface{}{
				"unit_name":   unitName,
				"credit_code": creditCode,
			}

			// 检查该企业在附表1中是否有记录
			if statDateMap, exists := table1DataMap[creditCode]; exists {
				// 复制所有stat_date的记录状态
				for statDate, hasRecord := range statDateMap {
					enterpriseStatus[statDate] = hasRecord
				}
			}

			enterpriseList = append(enterpriseList, enterpriseStatus)
		}
	}

	// 5. 收集所有年份
	allYears := make(map[string]bool)
	for _, statDateMap := range table1DataMap {
		for statDate := range statDateMap {
			allYears[statDate] = true
		}
	}

	// 转换为年份列表
	yearList := make([]string, 0)
	for year := range allYears {
		yearList = append(yearList, year)
	}

	// 6. 返回结果
	result.Ok = true
	result.Data = map[string]interface{}{
		"list":  enterpriseList,
		"years": yearList,
	}
	result.Message = "查询成功"
	return result
}

// QueryTable2Process 查询附表2的导入进度
func (a *App) QueryTable2Process() db.QueryResult {
	result := db.QueryResult{
		Ok:   false,
		Data: make([]map[string]interface{}, 0),
	}

	// 1. 查询重点装置清单表作为基准
	equipmentQuery := `
		SELECT 
			unit_name,
			credit_code,
			equip_type,
			equip_no
		FROM key_equipment_list 
		ORDER BY unit_name, equip_type, equip_no
	`
	equipmentResult, err := a.db.Query(equipmentQuery)
	if err != nil {
		result.Message = "查询重点装置清单失败: " + err.Error()
		return result
	}

	result.Ok = true
	if !equipmentResult.Ok || equipmentResult.Data == nil {
		return result
	}

	// 2. 查询附表2数据，按credit_code和stat_date分组
	table2Query := `
		SELECT 
			credit_code,
			stat_date,
			COUNT(1) as record_count
		FROM critical_coal_equipment_consumption 
		GROUP BY credit_code, stat_date
		ORDER BY credit_code, stat_date
	`
	table2Result, err := a.db.Query(table2Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询附表2数据失败: " + err.Error()
		return result
	}

	// 3. 构建装置记录状态映射
	table2DataMap := make(map[string]map[string]bool) // credit_code -> stat_date -> hasRecord

	// 处理附表2数据
	if table2Result.Ok && table2Result.Data != nil {
		if data, ok := table2Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				creditCode := ""
				if code, ok := row["credit_code"].(string); ok {
					creditCode = code
				}

				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				// 初始化装置数据映射
				if table2DataMap[creditCode] == nil {
					table2DataMap[creditCode] = make(map[string]bool)
				}
				table2DataMap[creditCode][statDate] = true
			}
		}
	}

	// 4. 处理重点装置清单数据，构建最终结果
	equipmentList := make([]map[string]interface{}, 0)
	if data, ok := equipmentResult.Data.([]map[string]interface{}); ok {
		for _, row := range data {
			unitName := ""
			if name, ok := row["unit_name"].(string); ok {
				unitName = name
			}

			creditCode := ""
			if code, ok := row["credit_code"].(string); ok {
				creditCode = code
			}

			equipType := ""
			if equipTypeVal, ok := row["equip_type"].(string); ok {
				equipType = equipTypeVal
			}

			equipNo := ""
			if equipNoVal, ok := row["equip_no"].(string); ok {
				equipNo = equipNoVal
			}

			// 创建装置记录状态
			equipmentStatus := map[string]interface{}{
				"unit_name":   unitName,
				"credit_code": creditCode,
				"equip_type":  equipType,
				"equip_no":    equipNo,
			}

			// 检查该装置在附表2中是否有记录
			if statDateMap, exists := table2DataMap[creditCode]; exists {
				// 复制所有stat_date的记录状态
				for statDate, hasRecord := range statDateMap {
					equipmentStatus[statDate] = hasRecord
				}
			}

			equipmentList = append(equipmentList, equipmentStatus)
		}
	}

	// 5. 收集所有年份
	allYears := make(map[string]bool)
	for _, statDateMap := range table2DataMap {
		for statDate := range statDateMap {
			allYears[statDate] = true
		}
	}

	// 转换为年份列表
	yearList := make([]string, 0)
	for year := range allYears {
		yearList = append(yearList, year)
	}

	// 6. 返回结果
	result.Ok = true
	result.Data = map[string]interface{}{
		"list":  equipmentList,
		"years": yearList,
	}
	result.Message = "查询成功"
	return result
}

// QueryTable3Process 查询附表3的导入进度
func (a *App) QueryTable3Process() db.QueryResult {
	result := db.QueryResult{}

	// 检查数据库连接
	if a.db == nil {
		result.Ok = false
		result.Message = "数据库未初始化"
		return result
	}

	// 查询附表3数据，按stat_date分组
	table3Query := `
		SELECT 
			stat_date,
			COUNT(1) as record_count
		FROM fixed_assets_investment_project 
		GROUP BY stat_date
		ORDER BY stat_date
	`
	table3Result, err := a.db.Query(table3Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询附表3数据失败: " + err.Error()
		return result
	}

	// 处理附表3数据
	table3List := make([]map[string]interface{}, 0)
	if table3Result.Ok && table3Result.Data != nil {
		if data, ok := table3Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				recordCount := int64(0)
				if count, ok := row["record_count"].(int64); ok {
					recordCount = count
				}

				table3List = append(table3List, map[string]interface{}{
					"stat_date":    statDate,
					"record_count": recordCount,
				})
			}
		}
	}

	// 返回结果
	result.Ok = true
	result.Data = table3List
	result.Message = "查询成功"

	return result
}

// QueryAttachment2Process 查询附件2的导入进度
func (a *App) QueryTableAttachment2Process() db.QueryResult {
	result := db.QueryResult{}

	// 检查数据库连接
	if a.db == nil {
		result.Ok = false
		result.Message = "数据库未初始化"
		return result
	}

	// 1. 获取当前用户区域配置
	areaConfigResult := a.GetAreaConfig()
	if !areaConfigResult.Ok {
		result.Ok = false
		result.Message = "获取区域配置失败: " + areaConfigResult.Message
		return result
	}

	areaConfig, ok := areaConfigResult.Data.(map[string]interface{})
	if !ok {
		result.Ok = false
		result.Message = "区域配置数据格式错误"
		return result
	}

	// 2. 根据区域级别确定分组字段和区域名称字段
	var groupByField string
	var areaNameField string
	var areaLevel string

	if areaConfig["country_name"] != nil {
		groupByField = "stat_date, country_name"
		areaNameField = "country_name"
		areaLevel = "县区"
	} else if areaConfig["city_name"] != nil {
		groupByField = "stat_date, city_name"
		areaNameField = "city_name"
		areaLevel = "城市"
	} else if areaConfig["province_name"] != nil {
		groupByField = "stat_date, province_name"
		areaNameField = "province_name"
		areaLevel = "省份"
	} else {
		result.Ok = false
		result.Message = "未配置区域信息"
		return result
	}

	// 3. 查询附件2数据，按区域级别和stat_date分组
	attachment2Query := fmt.Sprintf(`
		SELECT 
			stat_date,
			%s as area_name,
			COUNT(1) as record_count
		FROM coal_consumption_report 
		GROUP BY %s
		ORDER BY stat_date
	`, areaNameField, groupByField)

	attachment2Result, err := a.db.Query(attachment2Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询附件2数据失败: " + err.Error()
		return result
	}

	// 4. 构建区域记录状态映射
	attachment2DataMap := make(map[string]map[string]bool) // area_name -> stat_date -> hasRecord

	// 处理附件2数据
	if attachment2Result.Ok && attachment2Result.Data != nil {
		if data, ok := attachment2Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				areaName := ""
				if name, ok := row["area_name"].(string); ok {
					areaName = name
				}

				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				// 初始化区域数据映射
				if attachment2DataMap[areaName] == nil {
					attachment2DataMap[areaName] = make(map[string]bool)
				}
				attachment2DataMap[areaName][statDate] = true
			}
		}
	}

	// 5. 构建最终结果
	attachment2List := make([]map[string]interface{}, 0)

	// 为每个区域创建记录
	for areaName, statDateMap := range attachment2DataMap {
		areaStatus := map[string]interface{}{
			"area_name": areaName,
		}

		// 复制所有stat_date的记录状态
		for statDate, hasRecord := range statDateMap {
			areaStatus[statDate] = hasRecord
		}

		attachment2List = append(attachment2List, areaStatus)
	}

	// 6. 收集所有年份
	allYears := make(map[string]bool)
	for _, statDateMap := range attachment2DataMap {
		for statDate := range statDateMap {
			allYears[statDate] = true
		}
	}

	yearList := make([]string, 0)
	for statDate := range allYears {
		yearList = append(yearList, statDate)
	}

	// 7. 返回结果
	result.Ok = true
	result.Data = map[string]interface{}{
		"list":       attachment2List,
		"years":      yearList,
		"area_level": areaLevel,
	}
	result.Message = "查询成功"

	return result
}

// getAreaDisplayName 获取区域显示名称
func getAreaDisplayName(provinceName, cityName, countryName string) string {
	if countryName != "" {
		return countryName
	} else if cityName != "" {
		return cityName
	} else if provinceName != "" {
		return provinceName
	}
	return "未知区域"
}

// getAreaLevel 获取区域级别
func getAreaLevel(countryName, cityName, provinceName string) string {
	if countryName != "" {
		return "县"
	} else if cityName != "" {
		return "市"
	} else if provinceName != "" {
		return "省"
	}
	return "未知"
}
