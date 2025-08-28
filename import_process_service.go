package main

import (
	"fmt"
	"shuji/db"
	"strings"
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

	// 1. 获取当前用户区域信息
	targetLocation, dataLevel, _, err := a.getCurrentUserLocationData()
	if err != nil {
		result.Ok = false
		result.Message = "获取区域信息失败: " + err.Error()
		return result
	}

	if targetLocation == nil {
		result.Ok = false
		result.Message = "未找到对应的区域信息"
		return result
	}

	// 2. 根据区域级别决定查询逻辑
	if dataLevel == 3 {
		ret := a.QueryDataTable3()
		if !ret.Ok || ret.Data == nil {
			result.Ok = false
			result.Message = "查询附表3数据失败: " + ret.Message
			return result
			return ret
		}
		// 这里ret.Data是[]map[string]interface{}，需要包装成本函数的Data结构
		result.Ok = true
		result.Data = map[string]interface{}{
			"list":       ret.Data,
			"area_level": 3,
		}
		result.Message = "查询成功"
		return result

	} else if dataLevel == 2 {
		// 市级别：以县(或区)分组
		return a.queryTable3CityLevel(targetLocation)
	} else {
		// 省级别：以市分组
		return a.queryTable3ProvinceLevel(targetLocation)
	}
}

// queryTable3CityLevel 市级别查询：以县(或区)分组
func (a *App) queryTable3CityLevel(targetLocation interface{}) db.QueryResult {
	result := db.QueryResult{}

	// 1. 获取该市下的所有县区
	countyList := make([]string, 0)
	if targetLocationMap, ok := targetLocation.(map[string]interface{}); ok {
		if children, exists := targetLocationMap["children"]; exists && children != nil {
			if childrenList, ok := children.([]interface{}); ok {
				for _, county := range childrenList {
					if countyMap, ok := county.(map[string]interface{}); ok {
						if name, exists := countyMap["name"]; exists && name != nil {
							countyList = append(countyList, fmt.Sprintf("%v", name))
						}
					}
				}
			}
		}
	}

	// 2. 查询并解析附表3数据
	importedCounties, err := a.queryAndParseTable3Data()
	if err != nil {
		result.Ok = false
		result.Message = "查询附表3数据失败: " + err.Error()
		return result
	}

	// 4. 构建最终结果
	areaList := make([]map[string]interface{}, 0)

	// 为每个县区创建记录
	for _, countyName := range countyList {
		areaStatus := map[string]interface{}{
			"area_name": countyName,
			"is_import": importedCounties[countyName],
		}
		areaList = append(areaList, areaStatus)
	}

	// 返回结果
	result.Ok = true
	result.Data = map[string]interface{}{
		"list":       areaList,
		"area_level": 2,
	}
	result.Message = "查询成功"

	return result
}

// queryTable3ProvinceLevel 省级别查询：以市分组
func (a *App) queryTable3ProvinceLevel(targetLocation interface{}) db.QueryResult {
	result := db.QueryResult{}

	// 1. 获取该省下的所有市
	cityList := make([]string, 0)
	if targetLocationMap, ok := targetLocation.(map[string]interface{}); ok {
		if children, exists := targetLocationMap["children"]; exists && children != nil {
			if childrenList, ok := children.([]interface{}); ok {
				for _, city := range childrenList {
					if cityMap, ok := city.(map[string]interface{}); ok {
						if name, exists := cityMap["name"]; exists && name != nil {
							cityList = append(cityList, fmt.Sprintf("%v", name))
						}
					}
				}
			}
		}
	}

	if len(cityList) == 0 {
		result.Ok = false
		result.Message = "未找到该省下的市信息"
		return result
	}

	// 2. 查询并解析附表3数据
	importedCities, err := a.queryAndParseTable3Data()
	if err != nil {
		result.Ok = false
		result.Message = "查询附表3数据失败: " + err.Error()
		return result
	}

	// 4. 构建最终结果
	areaList := make([]map[string]interface{}, 0)

	// 为每个市创建记录
	for _, cityName := range cityList {
		areaStatus := map[string]interface{}{
			"area_name": cityName,
			"is_import": importedCities[cityName],
		}
		areaList = append(areaList, areaStatus)
	}

	// 返回结果
	result.Ok = true
	result.Data = map[string]interface{}{
		"list":       areaList,
		"area_level": 1,
	}
	result.Message = "查询成功"

	return result
}

// queryAndParseTable3Data 查询并解析附表3数据
func (a *App) queryAndParseTable3Data() (map[string]bool, error) {
	// 查询附表3数据，按examination_authority分组
	table3Query := `
		SELECT 
			examination_authority,
			COUNT(1) as record_count
		FROM fixed_assets_investment_project 
		WHERE examination_authority IS NOT NULL AND examination_authority != ''
		GROUP BY examination_authority
	`
	table3Result, err := a.db.Query(table3Query)
	if err != nil {
		return nil, err
	}

	// 解析节能审查机关
	importedAreas := make(map[string]bool)
	if table3Result.Ok && table3Result.Data != nil {
		if data, ok := table3Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				examinationAuthority := ""
				if authority, ok := row["examination_authority"].(string); ok {
					examinationAuthority = authority
				}

				// 提取名称，移除"发改委"后缀
				areaName := a.extractAreaFromAuthority(examinationAuthority)
				if areaName != "" {
					importedAreas[areaName] = true
				}
			}
		}
	}

	return importedAreas, nil
}

// extractAreaFromAuthority 从节能审查机关中提取区域名称
func (a *App) extractAreaFromAuthority(authority string) string {
	if authority == "" {
		return ""
	}

	// 移除"发改委"后缀
	authority = strings.TrimSuffix(authority, "发改委")
	authority = strings.TrimSuffix(authority, "发展和改革委员会")
	authority = strings.TrimSuffix(authority, "发展改革委")

	return authority
}

// QueryAttachment2Process 查询附件2的导入进度
func (a *App) QueryTableAttachment2Process() db.QueryResult {
	result := db.QueryResult{}

	// 1. 获取当前用户区域信息
	targetLocation, dataLevel, _, err := a.getCurrentUserLocationData()
	if err != nil {
		result.Ok = false
		result.Message = "获取区域信息失败: " + err.Error()
		return result
	}

	if targetLocation == nil {
		result.Ok = false
		result.Message = "未找到对应的区域信息"
		return result
	}

	// 2. 根据区域级别决定查询逻辑
	if dataLevel == 3 {
		// 县级别：查询所有表数据，返回所有字段
		return a.QueryDataAttachment2()
	} else if dataLevel == 2 {
		// 市级别：以县+年份分组
		return a.queryAttachment2CityLevel(targetLocation)
	} else {
		// 省级别：以市+年份分组
		return a.queryAttachment2ProvinceLevel(targetLocation)
	}
}

// queryAttachment2CityLevel 市级别查询：以县+年份分组
func (a *App) queryAttachment2CityLevel(targetLocation interface{}) db.QueryResult {
	result := db.QueryResult{}

	// 1. 获取该市下的所有县区
	countyList := make([]string, 0)
	if targetLocationMap, ok := targetLocation.(map[string]interface{}); ok {
		if children, exists := targetLocationMap["children"]; exists && children != nil {
			if childrenList, ok := children.([]interface{}); ok {
				for _, county := range childrenList {
					if countyMap, ok := county.(map[string]interface{}); ok {
						if name, exists := countyMap["name"]; exists && name != nil {
							countyList = append(countyList, fmt.Sprintf("%v", name))
						}
					}
				}
			}
		}
	}

	if len(countyList) == 0 {
		result.Ok = false
		result.Message = "未找到该市下的县区信息"
		return result
	}

	// 2. 查询附件2数据，按县和stat_date分组
	attachment2Query := `
		SELECT 
			country_name as area_name,
			stat_date,
			COUNT(1) as record_count
		FROM coal_consumption_report 
		WHERE country_name IS NOT NULL AND country_name != ''
		GROUP BY country_name, stat_date
		ORDER BY country_name, stat_date`

	attachment2Result, err := a.db.Query(attachment2Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询附件2数据失败: " + err.Error()
		return result
	}

	// 3. 构建区域记录状态映射和年份集合
	attachment2DataMap := make(map[string]map[string]bool) // area_name -> stat_date -> hasRecord
	allYears := make(map[string]bool)
	areaList := make([]map[string]interface{}, 0, len(countyList))

	// 初始化所有县区的记录
	for _, countyName := range countyList {
		areaStatus := map[string]interface{}{
			"area_name": countyName,
		}
		areaList = append(areaList, areaStatus)
	}

	// 处理附件2数据，同时收集年份
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

				// 收集年份
				allYears[statDate] = true

				// 初始化区域数据映射
				if attachment2DataMap[areaName] == nil {
					attachment2DataMap[areaName] = make(map[string]bool)
				}
				attachment2DataMap[areaName][statDate] = true
			}
		}
	}

	// 4. 为每个县区添加年份数据
	for i, countyName := range countyList {
		if statDateMap, exists := attachment2DataMap[countyName]; exists {
			for statDate, hasRecord := range statDateMap {
				areaList[i][statDate] = hasRecord
			}
		}
	}

	// 5. 转换年份集合为列表
	yearList := make([]string, 0, len(allYears))
	for statDate := range allYears {
		yearList = append(yearList, statDate)
	}

	// 返回结果
	result.Ok = true
	result.Data = map[string]interface{}{
		"list":       areaList,
		"years":      yearList,
		"area_level": 2,
	}
	result.Message = "查询成功"

	return result
}

// queryAttachment2ProvinceLevel 省级别查询：以市+年份分组
func (a *App) queryAttachment2ProvinceLevel(targetLocation interface{}) db.QueryResult {
	result := db.QueryResult{}

	// 1. 获取该省下的所有市
	cityList := make([]string, 0)
	if targetLocationMap, ok := targetLocation.(map[string]interface{}); ok {
		if children, exists := targetLocationMap["children"]; exists && children != nil {
			if childrenList, ok := children.([]interface{}); ok {
				for _, city := range childrenList {
					if cityMap, ok := city.(map[string]interface{}); ok {
						if name, exists := cityMap["name"]; exists && name != nil {
							cityList = append(cityList, fmt.Sprintf("%v", name))
						}
					}
				}
			}
		}
	}

	if len(cityList) == 0 {
		result.Ok = false
		result.Message = "未找到该省下的市信息"
		return result
	}

	// 2. 查询附件2数据，按市和stat_date分组
	attachment2Query := `
		SELECT 
			city_name as area_name,
			stat_date,
			COUNT(1) as record_count
		FROM coal_consumption_report 
		WHERE city_name IS NOT NULL AND city_name != ''
		GROUP BY city_name, stat_date
		ORDER BY city_name, stat_date`

	attachment2Result, err := a.db.Query(attachment2Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询附件2数据失败: " + err.Error()
		return result
	}

	// 3. 构建区域记录状态映射和年份集合
	attachment2DataMap := make(map[string]map[string]bool) // area_name -> stat_date -> hasRecord
	allYears := make(map[string]bool)
	areaList := make([]map[string]interface{}, 0, len(cityList))

	// 初始化所有市的记录
	for _, cityName := range cityList {
		areaStatus := map[string]interface{}{
			"area_name": cityName,
		}
		areaList = append(areaList, areaStatus)
	}

	// 处理附件2数据，同时收集年份
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

				// 收集年份
				allYears[statDate] = true

				// 初始化区域数据映射
				if attachment2DataMap[areaName] == nil {
					attachment2DataMap[areaName] = make(map[string]bool)
				}
				attachment2DataMap[areaName][statDate] = true
			}
		}
	}

	// 4. 为每个市添加年份数据
	for i, cityName := range cityList {
		if statDateMap, exists := attachment2DataMap[cityName]; exists {
			for statDate, hasRecord := range statDateMap {
				areaList[i][statDate] = hasRecord
			}
		}
	}

	// 5. 转换年份集合为列表
	yearList := make([]string, 0, len(allYears))
	for statDate := range allYears {
		yearList = append(yearList, statDate)
	}

	// 返回结果
	result.Ok = true
	result.Data = map[string]interface{}{
		"list":       areaList,
		"years":      yearList,
		"area_level": 1,
	}
	result.Message = "查询成功"

	return result
}
