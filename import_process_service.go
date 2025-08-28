package main

import (
	"fmt"
	"shuji/db"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Excel样式常量
const (
	HeaderBackgroundColor      = "D3D3D3" // 浅灰色
	HeaderFontColor            = "000000" // 黑色
	ImportedBackgroundColor    = "00B050" // 绿色
	NotImportedBackgroundColor = "FF0000" // 红色
	WhiteFontColor             = "FFFFFF" // 白色
)

// createHeaderStyle 创建表头样式
func createHeaderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  12,
			Color: HeaderFontColor,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{HeaderBackgroundColor},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
}

// createDataStyle 创建数据样式
func createDataStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 11,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
}

// createImportedStyle 创建已导入样式（绿色）
func createImportedStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:  11,
			Color: WhiteFontColor,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{ImportedBackgroundColor},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
}

// createNotImportedStyle 创建未导入样式（红色）
func createNotImportedStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:  11,
			Color: WhiteFontColor,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{NotImportedBackgroundColor},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
}

// applyCommonStyles 应用通用样式设置
func applyCommonStyles(f *excelize.File, headers []string) (int, int, int, int, error) {
	// 创建样式
	headerStyle, err := createHeaderStyle(f)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	dataStyle, err := createDataStyle(f)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	importedStyle, err := createImportedStyle(f)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	notImportedStyle, err := createNotImportedStyle(f)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// 应用表头样式
	for i := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellStyle("Sheet1", cellName, cellName, headerStyle)
	}

	// 设置行高
	f.SetRowHeight("Sheet1", 1, 25) // 表头行高

	return headerStyle, dataStyle, importedStyle, notImportedStyle, nil
}

// 导入进度

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

		ret := a.QueryDataAttachment2()
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

// ExportTable1ProgressToExcel 导出附表1导入进度到Excel
func (a *App) ExportTable1ProgressToExcel(filePath string) db.QueryResult {
	result := db.QueryResult{}

	// 1. 获取附表1导入进度数据
	progressResult := a.QueryTable1Process()
	if !progressResult.Ok {
		result.Ok = false
		result.Message = "获取导入进度失败: " + progressResult.Message
		return result
	}

	if progressResult.Data == nil {
		result.Ok = false
		result.Message = "获取导入进度失败: 导入进度没有数据"
		return result
	}

	progressData, ok := progressResult.Data.(map[string]interface{})
	if !ok {
		result.Ok = false
		result.Message = "导入进度数据格式错误"
		return result
	}

	list, ok := progressData["list"].([]map[string]interface{})
	if !ok {
		result.Ok = false
		result.Message = "导入进度列表数据格式错误"
		return result
	}

	years, ok := progressData["years"].([]string)
	if !ok {
		result.Ok = false
		result.Message = "导入进度年份数据格式错误"
		return result
	}

	// 2. 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 3. 设置表头
	headers := []string{"企业", "企业代码"}
	for _, year := range years {
		headers = append(headers, year+"年数据")
	}

	// 写入表头
	for i, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cellName, header)
	}

	// 4. 设置样式
	_, dataStyle, importedStyle, notImportedStyle, err := applyCommonStyles(f, headers)
	if err != nil {
		result.Ok = false
		result.Message = "创建样式失败: " + err.Error()
		return result
	}

	// 设置列宽
	f.SetColWidth("Sheet1", "A", "A", 25) // 企业名称
	f.SetColWidth("Sheet1", "B", "B", 25) // 企业代码
	for i := range years {
		colName, _ := excelize.ColumnNumberToName(i + 3)
		f.SetColWidth("Sheet1", colName, colName, 15) // 年份列
	}

	// 5. 写入数据
	for rowIndex, item := range list {
		row := rowIndex + 2 // 从第2行开始写入数据

		// 企业名称
		if unitName, ok := item["unit_name"].(string); ok {
			cellName, _ := excelize.CoordinatesToCellName(1, row)
			f.SetCellValue("Sheet1", cellName, unitName)
		}

		// 企业代码
		if creditCode, ok := item["credit_code"].(string); ok {
			cellName, _ := excelize.CoordinatesToCellName(2, row)
			f.SetCellValue("Sheet1", cellName, creditCode)
		}

		// 年份数据
		for yearIndex, year := range years {
			col := yearIndex + 3 // 从第3列开始写入年份数据
			cellName, _ := excelize.CoordinatesToCellName(col, row)

			if hasData, ok := item[year].(bool); ok && hasData {
				f.SetCellValue("Sheet1", cellName, "已导入")
				f.SetCellStyle("Sheet1", cellName, cellName, importedStyle)
			} else {
				f.SetCellValue("Sheet1", cellName, "未导入")
				f.SetCellStyle("Sheet1", cellName, cellName, notImportedStyle)
			}
		}

		// 应用数据样式到前两列（企业名称和企业代码）
		startCell, _ := excelize.CoordinatesToCellName(1, row)
		endCell, _ := excelize.CoordinatesToCellName(2, row)
		f.SetCellStyle("Sheet1", startCell, endCell, dataStyle)

		// 设置数据行高
		f.SetRowHeight("Sheet1", row, 20)
	}

	// 5. 保存文件
	err = f.SaveAs(filePath)
	if err != nil {
		result.Ok = false
		result.Message = "保存Excel文件失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Message = "导入进度导出成功"
	return result
}

// ExportTable2ProgressToExcel 导出附表2导入进度到Excel
func (a *App) ExportTable2ProgressToExcel(filePath string) db.QueryResult {
	result := db.QueryResult{}

	// 1. 获取附表2导入进度数据
	progressResult := a.QueryTable2Process()
	if !progressResult.Ok {
		result.Ok = false
		result.Message = "获取附表2导入进度失败: " + progressResult.Message
		return result
	}

	progressData, ok := progressResult.Data.(map[string]interface{})
	if !ok {
		result.Ok = false
		result.Message = "附表2导入进度数据格式错误"
		return result
	}

	list, ok := progressData["list"].([]map[string]interface{})
	if !ok {
		result.Ok = false
		result.Message = "附表2导入进度列表数据格式错误"
		return result
	}

	years, ok := progressData["years"].([]string)
	if !ok {
		result.Ok = false
		result.Message = "附表2导入进度年份数据格式错误"
		return result
	}

	// 2. 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	// 3. 设置表头
	headers := []string{"企业"}
	for _, year := range years {
		headers = append(headers, year+"年数据")
	}

	// 写入表头
	for i, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Sheet1", cellName, header)
	}

	// 4. 设置样式
	_, dataStyle, importedStyle, notImportedStyle, err := applyCommonStyles(f, headers)
	if err != nil {
		result.Ok = false
		result.Message = "创建样式失败: " + err.Error()
		return result
	}

	// 设置列宽
	f.SetColWidth("Sheet1", "A", "A", 25) // 企业
	for i := range years {
		colName, _ := excelize.ColumnNumberToName(i + 2)
		f.SetColWidth("Sheet1", colName, colName, 15) // 年份列
	}

	// 4. 写入数据
	for rowIndex, item := range list {
		row := rowIndex + 2 // 从第2行开始写入数据

		// 企业名称
		if areaName, ok := item["area_name"].(string); ok {
			cellName, _ := excelize.CoordinatesToCellName(1, row)
			f.SetCellValue("Sheet1", cellName, areaName)
		}

		// 年份数据
		for yearIndex, year := range years {
			col := yearIndex + 2 // 从第2列开始写入年份数据
			cellName, _ := excelize.CoordinatesToCellName(col, row)

			if hasData, ok := item[year].(bool); ok && hasData {
				f.SetCellValue("Sheet1", cellName, "已导入")
				f.SetCellStyle("Sheet1", cellName, cellName, importedStyle)
			} else {
				f.SetCellValue("Sheet1", cellName, "未导入")
				f.SetCellStyle("Sheet1", cellName, cellName, notImportedStyle)
			}
		}

		// 应用数据样式到第一列（企业名称）
		startCell, _ := excelize.CoordinatesToCellName(1, row)
		endCell, _ := excelize.CoordinatesToCellName(1, row)
		f.SetCellStyle("Sheet1", startCell, endCell, dataStyle)

		// 设置数据行高
		f.SetRowHeight("Sheet1", row, 20)
	}

	// 5. 保存文件
	err = f.SaveAs(filePath)
	if err != nil {
		result.Ok = false
		result.Message = "保存Excel文件失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Message = "附表2导入进度导出成功"
	return result
}

// ExportTable3ProgressToExcel 导出附表3导入进度到Excel
func (a *App) ExportTable3ProgressToExcel(filePath string) db.QueryResult {
	result := db.QueryResult{}

	// 1. 获取附表3导入进度数据
	progressResult := a.QueryTable3Process()
	if !progressResult.Ok {
		result.Ok = false
		result.Message = "获取附表3导入进度失败: " + progressResult.Message
		return result
	}

	progressData, ok := progressResult.Data.(map[string]interface{})
	if !ok {
		result.Ok = false
		result.Message = "附表3导入进度数据格式错误"
		return result
	}

	areaLevel, ok := progressData["area_level"].(int)
	if !ok {
		result.Ok = false
		result.Message = "附表3区域级别数据格式错误"
		return result
	}

	list, ok := progressData["list"].([]map[string]interface{})
	if !ok {
		result.Ok = false
		result.Message = "附表3导入进度列表数据格式错误"
		return result
	}

	// 2. 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	if areaLevel == 3 {
		// 县级别：导出所有数据
		// 3. 设置表头
		headers := []string{
			"项目名称", "项目代码", "建设单位", "主要建设内容", "项目所在省", "项目所在地市", "项目所在区县",
			"所属行业大类", "所属行业小类", "节能审查批复时间", "拟投产时间", "实际投产时间", "节能审查机关", "审查意见文号",
			"年综合能源消费量（万吨标准煤）-当量值", "年综合能源消费量（万吨标准煤）-等价值",
			"年煤品消费量（万吨，实物量）-煤品消费总量", "年煤品消费量（万吨，实物量）-煤炭消费量", "年煤品消费量（万吨，实物量）-焦炭消费量", "年煤品消费量（万吨，实物量）-兰炭消费量",
			"年煤品消费量（万吨标准煤，折标量）-煤品消费总量", "年煤品消费量（万吨标准煤，折标量）-煤炭消费量", "年煤品消费量（万吨标准煤，折标量）-焦炭消费量", "年煤品消费量（万吨标准煤，折标量）-兰炭消费量",
			"煤炭消费替代情况-是否煤炭消费替代", "煤炭消费替代情况-煤炭消费替代来源", "煤炭消费替代情况-煤炭消费替代量",
			"原料用煤情况-年原料用煤量（万吨，实物量）", "原料用煤情况-年原料用煤量（万吨标准煤，折标量）",
		}

		// 写入表头
		for i, header := range headers {
			cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue("Sheet1", cellName, header)
		}

		// 4. 设置样式
		_, dataStyle, _, _, err := applyCommonStyles(f, headers)
		if err != nil {
			result.Ok = false
			result.Message = "创建样式失败: " + err.Error()
			return result
		}

		// 设置列宽
		f.SetColWidth("Sheet1", "A", "A", 25)   // 项目名称
		f.SetColWidth("Sheet1", "B", "B", 20)   // 项目代码
		f.SetColWidth("Sheet1", "C", "C", 25)   // 建设单位
		f.SetColWidth("Sheet1", "D", "D", 30)   // 主要建设内容
		f.SetColWidth("Sheet1", "E", "E", 15)   // 项目所在省
		f.SetColWidth("Sheet1", "F", "F", 15)   // 项目所在地市
		f.SetColWidth("Sheet1", "G", "G", 15)   // 项目所在区县
		f.SetColWidth("Sheet1", "H", "H", 20)   // 所属行业大类
		f.SetColWidth("Sheet1", "I", "I", 20)   // 所属行业小类
		f.SetColWidth("Sheet1", "J", "J", 20)   // 节能审查批复时间
		f.SetColWidth("Sheet1", "K", "K", 20)   // 拟投产时间
		f.SetColWidth("Sheet1", "L", "L", 20)   // 实际投产时间
		f.SetColWidth("Sheet1", "M", "M", 25)   // 节能审查机关
		f.SetColWidth("Sheet1", "N", "N", 20)   // 审查意见文号
		f.SetColWidth("Sheet1", "O", "O", 15)   // 年综合能源消费量-当量值
		f.SetColWidth("Sheet1", "P", "P", 15)   // 年综合能源消费量-等价值
		f.SetColWidth("Sheet1", "Q", "Q", 15)   // 年煤品消费量-煤品消费总量
		f.SetColWidth("Sheet1", "R", "R", 15)   // 年煤品消费量-煤炭消费量
		f.SetColWidth("Sheet1", "S", "S", 15)   // 年煤品消费量-焦炭消费量
		f.SetColWidth("Sheet1", "T", "T", 15)   // 年煤品消费量-兰炭消费量
		f.SetColWidth("Sheet1", "U", "U", 15)   // 年煤品消费量折标量-煤品消费总量
		f.SetColWidth("Sheet1", "V", "V", 15)   // 年煤品消费量折标量-煤炭消费量
		f.SetColWidth("Sheet1", "W", "W", 15)   // 年煤品消费量折标量-焦炭消费量
		f.SetColWidth("Sheet1", "X", "X", 15)   // 年煤品消费量折标量-兰炭消费量
		f.SetColWidth("Sheet1", "Y", "Y", 15)   // 煤炭消费替代情况-是否煤炭消费替代
		f.SetColWidth("Sheet1", "Z", "Z", 20)   // 煤炭消费替代情况-煤炭消费替代来源
		f.SetColWidth("Sheet1", "AA", "AA", 20) // 煤炭消费替代情况-煤炭消费替代量
		f.SetColWidth("Sheet1", "AB", "AB", 20) // 原料用煤情况-年原料用煤量实物量
		f.SetColWidth("Sheet1", "AC", "AC", 20) // 原料用煤情况-年原料用煤量折标量

		// 4. 写入数据
		for rowIndex, item := range list {
			row := rowIndex + 2 // 从第2行开始写入数据
			col := 1

			// 项目名称
			if value, ok := item["project_name"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 项目代码
			if value, ok := item["project_code"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 建设单位
			if value, ok := item["construction_unit"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 主要建设内容
			if value, ok := item["main_construction_content"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 项目所在省
			if value, ok := item["province_name"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 项目所在地市
			if value, ok := item["city_name"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 项目所在区县
			if value, ok := item["country_name"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 所属行业大类
			if value, ok := item["trade_a"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 所属行业小类
			if value, ok := item["trade_c"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 节能审查批复时间
			if value, ok := item["examination_approval_time"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 拟投产时间
			if value, ok := item["scheduled_time"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 实际投产时间
			if value, ok := item["actual_time"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 节能审查机关
			if value, ok := item["examination_authority"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 审查意见文号
			if value, ok := item["document_number"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年综合能源消费量（万吨标准煤）-当量值
			if value, ok := item["equivalent_value"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年综合能源消费量（万吨标准煤）-等价值
			if value, ok := item["equivalent_cost"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年煤品消费量（万吨，实物量）-煤品消费总量
			if value, ok := item["pq_total_coal_consumption"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年煤品消费量（万吨，实物量）-煤炭消费量
			if value, ok := item["pq_coal_consumption"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年煤品消费量（万吨，实物量）-焦炭消费量
			if value, ok := item["pq_coke_consumption"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年煤品消费量（万吨，实物量）-兰炭消费量
			if value, ok := item["pq_blue_coke_consumption"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年煤品消费量（万吨标准煤，折标量）-煤品消费总量
			if value, ok := item["sce_total_coal_consumption"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年煤品消费量（万吨标准煤，折标量）-煤炭消费量
			if value, ok := item["sce_coal_consumption"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年煤品消费量（万吨标准煤，折标量）-焦炭消费量
			if value, ok := item["sce_coke_consumption"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年煤品消费量（万吨标准煤，折标量）-兰炭消费量
			if value, ok := item["sce_blue_coke_consumption"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 煤炭消费替代情况-是否煤炭消费替代
			if value, ok := item["is_substitution"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 煤炭消费替代情况-煤炭消费替代来源
			if value, ok := item["substitution_source"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 煤炭消费替代情况-煤炭消费替代量
			if value, ok := item["substitution_quantity"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 原料用煤情况-年原料用煤量（万吨，实物量）
			if value, ok := item["pq_annual_coal_quantity"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 原料用煤情况-年原料用煤量（万吨标准煤，折标量）
			if value, ok := item["sce_annual_coal_quantity"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}

			// 应用数据样式到整行
			startCell, _ := excelize.CoordinatesToCellName(1, row)
			endCell, _ := excelize.CoordinatesToCellName(len(headers), row)
			f.SetCellStyle("Sheet1", startCell, endCell, dataStyle)

			// 设置数据行高
			f.SetRowHeight("Sheet1", row, 20)
		}
	} else {
		// 市级别或省级别：导出区域导入状态
		// 3. 设置表头
		areaLevelName := "区域名称"
		if areaLevel == 1 {
			areaLevelName = "城市"
		} else if areaLevel == 2 {
			areaLevelName = "区县"
		}

		headers := []string{areaLevelName, "是否导入"}

		// 写入表头
		for i, header := range headers {
			cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue("Sheet1", cellName, header)
		}

		// 4. 设置样式
		_, dataStyle, importedStyle, notImportedStyle, err := applyCommonStyles(f, headers)
		if err != nil {
			result.Ok = false
			result.Message = "创建样式失败: " + err.Error()
			return result
		}

		// 设置列宽
		f.SetColWidth("Sheet1", "A", "A", 25) // 区域名称
		f.SetColWidth("Sheet1", "B", "B", 15) // 是否导入

		// 5. 写入数据
		for rowIndex, item := range list {
			row := rowIndex + 2 // 从第2行开始写入数据

			// 区域名称
			if areaName, ok := item["area_name"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(1, row)
				f.SetCellValue("Sheet1", cellName, areaName)
			}

			// 是否导入
			if isImport, ok := item["is_import"].(bool); ok {
				cellName, _ := excelize.CoordinatesToCellName(2, row)
				if isImport {
					f.SetCellValue("Sheet1", cellName, "已导入")
					f.SetCellStyle("Sheet1", cellName, cellName, importedStyle)
				} else {
					f.SetCellValue("Sheet1", cellName, "未导入")
					f.SetCellStyle("Sheet1", cellName, cellName, notImportedStyle)
				}
			}

			// 应用数据样式到第一列（区域名称）
			startCell, _ := excelize.CoordinatesToCellName(1, row)
			endCell, _ := excelize.CoordinatesToCellName(1, row)
			f.SetCellStyle("Sheet1", startCell, endCell, dataStyle)

			// 设置数据行高
			f.SetRowHeight("Sheet1", row, 20)
		}
	}

	// 5. 保存文件
	err := f.SaveAs(filePath)
	if err != nil {
		result.Ok = false
		result.Message = "保存Excel文件失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Message = "附表3导入进度导出成功"
	return result
}

// ExportAttachment2ProgressToExcel 导出附件2导入进度到Excel
func (a *App) ExportAttachment2ProgressToExcel(filePath string) db.QueryResult {
	result := db.QueryResult{}

	// 1. 获取附件2导入进度数据
	progressResult := a.QueryTableAttachment2Process()
	if !progressResult.Ok {
		result.Ok = false
		result.Message = "获取附件2导入进度失败: " + progressResult.Message
		return result
	}

	progressData, ok := progressResult.Data.(map[string]interface{})
	if !ok {
		result.Ok = false
		result.Message = "附件2导入进度数据格式错误"
		return result
	}

	areaLevel, ok := progressData["area_level"].(int)
	if !ok {
		result.Ok = false
		result.Message = "附件2区域级别数据格式错误"
		return result
	}

	list, ok := progressData["list"].([]map[string]interface{})
	if !ok {
		result.Ok = false
		result.Message = "附件2导入进度列表数据格式错误"
		return result
	}

	// 2. 创建Excel文件
	f := excelize.NewFile()
	defer f.Close()

	if areaLevel == 3 {
		// 县级别：导出所有数据
		// 3. 设置表头
		headers := []string{
			"省（市、区）", "地市（州）", "县（区）", "年份",
			"分品种煤炭消费摸底-煤合计", "分品种煤炭消费摸底-原煤", "分品种煤炭消费摸底-洗精煤", "分品种煤炭消费摸底-其他煤炭",
			"分用途煤炭消费摸底-能源加工转换-1.火力发电", "分用途煤炭消费摸底-能源加工转换-2.供热", "分用途煤炭消费摸底-能源加工转换-3.煤炭洗选", "分用途煤炭消费摸底-能源加工转换-4.炼焦", "分用途煤炭消费摸底-能源加工转换-5.炼油及煤制油", "分用途煤炭消费摸底-能源加工转换-6.制气",
			"分用途煤炭消费摸底-终端消费-1.工业", "分用途煤炭消费摸底-终端消费-#用作原料、材料", "分用途煤炭消费摸底-终端消费-2.其他用途",
			"焦炭消费摸底-焦炭",
		}

		// 写入表头
		for i, header := range headers {
			cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue("Sheet1", cellName, header)
		}

		// 4. 设置样式
		_, dataStyle, _, _, err := applyCommonStyles(f, headers)
		if err != nil {
			result.Ok = false
			result.Message = "创建样式失败: " + err.Error()
			return result
		}

		// 设置列宽
		f.SetColWidth("Sheet1", "A", "A", 15) // 省（市、区）
		f.SetColWidth("Sheet1", "B", "B", 15) // 地市（州）
		f.SetColWidth("Sheet1", "C", "C", 15) // 县（区）
		f.SetColWidth("Sheet1", "D", "D", 10) // 年份
		f.SetColWidth("Sheet1", "E", "E", 15) // 分品种煤炭消费摸底-煤合计
		f.SetColWidth("Sheet1", "F", "F", 15) // 分品种煤炭消费摸底-原煤
		f.SetColWidth("Sheet1", "G", "G", 15) // 分品种煤炭消费摸底-洗精煤
		f.SetColWidth("Sheet1", "H", "H", 15) // 分品种煤炭消费摸底-其他煤炭
		f.SetColWidth("Sheet1", "I", "I", 15) // 分用途煤炭消费摸底-能源加工转换-1.火力发电
		f.SetColWidth("Sheet1", "J", "J", 15) // 分用途煤炭消费摸底-能源加工转换-2.供热
		f.SetColWidth("Sheet1", "K", "K", 15) // 分用途煤炭消费摸底-能源加工转换-3.煤炭洗选
		f.SetColWidth("Sheet1", "L", "L", 15) // 分用途煤炭消费摸底-能源加工转换-4.炼焦
		f.SetColWidth("Sheet1", "M", "M", 15) // 分用途煤炭消费摸底-能源加工转换-5.炼油及煤制油
		f.SetColWidth("Sheet1", "N", "N", 15) // 分用途煤炭消费摸底-能源加工转换-6.制气
		f.SetColWidth("Sheet1", "O", "O", 15) // 分用途煤炭消费摸底-终端消费-1.工业
		f.SetColWidth("Sheet1", "P", "P", 15) // 分用途煤炭消费摸底-终端消费-#用作原料、材料
		f.SetColWidth("Sheet1", "Q", "Q", 15) // 分用途煤炭消费摸底-终端消费-2.其他用途
		f.SetColWidth("Sheet1", "R", "R", 15) // 焦炭消费摸底-焦炭

		// 4. 写入数据
		for rowIndex, item := range list {
			row := rowIndex + 2 // 从第2行开始写入数据
			col := 1

			// 省（市、区）
			if value, ok := item["province_name"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 地市（州）
			if value, ok := item["city_name"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 县（区）
			if value, ok := item["country_name"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 年份
			if value, ok := item["stat_date"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分品种煤炭消费摸底-煤合计
			if value, ok := item["total_coal"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分品种煤炭消费摸底-原煤
			if value, ok := item["raw_coal"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分品种煤炭消费摸底-洗精煤
			if value, ok := item["washed_coal"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分品种煤炭消费摸底-其他煤炭
			if value, ok := item["other_coal"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分用途煤炭消费摸底-能源加工转换-1.火力发电
			if value, ok := item["power_generation"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分用途煤炭消费摸底-能源加工转换-2.供热
			if value, ok := item["heating"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分用途煤炭消费摸底-能源加工转换-3.煤炭洗选
			if value, ok := item["coal_washing"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分用途煤炭消费摸底-能源加工转换-4.炼焦
			if value, ok := item["coking"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分用途煤炭消费摸底-能源加工转换-5.炼油及煤制油
			if value, ok := item["oil_refining"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分用途煤炭消费摸底-能源加工转换-6.制气
			if value, ok := item["gas_production"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分用途煤炭消费摸底-终端消费-1.工业
			if value, ok := item["industry"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分用途煤炭消费摸底-终端消费-#用作原料、材料
			if value, ok := item["raw_materials"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 分用途煤炭消费摸底-终端消费-2.其他用途
			if value, ok := item["other_uses"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}
			col++

			// 焦炭消费摸底-焦炭
			if value, ok := item["coke"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellValue("Sheet1", cellName, value)
			}

			// 应用数据样式到整行
			startCell, _ := excelize.CoordinatesToCellName(1, row)
			endCell, _ := excelize.CoordinatesToCellName(len(headers), row)
			f.SetCellStyle("Sheet1", startCell, endCell, dataStyle)

			// 设置数据行高
			f.SetRowHeight("Sheet1", row, 20)
		}
	} else {
		// 市级别或省级别：导出区域年份数据
		years, ok := progressData["years"].([]string)
		if !ok {
			result.Ok = false
			result.Message = "附件2年份数据格式错误"
			return result
		}

		// 3. 设置表头
		areaLevelName := "区域名称"
		if areaLevel == 1 {
			areaLevelName = "城市"
		} else if areaLevel == 2 {
			areaLevelName = "区县"
		}

		headers := []string{areaLevelName}
		for _, year := range years {
			headers = append(headers, year+"年数据")
		}

		// 写入表头
		for i, header := range headers {
			cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue("Sheet1", cellName, header)
		}

		// 4. 设置样式
		_, dataStyle, importedStyle, notImportedStyle, err := applyCommonStyles(f, headers)
		if err != nil {
			result.Ok = false
			result.Message = "创建样式失败: " + err.Error()
			return result
		}

		// 设置列宽
		f.SetColWidth("Sheet1", "A", "A", 25) // 区域名称
		for i := range years {
			colName, _ := excelize.ColumnNumberToName(i + 2)
			f.SetColWidth("Sheet1", colName, colName, 15) // 年份列
		}

		// 5. 写入数据
		for rowIndex, item := range list {
			row := rowIndex + 2 // 从第2行开始写入数据

			// 区域名称
			if areaName, ok := item["area_name"].(string); ok {
				cellName, _ := excelize.CoordinatesToCellName(1, row)
				f.SetCellValue("Sheet1", cellName, areaName)
			}

			// 年份数据
			for yearIndex, year := range years {
				col := yearIndex + 2 // 从第2列开始写入年份数据
				cellName, _ := excelize.CoordinatesToCellName(col, row)

				if hasData, ok := item[year].(bool); ok && hasData {
					f.SetCellValue("Sheet1", cellName, "已导入")
					f.SetCellStyle("Sheet1", cellName, cellName, importedStyle)
				} else {
					f.SetCellValue("Sheet1", cellName, "未导入")
					f.SetCellStyle("Sheet1", cellName, cellName, notImportedStyle)
				}
			}

			// 应用数据样式到第一列（区域名称）
			startCell, _ := excelize.CoordinatesToCellName(1, row)
			endCell, _ := excelize.CoordinatesToCellName(1, row)
			f.SetCellStyle("Sheet1", startCell, endCell, dataStyle)

					// 设置数据行高
		f.SetRowHeight("Sheet1", row, 20)
		}
	}

	// 5. 保存文件
	err := f.SaveAs(filePath)
	if err != nil {
		result.Ok = false
		result.Message = "保存Excel文件失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Message = "附件2导入进度导出成功"
	return result
}
