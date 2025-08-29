package main

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"shuji/db"
	"strings"
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

// ConflictInfo 冲突信息结构
type ConflictInfo struct {
	HasConflict          bool              `json:"hasConflict"`
	TotalConflictCount   int               `json:"totalConflictCount"`
	Table1Conflicts      TableConflictInfo `json:"table1Conflicts"`
	Table2Conflicts      TableConflictInfo `json:"table2Conflicts"`
	Table3Conflicts      TableConflictInfo `json:"table3Conflicts"`
	Attachment2Conflicts TableConflictInfo `json:"attachment2Conflicts"`
}

// TableConflictInfo 表冲突信息
type TableConflictInfo struct {
	HasConflict   bool             `json:"hasConflict"`
	FileNames     []string         `json:"fileNames"`
	ConflictCount int              `json:"conflictCount"`
	Conflicts     []ConflictDetail `json:"conflicts"`
}

// ConflictDetail 冲突详情
type ConflictDetail struct {
	ObjId          string               `json:"obj_id,omitempty"`
	CreditCode     string               `json:"credit_code,omitempty"`
	StatDate       string               `json:"stat_date,omitempty"`
	UnitName       string               `json:"unit_name,omitempty"`
	ProjectName    string               `json:"project_name,omitempty"`
	ProjectCode    string               `json:"project_code,omitempty"`
	DocumentNumber string               `json:"document_number,omitempty"`
	ProvinceName   string               `json:"province_name,omitempty"`
	CityName       string               `json:"city_name,omitempty"`
	CountryName    string               `json:"country_name,omitempty"`
	Conflict       []ConflictSourceInfo `json:"conflict"`
}

// ConflictSourceInfo 冲突源信息
type ConflictSourceInfo struct {
	FilePath  string   `json:"filePath"`
	FileName  string   `json:"fileName"`
	TableType string   `json:"tableType"`
	ObjIds    []string `json:"obj_ids"`
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
	fmt.Println("newDb dbTempPath", dbTempPath)
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

	for _, sourceDbPath := range sourceDbPath {
		fileName := filepath.Base(sourceDbPath)
		dbDstPath := GetPath(filepath.Join(DATA_DIR_NAME, time.Now().Format("20060102150405")+"_"+fileName))
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

		sourceDbs = append(sourceDbs, sourceDb)
		sourceDbPaths = append(sourceDbPaths, dbDstPath)
		originalSourcePaths = append(originalSourcePaths, sourceDbPath)
	}

	// 如果没有成功打开任何数据库文件，则返回错误
	if len(sourceDbs) == 0 {
		result.Ok = false
		result.Message = fmt.Sprintf("所有数据库文件打开失败: %v", failedFiles)
		// 删除所有数据库文件
		a.Removefile(dbTempPath)
		for i := range sourceDbPaths {
			a.Removefile(sourceDbPaths[i])
		}
		return result
	}

	// 4. 检查数据冲突
	// 检查表1冲突（规上企业煤炭消费信息主表）
	table1Conflicts, table1NonConflictData := a.checkTable1Conflicts(newDb, sourceDbs, sourceDbPaths, originalSourcePaths)
	table1ConflictCount := len(table1Conflicts.Conflicts)

	// 检查表2冲突（重点耗煤装置煤炭消耗信息表）
	table2Conflicts, table2NonConflictData := a.checkTable2Conflicts(newDb, sourceDbs, sourceDbPaths, originalSourcePaths)
	table2ConflictCount := len(table2Conflicts.Conflicts)

	// 检查表3冲突（固定资产投资项目节能审查煤炭消费情况汇总表）
	table3Conflicts, table3NonConflictData := a.checkTable3Conflicts(newDb, sourceDbs, sourceDbPaths, originalSourcePaths)
	table3ConflictCount := len(table3Conflicts.Conflicts)

	// 检查附件2冲突（煤炭消费状况表）
	attachment2Conflicts, attachment2NonConflictData := a.checkAttachment2Conflicts(newDb, sourceDbs, sourceDbPaths, originalSourcePaths)
	attachment2ConflictCount := len(attachment2Conflicts.Conflicts)

	totalConflictCount := table1ConflictCount + table2ConflictCount + table3ConflictCount + attachment2ConflictCount

	// 5. 记录冲突信息（如果有冲突）
	var conflictInfo map[string]interface{}
	if totalConflictCount > 0 {
		conflictInfo = map[string]interface{}{
			"hasConflict": true,
			"table1Conflicts": map[string]interface{}{
				"conflicts":     table1Conflicts.Conflicts,
				"conflictCount": table1ConflictCount,
				"fileNames":     table1Conflicts.FileNames,
			},
			"table2Conflicts": map[string]interface{}{
				"conflicts":     table2Conflicts.Conflicts,
				"conflictCount": table2ConflictCount,
				"fileNames":     table2Conflicts.FileNames,
			},
			"table3Conflicts": map[string]interface{}{
				"conflicts":     table3Conflicts.Conflicts,
				"conflictCount": table3ConflictCount,
				"fileNames":     table3Conflicts.FileNames,
			},
			"attachment2Conflicts": map[string]interface{}{
				"conflicts":     attachment2Conflicts.Conflicts,
				"conflictCount": attachment2ConflictCount,
				"fileNames":     attachment2Conflicts.FileNames,
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
		// 先关闭所有源数据库连接
		for i, sourceDb := range sourceDbs {
			if sourceDb != nil {
				sourceDb.Close()
				fmt.Printf("事务失败关闭源数据库连接: %s\n", sourceDbPaths[i])
			}
		}
		// 删除所有数据库文件
		a.Removefile(dbTempPath)
		for i := range sourceDbPaths {
			removeResult := a.Removefile(sourceDbPaths[i])
			fmt.Printf("删除源临时文件 %s: %v\n", sourceDbPaths[i], removeResult)
		}
		return result
	}

	// 确保事务回滚（如果出错）
	var txErr error
	defer func() {
		if txErr != nil {
			tx.Rollback()
			// 先关闭所有源数据库连接
			for i, sourceDb := range sourceDbs {
				if sourceDb != nil {
					sourceDb.Close()
					fmt.Printf("事务回滚关闭源数据库连接: %s\n", sourceDbPaths[i])
				}
			}
			// 事务回滚时删除所有临时文件（因为事务失败，没有成功导入数据）
			a.Removefile(dbTempPath)
			for i := range sourceDbPaths {
				removeResult := a.Removefile(sourceDbPaths[i])
				fmt.Printf("事务回滚删除源临时文件 %s: %v\n", sourceDbPaths[i], removeResult)
			}
		}
	}()

	// 统计实际冲突数量
	actualConflictCount := 0

	// 合并所有表的数据（使用已获取的非冲突数据）
	fmt.Printf("开始合并表1数据，非冲突数据数量: %d\n", len(table1NonConflictData))
	table1Result := a.mergeTable1DataWithTx(tx, table1NonConflictData)
	fmt.Printf("表1合并结果: Ok=%v, SuccessCount=%d, ErrorCount=%d, Message=%s\n",
		table1Result.Ok, table1Result.SuccessCount, table1Result.ErrorCount, table1Result.Message)
	if table1Result.Ok {
		successCount += table1Result.SuccessCount
		actualConflictCount += table1Result.ConflictCount
	} else {
		errorCount += table1Result.ErrorCount
	}

	table2Result := a.mergeTable2DataWithTx(tx, table2NonConflictData)
	if table2Result.Ok {
		successCount += table2Result.SuccessCount
		actualConflictCount += table2Result.ConflictCount
	} else {
		errorCount += table2Result.ErrorCount
	}

	table3Result := a.mergeTable3DataWithTx(tx, table3NonConflictData)
	if table3Result.Ok {
		successCount += table3Result.SuccessCount
		actualConflictCount += table3Result.ConflictCount
	} else {
		errorCount += table3Result.ErrorCount
	}

	attachment2Result := a.mergeAttachment2DataWithTx(tx, attachment2NonConflictData)
	if attachment2Result.Ok {
		successCount += attachment2Result.SuccessCount
		actualConflictCount += attachment2Result.ConflictCount
	} else {
		errorCount += attachment2Result.ErrorCount
	}

	// 提交事务
	txErr = tx.Commit()
	if txErr != nil {
		result.Ok = false
		result.Message = "提交事务失败: " + txErr.Error()
		// 先关闭所有源数据库连接
		for i, sourceDb := range sourceDbs {
			if sourceDb != nil {
				sourceDb.Close()
				fmt.Printf("事务提交失败关闭源数据库连接: %s\n", sourceDbPaths[i])
			}
		}
		// 事务提交失败时删除所有临时文件
		a.Removefile(dbTempPath)
		for i := range sourceDbPaths {
			removeResult := a.Removefile(sourceDbPaths[i])
			fmt.Printf("事务提交失败删除源临时文件 %s: %v\n", sourceDbPaths[i], removeResult)
		}
		return result
	}

	// 先关闭所有源数据库连接，然后删除源临时数据库文件
	for i, sourceDb := range sourceDbs {
		if sourceDb != nil {
			sourceDb.Close()
			fmt.Printf("关闭源数据库连接: %s\n", sourceDbPaths[i])
		}
	}

	// 删除源临时数据库文件
	for i := range sourceDbPaths {
		removeResult := a.Removefile(sourceDbPaths[i])
		fmt.Printf("删除源临时文件 %s: %v\n", sourceDbPaths[i], removeResult)
	}

	// 只有在完全没有成功导入数据时才删除目标数据库临时文件
	// 即使存在冲突也不删除，因为用户可能需要通过冲突解决界面确认是否覆盖
	if successCount == 0 && totalConflictCount == 0 {
		a.Removefile(dbTempPath)
	}

	// 8. 返回合并结果
	result.Ok = true
	result.Message = fmt.Sprintf("数据库合并完成。成功合并: %d 条数据，错误: %d 条，冲突: %d 条", successCount, errorCount, totalConflictCount)
	result.Data = map[string]interface{}{
		"successCount":       successCount,
		"errorCount":         errorCount,
		"totalConflictCount": totalConflictCount,
		"failedFiles":        failedFiles,
		"targetDbPath":       dbTempPath,
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
func (a *App) checkTable1Conflicts(targetDb *db.Database, sourceDbs []*db.Database, sourceDbPaths []string, originalSourcePaths []string) (TableConflictInfo, []map[string]interface{}) {
	conflictInfo := TableConflictInfo{}
	var nonConflictData []map[string]interface{}

	// 1. 获取目标数据库中的表1数据，构建映射
	targetQuery := `SELECT obj_id, credit_code, stat_date, unit_name 
					FROM enterprise_coal_consumption_main`
	targetResult, err := targetDb.Query(targetQuery)
	if err != nil {
		return conflictInfo, nonConflictData
	}

	// 构建目标数据映射，用于快速查找
	targetDataMap := make(map[string]map[string]interface{})
	if targetResult.Ok && targetResult.Data != nil {
		if data, ok := targetResult.Data.([]map[string]interface{}); ok {
			for _, targetRow := range data {
				key := fmt.Sprintf("%v_%v", targetRow["credit_code"], targetRow["stat_date"])
				targetDataMap[key] = targetRow
			}
		}
	}

	fmt.Printf("表1目标数据映射构建完成，共 %d 条记录\n", len(targetDataMap))

	// 2. 遍历每个源数据库，查询源数据并检查冲突
	fileNamesSet := make(map[string]bool)
	conflictMap := make(map[string][]ConflictSourceInfo) // key: conflictKey, value: conflict sources

	for i, sourceDb := range sourceDbs {
		sourceQuery := `SELECT * FROM enterprise_coal_consumption_main`
		sourceResult, err := sourceDb.Query(sourceQuery)
		if err != nil {
			continue
		}

		if sourceResult.Ok && sourceResult.Data != nil {
			if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
				filePath := originalSourcePaths[i]
				sourceFileName := filepath.Base(filePath)
				fileNamesSet[sourceFileName] = true

				fmt.Printf("处理源文件 %s，共 %d 条记录\n", sourceFileName, len(data))

				// 遍历源数据，检查冲突
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
					objId := fmt.Sprintf("%v", row["obj_id"])

					// 检查是否与目标数据冲突
					if targetRow, exists := targetDataMap[key]; exists {
						// 发现冲突
						fmt.Printf("发现冲突 - key: %s, 目标: %v, 源: %v\n", key, targetRow["unit_name"], row["unit_name"])

						// 添加到冲突映射
						conflictMap[key] = append(conflictMap[key], ConflictSourceInfo{
							FilePath:  filePath,
							FileName:  sourceFileName,
							TableType: "table1",
							ObjIds:    []string{objId},
						})
					} else {
						// 没有冲突，添加到待插入列表
						fmt.Printf("无冲突数据 - key: %s, 数据: %v\n", key, row["unit_name"])
						nonConflictData = append(nonConflictData, row)
					}
				}
			}
		}
	}

	// 3. 构建冲突详情
	for key, conflictSources := range conflictMap {
		if targetRow, exists := targetDataMap[key]; exists {
			conflictDetail := ConflictDetail{
				ObjId:      fmt.Sprintf("%v", targetRow["obj_id"]),
				CreditCode: fmt.Sprintf("%v", targetRow["credit_code"]),
				StatDate:   fmt.Sprintf("%v", targetRow["stat_date"]),
				UnitName:   fmt.Sprintf("%v", targetRow["unit_name"]),
				Conflict:   conflictSources,
			}
			conflictInfo.Conflicts = append(conflictInfo.Conflicts, conflictDetail)
		}
	}

	fmt.Printf("表1冲突检查完成 - 冲突数: %d, 非冲突数据数: %d\n", len(conflictInfo.Conflicts), len(nonConflictData))
	if len(nonConflictData) > 0 {
		fmt.Printf("表1非冲突数据示例: obj_id=%v, unit_name=%v, stat_date=%v\n",
			nonConflictData[0]["obj_id"], nonConflictData[0]["unit_name"], nonConflictData[0]["stat_date"])
	}

	// 设置文件名列表
	conflictInfo.FileNames = make([]string, 0, len(fileNamesSet))
	for fileName := range fileNamesSet {
		conflictInfo.FileNames = append(conflictInfo.FileNames, fileName)
	}

	conflictInfo.ConflictCount = len(conflictInfo.Conflicts)
	conflictInfo.HasConflict = conflictInfo.ConflictCount > 0

	return conflictInfo, nonConflictData
}

// checkTable2Conflicts 检查表2冲突（统一信用代码+年份）
func (a *App) checkTable2Conflicts(targetDb *db.Database, sourceDbs []*db.Database, sourceDbPaths []string, originalSourcePaths []string) (TableConflictInfo, []map[string]interface{}) {
	conflictInfo := TableConflictInfo{}
	var nonConflictData []map[string]interface{}

	// 1. 获取目标数据库中的表2数据，构建映射
	targetQuery := `SELECT obj_id, credit_code, stat_date, unit_name 
					FROM critical_coal_equipment_consumption`
	targetResult, err := targetDb.Query(targetQuery)
	if err != nil {
		return conflictInfo, nonConflictData
	}

	// 构建目标数据映射，用于快速查找
	targetDataMap := make(map[string]map[string]interface{})
	if targetResult.Ok && targetResult.Data != nil {
		if data, ok := targetResult.Data.([]map[string]interface{}); ok {
			for _, targetRow := range data {
				key := fmt.Sprintf("%v_%v", targetRow["credit_code"], targetRow["stat_date"])
				targetDataMap[key] = targetRow
			}
		}
	}

	fmt.Printf("表2目标数据映射构建完成，共 %d 条记录\n", len(targetDataMap))

	// 2. 遍历每个源数据库，查询源数据并检查冲突
	fileNamesSet := make(map[string]bool)
	conflictMap := make(map[string][]ConflictSourceInfo) // key: conflictKey, value: conflict sources

	for i, sourceDb := range sourceDbs {
		sourceQuery := `SELECT * FROM critical_coal_equipment_consumption`
		sourceResult, err := sourceDb.Query(sourceQuery)
		if err != nil {
			continue
		}

		if sourceResult.Ok && sourceResult.Data != nil {
			if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
				filePath := originalSourcePaths[i]
				sourceFileName := filepath.Base(filePath)
				fileNamesSet[sourceFileName] = true

				fmt.Printf("处理源文件 %s，共 %d 条记录\n", sourceFileName, len(data))

				// 收集当前文件的obj_ids
				var fileObjIds []string

				// 遍历源数据，检查冲突
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
					objId := fmt.Sprintf("%v", row["obj_id"])
					fileObjIds = append(fileObjIds, objId)

					// 检查是否与目标数据冲突
					if targetRow, exists := targetDataMap[key]; exists {
						// 发现冲突
						fmt.Printf("发现冲突 - key: %s, 目标: %v, 源: %v\n", key, targetRow["unit_name"], row["unit_name"])

						// 添加到冲突映射
						conflictMap[key] = append(conflictMap[key], ConflictSourceInfo{
							FilePath:  filePath,
							FileName:  sourceFileName,
							TableType: "table2",
							ObjIds:    []string{objId},
						})
					} else {
						// 没有冲突，添加到待插入列表
						fmt.Printf("无冲突数据 - key: %s, 数据: %v\n", key, row["unit_name"])
						nonConflictData = append(nonConflictData, row)
					}
				}

				// 合并相同冲突键的obj_ids
				for key, conflicts := range conflictMap {
					// 找到当前文件的冲突
					for j, conflict := range conflicts {
						if conflict.FilePath == filePath {
							// 更新obj_ids，包含当前文件的所有冲突记录
							var conflictObjIds []string
							for _, row := range data {
								rowKey := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
								if rowKey == key {
									conflictObjIds = append(conflictObjIds, fmt.Sprintf("%v", row["obj_id"]))
								}
							}
							conflicts[j].ObjIds = conflictObjIds
							break
						}
					}
				}
			}
		}
	}

	// 3. 构建冲突详情
	for key, conflictSources := range conflictMap {
		if targetRow, exists := targetDataMap[key]; exists {
			conflictDetail := ConflictDetail{
				ObjId:      fmt.Sprintf("%v", targetRow["obj_id"]),
				CreditCode: fmt.Sprintf("%v", targetRow["credit_code"]),
				StatDate:   fmt.Sprintf("%v", targetRow["stat_date"]),
				UnitName:   fmt.Sprintf("%v", targetRow["unit_name"]),
				Conflict:   conflictSources,
			}
			conflictInfo.Conflicts = append(conflictInfo.Conflicts, conflictDetail)
		}
	}

	fmt.Printf("表2冲突检查完成 - 冲突数: %d, 非冲突数据数: %d\n", len(conflictInfo.Conflicts), len(nonConflictData))

	// 设置文件名列表
	conflictInfo.FileNames = make([]string, 0, len(fileNamesSet))
	for fileName := range fileNamesSet {
		conflictInfo.FileNames = append(conflictInfo.FileNames, fileName)
	}

	conflictInfo.ConflictCount = len(conflictInfo.Conflicts)
	conflictInfo.HasConflict = conflictInfo.ConflictCount > 0

	return conflictInfo, nonConflictData
}

// checkTable3Conflicts 检查表3冲突（项目代码+审查意见文号）
func (a *App) checkTable3Conflicts(targetDb *db.Database, sourceDbs []*db.Database, sourceDbPaths []string, originalSourcePaths []string) (TableConflictInfo, []map[string]interface{}) {
	conflictInfo := TableConflictInfo{}
	var nonConflictData []map[string]interface{}

	// 获取目标数据库中的表3数据
	targetQuery := `SELECT obj_id, project_code, document_number, project_name, construction_unit, province_name, city_name, country_name 
					FROM fixed_assets_investment_project`
	targetResult, err := targetDb.Query(targetQuery)
	if err != nil {
		return conflictInfo, nonConflictData
	}

	// 构建源数据映射，按冲突键分组
	sourceDataMap := make(map[string][]ConflictSourceInfo)
	sourceDataMapFull := make(map[string][]map[string]interface{}) // 存储完整数据
	fileNamesSet := make(map[string]bool)

	// 存储每个文件的obj_ids映射
	fileObjIdsMap := make(map[string][]string) // key: filePath, value: objIds

	// 遍历每个源数据库，获取所有源数据
	for i, sourceDb := range sourceDbs {
		sourceQuery := `SELECT * FROM fixed_assets_investment_project`
		sourceResult, err := sourceDb.Query(sourceQuery)
		if err != nil {
			continue
		}

		if sourceResult.Ok && sourceResult.Data != nil {
			if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
				filePath := originalSourcePaths[i]
				sourceFileName := filepath.Base(filePath)
				fileNamesSet[sourceFileName] = true

				// 获取该文件中所有记录的obj_ids
				var fileObjIds []string

				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["project_code"], row["document_number"])

					// 获取obj_id
					fileObjIds = append(fileObjIds, fmt.Sprintf("%v", row["obj_id"]))

					// 同时存储完整数据
					sourceDataMapFull[key] = append(sourceDataMapFull[key], row)
				}

				// 存储该文件的所有obj_ids
				fileObjIdsMap[filePath] = fileObjIds
			}
		}
	}

	// 为每个冲突键创建ConflictSourceInfo
	for key, dataList := range sourceDataMapFull {
		// 按文件路径分组obj_ids
		fileObjIdsMapForKey := make(map[string][]string)

		for _, row := range dataList {
			// 找到这个数据来自哪个文件
			for filePath, allObjIds := range fileObjIdsMap {
				objId := fmt.Sprintf("%v", row["obj_id"])
				for _, id := range allObjIds {
					if id == objId {
						fileObjIdsMapForKey[filePath] = append(fileObjIdsMapForKey[filePath], objId)
						break
					}
				}
			}
		}

		// 为每个文件创建ConflictSourceInfo
		for filePath, objIds := range fileObjIdsMapForKey {
			sourceDataMap[key] = append(sourceDataMap[key], ConflictSourceInfo{
				FilePath:  filePath,
				FileName:  filepath.Base(filePath),
				TableType: "table3",
				ObjIds:    objIds,
			})
		}
	}

	// 构建目标数据映射，用于快速查找
	targetDataMap := make(map[string]map[string]interface{})
	if targetResult.Ok && targetResult.Data != nil {
		if data, ok := targetResult.Data.([]map[string]interface{}); ok {
			for _, targetRow := range data {
				key := fmt.Sprintf("%v_%v", targetRow["project_code"], targetRow["document_number"])
				targetDataMap[key] = targetRow
			}
		}
	}

	// 遍历源数据，检查是否有冲突，同时获取非冲突数据
	conflictKeys := make(map[string]bool)
	for key, dataList := range sourceDataMapFull {
		// 检查目标数据中是否有相同键的数据
		if targetRow, exists := targetDataMap[key]; exists {
			// 发现冲突，创建冲突详情
			conflictDetail := ConflictDetail{
				ObjId:          fmt.Sprintf("%v", targetRow["obj_id"]),
				ProjectName:    fmt.Sprintf("%v", targetRow["project_name"]),
				ProjectCode:    fmt.Sprintf("%v", targetRow["project_code"]),
				DocumentNumber: fmt.Sprintf("%v", targetRow["document_number"]),
				Conflict:       sourceDataMap[key],
			}
			conflictInfo.Conflicts = append(conflictInfo.Conflicts, conflictDetail)
			conflictKeys[key] = true
		} else {
			// 没有冲突，这些数据需要插入
			nonConflictData = append(nonConflictData, dataList...)
		}
	}

	// 设置文件名列表
	conflictInfo.FileNames = make([]string, 0, len(fileNamesSet))
	for fileName := range fileNamesSet {
		conflictInfo.FileNames = append(conflictInfo.FileNames, fileName)
	}

	conflictInfo.ConflictCount = len(conflictInfo.Conflicts)
	conflictInfo.HasConflict = conflictInfo.ConflictCount > 0

	fmt.Printf("附件2冲突检查完成 - 冲突数: %d, 非冲突数据数: %d\n", len(conflictInfo.Conflicts), len(nonConflictData))
	if len(nonConflictData) > 0 {
		fmt.Printf("附件2非冲突数据示例: obj_id=%v, unit_name=%v, stat_date=%v\n",
			nonConflictData[0]["obj_id"], nonConflictData[0]["unit_name"], nonConflictData[0]["stat_date"])
	}

	return conflictInfo, nonConflictData
}

// checkAttachment2Conflicts 检查附件2冲突（省+市+县+年份）
func (a *App) checkAttachment2Conflicts(targetDb *db.Database, sourceDbs []*db.Database, sourceDbPaths []string, originalSourcePaths []string) (TableConflictInfo, []map[string]interface{}) {
	conflictInfo := TableConflictInfo{}
	var nonConflictData []map[string]interface{}

	// 获取目标数据库中的附件2数据
	targetQuery := `SELECT obj_id, province_name, city_name, country_name, stat_date, unit_name, unit_level 
					FROM coal_consumption_report`
	targetResult, err := targetDb.Query(targetQuery)
	if err != nil {
		return conflictInfo, nonConflictData
	}

	// 构建源数据映射，按冲突键分组
	sourceDataMap := make(map[string][]ConflictSourceInfo)
	sourceDataMapFull := make(map[string][]map[string]interface{}) // 存储完整数据
	fileNamesSet := make(map[string]bool)

	// 存储每个文件的obj_ids映射
	fileObjIdsMap := make(map[string][]string) // key: filePath, value: objIds

	// 遍历每个源数据库，获取所有源数据
	for i, sourceDb := range sourceDbs {
		sourceQuery := `SELECT * FROM coal_consumption_report`
		sourceResult, err := sourceDb.Query(sourceQuery)
		if err != nil {
			continue
		}

		if sourceResult.Ok && sourceResult.Data != nil {
			if data, ok := sourceResult.Data.([]map[string]interface{}); ok {
				filePath := originalSourcePaths[i]
				sourceFileName := filepath.Base(filePath)
				fileNamesSet[sourceFileName] = true

				// 获取该文件中所有记录的obj_ids
				var fileObjIds []string

				for _, row := range data {
					key := fmt.Sprintf("%v_%v_%v_%v", row["province_name"], row["city_name"], row["country_name"], row["stat_date"])

					// 获取obj_id
					fileObjIds = append(fileObjIds, fmt.Sprintf("%v", row["obj_id"]))

					// 同时存储完整数据
					sourceDataMapFull[key] = append(sourceDataMapFull[key], row)
				}

				// 存储该文件的所有obj_ids
				fileObjIdsMap[filePath] = fileObjIds
			}
		}
	}

	// 为每个冲突键创建ConflictSourceInfo
	for key, dataList := range sourceDataMapFull {
		// 按文件路径分组obj_ids
		fileObjIdsMapForKey := make(map[string][]string)

		for _, row := range dataList {
			// 找到这个数据来自哪个文件
			for filePath, allObjIds := range fileObjIdsMap {
				objId := fmt.Sprintf("%v", row["obj_id"])
				for _, id := range allObjIds {
					if id == objId {
						fileObjIdsMapForKey[filePath] = append(fileObjIdsMapForKey[filePath], objId)
						break
					}
				}
			}
		}

		// 为每个文件创建ConflictSourceInfo
		for filePath, objIds := range fileObjIdsMapForKey {
			sourceDataMap[key] = append(sourceDataMap[key], ConflictSourceInfo{
				FilePath:  filePath,
				FileName:  filepath.Base(filePath),
				TableType: "attachment2",
				ObjIds:    objIds,
			})
		}
	}

	// 构建目标数据映射，用于快速查找
	targetDataMap := make(map[string]map[string]interface{})
	if targetResult.Ok && targetResult.Data != nil {
		if data, ok := targetResult.Data.([]map[string]interface{}); ok {
			for _, targetRow := range data {
				key := fmt.Sprintf("%v_%v_%v_%v", targetRow["province_name"], targetRow["city_name"], targetRow["country_name"], targetRow["stat_date"])
				targetDataMap[key] = targetRow
			}
		}
	}

	// 遍历源数据，检查是否有冲突，同时获取非冲突数据
	conflictKeys := make(map[string]bool)
	for key, dataList := range sourceDataMapFull {
		// 检查目标数据中是否有相同键的数据
		if targetRow, exists := targetDataMap[key]; exists {
			// 发现冲突，创建冲突详情
			conflictDetail := ConflictDetail{
				ObjId:        fmt.Sprintf("%v", targetRow["obj_id"]),
				ProvinceName: fmt.Sprintf("%v", targetRow["province_name"]),
				CityName:     fmt.Sprintf("%v", targetRow["city_name"]),
				CountryName:  fmt.Sprintf("%v", targetRow["country_name"]),
				StatDate:     fmt.Sprintf("%v", targetRow["stat_date"]),
				Conflict:     sourceDataMap[key],
			}
			conflictInfo.Conflicts = append(conflictInfo.Conflicts, conflictDetail)
			conflictKeys[key] = true
		} else {
			// 没有冲突，这些数据需要插入
			nonConflictData = append(nonConflictData, dataList...)
		}
	}

	// 设置文件名列表
	conflictInfo.FileNames = make([]string, 0, len(fileNamesSet))
	for fileName := range fileNamesSet {
		conflictInfo.FileNames = append(conflictInfo.FileNames, fileName)
	}

	conflictInfo.ConflictCount = len(conflictInfo.Conflicts)
	conflictInfo.HasConflict = conflictInfo.ConflictCount > 0

	return conflictInfo, nonConflictData
}

// mergeTable1DataWithTx 使用事务合并表1数据
func (a *App) mergeTable1DataWithTx(tx *sql.Tx, nonConflictData []map[string]interface{}) MergeResult {
	fmt.Printf("=== 进入表1合并函数，数据数量: %d ===\n", len(nonConflictData))
	result := MergeResult{}

	if len(nonConflictData) == 0 {
		fmt.Printf("表1没有数据需要插入\n")
		return result
	}

	// 直接插入没有冲突的数据
	insertQuery := `INSERT INTO enterprise_coal_consumption_main (
		obj_id, unit_name, stat_date, sg_code, tel, credit_code, create_time,
		trade_a, trade_b, trade_c, province_code, province_name, city_code, city_name,
		country_code, country_name, annual_energy_equivalent_value, annual_energy_equivalent_cost,
		annual_raw_material_energy, annual_total_coal_consumption, annual_total_coal_products,
		annual_raw_coal, annual_raw_coal_consumption, annual_clean_coal_consumption,
		annual_other_coal_consumption, annual_coke_consumption, create_user, is_confirm, is_check
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	fmt.Printf("表1没有冲突的数据数量: %d\n", len(nonConflictData))
	for i, row := range nonConflictData {
		// 使用源数据的obj_id、create_time、create_user，不生成新的
		fmt.Printf("表1插入第 %d 条数据: obj_id=%v, unit_name=%v, stat_date=%v\n",
			i+1, row["obj_id"], row["unit_name"], row["stat_date"])

		// 插入数据
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
			fmt.Printf("表1插入失败: %v\n", err)
		} else {
			result.SuccessCount++
			fmt.Printf("表1插入成功: 第 %d 条\n", i+1)
		}
	}

	result.Ok = true
	fmt.Printf("=== 表1合并函数结束，成功: %d, 失败: %d ===\n", result.SuccessCount, result.ErrorCount)
	return result
}

// mergeTable2DataWithTx 使用事务合并表2数据
func (a *App) mergeTable2DataWithTx(tx *sql.Tx, nonConflictData []map[string]interface{}) MergeResult {
	result := MergeResult{}

	if len(nonConflictData) == 0 {
		return result
	}

	// 直接插入没有冲突的数据
	insertQuery := `INSERT INTO critical_coal_equipment_consumption (
		obj_id, stat_date, create_time, sg_code, unit_name, credit_code, trade_a, trade_b, trade_c, trade_d,
		province_code, province_name, city_code, city_name, country_code, country_name, unit_addr,
		coal_type, coal_no, usage_time, design_life, enecrgy_efficienct_bmk, capacity_unit, capacity,
		use_info, status, annual_coal_consumption, row_no, create_user, is_confirm, is_check
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, row := range nonConflictData {
		// 使用源数据的obj_id、create_time、create_user，不生成新的

		// 插入数据
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

	result.Ok = true
	return result
}

// mergeTable3DataWithTx 使用事务合并表3数据
func (a *App) mergeTable3DataWithTx(tx *sql.Tx, nonConflictData []map[string]interface{}) MergeResult {
	result := MergeResult{}

	if len(nonConflictData) == 0 {
		return result
	}

	// 直接插入没有冲突的数据
	insertQuery := `INSERT INTO fixed_assets_investment_project (
		obj_id, stat_date, sg_code, project_name, project_code, construction_unit, main_construction_content,
		unit_id, province_name, city_name, country_name, trade_a, trade_c, examination_approval_time,
		scheduled_time, actual_time, examination_authority, document_number, equivalent_value, equivalent_cost,
		pq_total_coal_consumption, pq_coal_consumption, pq_coke_consumption, pq_blue_coke_consumption,
		sce_total_coal_consumption, sce_coal_consumption, sce_coke_consumption, sce_blue_coke_consumption,
		is_substitution, substitution_source, substitution_quantity, pq_annual_coal_quantity, sce_annual_coal_quantity,
		create_time, create_user, is_confirm, is_check
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, row := range nonConflictData {
		// 使用源数据的obj_id、create_time、create_user，不生成新的

		// 插入数据
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

	result.Ok = true
	return result
}

// mergeAttachment2DataWithTx 使用事务合并附件2数据
func (a *App) mergeAttachment2DataWithTx(tx *sql.Tx, nonConflictData []map[string]interface{}) MergeResult {
	result := MergeResult{}

	if len(nonConflictData) == 0 {
		return result
	}

	// 直接插入没有冲突的数据
	insertQuery := `INSERT INTO coal_consumption_report (
		obj_id, stat_date, sg_code, unit_id, unit_name, unit_level, province_name, city_name, country_name,
		total_coal, raw_coal, washed_coal, other_coal, power_generation, heating, coal_washing, coking,
		oil_refining, gas_production, industry, raw_materials, other_uses, coke, create_user, create_time, is_confirm, is_check
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	fmt.Printf("附件2没有冲突的数据数量: %d\n", len(nonConflictData))
	for i, row := range nonConflictData {
		// 使用源数据的obj_id、create_time、create_user，不生成新的
		fmt.Printf("附件2插入第 %d 条数据: obj_id=%v, unit_name=%v, stat_date=%v\n",
			i+1, row["obj_id"], row["unit_name"], row["stat_date"])

		// 插入数据
		_, err := tx.Exec(insertQuery,
			row["obj_id"], row["stat_date"], row["sg_code"], row["unit_id"], row["unit_name"], row["unit_level"],
			row["province_name"], row["city_name"], row["country_name"], row["total_coal"], row["raw_coal"],
			row["washed_coal"], row["other_coal"], row["power_generation"], row["heating"], row["coal_washing"],
			row["coking"], row["oil_refining"], row["gas_production"], row["industry"], row["raw_materials"],
			row["other_uses"], row["coke"], row["create_user"], row["create_time"], row["is_confirm"], row["is_check"])

		if err != nil {
			result.ErrorCount++
			result.Message = "插入数据失败: " + err.Error()
			fmt.Printf("附件2插入失败: %v\n", err)
		} else {
			result.SuccessCount++
			fmt.Printf("附件2插入成功: 第 %d 条\n", i+1)
		}
	}

	result.Ok = true
	return result
}

// 合并冲突数据
func (a *App) MergeConflictData(dbFilePath string, conflicts map[string][]ConflictSourceInfo) db.QueryResult {
	result := db.QueryResult{}

	if len(conflicts) == 0 {
		result.Ok = false
		result.Message = "没有冲突数据需要合并"
		return result
	}

	// 打开目标数据库
	targetDb, err := db.NewDatabase(dbFilePath, DB_PASSWORD)
	if err != nil {
		result.Ok = false
		result.Message = "打开目标数据库失败: " + err.Error()
		return result
	}
	defer targetDb.Close()

	// 开始事务
	tx, err := targetDb.Begin()
	if err != nil {
		result.Ok = false
		result.Message = "开始事务失败: " + err.Error()
		return result
	}

	// 确保事务回滚（如果出错）
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	totalSuccessCount := 0
	totalErrorCount := 0
	tableResults := make(map[string]map[string]interface{})

	// 处理所有表的冲突数据
	for tableType, tableConflicts := range conflicts {
		if len(tableConflicts) == 0 {
			continue
		}

		var successCount, errorCount int
		var tableErr error

		// 根据表类型处理冲突数据
		switch tableType {
		case "table1":
			successCount, errorCount, tableErr = a.mergeTable1ConflictData(tx, tableConflicts)
		case "table2":
			successCount, errorCount, tableErr = a.mergeTable2ConflictData(tx, tableConflicts)
		case "table3":
			successCount, errorCount, tableErr = a.mergeTable3ConflictData(tx, tableConflicts)
		case "attachment2":
			successCount, errorCount, tableErr = a.mergeAttachment2ConflictData(tx, tableConflicts)
		default:
			// 跳过不支持的表类型
			continue
		}

		// 记录每个表的结果
		tableResults[tableType] = map[string]interface{}{
			"successCount": successCount,
			"errorCount":   errorCount,
			"error":        tableErr,
		}

		totalSuccessCount += successCount
		totalErrorCount += errorCount

		// 如果某个表处理出错，记录错误但不中断整个事务
		if tableErr != nil {
			err = fmt.Errorf("表 %s 处理失败: %v", tableType, tableErr)
		}
	}

	// 提交事务
	commitErr := tx.Commit()
	if commitErr != nil {
		result.Ok = false
		result.Message = "提交事务失败: " + commitErr.Error()
		return result
	}

	result.Ok = true
	result.Message = fmt.Sprintf("成功合并 %d 条冲突数据，错误: %d 条", totalSuccessCount, totalErrorCount)
	result.Data = map[string]interface{}{
		"totalSuccessCount": totalSuccessCount,
		"totalErrorCount":   totalErrorCount,
		"tableResults":      tableResults,
	}

	return result
}

// mergeTable1ConflictData 合并表1冲突数据
func (a *App) mergeTable1ConflictData(tx *sql.Tx, conflicts []ConflictSourceInfo) (int, int, error) {
	successCount := 0
	errorCount := 0

	for _, conflict := range conflicts {
		// 打开源数据库
		sourceDb, err := db.NewDatabase(conflict.FilePath, DB_PASSWORD)
		if err != nil {
			errorCount++
			continue
		}
		defer sourceDb.Close()

		// 查询源数据
		query := `SELECT * FROM enterprise_coal_consumption_main WHERE obj_id IN (`
		placeholders := strings.Repeat("?,", len(conflict.ObjIds))
		placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
		query += placeholders + ")"

		args := make([]interface{}, len(conflict.ObjIds))
		for i, objId := range conflict.ObjIds {
			args[i] = objId
		}

		result, err := sourceDb.Query(query, args...)
		if err != nil {
			errorCount++
			continue
		}

		if !result.Ok || result.Data == nil {
			errorCount++
			continue
		}

		data, ok := result.Data.([]map[string]interface{})
		if !ok {
			errorCount++
			continue
		}

		// 更新目标数据库中的数据
		for _, row := range data {
			// 更新数据，使用obj_id进行精确更新，不更新create_time、create_user、obj_id
			updateQuery := `UPDATE enterprise_coal_consumption_main SET 
				unit_name = ?, stat_date = ?, sg_code = ?, tel = ?, credit_code = ?,
				trade_a = ?, trade_b = ?, trade_c = ?, province_code = ?, province_name = ?, city_code = ?, city_name = ?,
				country_code = ?, country_name = ?, annual_energy_equivalent_value = ?, annual_energy_equivalent_cost = ?,
				annual_raw_material_energy = ?, annual_total_coal_consumption = ?, annual_total_coal_products = ?,
				annual_raw_coal = ?, annual_raw_coal_consumption = ?, annual_clean_coal_consumption = ?,
				annual_other_coal_consumption = ?, annual_coke_consumption = ?, is_confirm = ?, is_check = ?
				WHERE obj_id = ?`

			_, err := tx.Exec(updateQuery,
				row["unit_name"], row["stat_date"], row["sg_code"], row["tel"], row["credit_code"],
				row["trade_a"], row["trade_b"], row["trade_c"], row["province_code"], row["province_name"], row["city_code"], row["city_name"],
				row["country_code"], row["country_name"], row["annual_energy_equivalent_value"], row["annual_energy_equivalent_cost"],
				row["annual_raw_material_energy"], row["annual_total_coal_consumption"], row["annual_total_coal_products"],
				row["annual_raw_coal"], row["annual_raw_coal_consumption"], row["annual_clean_coal_consumption"],
				row["annual_other_coal_consumption"], row["annual_coke_consumption"], row["is_confirm"], row["is_check"],
				row["obj_id"])

			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		}
	}

	return successCount, errorCount, nil
}

// mergeTable2ConflictData 合并表2冲突数据
func (a *App) mergeTable2ConflictData(tx *sql.Tx, conflicts []ConflictSourceInfo) (int, int, error) {
	successCount := 0
	errorCount := 0

	for _, conflict := range conflicts {
		// 打开源数据库
		sourceDb, err := db.NewDatabase(conflict.FilePath, DB_PASSWORD)
		if err != nil {
			errorCount++
			continue
		}
		defer sourceDb.Close()

		// 查询源数据
		query := `SELECT * FROM critical_coal_equipment_consumption WHERE obj_id IN (`
		placeholders := strings.Repeat("?,", len(conflict.ObjIds))
		placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
		query += placeholders + ")"

		args := make([]interface{}, len(conflict.ObjIds))
		for i, objId := range conflict.ObjIds {
			args[i] = objId
		}

		result, err := sourceDb.Query(query, args...)
		if err != nil {
			errorCount++
			continue
		}

		if !result.Ok || result.Data == nil {
			errorCount++
			continue
		}

		data, ok := result.Data.([]map[string]interface{})
		if !ok {
			errorCount++
			continue
		}

		// 更新目标数据库中的数据
		for _, row := range data {
			// 更新数据，使用obj_id进行精确更新，不更新create_time、create_user、obj_id
			updateQuery := `UPDATE critical_coal_equipment_consumption SET 
				stat_date = ?, sg_code = ?, unit_name = ?, credit_code = ?, trade_a = ?, trade_b = ?, trade_c = ?, trade_d = ?,
				province_code = ?, province_name = ?, city_code = ?, city_name = ?, country_code = ?, country_name = ?, unit_addr = ?,
				coal_type = ?, coal_no = ?, usage_time = ?, design_life = ?, enecrgy_efficienct_bmk = ?, capacity_unit = ?, capacity = ?,
				use_info = ?, status = ?, annual_coal_consumption = ?, row_no = ?, is_confirm = ?, is_check = ?
				WHERE obj_id = ?`

			_, err := tx.Exec(updateQuery,
				row["stat_date"], row["sg_code"], row["unit_name"], row["credit_code"], row["trade_a"], row["trade_b"], row["trade_c"], row["trade_d"],
				row["province_code"], row["province_name"], row["city_code"], row["city_name"], row["country_code"], row["country_name"], row["unit_addr"],
				row["coal_type"], row["coal_no"], row["usage_time"], row["design_life"], row["enecrgy_efficienct_bmk"], row["capacity_unit"], row["capacity"],
				row["use_info"], row["status"], row["annual_coal_consumption"], row["row_no"], row["is_confirm"], row["is_check"],
				row["obj_id"])

			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		}
	}

	return successCount, errorCount, nil
}

// mergeTable3ConflictData 合并表3冲突数据
func (a *App) mergeTable3ConflictData(tx *sql.Tx, conflicts []ConflictSourceInfo) (int, int, error) {
	successCount := 0
	errorCount := 0

	for _, conflict := range conflicts {
		// 打开源数据库
		sourceDb, err := db.NewDatabase(conflict.FilePath, DB_PASSWORD)
		if err != nil {
			errorCount++
			continue
		}
		defer sourceDb.Close()

		// 查询源数据
		query := `SELECT * FROM fixed_assets_investment_project WHERE obj_id IN (`
		placeholders := strings.Repeat("?,", len(conflict.ObjIds))
		placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
		query += placeholders + ")"

		args := make([]interface{}, len(conflict.ObjIds))
		for i, objId := range conflict.ObjIds {
			args[i] = objId
		}

		result, err := sourceDb.Query(query, args...)
		if err != nil {
			errorCount++
			continue
		}

		if !result.Ok || result.Data == nil {
			errorCount++
			continue
		}

		data, ok := result.Data.([]map[string]interface{})
		if !ok {
			errorCount++
			continue
		}

		// 更新目标数据库中的数据
		for _, row := range data {
			// 更新数据，使用obj_id进行精确更新，不更新create_time、create_user、obj_id
			updateQuery := `UPDATE fixed_assets_investment_project SET 
				stat_date = ?, sg_code = ?, project_name = ?, project_code = ?, construction_unit = ?, main_construction_content = ?,
				unit_id = ?, province_name = ?, city_name = ?, country_name = ?, trade_a = ?, trade_c = ?, examination_approval_time = ?,
				scheduled_time = ?, actual_time = ?, examination_authority = ?, document_number = ?, equivalent_value = ?, equivalent_cost = ?,
				pq_total_coal_consumption = ?, pq_coal_consumption = ?, pq_coke_consumption = ?, pq_blue_coke_consumption = ?,
				sce_total_coal_consumption = ?, sce_coal_consumption = ?, sce_coke_consumption = ?, sce_blue_coke_consumption = ?,
				is_substitution = ?, substitution_source = ?, substitution_quantity = ?, pq_annual_coal_quantity = ?, sce_annual_coal_quantity = ?,
				is_confirm = ?, is_check = ?
				WHERE obj_id = ?`

			_, err := tx.Exec(updateQuery,
				row["stat_date"], row["sg_code"], row["project_name"], row["project_code"], row["construction_unit"], row["main_construction_content"],
				row["unit_id"], row["province_name"], row["city_name"], row["country_name"], row["trade_a"], row["trade_c"], row["examination_approval_time"],
				row["scheduled_time"], row["actual_time"], row["examination_authority"], row["document_number"], row["equivalent_value"], row["equivalent_cost"],
				row["pq_total_coal_consumption"], row["pq_coal_consumption"], row["pq_coke_consumption"], row["pq_blue_coke_consumption"],
				row["sce_total_coal_consumption"], row["sce_coal_consumption"], row["sce_coke_consumption"], row["sce_blue_coke_consumption"],
				row["is_substitution"], row["substitution_source"], row["substitution_quantity"], row["pq_annual_coal_quantity"], row["sce_annual_coal_quantity"],
				row["is_confirm"], row["is_check"],
				row["obj_id"])

			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		}
	}

	return successCount, errorCount, nil
}

// mergeAttachment2ConflictData 合并附件2冲突数据
func (a *App) mergeAttachment2ConflictData(tx *sql.Tx, conflicts []ConflictSourceInfo) (int, int, error) {
	successCount := 0
	errorCount := 0

	for _, conflict := range conflicts {
		// 打开源数据库
		sourceDb, err := db.NewDatabase(conflict.FilePath, DB_PASSWORD)
		if err != nil {
			errorCount++
			continue
		}
		defer sourceDb.Close()

		// 查询源数据
		query := `SELECT * FROM coal_consumption_report WHERE obj_id IN (`
		placeholders := strings.Repeat("?,", len(conflict.ObjIds))
		placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号
		query += placeholders + ")"

		args := make([]interface{}, len(conflict.ObjIds))
		for i, objId := range conflict.ObjIds {
			args[i] = objId
		}

		result, err := sourceDb.Query(query, args...)
		if err != nil {
			errorCount++
			continue
		}

		if !result.Ok || result.Data == nil {
			errorCount++
			continue
		}

		data, ok := result.Data.([]map[string]interface{})
		if !ok {
			errorCount++
			continue
		}

		// 更新目标数据库中的数据
		for _, row := range data {
			// 更新数据，使用obj_id进行精确更新，不更新create_time、create_user、obj_id
			updateQuery := `UPDATE coal_consumption_report SET 
				stat_date = ?, sg_code = ?, unit_id = ?, unit_name = ?, unit_level = ?, province_name = ?, city_name = ?, country_name = ?,
				total_coal = ?, raw_coal = ?, washed_coal = ?, other_coal = ?, power_generation = ?, heating = ?, coal_washing = ?, coking = ?,
				oil_refining = ?, gas_production = ?, industry = ?, raw_materials = ?, other_uses = ?, coke = ?, is_confirm = ?, is_check = ?
				WHERE obj_id = ?`

			_, err := tx.Exec(updateQuery,
				row["stat_date"], row["sg_code"], row["unit_id"], row["unit_name"], row["unit_level"],
				row["province_name"], row["city_name"], row["country_name"], row["total_coal"], row["raw_coal"],
				row["washed_coal"], row["other_coal"], row["power_generation"], row["heating"], row["coal_washing"],
				row["coking"], row["oil_refining"], row["gas_production"], row["industry"], row["raw_materials"],
				row["other_uses"], row["coke"], row["is_confirm"], row["is_check"],
				row["obj_id"])

			if err != nil {
				errorCount++
			} else {
				successCount++
			}
		}
	}

	return successCount, errorCount, nil
}
