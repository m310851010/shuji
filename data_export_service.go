package main

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"shuji/db"
	"time"
)

// 数据导出服务

// ExportDataItem 导出数据项结构
type ExportDataItem struct {
	StatDate     string `json:"stat_date"`      // 年份
	IsConfirmYes int    `json:"is_confirm_yes"` // 已确认记录数
	IsConfirmNo  int    `json:"is_confirm_no"`  // 未确认记录数
	Count        int    `json:"count"`          // 总记录数
	IsCheckedYes int    `json:"is_checked_yes"` // 已检查记录数
	IsCheckedNo  int    `json:"is_checked_no"`  // 未检查记录数
}

// ExportResult 导出结果结构
type ExportResult struct {
	Table1      []ExportDataItem `json:"table1"`      // 企业清单数据
	Table2      []ExportDataItem `json:"table2"`      // 设备清单数据
	Table3      []ExportDataItem `json:"table3"`      // 附表3数据
	Attachment2 []ExportDataItem `json:"attachment2"` // 附件2数据
}

func (a *App) QueryExportData() db.QueryResult {
	result := db.QueryResult{}

	// 检查数据库连接
	if a.db == nil {
		result.Ok = false
		result.Message = "数据库未初始化"
		return result
	}

	// 1. 查询企业清单记录数
	table1CountResult, err := a.db.Count("enterprise_list", "")
	if err != nil {
		result.Ok = false
		result.Message = "查询企业清单记录数失败: " + err.Error()
		return result
	}

	table1Count := int(table1CountResult.Data.(map[string]interface{})["count"].(int64))

	// 2. 查询表1记录数，用stat_date,is_confirm分组
	table1Query := fmt.Sprintf(`
		SELECT
			stat_date,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as is_confirm_yes,
			COUNT(1) as total_count
		FROM enterprise_coal_consumption_main
		GROUP BY stat_date
		ORDER BY stat_date
	`, ENCRYPTED_ONE)
	table1Result, err := a.db.Query(table1Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询表1数据失败: " + err.Error()
		return result
	}

	// 3. 处理表1数据，按年份分组
	table1List := make([]ExportDataItem, 0)

	if table1Result.Ok && table1Result.Data != nil {
		if data, ok := table1Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				totalCount := 0
				if count, ok := row["total_count"].(int64); ok {
					totalCount = int(count)
				}

				isConfirmYes := int(row["is_confirm_yes"].(int64))

				item := &ExportDataItem{
					StatDate:     statDate,
					IsConfirmYes: isConfirmYes,
					IsConfirmNo:  totalCount - isConfirmYes,
					Count:        table1Count,
					IsCheckedYes: totalCount,
					IsCheckedNo:  0,
				}

				table1List = append(table1List, *item)
			}

			if len(table1List) == 0 {
				item := &ExportDataItem{
					StatDate:     "--",
					IsConfirmYes: 0,
					IsConfirmNo:  0,
					IsCheckedYes: 0,
					IsCheckedNo:  0,
					Count:        table1Count,
				}
				table1List = append(table1List, *item)
			}
		}
	}

	// 5. 查询设备清单记录数
	equipCountResult, err := a.db.QueryRow("SELECT COUNT(distinct credit_code) as count FROM key_equipment_list ")
	if err != nil {
		result.Ok = false
		result.Message = "查询设备清单记录数失败: " + err.Error()
		return result
	}

	equipCount := int(equipCountResult.Data.(map[string]interface{})["count"].(int64))

	// 6. 查询表2记录数，用stat_date,is_confirm分组
	table2Query := fmt.Sprintf(`
		SELECT
		stat_date, SUM(CASE WHEN confirm_yes = _count and _count > 0 THEN 1 ELSE 0 END)  is_confirm_yes,
		COUNT(1) as total_count
		from 
		( SELECT 
					stat_date, credit_code,
					SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as confirm_yes,
					COUNT(1) as _count
				FROM critical_coal_equipment_consumption 
				GROUP BY stat_date, credit_code
		) t  group by stat_date
	`, ENCRYPTED_ONE)
	table2Result, err := a.db.Query(table2Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询表2数据失败: " + err.Error()
		return result
	}

	// 7. 处理表2数据，按年份分组（现在按credit_code+stat_date分组，但最终仍按年份汇总）
	equipList := make([]ExportDataItem, 0)

	if table2Result.Ok && table2Result.Data != nil {
		if data, ok := table2Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				fmt.Println("row", row)
				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				totalCount := 0
				if count, ok := row["total_count"].(int64); ok {
					totalCount = int(count)
				}

				isConfirmYes := int(row["is_confirm_yes"].(int64))

				item := &ExportDataItem{
					StatDate:     statDate,
					IsConfirmYes: isConfirmYes,
					IsConfirmNo:  totalCount - isConfirmYes,
					Count:        equipCount,
					IsCheckedYes: totalCount,
					IsCheckedNo:  0,
				}
				equipList = append(equipList, *item)
			}

			if len(equipList) == 0 {
				item := &ExportDataItem{
					StatDate:     "--",
					IsConfirmYes: 0,
					IsConfirmNo:  0,
					IsCheckedYes: 0,
					IsCheckedNo:  0,
					Count:        equipCount,
				}
				equipList = append(equipList, *item)
			}
		}
	}

	targetLocation, dataLevel, _, _ := a.getCurrentUserLocationData()
	// 获取该市下的所有县区
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

	countyCount := len(countyList) + 1

	// 9. 处理附表3数据，根据用户级别分省级和市级
	table3List, err := a.processTable3Data(countyCount, dataLevel)
	if err != nil {
		result.Ok = false
		result.Message = "处理附表3数据失败: " + err.Error()
		return result
	}

	// 10. 处理附件2数据，根据用户级别分省级和市级
	attachment2List, err := a.processAttachment2Data(countyCount, dataLevel)
	if err != nil {
		result.Ok = false
		result.Message = "处理附件2数据失败: " + err.Error()
		return result
	}

	// 11. 合并所有数据
	exportResult := ExportResult{
		Table1:      table1List,
		Table2:      equipList,
		Table3:      table3List,
		Attachment2: attachment2List,
	}

	result.Ok = true
	result.Data = exportResult
	result.Message = "数据查询成功"
	return result
}

func (a *App) CopySystemDb(fileName string) (*db.Database, string, error) {
	// 使用包装函数来处理异常
	return a.copySystemDbWithRecover(fileName)
}

// copySystemDbWithRecover 带异常处理的复制系统数据库函数
func (a *App) copySystemDbWithRecover(fileName string) (*db.Database, string, error) {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("CopySystemDb 发生异常: %v", r)
		}
	}()

	dbDstPath := GetPath(filepath.Join(DATA_DIR_NAME, DB_FILE_NAME))
	dbTempPath := GetPath(filepath.Join(DATA_DIR_NAME, fileName+time.Now().Format("20060102150405")))

	copyResult := a.Copyfile(dbDstPath, dbTempPath)
	if !copyResult.Ok {
		return nil, "", fmt.Errorf(copyResult.Data)
	}

	// 1.创建数据库连接
	newDb, err := db.NewDatabase(dbTempPath, DB_PASSWORD)
	if err != nil {
		return nil, "", err
	}

	// 2.把表data_import_record清空
	_, err = newDb.Exec("DELETE FROM data_import_record")
	if err != nil {
		newDb.Close()
		return nil, "", err
	}
	return newDb, dbTempPath, nil
}

// ExportData 导出数据
func (a *App) ExportDBData(filePath string) db.QueryResult {
	// 使用包装函数来处理异常
	return a.exportDBDataWithRecover(filePath)
}

// exportDBDataWithRecover 带异常处理的数据导出函数
func (a *App) exportDBDataWithRecover(filePath string) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ExportDBData 发生异常: %v", r)
		}
	}()

	result := db.QueryResult{}

	newDb, dbTempPath, err := a.CopySystemDb("export_")
	if err != nil {
		result.Ok = false
		result.Message = "复制数据库文件失败: " + err.Error()
		return result
	}

	newDb.Close()

	moveResult := a.Movefile(dbTempPath, filePath)
	if !moveResult.Ok {
		a.Removefile(dbTempPath) // 删除临时文件
		result.Ok = false
		result.Message = "创建数据库文件失败:" + moveResult.Data
		return result
	}

	result.Ok = true
	result.Message = "数据导出成功"
	return result
}

// 获取当前用户区域数据
// 返回：targetLocation, dataLevel, areaName, error
func (a *App) getCurrentUserLocationData() (interface{}, int, string, error) {
	// 1. 获取中国区域地图数据
	areaData, err := a.ReadFile(CHINA_AREA_FILE_PATH, true)
	if err != nil {
		return nil, 0, "", fmt.Errorf("读取中国区域地图文件失败: %v", err)
	}

	// 解析JSON数据为数组
	var areaList []interface{}
	err = json.Unmarshal(areaData, &areaList)
	if err != nil {
		return nil, 0, "", fmt.Errorf("解析中国区域地图数据失败: %v", err)
	}

	// 2. 获取当前用户区域配置
	areaConfigResult := a.GetAreaConfig()
	if !areaConfigResult.Ok {
		return nil, 0, "", fmt.Errorf("获取区域配置失败: %s", areaConfigResult.Message)
	}

	areaConfig, ok := areaConfigResult.Data.(map[string]interface{})
	if !ok {
		return nil, 0, "", fmt.Errorf("区域配置数据格式错误")
	}

	areaName := ""
	cityName := ""
	countryName := ""
	provinceName := ""

	dataLevel := 0
	if areaConfig["province_name"] != nil && areaConfig["province_name"] != "" {
		provinceName = fmt.Sprintf("%v", areaConfig["province_name"])
		areaName = provinceName
		dataLevel = 1
	}

	if areaConfig["city_name"] != nil && areaConfig["city_name"] != "" {
		cityName = fmt.Sprintf("%v", areaConfig["city_name"])
		areaName = cityName
		dataLevel = 2
	}

	if areaConfig["country_name"] != nil && areaConfig["country_name"] != "" {
		countryName = fmt.Sprintf("%v", areaConfig["country_name"])
		areaName = countryName
		dataLevel = 3
	}

	// 3. 根据当前用户区域查找对应的LocationItem
	var targetLocation interface{}

	// 查找省份
	for _, province := range areaList {
		provinceMap, ok := province.(map[string]interface{})
		if !ok {
			continue
		}

		provinceNameFromData := ""
		if name, exists := provinceMap["name"]; exists && name != nil {
			provinceNameFromData = fmt.Sprintf("%v", name)
		}

		if provinceNameFromData == provinceName {
			// 如果市为空，返回省份的children数量
			if cityName == "" {
				targetLocation = province
				break
			}

			// 查找市
			if children, exists := provinceMap["children"]; exists && children != nil {
				if childrenList, ok := children.([]interface{}); ok {
					for _, city := range childrenList {
						cityMap, ok := city.(map[string]interface{})
						if !ok {
							continue
						}

						cityNameFromData := ""
						if name, exists := cityMap["name"]; exists && name != nil {
							cityNameFromData = fmt.Sprintf("%v", name)
						}

						if cityNameFromData == cityName {
							// 如果县为空，返回市的children数量
							if countryName == "" {
								targetLocation = city
								break
							}

							// 查找县
							if cityChildren, exists := cityMap["children"]; exists && cityChildren != nil {
								if cityChildrenList, ok := cityChildren.([]interface{}); ok {
									for _, country := range cityChildrenList {
										countryMap, ok := country.(map[string]interface{})
										if !ok {
											continue
										}

										countryNameFromData := ""
										if name, exists := countryMap["name"]; exists && name != nil {
											countryNameFromData = fmt.Sprintf("%v", name)
										}

										if countryNameFromData == countryName {
											targetLocation = country
											break
										}
									}
								}
							}
							break
						}
					}
				}
			}
			break
		}
	}

	// 如果找不到匹配的区域，尝试使用第一个省份作为默认值
	if targetLocation == nil && len(areaList) > 0 {
		targetLocation = areaList[0]
		if provinceName == "" {
			if firstProvince, ok := areaList[0].(map[string]interface{}); ok {
				if name, exists := firstProvince["name"]; exists {
					areaName = fmt.Sprintf("%v", name)
					dataLevel = 1
				}
			}
		}
	}

	return targetLocation, dataLevel, areaName, nil
}

// processTable3Data 处理附表3数据
func (a *App) processTable3Data(countyCount int, dataLevel int) ([]ExportDataItem, error) {
	// 批量查询所有县区的校核状态
	query := fmt.Sprintf(`
		SELECT 
			examination_authority,
			COUNT(1) as total_count,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as is_confirm_yes
		FROM fixed_assets_investment_project 
		GROUP BY examination_authority
	`, ENCRYPTED_ONE)

	result, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}

	Table3Count := countyCount
	if dataLevel == 3 {
		Table3Count = 0
	}
	table3List := make([]ExportDataItem, 0)

	item := &ExportDataItem{
		StatDate:     "--",
		IsConfirmYes: 0,
		IsConfirmNo:  0,
		Count:        Table3Count,
		IsCheckedYes: 0,
		IsCheckedNo:  0,
	}

	if result.Ok && result.Data != nil {
		if data, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				totalCount := 0
				if count, ok := row["total_count"].(int64); ok {
					totalCount = int(count)
				}

				isConfirmYes := int(row["is_confirm_yes"].(int64))
				if isConfirmYes == totalCount && totalCount > 0 {
					item.IsConfirmYes++
				}
			}

			item.IsCheckedYes = len(data)
			item.IsConfirmNo = item.IsCheckedYes - item.IsConfirmYes
		}
	}

	table3List = append(table3List, *item)
	return table3List, nil
}

// processAttachment2Data 处理附件2数据，根据用户级别分省级和市级
func (a *App) processAttachment2Data(countyCount int, dataLevel int) ([]ExportDataItem, error) {

	// 批量查询所有县区的数据状态
	query := fmt.Sprintf(`
		SELECT 
			stat_date, SUM(CASE WHEN confirm_yes = _count and _count > 0 THEN 1 ELSE 0 END)  is_confirm_yes,
			COUNT(1) as total_count
			FROM 
			( SELECT 
						country_name, stat_date,
						SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as confirm_yes,
						COUNT(1) as _count
					FROM coal_consumption_report 
					GROUP BY country_name, stat_date
			
			) t  GROUP BY stat_date
	`, ENCRYPTED_ONE)

	result, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}

	// 市级逻辑：总数为下辖县区数量+1(本市)
	Attachment2Count := countyCount
	if dataLevel == 3 {
		Attachment2Count = 1
	}
	attachment2List := make([]ExportDataItem, 0)

	if result.Ok && result.Data != nil {
		if data, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				totalCount := 0
				if count, ok := row["total_count"].(int64); ok {
					totalCount = int(count)
				}

				isConfirmYes := int(row["is_confirm_yes"].(int64))

				item := &ExportDataItem{
					StatDate:     statDate,
					IsConfirmYes: isConfirmYes,
					IsConfirmNo:  totalCount - isConfirmYes,
					Count:        Attachment2Count,
					IsCheckedYes: totalCount,
					IsCheckedNo:  0,
				}
				attachment2List = append(attachment2List, *item)
			}

			if len(data) == 0 {
				item := &ExportDataItem{
					StatDate:     "--",
					IsConfirmYes: 0,
					IsConfirmNo:  0,
					IsCheckedYes: 0,
					IsCheckedNo:  0,
					Count:        Attachment2Count,
				}
				attachment2List = append(attachment2List, *item)
			}
		}
	}

	return attachment2List, nil
}
