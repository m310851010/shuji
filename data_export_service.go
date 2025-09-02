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

	// 7. 处理表2数据，按年份分组（现在按credit_code+stat_date分组，但最终仍按年份汇总）
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

				// 按年份汇总数据
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

	// 9. 处理附表3数据，根据用户级别分省级和市级
	table3List, err := a.processTable3Data()
	if err != nil {
		result.Ok = false
		result.Message = "处理附表3数据失败: " + err.Error()
		return result
	}

	// 10. 处理附件2数据，根据用户级别分省级和市级
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

// processTable3Data 处理附表3数据，根据用户级别分省级和市级
func (a *App) processTable3Data() ([]ExportDataItem, error) {
	targetLocation, dataLevel, _, err := a.getCurrentUserLocationData()
	if err != nil {
		return nil, fmt.Errorf("获取当前用户区域数据失败: %v", err)
	}

	if targetLocation == nil {
		return nil, fmt.Errorf("未找到对应的区域信息")
	}

	if dataLevel == 1 {
		// 省级：没数据时总数为0，有数据时总数为0但导入进度为1
		return a.processTable3ProvinceLevel()
	} else if dataLevel == 2 {
		// 市级：总数为下辖县区数量，统计有数据的县区数量
		return a.processTable3CityLevel(targetLocation)
	} else {
		// 县级：直接查询所有数据
		return a.processTable3CountyLevel()
	}
}

// processTable3ProvinceLevel 处理附表3省级数据
func (a *App) processTable3ProvinceLevel() ([]ExportDataItem, error) {
	// 查询附表3是否有数据
	table3Query := fmt.Sprintf(`
		SELECT 
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as is_confirm_yes,
			COUNT(1) as total_count
		FROM fixed_assets_investment_project
	`, ENCRYPTED_ONE)

	table3Result, err := a.db.Query(table3Query)
	if err != nil {
		return nil, fmt.Errorf("查询附表3数据失败: %v", err)
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

			// 省级逻辑：没数据时总数为0，有数据时总数为0但导入进度为1
			table3.Count = 0 // 总数始终为0
			if totalCount > 0 {
				// 有数据时，导入进度为1，自动校验为1
				table3.IsConfirmYes = 1
				table3.IsConfirmNo = 0
				// 人工校验：全部数据is_confirm为已校核则为1，否则为0
				if isConfirmYes == totalCount {
					table3.IsCheckedYes = 1
					table3.IsCheckedNo = 0
				} else {
					table3.IsCheckedYes = 0
					table3.IsCheckedNo = 1
				}
			} else {
				// 没数据时，所有数量都是0
				table3.IsConfirmYes = 0
				table3.IsConfirmNo = 0
				table3.IsCheckedYes = 0
				table3.IsCheckedNo = 0
			}
		}
	}

	return []ExportDataItem{table3}, nil
}

// processTable3CityLevel 处理附表3市级数据
func (a *App) processTable3CityLevel(targetLocation interface{}) ([]ExportDataItem, error) {
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

	// 批量查询县区校核状态
	checkedCounties, isConfirmCounties, err := a.batchCheckTable3CountiesChecked(countyList)
	if err != nil {
		return nil, fmt.Errorf("批量查询县区校核状态失败: %v", err)
	}

	// 统计已确认的县区数量和已校核的县区数量
	confirmedCount := 0
	checkedCount := 0

	for _, countyName := range countyList {
		// 模型校验：有数据就表示模型验证通过
		if checkedCounties[countyName] {
			checkedCount++
		}

		// 人工确认：全部数据已确认
		if isConfirmCounties[countyName] {
			confirmedCount++
		}
	}

	// 市级逻辑：总数为下辖县区数量
	totalCount := len(countyList)

	table3 := ExportDataItem{
		Count:        totalCount,
		IsConfirmYes: confirmedCount, // 已确认数量是已确认的下辖县区数量
		IsConfirmNo:  totalCount - confirmedCount,
		IsCheckedYes: checkedCount, // 人工校验数量是已校核的下辖县区数量
		IsCheckedNo:  totalCount - checkedCount,
	}

	return []ExportDataItem{table3}, nil
}

// processTable3CountyLevel 处理附表3县级数据
func (a *App) processTable3CountyLevel() ([]ExportDataItem, error) {
	// 县级直接查询所有数据
	table3Query := fmt.Sprintf(`
		SELECT 
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as is_confirm_yes,
			COUNT(1) as total_count
		FROM fixed_assets_investment_project
	`, ENCRYPTED_ONE)

	table3Result, err := a.db.Query(table3Query)
	if err != nil {
		return nil, fmt.Errorf("查询附表3数据失败: %v", err)
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

	return []ExportDataItem{table3}, nil
}

// batchCheckTable3CountiesChecked 批量检查附表3县区是否已校核
func (a *App) batchCheckTable3CountiesChecked(countyList []string) (map[string]bool, map[string]bool, error) {

	// 批量查询所有县区的校核状态
	query := fmt.Sprintf(`
		SELECT 
			examination_authority,
			COUNT(1) as total_count,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as confirmed_count
		FROM fixed_assets_investment_project 
		GROUP BY examination_authority
	`, ENCRYPTED_ONE)

	result, err := a.db.Query(query)
	if err != nil {
		return nil, nil, err
	}

	checkedCounties := make(map[string]bool)
	isConfirmCounties := make(map[string]bool)

	// 初始化所有县区为未校核
	for _, countyName := range countyList {
		checkedCounties[countyName] = false
		isConfirmCounties[countyName] = false
	}

	if result.Ok && result.Data != nil {
		if data, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				examinationAuthority := ""
				if authority, ok := row["examination_authority"].(string); ok {
					examinationAuthority = authority
				}

				totalCount := int(row["total_count"].(int64))
				confirmedCount := 0
				if row["confirmed_count"] != nil {
					confirmedCount = int(row["confirmed_count"].(int64))
				}

				// 提取区域名称
				areaName := a.extractAreaFromAuthority(examinationAuthority)
				if areaName != "" {
					// 如果总数大于0且全部已确认，则认为已校核
					isConfirmCounties[areaName] = totalCount > 0 && confirmedCount == totalCount
					checkedCounties[areaName] = totalCount > 0
				}
			}
		}
	}

	return checkedCounties, isConfirmCounties, nil
}

// processAttachment2Data 处理附件2数据，根据用户级别分省级和市级
func (a *App) processAttachment2Data() ([]ExportDataItem, error) {
	targetLocation, dataLevel, _, err := a.getCurrentUserLocationData()
	if err != nil {
		return nil, fmt.Errorf("获取当前用户区域数据失败: %v", err)
	}

	if targetLocation == nil {
		return nil, fmt.Errorf("未找到对应的区域信息")
	}

	if dataLevel == 1 {
		// 省级：没数据时总数为1，有数据时总数为1
		return a.processAttachment2ProvinceLevel()
	} else if dataLevel == 2 {
		// 市级：总数为下辖县区数量+1，统计有数据的县区和市
		return a.processAttachment2CityLevel(targetLocation)
	} else {
		// 县级：直接查询所有数据
		return a.processAttachment2CountyLevel()
	}
}

// processAttachment2ProvinceLevel 处理附件2省级数据
func (a *App) processAttachment2ProvinceLevel() ([]ExportDataItem, error) {
	// 查询附件2数据，按年份分组
	attachment2Query := fmt.Sprintf(`
		SELECT 
			stat_date,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as is_confirm_yes,
			COUNT(1) as total_count
		FROM coal_consumption_report
		GROUP BY stat_date
		ORDER BY stat_date
	`, ENCRYPTED_ONE)

	attachment2Result, err := a.db.Query(attachment2Query)
	if err != nil {
		return nil, fmt.Errorf("查询附件2数据失败: %v", err)
	}

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
				if count, ok := row["total_count"].(int64); ok {
					totalCount = int(count)
				}

				isConfirmYes := 0
				if row["is_confirm_yes"] == nil {
					isConfirmYes = 0
				} else {
					isConfirmYes = int(row["is_confirm_yes"].(int64))
				}

				// 省级逻辑：没数据时总数为1，有数据时总数为1
				attachment2 := ExportDataItem{
					StatDate: statDate,
					Count:    1, // 总数始终为1
				}

				if totalCount > 0 {
					// 有数据时，导入进度为1，自动校验为1
					attachment2.IsConfirmYes = 1
					attachment2.IsConfirmNo = 0
					// 人工校验：全部数据is_confirm为已校核则为1，否则为0
					if isConfirmYes == totalCount {
						attachment2.IsCheckedYes = 1
						attachment2.IsCheckedNo = 0
					} else {
						attachment2.IsCheckedYes = 0
						attachment2.IsCheckedNo = 1
					}
				} else {
					// 没数据时，导入进度为0，自动校验为0，人工校验为0
					attachment2.IsConfirmYes = 0
					attachment2.IsConfirmNo = 1
					attachment2.IsCheckedYes = 0
					attachment2.IsCheckedNo = 1
				}

				attachment2YearMap[statDate] = &attachment2
			}
		}
	}

	// 转换为列表
	for _, item := range attachment2YearMap {
		attachment2List = append(attachment2List, *item)
	}

	return attachment2List, nil
}

// processAttachment2CountyLevel 处理附件2县级数据
func (a *App) processAttachment2CountyLevel() ([]ExportDataItem, error) {
	// 县级直接查询所有数据，按年份分组
	attachment2Query := fmt.Sprintf(`
		SELECT 
			stat_date,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as is_confirm_yes,
			COUNT(1) as total_count
		FROM coal_consumption_report
		GROUP BY stat_date
		ORDER BY stat_date
	`, ENCRYPTED_ONE)

	attachment2Result, err := a.db.Query(attachment2Query)
	if err != nil {
		return nil, fmt.Errorf("查询附件2数据失败: %v", err)
	}

	attachment2List := make([]ExportDataItem, 0)
	attachment2YearMap := make(map[string]*ExportDataItem)

	if attachment2Result.Ok && attachment2Result.Data != nil {
		if data, ok := attachment2Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				statDate := ""
				if date, ok := row["stat_date"].(string); ok {
					statDate = date
				}

				// 如果stat_date为空，跳过这条记录
				if statDate == "" {
					continue
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

				// 县级逻辑：Count=1, 只有当表中的数据is_confirm全部为ENCRYPTED_ONE时, IsConfirmYes计为1, IsConfirmNo:0,否则IsConfirmYes计为0,IsConfirmNo:0
				attachment2 := ExportDataItem{
					StatDate: statDate,
					Count:    1, // 总数始终为1
				}

				if totalCount > 0 {
					// 表中有数据,IsCheckedYes=1,IsCheckedNo=0 (有数据就表示已校核)
					attachment2.IsCheckedYes = 1
					attachment2.IsCheckedNo = 0

					// 只有当表中的数据is_confirm全部为ENCRYPTED_ONE时, IsConfirmYes计为1, IsConfirmNo:0,否则IsConfirmYes计为0,IsConfirmNo:0
					if isConfirmYes == totalCount {
						attachment2.IsConfirmYes = 1
						attachment2.IsConfirmNo = 0
					} else {
						attachment2.IsConfirmYes = 0
						attachment2.IsConfirmNo = 0
					}
				} else {
					// 表中没有数据
					attachment2.IsCheckedYes = 0
					attachment2.IsCheckedNo = 1
					attachment2.IsConfirmYes = 0
					attachment2.IsConfirmNo = 0
				}

				attachment2YearMap[statDate] = &attachment2
			}
		}
	}

	// 如果没有查询到任何数据，也要创建一个默认记录
	if len(attachment2YearMap) == 0 {
		// 创建一个默认的年份记录（比如当前年份）
		defaultYear := time.Now().Format("2006")
		defaultItem := ExportDataItem{
			StatDate:     defaultYear,
			Count:        1, // 县级总数始终为1
			IsConfirmYes: 0,
			IsConfirmNo:  0,
			IsCheckedYes: 0,
			IsCheckedNo:  1, // 没有数据时，未检查为1
		}
		attachment2List = append(attachment2List, defaultItem)
	} else {
		// 转换为列表
		for _, item := range attachment2YearMap {
			attachment2List = append(attachment2List, *item)
		}
	}

	return attachment2List, nil
}

// processAttachment2CityLevel 处理附件2市级数据
func (a *App) processAttachment2CityLevel(targetLocation interface{}) ([]ExportDataItem, error) {
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

	// 获取当前市名称

	// 批量查询县区校核状态（参考附表3的逻辑）
	checkedCounties, isConfirmCounties, err := a.batchCheckAttachment2CountiesChecked(countyList)
	if err != nil {
		return nil, fmt.Errorf("批量查询县区校核状态失败: %v", err)
	}

	// 统计已确认的县区数量和已校核的县区数量
	confirmedCount := 0
	checkedCount := 0
	for _, countyName := range countyList {
		if checkedCounties[countyName] {
			checkedCount++ // 已校核的县区数量
		}
		if isConfirmCounties[countyName] {
			confirmedCount++
		}
		// 如果该县区不是全部数据已确认，则计为0（不增加confirmedCount）
	}

	// 市级逻辑：总数为下辖县区数量+1(本市)
	totalCount := len(countyList) + 1

	// 查询本市的数据确认状态
	cityConfirmedCount, err := a.queryCityAttachment2ConfirmedCount()
	if err != nil {
		return nil, fmt.Errorf("查询本市确认状态失败: %v", err)
	}

	attachment2 := ExportDataItem{
		Count:        totalCount,
		IsConfirmYes: confirmedCount + cityConfirmedCount, // 已确认数量 = 所有已确认县区的数量(0或1) + 所有已人工确认县区的数量(0或1) + 本市是否已确认(0或1)
		IsConfirmNo:  totalCount - (confirmedCount + cityConfirmedCount),
		IsCheckedYes: checkedCount, // 人工校验数量是已校核的下辖县区数量
		IsCheckedNo:  totalCount - checkedCount,
	}

	return []ExportDataItem{attachment2}, nil
}

// queryAndParseAttachment2Data 查询并解析附件2数据（参考附表3的逻辑）
func (a *App) queryAndParseAttachment2Data() (map[string]bool, error) {
	// 查询附件2数据，按country_name分组
	attachment2Query := `
		SELECT 
			country_name,
			COUNT(1) as record_count
		FROM coal_consumption_report 
		WHERE country_name IS NOT NULL AND country_name != ''
		GROUP BY country_name
	`
	attachment2Result, err := a.db.Query(attachment2Query)
	if err != nil {
		return nil, err
	}

	importedCounties := make(map[string]bool)

	if attachment2Result.Ok && attachment2Result.Data != nil {
		if data, ok := attachment2Result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				countryName := ""
				if name, ok := row["country_name"].(string); ok {
					countryName = name
				}

				recordCount := 0
				if count, ok := row["record_count"].(int64); ok {
					recordCount = int(count)
				}

				// 如果有记录，则认为该县区已导入数据
				if recordCount > 0 {
					importedCounties[countryName] = true
				}
			}
		}
	}

	return importedCounties, nil
}

// batchCheckAttachment2CountiesChecked 批量检查附件2县区是否已校核（模型校验：有数据就表示模型验证通过）
func (a *App) batchCheckAttachment2CountiesChecked(countyList []string) (map[string]bool, map[string]bool, error) {
	if len(countyList) == 0 {
		return make(map[string]bool), make(map[string]bool), nil
	}

	// 批量查询所有县区的数据状态
	query := fmt.Sprintf(`
		SELECT 
			country_name,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as checked_count,
			COUNT(1) as total_count
		FROM coal_consumption_report 
		GROUP BY country_name
	`, ENCRYPTED_ONE)

	result, err := a.db.Query(query)
	if err != nil {
		return nil, nil, err
	}

	checkedCounties := make(map[string]bool)
	isConfirmCounties := make(map[string]bool)

	// 初始化所有县区为未校核
	for _, countyName := range countyList {
		checkedCounties[countyName] = false
		isConfirmCounties[countyName] = false
	}

	if result.Ok && result.Data != nil {
		if data, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				countryName := ""
				if name, ok := row["country_name"].(string); ok {
					countryName = name
				}

				totalCount := int(row["total_count"].(int64))
				checkedCount := 0
				if row["checked_count"] != nil {
					checkedCount = int(row["checked_count"].(int64))
				}
				fmt.Println("countryName==", countryName, "totalCount==", totalCount, "checkedCount==", checkedCount)
				// 只更新countyList中存在的县区
				if _, exists := checkedCounties[countryName]; exists {
					// 模型校验：有数据就表示模型验证通过
					checkedCounties[countryName] = totalCount > 0
				}
				if _, exists := isConfirmCounties[countryName]; exists {
					isConfirmCounties[countryName] = totalCount > 0 && checkedCount == totalCount
				}
			}
		}
	}

	return checkedCounties, isConfirmCounties, nil
}

// batchCheckAttachment2CountiesConfirmed 批量检查附件2县区是否已人工确认
func (a *App) batchCheckAttachment2CountiesConfirmed(countyList []string) (map[string]bool, error) {
	if len(countyList) == 0 {
		return make(map[string]bool), nil
	}

	// 批量查询所有县区的人工确认状态
	query := fmt.Sprintf(`
		SELECT 
			country_name,
			COUNT(1) as total_count,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as confirmed_count
		FROM coal_consumption_report 
		GROUP BY country_name
	`, ENCRYPTED_ONE)

	result, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}

	confirmedCounties := make(map[string]bool)

	// 初始化所有县区为未确认
	for _, countyName := range countyList {
		confirmedCounties[countyName] = false
	}

	if result.Ok && result.Data != nil {
		if data, ok := result.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				countryName := ""
				if name, ok := row["country_name"].(string); ok {
					countryName = name
				}

				totalCount := int(row["total_count"].(int64))
				confirmedCount := 0
				if row["confirmed_count"] != nil {
					confirmedCount = int(row["confirmed_count"].(int64))
				}

				// 只更新countyList中存在的县区
				if _, exists := confirmedCounties[countryName]; exists {
					// 人工确认：全部数据已确认
					confirmedCounties[countryName] = totalCount > 0 && confirmedCount == totalCount
				}
			}
		}
	}

	return confirmedCounties, nil
}

// queryCityAttachment2ConfirmedCount 查询本市附件2数据的确认状态
func (a *App) queryCityAttachment2ConfirmedCount() (int, error) {
	// 查询本市的数据确认状态（city_name不为空，country_name为空）
	query := fmt.Sprintf(`
		SELECT 
			COUNT(1) as total_count,
			SUM(CASE WHEN is_confirm = '%s' THEN 1 ELSE 0 END) as confirmed_count
		FROM coal_consumption_report 
		WHERE city_name IS NOT NULL AND city_name != '' 
		AND (country_name IS NULL OR country_name = '')
	`, ENCRYPTED_ONE)

	result, err := a.db.Query(query)
	if err != nil {
		return 0, err
	}

	if result.Ok && result.Data != nil {
		if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
			row := data[0]
			totalCount := 0
			if count, ok := row["total_count"].(int64); ok {
				totalCount = int(count)
			}

			confirmedCount := 0
			if row["confirmed_count"] != nil {
				if count, ok := row["confirmed_count"].(int64); ok {
					confirmedCount = int(count)
				}
			}

			// 如果有数据且全部已确认，返回1，否则返回0
			if totalCount > 0 && confirmedCount == totalCount {
				return 1, nil
			}
		}
	}

	return 0, nil
}
