package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"shuji/db"
	"strconv"
	"time"
)

// MergeResult 合并结果
type MergeResult struct {
	Ok            bool   `json:"ok"`
	Message       string `json:"message"`
	SuccessCount  int    `json:"successCount"`
	ConflictCount int    `json:"conflictCount"`
	ErrorCount    int    `json:"errorCount"`
}

// ConflictData 冲突数据
type ConflictData struct {
	TableName    string                   `json:"tableName"`
	SourceData   []map[string]interface{} `json:"sourceData"`
	TargetData   []map[string]interface{} `json:"targetData"`
	SourceFile   string                   `json:"sourceFile"`
	ConflictKeys []string                 `json:"conflictKeys"`
}

// MergeDatabase 合并数据库
func (a *App) MergeDatabase(province string, city string, country string, sourceDbPath []string) db.QueryResult {
	result := db.QueryResult{}

	// 1. 验证区域一致性
	areaValidation := a.validateAreaConsistency(province, city, country)
	if !areaValidation.Ok {
		result.Ok = false
		result.Message = areaValidation.Message
		return result
	}

	// 2. 复制系统数据库到临时文件并打开数据库连接
	newDb, dbTempPath, err := a.CopySystemDb("merge_")
	if err != nil {
		result.Ok = false
		result.Message = "复制数据库文件失败: " + err.Error()
		return result
	}
	defer newDb.Close()

	// 3. 复制源数据库到临时文件并打开数据库连接
	var sourceDbs []*db.Database
	var sourceDbPaths []string
	var originalSourcePaths []string
	var failedFiles []string

	for i, sourceDbPath := range sourceDbPath {
		fileName := filepath.Base(sourceDbPath)
		dbDstPath := GetPath(filepath.Join(DATA_DIR_NAME, "merge_"+strconv.Itoa(i)+"_"+fileName))
		copyResult := a.Copyfile(sourceDbPath, dbDstPath)
		fmt.Println("复制文件copyResult", copyResult)
		if !copyResult.Ok {
			failedFiles = append(failedFiles, sourceDbPath)
			continue
		}

		fmt.Println("链接数据库", dbDstPath)
		// 创建数据库连接
		sourceDb, err := db.NewDatabase(dbDstPath, DB_PASSWORD)
		fmt.Println("链接数据库sourceDb", sourceDb)
		if err != nil {
			failedFiles = append(failedFiles, sourceDbPath)
			continue
		}
		defer sourceDb.Close()

		sourceDbs = append(sourceDbs, sourceDb)
		sourceDbPaths = append(sourceDbPaths, dbDstPath)
		originalSourcePaths = append(originalSourcePaths, sourceDbPath)
	}

	// 如果没有成功打开任何数据库文件，则返回错误
	if len(sourceDbs) == 0 {
		result.Ok = false
		result.Message = fmt.Sprintf("所有数据库文件打开失败: %v", failedFiles)
		a.Removefile(dbTempPath) // 删除临时文件
		return result
	}

	// 4. 检查数据冲突
	// 检查表1冲突（规上企业煤炭消费信息主表）
	table1Conflicts := a.checkTable1Conflicts(newDb, sourceDbs, sourceDbPaths, originalSourcePaths)
	table1ConflictCount := len(table1Conflicts)

	// 检查表2冲突（重点耗煤装置煤炭消耗信息表）
	table2Conflicts := a.checkTable2Conflicts(newDb, sourceDbs, sourceDbPaths, originalSourcePaths)
	table2ConflictCount := len(table2Conflicts)

	// 检查表3冲突（固定资产投资项目节能审查煤炭消费情况汇总表）
	table3Conflicts := a.checkTable3Conflicts(newDb, sourceDbs, sourceDbPaths, originalSourcePaths)
	table3ConflictCount := len(table3Conflicts)

	// 检查附件2冲突（煤炭消费状况表）
	attachment2Conflicts := a.checkAttachment2Conflicts(newDb, sourceDbs, sourceDbPaths, originalSourcePaths)
	attachment2ConflictCount := len(attachment2Conflicts)

	totalConflictCount := table1ConflictCount + table2ConflictCount + table3ConflictCount + attachment2ConflictCount

	// 5. 记录冲突信息（如果有冲突）
	var conflictInfo map[string]interface{}
	if totalConflictCount > 0 {
		// 收集每个表的冲突文件名
		table1FileNames := a.collectConflictFileNames(table1Conflicts)
		table2FileNames := a.collectConflictFileNames(table2Conflicts)
		table3FileNames := a.collectConflictFileNames(table3Conflicts)
		attachment2FileNames := a.collectConflictFileNames(attachment2Conflicts)

		conflictInfo = map[string]interface{}{
			"hasConflict": true,
			"table1Conflicts": map[string]interface{}{
				"conflicts":     table1Conflicts,
				"conflictCount": table1ConflictCount,
				"fileNames":     table1FileNames,
			},
			"table2Conflicts": map[string]interface{}{
				"conflicts":     table2Conflicts,
				"conflictCount": table2ConflictCount,
				"fileNames":     table2FileNames,
			},
			"table3Conflicts": map[string]interface{}{
				"conflicts":     table3Conflicts,
				"conflictCount": table3ConflictCount,
				"fileNames":     table3FileNames,
			},
			"attachment2Conflicts": map[string]interface{}{
				"conflicts":     attachment2Conflicts,
				"conflictCount": attachment2ConflictCount,
				"fileNames":     attachment2FileNames,
			},
			"totalConflictCount": totalConflictCount,
		}
	}

	// 6. 如果没有冲突，执行数据合并（使用事务）
	successCount := 0
	errorCount := 0

	// 开始事务
	tx, err := newDb.Begin()
	if err != nil {
		result.Ok = false
		result.Message = "开始事务失败: " + err.Error()
		a.Removefile(dbTempPath) // 删除临时文件
		return result
	}

	// 确保事务回滚（如果出错）
	var txErr error
	defer func() {
		if txErr != nil {
			tx.Rollback()
		}
	}()

	for i, sourceDb := range sourceDbs {
		// 检查当前数据库文件是否有冲突
		hasConflict := false
		originalFileName := filepath.Base(originalSourcePaths[i])

		// 检查表1冲突
		for _, conflict := range table1Conflicts {
			if len(conflict.SourceData) > 0 {
				conflictData := conflict.SourceData[0]
				if _, exists := conflictData[originalFileName]; exists {
					hasConflict = true
					break
				}
			}
		}

		// 检查表2冲突
		if !hasConflict {
			for _, conflict := range table2Conflicts {
				if len(conflict.SourceData) > 0 {
					conflictData := conflict.SourceData[0]
					if _, exists := conflictData[originalFileName]; exists {
						hasConflict = true
						break
					}
				}
			}
		}

		// 检查表3冲突
		if !hasConflict {
			for _, conflict := range table3Conflicts {
				if len(conflict.SourceData) > 0 {
					conflictData := conflict.SourceData[0]
					if _, exists := conflictData[originalFileName]; exists {
						hasConflict = true
						break
					}
				}
			}
		}

		// 检查附件2冲突
		if !hasConflict {
			for _, conflict := range attachment2Conflicts {
				if len(conflict.SourceData) > 0 {
					conflictData := conflict.SourceData[0]
					if _, exists := conflictData[originalFileName]; exists {
						hasConflict = true
						break
					}
				}
			}
		}

		// 如果有冲突，跳过这个数据库文件,不合并数据
		if hasConflict {
			continue
		}

		// 合并表1数据
		table1Result := a.mergeTable1DataWithTx(tx, sourceDb)
		if table1Result.Ok {
			successCount += table1Result.SuccessCount
		} else {
			errorCount += table1Result.ErrorCount
		}

		// 合并表2数据
		table2Result := a.mergeTable2DataWithTx(tx, sourceDb)
		if table2Result.Ok {
			successCount += table2Result.SuccessCount
		} else {
			errorCount += table2Result.ErrorCount
		}

		// 合并表3数据
		table3Result := a.mergeTable3DataWithTx(tx, sourceDb)
		if table3Result.Ok {
			successCount += table3Result.SuccessCount
		} else {
			errorCount += table3Result.ErrorCount
		}

		// 合并附件2数据
		attachment2Result := a.mergeAttachment2DataWithTx(tx, sourceDb)
		if attachment2Result.Ok {
			successCount += attachment2Result.SuccessCount
		} else {
			errorCount += attachment2Result.ErrorCount
		}

		// 删除源临时数据库文件
		os.Remove(sourceDbPaths[i])
	}

	// 提交事务
	txErr = tx.Commit()
	if txErr != nil {
		result.Ok = false
		result.Message = "提交事务失败: " + txErr.Error()
		a.Removefile(dbTempPath) // 删除临时文件
		return result
	}

	// 8. 返回合并结果
	result.Ok = true
	result.Message = fmt.Sprintf("数据库合并完成。成功合并: %d 条数据，错误: %d 条", successCount, errorCount)
	result.Data = map[string]interface{}{
		"successCount":       successCount,
		"errorCount":         errorCount,
		"totalConflictCount": totalConflictCount,
		"failedFiles":        failedFiles,
	}

	// 如果有冲突信息，添加到返回结果中
	if conflictInfo != nil {
		// 合并冲突信息到现有结果中，而不是替换
		if resultData, ok := result.Data.(map[string]interface{}); ok {
			for key, value := range conflictInfo {
				resultData[key] = value
			}
			result.Data = resultData
		}
	}

	return result
}

// collectConflictFileNames 收集冲突文件名
func (a *App) collectConflictFileNames(conflicts []ConflictData) []string {
	fileNames := make(map[string]bool)

	for _, conflict := range conflicts {
		if len(conflict.SourceData) > 0 {
			conflictData := conflict.SourceData[0]
			for key, value := range conflictData {
				log.Println("key===", key, "value===", value)

				// 跳过业务字段，只收集文件名
				if key != "unit_name" && key != "credit_code" && key != "stat_date" &&
					key != "project_name" && key != "project_code" && key != "document_number" &&
					key != "province_name" && key != "city_name" && key != "country_name" {
					fileNames[key] = true
				}
			}
		}
	}

	// 转换为字符串数组
	result := make([]string, 0, len(fileNames))
	for fileName := range fileNames {
		result = append(result, fileName)
	}

	return result
}

// validateAreaConsistency 验证区域一致性
func (a *App) validateAreaConsistency(province, city, country string) db.QueryResult {
	// 获取当前系统配置的区域信息
	areaConfigResult := a.GetAreaConfig()
	if !areaConfigResult.Ok {
		return db.QueryResult{Ok: false, Message: "获取区域配置失败: " + areaConfigResult.Message}
	}

	// 从返回结果中提取区域信息
	areaConfigData, ok := areaConfigResult.Data.(map[string]interface{})
	if !ok {
		return db.QueryResult{Ok: false, Message: "区域配置数据格式错误"}
	}

	configProvince := fmt.Sprintf("%v", areaConfigData["province_name"])
	configCity := fmt.Sprintf("%v", areaConfigData["city_name"])
	configCountry := fmt.Sprintf("%v", areaConfigData["country_name"])

	// 根据传入参数进行区域一致性验证
	// 如果country有值就比较country
	if country != "" {
		if configCountry != country {
			return db.QueryResult{Ok: false, Message: "上传数据不在同一个县、无法合并"}
		}
	}

	// 如果city有值就比较city
	if city != "" {
		if configCity != city {
			return db.QueryResult{Ok: false, Message: "上传数据不在同一个市、无法合并"}
		}
	}

	// 如果province有值就比较province
	if province != "" {
		if configProvince != province {
			return db.QueryResult{Ok: false, Message: "上传数据不在同一个省、无法合并"}
		}
	}

	return db.QueryResult{Ok: true, Message: "区域一致性验证通过"}
}

// checkTable1Conflicts 检查表1冲突（统一信用代码+年份）
func (a *App) checkTable1Conflicts(targetDb *db.Database, sourceDbs []*db.Database, sourceDbPaths []string, originalSourcePaths []string) []ConflictData {
	var conflicts []ConflictData

	// 获取目标数据库中的表1数据
	targetQuery := `SELECT credit_code, stat_date, unit_name, province_name, city_name, country_name 
					FROM enterprise_coal_consumption_main`
	targetResult, err := targetDb.Query(targetQuery)
	if err != nil {
		return conflicts
	}

	targetData := make(map[string]map[string]interface{})
	if targetResult.Ok && targetResult.Data != nil {
		if data, ok := targetResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
				targetData[key] = row
			}
		}
	}

	// 按冲突键分组收集冲突数据
	conflictGroups := make(map[string]map[string]interface{})

	// 检查每个源数据库
	for i, sourceDb := range sourceDbs {
		sourceQuery := `SELECT credit_code, stat_date, unit_name, province_name, city_name, country_name 
						FROM enterprise_coal_consumption_main`
		sourceResult, err := sourceDb.Query(sourceQuery)
		if err != nil {
			continue
		}

		if sourceResult.Ok && sourceResult.Data != nil {
			if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
					if targetRow, exists := targetData[key]; exists {
						// 如果这个冲突键还没有记录，创建新的冲突组
						if _, exists := conflictGroups[key]; !exists {
							conflictGroups[key] = map[string]interface{}{
								"unit_name":   targetRow["unit_name"],
								"credit_code": targetRow["credit_code"],
								"stat_date":   targetRow["stat_date"],
							}
						}

						// 直接添加文件名和路径作为键值对
						sourceFileName := filepath.Base(originalSourcePaths[i])
						conflictGroups[key][sourceFileName] = originalSourcePaths[i]
					}
				}
			}
		}
	}

	// 将分组数据转换为冲突数据格式
	for _, conflictGroup := range conflictGroups {
		conflicts = append(conflicts, ConflictData{
			TableName:  "规上企业煤炭消费信息主表",
			SourceData: []map[string]interface{}{conflictGroup},
		})
	}

	return conflicts
}

// checkTable2Conflicts 检查表2冲突（统一信用代码+年份）
func (a *App) checkTable2Conflicts(targetDb *db.Database, sourceDbs []*db.Database, sourceDbPaths []string, originalSourcePaths []string) []ConflictData {
	var conflicts []ConflictData

	// 获取目标数据库中的表2数据
	targetQuery := `SELECT credit_code, stat_date, unit_name, province_name, city_name, country_name 
					FROM critical_coal_equipment_consumption`
	targetResult, err := targetDb.Query(targetQuery)
	if err != nil {
		return conflicts
	}

	targetData := make(map[string]map[string]interface{})
	if targetResult.Ok && targetResult.Data != nil {
		if data, ok := targetResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
				targetData[key] = row
			}
		}
	}

	// 按冲突键分组收集冲突数据
	conflictGroups := make(map[string]map[string]interface{})

	// 检查每个源数据库
	for i, sourceDb := range sourceDbs {
		sourceQuery := `SELECT credit_code, stat_date, unit_name, province_name, city_name, country_name 
						FROM critical_coal_equipment_consumption`
		sourceResult, err := sourceDb.Query(sourceQuery)
		if err != nil {
			continue
		}

		if sourceResult.Ok && sourceResult.Data != nil {
			if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
					if targetRow, exists := targetData[key]; exists {
						// 如果这个冲突键还没有记录，创建新的冲突组
						if _, exists := conflictGroups[key]; !exists {
							conflictGroups[key] = map[string]interface{}{
								"unit_name":   targetRow["unit_name"],
								"credit_code": targetRow["credit_code"],
								"stat_date":   targetRow["stat_date"],
							}
						}

						// 直接添加文件名和路径作为键值对
						sourceFileName := filepath.Base(originalSourcePaths[i])
						conflictGroups[key][sourceFileName] = originalSourcePaths[i]
					}
				}
			}
		}
	}

	// 将分组数据转换为冲突数据格式
	for _, conflictGroup := range conflictGroups {
		conflicts = append(conflicts, ConflictData{
			TableName:  "重点耗煤装置煤炭消耗信息表",
			SourceData: []map[string]interface{}{conflictGroup},
		})
	}

	return conflicts
}

// checkTable3Conflicts 检查表3冲突（项目代码+审查意见文号）
func (a *App) checkTable3Conflicts(targetDb *db.Database, sourceDbs []*db.Database, sourceDbPaths []string, originalSourcePaths []string) []ConflictData {
	var conflicts []ConflictData

	// 获取目标数据库中的表3数据
	targetQuery := `SELECT project_code, document_number, project_name, construction_unit, province_name, city_name, country_name 
					FROM fixed_assets_investment_project`
	targetResult, err := targetDb.Query(targetQuery)
	if err != nil {
		return conflicts
	}

	targetData := make(map[string]map[string]interface{})
	if targetResult.Ok && targetResult.Data != nil {
		if data, ok := targetResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				key := fmt.Sprintf("%v_%v", row["project_code"], row["document_number"])
				targetData[key] = row
			}
		}
	}

	// 按冲突键分组收集冲突数据
	conflictGroups := make(map[string]map[string]interface{})

	// 检查每个源数据库
	for i, sourceDb := range sourceDbs {
		sourceQuery := `SELECT project_code, document_number, project_name, construction_unit, province_name, city_name, country_name 
						FROM fixed_assets_investment_project`
		sourceResult, err := sourceDb.Query(sourceQuery)
		if err != nil {
			continue
		}

		if sourceResult.Ok && sourceResult.Data != nil {
			if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["project_code"], row["document_number"])
					if targetRow, exists := targetData[key]; exists {
						// 如果这个冲突键还没有记录，创建新的冲突组
						if _, exists := conflictGroups[key]; !exists {
							conflictGroups[key] = map[string]interface{}{
								"project_name":    targetRow["project_name"],
								"project_code":    targetRow["project_code"],
								"document_number": targetRow["document_number"],
							}
						}

						// 直接添加文件名和路径作为键值对
						sourceFileName := filepath.Base(originalSourcePaths[i])
						conflictGroups[key][sourceFileName] = originalSourcePaths[i]
					}
				}
			}
		}
	}

	// 将分组数据转换为冲突数据格式
	for _, conflictGroup := range conflictGroups {
		conflicts = append(conflicts, ConflictData{
			TableName:  "固定资产投资项目节能审查煤炭消费情况汇总表",
			SourceData: []map[string]interface{}{conflictGroup},
		})
	}

	return conflicts
}

// checkAttachment2Conflicts 检查附件2冲突（省+市+县+年份）
func (a *App) checkAttachment2Conflicts(targetDb *db.Database, sourceDbs []*db.Database, sourceDbPaths []string, originalSourcePaths []string) []ConflictData {
	var conflicts []ConflictData

	// 获取目标数据库中的附件2数据
	targetQuery := `SELECT province_name, city_name, country_name, stat_date, unit_name, unit_level 
					FROM coal_consumption_report`
	targetResult, err := targetDb.Query(targetQuery)
	if err != nil {
		return conflicts
	}

	targetData := make(map[string]map[string]interface{})
	if targetResult.Ok && targetResult.Data != nil {
		if data, ok := targetResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				key := fmt.Sprintf("%v_%v_%v_%v", row["province_name"], row["city_name"], row["country_name"], row["stat_date"])
				targetData[key] = row
			}
		}
	}

	// 按冲突键分组收集冲突数据
	conflictGroups := make(map[string]map[string]interface{})

	// 检查每个源数据库
	for i, sourceDb := range sourceDbs {
		sourceQuery := `SELECT province_name, city_name, country_name, stat_date, unit_name, unit_level 
						FROM coal_consumption_report`
		sourceResult, err := sourceDb.Query(sourceQuery)
		if err != nil {
			continue
		}

		if sourceResult.Ok && sourceResult.Data != nil {
			if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v_%v_%v", row["province_name"], row["city_name"], row["country_name"], row["stat_date"])
					if targetRow, exists := targetData[key]; exists {
						// 如果这个冲突键还没有记录，创建新的冲突组
						if _, exists := conflictGroups[key]; !exists {
							conflictGroups[key] = map[string]interface{}{
								"province_name": targetRow["province_name"],
								"city_name":     targetRow["city_name"],
								"country_name":  targetRow["country_name"],
								"stat_date":     targetRow["stat_date"],
							}
						}

						// 直接添加文件名和路径作为键值对
						sourceFileName := filepath.Base(originalSourcePaths[i])
						conflictGroups[key][sourceFileName] = originalSourcePaths[i]
					}
				}
			}
		}
	}

	// 将分组数据转换为冲突数据格式
	for _, conflictGroup := range conflictGroups {
		conflicts = append(conflicts, ConflictData{
			TableName:  "煤炭消费状况表",
			SourceData: []map[string]interface{}{conflictGroup},
		})
	}

	return conflicts
}

// mergeTable1Data 合并表1数据
func (a *App) mergeTable1Data(targetDb *db.Database, sourceDb *db.Database) MergeResult {
	result := MergeResult{}

	query := `SELECT * FROM enterprise_coal_consumption_main`
	sourceResult, err := sourceDb.Query(query)
	if err != nil {
		result.ErrorCount++
		return result
	}

	if sourceResult.Ok && sourceResult.Data != nil {
		if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				// 生成新的obj_id
				row["obj_id"] = a.generateUUID()
				row["create_time"] = time.Now().Format("2006-01-02 15:04:05")
				row["create_user"] = a.GetCurrentOSUser()

				// 插入数据
				insertQuery := `INSERT INTO enterprise_coal_consumption_main (
					obj_id, unit_name, stat_date, sg_code, tel, credit_code, create_time,
					trade_a, trade_b, trade_c, province_code, province_name, city_code, city_name,
					country_code, country_name, annual_energy_equivalent_value, annual_energy_equivalent_cost,
					annual_raw_material_energy, annual_total_coal_consumption, annual_total_coal_products,
					annual_raw_coal, annual_raw_coal_consumption, annual_clean_coal_consumption,
					annual_other_coal_consumption, annual_coke_consumption, create_user, is_confirm, is_check
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

				_, err := targetDb.Exec(insertQuery,
					row["obj_id"], row["unit_name"], row["stat_date"], row["sg_code"], row["tel"], row["credit_code"], row["create_time"],
					row["trade_a"], row["trade_b"], row["trade_c"], row["province_code"], row["province_name"], row["city_code"], row["city_name"],
					row["country_code"], row["country_name"], row["annual_energy_equivalent_value"], row["annual_energy_equivalent_cost"],
					row["annual_raw_material_energy"], row["annual_total_coal_consumption"], row["annual_total_coal_products"],
					row["annual_raw_coal"], row["annual_raw_coal_consumption"], row["annual_clean_coal_consumption"],
					row["annual_other_coal_consumption"], row["annual_coke_consumption"], row["create_user"], row["is_confirm"], row["is_check"])

				if err != nil {
					result.ErrorCount++
				} else {
					result.SuccessCount++
				}
			}
		}
	}

	result.Ok = true
	return result
}

// mergeTable2Data 合并表2数据
func (a *App) mergeTable2Data(targetDb *db.Database, sourceDb *db.Database) MergeResult {
	result := MergeResult{}

	query := `SELECT * FROM critical_coal_equipment_consumption`
	sourceResult, err := sourceDb.Query(query)
	if err != nil {
		result.ErrorCount++
		return result
	}

	if sourceResult.Ok && sourceResult.Data != nil {
		if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				// 生成新的obj_id
				row["obj_id"] = a.generateUUID()
				row["create_time"] = time.Now().Format("2006-01-02 15:04:05")
				row["create_user"] = a.GetCurrentOSUser()

				// 插入数据
				insertQuery := `INSERT INTO critical_coal_equipment_consumption (
					obj_id, stat_date, create_time, sg_code, unit_name, credit_code, trade_a, trade_b, trade_c, trade_d,
					province_code, province_name, city_code, city_name, country_code, country_name, unit_addr,
					coal_type, coal_no, usage_time, design_life, enecrgy_efficienct_bmk, capacity_unit, capacity,
					use_info, status, annual_coal_consumption, row_no, create_user, is_confirm, is_check
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

				_, err := targetDb.Exec(insertQuery,
					row["obj_id"], row["stat_date"], row["create_time"], row["sg_code"], row["unit_name"], row["credit_code"],
					row["trade_a"], row["trade_b"], row["trade_c"], row["trade_d"], row["province_code"], row["province_name"],
					row["city_code"], row["city_name"], row["country_code"], row["country_name"], row["unit_addr"],
					row["coal_type"], row["coal_no"], row["usage_time"], row["design_life"], row["enecrgy_efficienct_bmk"],
					row["capacity_unit"], row["capacity"], row["use_info"], row["status"], row["annual_coal_consumption"],
					row["row_no"], row["create_user"], row["is_confirm"], row["is_check"])

				if err != nil {
					result.ErrorCount++
				} else {
					result.SuccessCount++
				}
			}
		}
	}

	result.Ok = true
	return result
}

// mergeTable3Data 合并表3数据
func (a *App) mergeTable3Data(targetDb *db.Database, sourceDb *db.Database) MergeResult {
	result := MergeResult{}

	query := `SELECT * FROM fixed_assets_investment_project`
	sourceResult, err := sourceDb.Query(query)
	if err != nil {
		result.ErrorCount++
		return result
	}

	if sourceResult.Ok && sourceResult.Data != nil {
		if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				// 生成新的obj_id
				row["obj_id"] = a.generateUUID()
				row["create_time"] = time.Now().Format("2006-01-02 15:04:05")
				row["create_user"] = a.GetCurrentOSUser()

				// 插入数据
				insertQuery := `INSERT INTO fixed_assets_investment_project (
					obj_id, stat_date, sg_code, project_name, project_code, construction_unit, main_construction_content,
					unit_id, province_name, city_name, country_name, trade_a, trade_c, examination_approval_time,
					scheduled_time, actual_time, examination_authority, document_number, equivalent_value, equivalent_cost,
					pq_total_coal_consumption, pq_coal_consumption, pq_coke_consumption, pq_blue_coke_consumption,
					sce_total_coal_consumption, sce_coal_consumption, sce_coke_consumption, sce_blue_coke_consumption,
					is_substitution, substitution_source, substitution_quantity, pq_annual_coal_quantity, sce_annual_coal_quantity,
					create_time, create_user, is_confirm, is_check
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

				_, err := targetDb.Exec(insertQuery,
					row["obj_id"], row["stat_date"], row["sg_code"], row["project_name"], row["project_code"], row["construction_unit"], row["main_construction_content"],
					row["unit_id"], row["province_name"], row["city_name"], row["country_name"], row["trade_a"], row["trade_c"], row["examination_approval_time"],
					row["scheduled_time"], row["actual_time"], row["examination_authority"], row["document_number"], row["equivalent_value"], row["equivalent_cost"],
					row["pq_total_coal_consumption"], row["pq_coal_consumption"], row["pq_coke_consumption"], row["pq_blue_coke_consumption"],
					row["sce_total_coal_consumption"], row["sce_coal_consumption"], row["sce_coke_consumption"], row["sce_blue_coke_consumption"],
					row["is_substitution"], row["substitution_source"], row["substitution_quantity"], row["pq_annual_coal_quantity"], row["sce_annual_coal_quantity"],
					row["create_time"], row["create_user"], row["is_confirm"], row["is_check"])

				if err != nil {
					result.ErrorCount++
				} else {
					result.SuccessCount++
				}
			}
		}
	}

	result.Ok = true
	return result
}

// mergeAttachment2Data 合并附件2数据
func (a *App) mergeAttachment2Data(targetDb *db.Database, sourceDb *db.Database) MergeResult {
	result := MergeResult{}

	query := `SELECT * FROM coal_consumption_report`
	sourceResult, err := sourceDb.Query(query)
	if err != nil {
		result.ErrorCount++
		return result
	}

	if sourceResult.Ok && sourceResult.Data != nil {
		if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				// 生成新的obj_id
				row["obj_id"] = a.generateUUID()
				row["create_time"] = time.Now().Format("2006-01-02 15:04:05")
				row["create_user"] = a.GetCurrentOSUser()

				// 插入数据
				insertQuery := `INSERT INTO coal_consumption_report (
					obj_id, stat_date, sg_code, unit_id, unit_name, unit_level, province_name, city_name, country_name,
					total_coal, raw_coal, washed_coal, other_coal, power_generation, heating, coal_washing, coking,
					oil_refining, gas_production, industry, raw_materials, other_uses, coke, create_user, create_time, is_confirm, is_check
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

				_, err := targetDb.Exec(insertQuery,
					row["obj_id"], row["stat_date"], row["sg_code"], row["unit_id"], row["unit_name"], row["unit_level"],
					row["province_name"], row["city_name"], row["country_name"], row["total_coal"], row["raw_coal"],
					row["washed_coal"], row["other_coal"], row["power_generation"], row["heating"], row["coal_washing"],
					row["coking"], row["oil_refining"], row["gas_production"], row["industry"], row["raw_materials"],
					row["other_uses"], row["coke"], row["create_user"], row["create_time"], row["is_confirm"], row["is_check"])

				if err != nil {
					result.ErrorCount++
				} else {
					result.SuccessCount++
				}
			}
		}
	}

	result.Ok = true
	return result
}

// generateUUID 生成UUID
func (a *App) generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// mergeTable1DataWithTx 使用事务合并表1数据
func (a *App) mergeTable1DataWithTx(tx *sql.Tx, sourceDb *db.Database) MergeResult {
	result := MergeResult{}

	query := `SELECT * FROM enterprise_coal_consumption_main`
	sourceResult, err := sourceDb.Query(query)
	if err != nil {
		result.ErrorCount++
		result.Message = "查询源数据失败: " + err.Error()
		return result
	}

	if sourceResult.Ok && sourceResult.Data != nil {
		if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				// 生成新的obj_id
				row["obj_id"] = a.generateUUID()
				row["create_time"] = time.Now().Format("2006-01-02 15:04:05")
				row["create_user"] = a.GetCurrentOSUser()

				// 插入数据
				insertQuery := `INSERT INTO enterprise_coal_consumption_main (
					obj_id, unit_name, stat_date, sg_code, tel, credit_code, create_time,
					trade_a, trade_b, trade_c, province_code, province_name, city_code, city_name,
					country_code, country_name, annual_energy_equivalent_value, annual_energy_equivalent_cost,
					annual_raw_material_energy, annual_total_coal_consumption, annual_total_coal_products,
					annual_raw_coal, annual_raw_coal_consumption, annual_clean_coal_consumption,
					annual_other_coal_consumption, annual_coke_consumption, create_user, is_confirm, is_check
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

				_, err := tx.Exec(insertQuery,
					row["obj_id"], row["unit_name"], row["stat_date"], row["sg_code"], row["tel"], row["credit_code"], row["create_time"],
					row["trade_a"], row["trade_b"], row["trade_c"], row["province_code"], row["province_name"], row["city_code"], row["city_name"],
					row["country_code"], row["country_name"], row["annual_energy_equivalent_value"], row["annual_energy_equivalent_cost"],
					row["annual_raw_material_energy"], row["annual_total_coal_consumption"], row["annual_total_coal_products"],
					row["annual_raw_coal"], row["annual_raw_coal_consumption"], row["annual_clean_coal_consumption"],
					row["annual_other_coal_consumption"], row["annual_coke_consumption"], row["create_user"], row["is_confirm"], row["is_check"])

				if err != nil {
					result.ErrorCount++
					result.Message = "插入数据失败: " + err.Error()
				} else {
					result.SuccessCount++
				}
			}
		}
	}

	result.Ok = true
	return result
}

// mergeTable2DataWithTx 使用事务合并表2数据
func (a *App) mergeTable2DataWithTx(tx *sql.Tx, sourceDb *db.Database) MergeResult {
	result := MergeResult{}

	query := `SELECT * FROM critical_coal_equipment_consumption`
	sourceResult, err := sourceDb.Query(query)
	if err != nil {
		result.ErrorCount++
		result.Message = "查询源数据失败: " + err.Error()
		return result
	}

	if sourceResult.Ok && sourceResult.Data != nil {
		if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				// 生成新的obj_id
				row["obj_id"] = a.generateUUID()
				row["create_time"] = time.Now().Format("2006-01-02 15:04:05")
				row["create_user"] = a.GetCurrentOSUser()

				// 插入数据
				insertQuery := `INSERT INTO critical_coal_equipment_consumption (
					obj_id, stat_date, create_time, sg_code, unit_name, credit_code, trade_a, trade_b, trade_c, trade_d,
					province_code, province_name, city_code, city_name, country_code, country_name, unit_addr,
					coal_type, coal_no, usage_time, design_life, enecrgy_efficienct_bmk, capacity_unit, capacity,
					use_info, status, annual_coal_consumption, row_no, create_user, is_confirm, is_check
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

				_, err := tx.Exec(insertQuery,
					row["obj_id"], row["stat_date"], row["create_time"], row["sg_code"], row["unit_name"], row["credit_code"],
					row["trade_a"], row["trade_b"], row["trade_c"], row["trade_d"], row["province_code"], row["province_name"],
					row["city_code"], row["city_name"], row["country_code"], row["country_name"], row["unit_addr"],
					row["coal_type"], row["coal_no"], row["usage_time"], row["design_life"], row["enecrgy_efficienct_bmk"],
					row["capacity_unit"], row["capacity"], row["use_info"], row["status"], row["annual_coal_consumption"],
					row["row_no"], row["create_user"], row["is_confirm"], row["is_check"])

				if err != nil {
					result.ErrorCount++
					result.Message = "插入数据失败: " + err.Error()
				} else {
					result.SuccessCount++
				}
			}
		}
	}

	result.Ok = true
	return result
}

// mergeTable3DataWithTx 使用事务合并表3数据
func (a *App) mergeTable3DataWithTx(tx *sql.Tx, sourceDb *db.Database) MergeResult {
	result := MergeResult{}

	query := `SELECT * FROM fixed_assets_investment_project`
	sourceResult, err := sourceDb.Query(query)
	if err != nil {
		result.ErrorCount++
		result.Message = "查询源数据失败: " + err.Error()
		return result
	}

	if sourceResult.Ok && sourceResult.Data != nil {
		if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				// 生成新的obj_id
				row["obj_id"] = a.generateUUID()
				row["create_time"] = time.Now().Format("2006-01-02 15:04:05")
				row["create_user"] = a.GetCurrentOSUser()

				// 插入数据
				insertQuery := `INSERT INTO fixed_assets_investment_project (
					obj_id, stat_date, sg_code, project_name, project_code, construction_unit, main_construction_content,
					unit_id, province_name, city_name, country_name, trade_a, trade_c, examination_approval_time,
					scheduled_time, actual_time, examination_authority, document_number, equivalent_value, equivalent_cost,
					pq_total_coal_consumption, pq_coal_consumption, pq_coke_consumption, pq_blue_coke_consumption,
					sce_total_coal_consumption, sce_coal_consumption, sce_coke_consumption, sce_blue_coke_consumption,
					is_substitution, substitution_source, substitution_quantity, pq_annual_coal_quantity, sce_annual_coal_quantity,
					create_time, create_user, is_confirm, is_check
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

				_, err := tx.Exec(insertQuery,
					row["obj_id"], row["stat_date"], row["sg_code"], row["project_name"], row["project_code"], row["construction_unit"], row["main_construction_content"],
					row["unit_id"], row["province_name"], row["city_name"], row["country_name"], row["trade_a"], row["trade_c"], row["examination_approval_time"],
					row["scheduled_time"], row["actual_time"], row["examination_authority"], row["document_number"], row["equivalent_value"], row["equivalent_cost"],
					row["pq_total_coal_consumption"], row["pq_coal_consumption"], row["pq_coke_consumption"], row["pq_blue_coke_consumption"],
					row["sce_total_coal_consumption"], row["sce_coal_consumption"], row["sce_coke_consumption"], row["sce_blue_coke_consumption"],
					row["is_substitution"], row["substitution_source"], row["substitution_quantity"], row["pq_annual_coal_quantity"], row["sce_annual_coal_quantity"],
					row["create_time"], row["create_user"], row["is_confirm"], row["is_check"])

				if err != nil {
					result.ErrorCount++
					result.Message = "插入数据失败: " + err.Error()
				} else {
					result.SuccessCount++
				}
			}
		}
	}

	result.Ok = true
	return result
}

// mergeAttachment2DataWithTx 使用事务合并附件2数据
func (a *App) mergeAttachment2DataWithTx(tx *sql.Tx, sourceDb *db.Database) MergeResult {
	result := MergeResult{}

	query := `SELECT * FROM coal_consumption_report`
	sourceResult, err := sourceDb.Query(query)
	if err != nil {
		result.ErrorCount++
		result.Message = "查询源数据失败: " + err.Error()
		return result
	}

	if sourceResult.Ok && sourceResult.Data != nil {
		if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
			for _, row := range data {
				// 生成新的obj_id
				row["obj_id"] = a.generateUUID()
				row["create_time"] = time.Now().Format("2006-01-02 15:04:05")
				row["create_user"] = a.GetCurrentOSUser()

				// 插入数据
				insertQuery := `INSERT INTO coal_consumption_report (
					obj_id, stat_date, sg_code, unit_id, unit_name, unit_level, province_name, city_name, country_name,
					total_coal, raw_coal, washed_coal, other_coal, power_generation, heating, coal_washing, coking,
					oil_refining, gas_production, industry, raw_materials, other_uses, coke, create_user, create_time, is_confirm, is_check
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

				_, err := tx.Exec(insertQuery,
					row["obj_id"], row["stat_date"], row["sg_code"], row["unit_id"], row["unit_name"], row["unit_level"],
					row["province_name"], row["city_name"], row["country_name"], row["total_coal"], row["raw_coal"],
					row["washed_coal"], row["other_coal"], row["power_generation"], row["heating"], row["coal_washing"],
					row["coking"], row["oil_refining"], row["gas_production"], row["industry"], row["raw_materials"],
					row["other_uses"], row["coke"], row["create_user"], row["create_time"], row["is_confirm"], row["is_check"])

				if err != nil {
					result.ErrorCount++
					result.Message = "插入数据失败: " + err.Error()
				} else {
					result.SuccessCount++
				}
			}
		}
	}

	result.Ok = true
	return result
}
