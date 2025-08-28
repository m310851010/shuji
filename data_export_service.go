package main

import (
	"encoding/json"
	"fmt"
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
	yearMap := make(map[string]*ExportDataItem)

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

				isConfirmYes := 0
				// is_confirm_yes可能为空, 需要判断是否为空
				if row["is_confirm_yes"] == nil {
					isConfirmYes = 0
				} else {
					isConfirmYes = int(row["is_confirm_yes"].(int64))
				}

				isConfirmNo := totalCount - isConfirmYes

				if item, exists := yearMap[statDate]; exists {
					item.IsConfirmYes += isConfirmYes
					item.IsConfirmNo += isConfirmNo
				} else {
					yearMap[statDate] = &ExportDataItem{
						StatDate:     statDate,
						IsConfirmYes: isConfirmYes,
						IsConfirmNo:  isConfirmNo,
					}
				}
			}
		}
	}

	// 4. 把table1Count合并到table1List的count字段
	for _, item := range yearMap {
		item.Count = table1Count
		item.IsCheckedYes = item.IsConfirmYes + item.IsConfirmNo
		item.IsCheckedNo = 0
		table1List = append(table1List, *item)
	}

	// 5. 查询设备清单记录数
	equipCountResult, err := a.db.Count("key_equipment_list", "")
	if err != nil {
		result.Ok = false
		result.Message = "查询设备清单记录数失败: " + err.Error()
		return result
	}

	equipCount := int(equipCountResult.Data.(map[string]interface{})["count"].(int64))

	// 6. 查询表2记录数，用stat_date,is_confirm分组
	table2Query := fmt.Sprintf(`
		SELECT 
			stat_date,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as is_confirm_yes,
			COUNT(1) as total_count
		FROM critical_coal_equipment_consumption 
		GROUP BY stat_date
		ORDER BY stat_date
	`, ENCRYPTED_ONE)
	table2Result, err := a.db.Query(table2Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询表2数据失败: " + err.Error()
		return result
	}

	// 7. 处理表2数据，按年份分组
	equipList := make([]ExportDataItem, 0)
	equipYearMap := make(map[string]*ExportDataItem)

	if table2Result.Ok && table2Result.Data != nil {
		if data, ok := table2Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				totalCount := 0
				if count, ok := row["total_count"].(int64); ok {
					totalCount = int(count)
				}

				isConfirmYes := 0
				if row["is_confirm_yes"] == nil {
					isConfirmYes = 0
				} else {
					isConfirmYes = int(row["is_confirm_yes"].(int64))
				}

				isConfirmNo := totalCount - isConfirmYes

				if item, exists := equipYearMap[statDate]; exists {
					item.IsConfirmYes += isConfirmYes
					item.IsConfirmNo += isConfirmNo
				} else {
					equipYearMap[statDate] = &ExportDataItem{
						StatDate:     statDate,
						IsConfirmYes: isConfirmYes,
						IsConfirmNo:  isConfirmNo,
					}
				}
			}
		}
	}

	// 8. 把equipCount合并到equipList的count字段
	for _, item := range equipYearMap {
		item.Count = equipCount
		item.IsCheckedYes = item.IsConfirmYes + item.IsConfirmNo
		item.IsCheckedNo = 0
		equipList = append(equipList, *item)
	}

	// 9. 查询附表3确认记录数和总记录数
	table3Query := fmt.Sprintf(`
		SELECT 
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as is_confirm_yes,
			COUNT(1) as total_count
		FROM fixed_assets_investment_project
	`, ENCRYPTED_ONE)
	table3Result, err := a.db.Query(table3Query)
	if err != nil {
		result.Ok = false
		result.Message = "查询附表3数据失败: " + err.Error()
		return result
	}

	table3 := ExportDataItem{}
	if table3Result.Ok && table3Result.Data != nil {
		if data, ok := table3Result.Data.([]map[string]interface{}); ok && len(data) > 0 {
			row := data[0]
			totalCount := 0
			if count, ok := row["total_count"].(int64); ok {
				totalCount = int(count)
			}

			isConfirmYes := 0
			if row["is_confirm_yes"] == nil {
				isConfirmYes = 0
			} else {
				isConfirmYes = int(row["is_confirm_yes"].(int64))
			}

			isConfirmNo := totalCount - isConfirmYes

			table3.IsConfirmYes = isConfirmYes
			table3.IsConfirmNo = isConfirmNo
			table3.Count = totalCount
			table3.IsCheckedYes = isConfirmYes + isConfirmNo
			table3.IsCheckedNo = 0
		}
	}

	// 10. 处理附件2数据
	attachment2List, err := a.processAttachment2Data()
	if err != nil {
		result.Ok = false
		result.Message = "处理附件2数据失败: " + err.Error()
		return result
	}

	// 11. 合并所有数据
	exportResult := ExportResult{
		Table1:      table1List,
		Table2:      equipList,
		Table3:      []ExportDataItem{table3},
		Attachment2: attachment2List,
	}

	result.Ok = true
	result.Data = exportResult
	result.Message = "数据查询成功"
	return result
}

func (a *App) CopySystemDb(fileName string) (*db.Database, string, error) {
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

// processAttachment2Data 处理附件2数据
func (a *App) processAttachment2Data() ([]ExportDataItem, error) {
	// 1. 获取中国区域地图数据
	areaData, err := a.ReadFile(CHINA_AREA_FILE_PATH, true)
	if err != nil {
		return nil, fmt.Errorf("读取中国区域地图文件失败: %v", err)
	}

	// 解析JSON数据为数组
	var areaList []interface{}
	err = json.Unmarshal(areaData, &areaList)
	if err != nil {
		return nil, fmt.Errorf("解析中国区域地图数据失败: %v", err)
	}

	// 2. 获取当前用户区域配置
	areaConfigResult := a.GetAreaConfig()
	if !areaConfigResult.Ok {
		return nil, fmt.Errorf("获取区域配置失败: %s", areaConfigResult.Message)
	}

	areaConfig, ok := areaConfigResult.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("区域配置数据格式错误")
	}

	provinceName := ""
	cityName := ""
	countryName := ""

	if areaConfig["province_name"] != nil {
		provinceName = fmt.Sprintf("%v", areaConfig["province_name"])
	}
	if areaConfig["city_name"] != nil {
		cityName = fmt.Sprintf("%v", areaConfig["city_name"])
	}
	if areaConfig["country_name"] != nil {
		countryName = fmt.Sprintf("%v", areaConfig["country_name"])
	}

	// 3. 根据当前用户区域查找对应的LocationItem
	var targetLocation interface{}
	var count int

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
				if children, exists := provinceMap["children"]; exists && children != nil {
					if childrenList, ok := children.([]interface{}); ok {
						count = len(childrenList)
					} else {
						count = 0
					}
				} else {
					count = 0
				}
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
								if cityChildren, exists := cityMap["children"]; exists && cityChildren != nil {
									if cityChildrenList, ok := cityChildren.([]interface{}); ok {
										count = len(cityChildrenList)
									} else {
										count = 0
									}
								} else {
									count = 0
								}
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
											// 县级别，count为1
											count = 1
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

	if targetLocation == nil {
		return nil, fmt.Errorf("未找到对应的区域信息")
	}

	// 4. 查询附件2数据，按年份和区域级别分组
	var groupByField string
	if countryName != "" {
		// 县不为空，按县分组
		groupByField = "stat_date, country_name"
	} else if cityName != "" {
		// 市不为空，县为空，按市分组
		groupByField = "stat_date, city_name"
	} else {
		// 省不为空，市和县都为空，按省分组
		groupByField = "stat_date, province_name"
	}

	attachment2Query := fmt.Sprintf(`
		SELECT 
			stat_date,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as is_confirm_yes,
			COUNT(1) as total_count
		FROM coal_consumption_report
		GROUP BY %s
		ORDER BY stat_date
	`, ENCRYPTED_ONE, groupByField)

	attachment2Result, err := a.db.Query(attachment2Query)
	if err != nil {
		return nil, fmt.Errorf("查询附件2数据失败: %v", err)
	}

	// 5. 处理查询结果，按年份分组
	attachment2List := make([]ExportDataItem, 0)
	attachment2YearMap := make(map[string]*ExportDataItem)

	if attachment2Result.Ok && attachment2Result.Data != nil {
		if data, ok := attachment2Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				totalCount := 0
				if countVal, ok := row["total_count"].(int64); ok {
					totalCount = int(countVal)
				}

				isConfirmYes := 0
				if row["is_confirm_yes"] != nil {
					if confirmYes, ok := row["is_confirm_yes"].(int64); ok {
						isConfirmYes = int(confirmYes)
					}
				}

				isConfirmNo := totalCount - isConfirmYes

				if item, exists := attachment2YearMap[statDate]; exists {
					item.IsConfirmYes += isConfirmYes
					item.IsConfirmNo += isConfirmNo
				} else {
					attachment2YearMap[statDate] = &ExportDataItem{
						StatDate:     statDate,
						IsConfirmYes: isConfirmYes,
						IsConfirmNo:  isConfirmNo,
					}
				}
			}
		}
	}

	// 6. 设置count和检查状态
	for _, item := range attachment2YearMap {
		item.Count = count // 使用从区域地图计算出的count
		item.IsCheckedYes = item.IsConfirmYes + item.IsConfirmNo
		item.IsCheckedNo = 0
		attachment2List = append(attachment2List, *item)
	}

	return attachment2List, nil
}
