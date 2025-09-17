package main

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"shuji/db"
	"strings"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
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

// ConflictData 冲突数据结构
type ConflictData struct {
	FilePath   string      `json:"filePath"`
	TableType  string      `json:"tableType"`
	Conditions []Condition `json:"conditions"`
}

// Condition 冲突条件结构
type Condition struct {
	CreditCode     string `json:"credit_code,omitempty"`     // 统一信用代码（表1、表2）
	StatDate       string `json:"stat_date,omitempty"`       // 年份（表1、表2、附件2）
	ProjectCode    string `json:"project_code,omitempty"`    // 项目代码（表3）
	DocumentNumber string `json:"document_number,omitempty"` // 审查意见文号（表3）
	ProvinceName   string `json:"province_name,omitempty"`   // 省（附件2）
	CityName       string `json:"city_name,omitempty"`       // 市（附件2）
	CountryName    string `json:"country_name,omitempty"`    // 县（附件2）
}

// MergeDatabase 合并数据库
func (a *App) MergeDatabase(province string, city string, country string, sourceDbPath []string) db.QueryResult {
	// 使用包装函数来处理异常
	return a.mergeDatabaseWithRecover(province, city, country, sourceDbPath)
}

// mergeDatabaseWithRecover 带异常处理的合并数据库函数
func (a *App) mergeDatabaseWithRecover(province string, city string, country string, sourceDbPath []string) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MergeDatabase 发生异常: %v", r)
		}
	}()

	areaConfig := AreaConfig{
		ProvinceName: province,
		CityName: city,
		CountryName: country,
	}

	result := db.QueryResult{
		Ok: false,
	}

	// 2. 创建新的空数据库并初始化表结构
	newDb, dbTempPath, err := a.CreateNewDatabase("merge_")
	if err != nil {
		result.Message = "创建新数据库失败: " + err.Error()
		return result
	}
	defer newDb.Close()

	// 3. 复制源数据库到临时文件并打开数据库连接
	var sourceDbs []*db.Database
	var sourceDbPaths []string
	var originalSourcePaths []string
	var failedFiles []string

	now := time.Now().Unix()
	for i, sourceDbPath := range sourceDbPath {
		dbDstPath := GetPath(filepath.Join(DATA_DIR_NAME,  strconv.FormatInt(now + int64(i), 16)))
		copyResult := a.Copyfile(sourceDbPath, dbDstPath)
		if !copyResult.Ok {
			failedFiles = append(failedFiles, sourceDbPath)
			continue
		}

		// 创建数据库连接
		sourceDb, err := db.NewDatabase(dbDstPath, DB_PASSWORD)
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
		result.Message = fmt.Sprintf("所有数据库文件打开失败: %v", failedFiles)
		// 删除所有数据库文件
		a.Removefile(dbTempPath)
		for i := range sourceDbPaths {
			a.Removefile(sourceDbPaths[i])
		}
		return result
	}

	// 4. 检查数据冲突（只在上传文件之间检查冲突）
	// 检查表1冲突（规上企业煤炭消费信息主表）
	table1Conflicts, table1NonConflictData, err := a.checkTable1Conflicts(sourceDbs, originalSourcePaths, areaConfig)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	table1ConflictCount := len(table1Conflicts.Conflicts)

	// 检查表2冲突（重点耗煤装置煤炭消耗信息表）
	table2Conflicts, table2NonConflictData, err := a.checkTable2Conflicts(sourceDbs, originalSourcePaths, areaConfig)
	if err != nil {
		result.Message = err.Error()
		return result
	}
	table2ConflictCount := len(table2Conflicts.Conflicts)

	// 检查表3冲突（固定资产投资项目节能审查煤炭消费情况汇总表）
	table3Conflicts, table3NonConflictData := a.checkTable3Conflicts(sourceDbs, originalSourcePaths)
	table3ConflictCount := len(table3Conflicts.Conflicts)

	// 检查附件2冲突（煤炭消费状况表）
	attachment2Conflicts, attachment2NonConflictData, err := a.checkAttachment2Conflicts(sourceDbs, originalSourcePaths, areaConfig)
	if err != nil {
		result.Message = err.Error()
		return result
	}
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
			a.Removefile(sourceDbPaths[i])
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
				a.Removefile(sourceDbPaths[i])
			}
		}
	}()

	// 统计实际冲突数量
	actualConflictCount := 0

	// 合并所有表的数据（现在只在上传文件之间检查冲突，所有数据都可以合并）
	// 对于有文件间冲突的数据，我们需要处理冲突，这里先合并非冲突数据
	table1Result := a.mergeTable1DataWithTx(tx, table1NonConflictData, sourceDbs, originalSourcePaths)
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
			a.Removefile(sourceDbPaths[i])
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
		a.Removefile(sourceDbPaths[i])
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
func (a *App) validateAreaConsistency( expectedProvince string,  expectedCity string,  expectedCountry string, areaConfig AreaConfig) error {

	err := fmt.Errorf("待合并数据文件区域与所选区域不一致，请检查")
	// 如果country有值就比较country
	if areaConfig.CountryName != "" {
		if areaConfig.CountryName != expectedCountry {
			return err	
		}
	}

	// 如果city有值就比较city
	if areaConfig.CityName != "" {
		if areaConfig.CityName != expectedCity {
			return err
		}
	}

	// 如果province有值就比较province
	if areaConfig.ProvinceName != "" {
		if areaConfig.ProvinceName != expectedProvince {
			return err
		}
	}

	return nil
}

// CreateNewDatabase 创建新的空数据库并初始化表结构
func (a *App) CreateNewDatabase(prefix string) (*db.Database, string, error) {
	// 生成新的数据库文件路径
	timestamp := time.Now().Unix()
	dbFileName := fmt.Sprintf("%s%d.db", prefix, timestamp)
	dbTempPath := GetPath(filepath.Join(DATA_DIR_NAME, dbFileName))

	fmt.Printf("创建新的数据库文件路径: %s\n", dbTempPath)
	fmt.Printf("创建新的数据库文件路径: %s\n", FRONTEND_FILE_DIR_NAME + DB_FILE_NAME)
	// 抽取数据库文件
	a.extractEmbeddedFile(FRONTEND_FILE_DIR_NAME + DB_FILE_NAME, dbTempPath)

	// 创建新的数据库连接
	newDb, err := db.NewDatabase(dbTempPath, DB_PASSWORD)
	if err != nil {
		return nil, "", fmt.Errorf("创建数据库连接失败: %v", err)
	}

	return newDb, dbTempPath, nil
}

// checkTable1Conflicts 检查表1冲突（统一信用代码+年份）- 只在上传文件之间检查冲突
func (a *App) checkTable1Conflicts(sourceDbs []*db.Database, originalSourcePaths []string, areaConfig AreaConfig) (TableConflictInfo, []map[string]interface{}, error) {
	conflictInfo := TableConflictInfo{}
	var nonConflictData []map[string]interface{}

	// 存储所有源数据，用于文件间冲突检查
	allSourceData := make(map[string][]map[string]interface{}) // key: filePath, value: source data
	fileNamesSet := make(map[string]bool)
	
	// 1. 收集所有源数据
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
				allSourceData[filePath] = data
			}
		}
	}

	// 2. 检查文件间冲突
	conflictKeyMap := make(map[string]map[string]ConflictSourceInfo) // key: conflictKey, value: map[filePath]ConflictSourceInfo
	
	// 首先收集所有可能的冲突键
	allKeys := make(map[string][]string) // key: conflictKey, value: []filePath
	for filePath, data := range allSourceData {
		for _, row := range data {
			key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
			if allKeys[key] == nil {
				allKeys[key] = []string{}
			}
			allKeys[key] = append(allKeys[key], filePath)

			areaValidation := a.validateAreaConsistency(getStringValue(row["province_name"]), getStringValue(row["city_name"]), getStringValue(row["country_name"]), areaConfig)
			if areaValidation != nil {
				return conflictInfo, nonConflictData, areaValidation
			}
		}
	}
	
	// 然后处理真正的冲突（同一个键出现在多个文件中）
	for conflictKey, filePaths := range allKeys {
		// 去重文件路径
		uniqueFilePaths := make(map[string]bool)
		for _, filePath := range filePaths {
			uniqueFilePaths[filePath] = true
		}
		
		// 只有当同一个键出现在多个文件中时才算冲突
		if len(uniqueFilePaths) > 1 {
			// 创建冲突映射
			conflictKeyMap[conflictKey] = make(map[string]ConflictSourceInfo)
			
			// 为每个包含此冲突键的文件创建冲突记录
			for filePath := range uniqueFilePaths {
				fileName := filepath.Base(filePath)
				var objIds []string
				
				// 收集该文件中所有匹配此冲突键的obj_id
				if data, exists := allSourceData[filePath]; exists {
					for _, row := range data {
						key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
						if key == conflictKey {
							objIds = append(objIds, fmt.Sprintf("%v", row["obj_id"]))
						}
					}
				}
				
				conflictKeyMap[conflictKey][filePath] = ConflictSourceInfo{
					FilePath:  filePath,
					FileName:  fileName,
					TableType: "table1",
					ObjIds:    objIds,
				}
			}
		} else {
			// 没有冲突，添加到非冲突数据
			filePath := filePaths[0] // 只有一个文件包含此键
			if data, exists := allSourceData[filePath]; exists {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
					if key == conflictKey {
						nonConflictData = append(nonConflictData, row)
					}
				}
			}
		}
	}

	// 3. 构建冲突详情，按冲突键分组
	for conflictKey, fileConflicts := range conflictKeyMap {
		// 解析冲突键
		parts := strings.Split(conflictKey, "_")
		if len(parts) != 2 {
			continue
		}

		// 创建冲突详情（不再依赖目标数据库）
		conflictDetail := ConflictDetail{
			ObjId:      "", // 不再使用目标数据库的obj_id
			CreditCode: parts[0],
			StatDate:   parts[1],
			UnitName:   "", // 从冲突的文件中获取
			Conflict:   make([]ConflictSourceInfo, 0, len(fileConflicts)),
		}

		// 从冲突的文件中获取unit_name
		for _, fileConflict := range fileConflicts {
			if data, exists := allSourceData[fileConflict.FilePath]; exists {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
					if key == conflictKey {
						conflictDetail.UnitName = fmt.Sprintf("%v", row["unit_name"])
						break
					}
				}
			}
			conflictDetail.Conflict = append(conflictDetail.Conflict, fileConflict)
		}

		conflictInfo.Conflicts = append(conflictInfo.Conflicts, conflictDetail)
	}

	// 设置文件名列表
	conflictInfo.FileNames = make([]string, 0, len(fileNamesSet))
	for fileName := range fileNamesSet {
		conflictInfo.FileNames = append(conflictInfo.FileNames, fileName)
	}

	conflictInfo.ConflictCount = len(conflictInfo.Conflicts)
	conflictInfo.HasConflict = conflictInfo.ConflictCount > 0

	return conflictInfo, nonConflictData, nil
}

// checkTable2Conflicts 检查表2冲突（统一信用代码+年份）- 只在上传文件之间检查冲突
func (a *App) checkTable2Conflicts(sourceDbs []*db.Database, originalSourcePaths []string, areaConfig AreaConfig) (TableConflictInfo, []map[string]interface{}, error) {
	conflictInfo := TableConflictInfo{}
	var nonConflictData []map[string]interface{}

	// 存储所有源数据，用于文件间冲突检查
	allSourceData := make(map[string][]map[string]interface{}) // key: filePath, value: source data
	fileNamesSet := make(map[string]bool)
	
	// 1. 收集所有源数据
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
				allSourceData[filePath] = data
			}
		}
	}

	// 2. 检查文件间冲突
	conflictKeyMap := make(map[string]map[string]ConflictSourceInfo) // key: conflictKey, value: map[filePath]ConflictSourceInfo
	
	// 首先收集所有可能的冲突键
	allKeys := make(map[string][]string) // key: conflictKey, value: []filePath
	for filePath, data := range allSourceData {
		for _, row := range data {
			key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
			if allKeys[key] == nil {
				allKeys[key] = []string{}
			}
			allKeys[key] = append(allKeys[key], filePath)

			areaValidation := a.validateAreaConsistency(getStringValue(row["province_name"]), getStringValue(row["city_name"]), getStringValue(row["country_name"]), areaConfig)
			if areaValidation != nil {
				return conflictInfo, nonConflictData, areaValidation
			}
		}
	}
	
	// 然后处理真正的冲突（同一个键出现在多个文件中）
	for conflictKey, filePaths := range allKeys {
		// 去重文件路径
		uniqueFilePaths := make(map[string]bool)
		for _, filePath := range filePaths {
			uniqueFilePaths[filePath] = true
		}
		
		// 只有当同一个键出现在多个文件中时才算冲突
		if len(uniqueFilePaths) > 1 {
			// 创建冲突映射
			conflictKeyMap[conflictKey] = make(map[string]ConflictSourceInfo)
			
			// 为每个包含此冲突键的文件创建冲突记录
			for filePath := range uniqueFilePaths {
				fileName := filepath.Base(filePath)
				var objIds []string
				
				// 收集该文件中所有匹配此冲突键的obj_id
				if data, exists := allSourceData[filePath]; exists {
					for _, row := range data {
						key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
						if key == conflictKey {
							objIds = append(objIds, fmt.Sprintf("%v", row["obj_id"]))
						}
					}
				}
				
				conflictKeyMap[conflictKey][filePath] = ConflictSourceInfo{
					FilePath:  filePath,
					FileName:  fileName,
					TableType: "table2",
					ObjIds:    objIds,
				}
			}
		} else {
			// 没有冲突，添加到非冲突数据
			filePath := filePaths[0] // 只有一个文件包含此键
			if data, exists := allSourceData[filePath]; exists {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
					if key == conflictKey {
						nonConflictData = append(nonConflictData, row)
					}
				}
			}
		}
	}

	// 3. 构建冲突详情，按冲突键分组
	for conflictKey, fileConflicts := range conflictKeyMap {
		// 解析冲突键
		parts := strings.Split(conflictKey, "_")
		if len(parts) != 2 {
			continue
		}

		// 创建冲突详情（不再依赖目标数据库）
		conflictDetail := ConflictDetail{
			ObjId:      "", // 不再使用目标数据库的obj_id
			CreditCode: parts[0],
			StatDate:   parts[1],
			UnitName:   "", // 从冲突的文件中获取
			Conflict:   make([]ConflictSourceInfo, 0, len(fileConflicts)),
		}

		// 从冲突的文件中获取unit_name
		for _, fileConflict := range fileConflicts {
			if data, exists := allSourceData[fileConflict.FilePath]; exists {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["credit_code"], row["stat_date"])
					if key == conflictKey {
						conflictDetail.UnitName = fmt.Sprintf("%v", row["unit_name"])
						break
					}
				}
			}
			conflictDetail.Conflict = append(conflictDetail.Conflict, fileConflict)
		}

		conflictInfo.Conflicts = append(conflictInfo.Conflicts, conflictDetail)
	}

	// 设置文件名列表
	conflictInfo.FileNames = make([]string, 0, len(fileNamesSet))
	for fileName := range fileNamesSet {
		conflictInfo.FileNames = append(conflictInfo.FileNames, fileName)
	}

	conflictInfo.ConflictCount = len(conflictInfo.Conflicts)
	conflictInfo.HasConflict = conflictInfo.ConflictCount > 0

	return conflictInfo, nonConflictData, nil
}

// checkTable3Conflicts 检查表3冲突（项目代码+审查意见文号）- 只在上传文件之间检查冲突
func (a *App) checkTable3Conflicts(sourceDbs []*db.Database, originalSourcePaths []string) (TableConflictInfo, []map[string]interface{}) {
	conflictInfo := TableConflictInfo{}
	var nonConflictData []map[string]interface{}

	// 存储所有源数据，用于文件间冲突检查
	allSourceData := make(map[string][]map[string]interface{}) // key: filePath, value: source data
	fileNamesSet := make(map[string]bool)
	
	// 1. 收集所有源数据
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
				allSourceData[filePath] = data
			}
		}
	}

	// 2. 检查文件间冲突
	conflictKeyMap := make(map[string]map[string]ConflictSourceInfo) // key: conflictKey, value: map[filePath]ConflictSourceInfo
	
	// 首先收集所有可能的冲突键
	allKeys := make(map[string][]string) // key: conflictKey, value: []filePath
	for filePath, data := range allSourceData {
		for _, row := range data {
			key := fmt.Sprintf("%v_%v", row["project_code"], row["document_number"])
			if allKeys[key] == nil {
				allKeys[key] = []string{}
			}
			allKeys[key] = append(allKeys[key], filePath)
		}
	}
	
	// 然后处理真正的冲突（同一个键出现在多个文件中）
	for conflictKey, filePaths := range allKeys {
		// 去重文件路径
		uniqueFilePaths := make(map[string]bool)
		for _, filePath := range filePaths {
			uniqueFilePaths[filePath] = true
		}
		
		// 只有当同一个键出现在多个文件中时才算冲突
		if len(uniqueFilePaths) > 1 {
			// 创建冲突映射
			conflictKeyMap[conflictKey] = make(map[string]ConflictSourceInfo)
			
			// 为每个包含此冲突键的文件创建冲突记录
			for filePath := range uniqueFilePaths {
				fileName := filepath.Base(filePath)
				var objIds []string
				
				// 收集该文件中所有匹配此冲突键的obj_id
				if data, exists := allSourceData[filePath]; exists {
					for _, row := range data {
						key := fmt.Sprintf("%v_%v", row["project_code"], row["document_number"])
						if key == conflictKey {
							objIds = append(objIds, fmt.Sprintf("%v", row["obj_id"]))
						}
					}
				}
				
				conflictKeyMap[conflictKey][filePath] = ConflictSourceInfo{
					FilePath:  filePath,
					FileName:  fileName,
					TableType: "table3",
					ObjIds:    objIds,
				}
			}
		} else {
			// 没有冲突，添加到非冲突数据
			filePath := filePaths[0] // 只有一个文件包含此键
			if data, exists := allSourceData[filePath]; exists {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["project_code"], row["document_number"])
					if key == conflictKey {
						nonConflictData = append(nonConflictData, row)
					}
				}
			}
		}
	}

	// 3. 构建冲突详情，按冲突键分组
	for conflictKey, fileConflicts := range conflictKeyMap {
		// 解析冲突键
		parts := strings.Split(conflictKey, "_")
		if len(parts) != 2 {
			continue
		}

		// 创建冲突详情（不再依赖目标数据库）
		conflictDetail := ConflictDetail{
			ObjId:          "", // 不再使用目标数据库的obj_id
			ProjectCode:    parts[0],
			DocumentNumber: parts[1],
			ProjectName:    "", // 从冲突的文件中获取
			Conflict:       make([]ConflictSourceInfo, 0, len(fileConflicts)),
		}

		// 从冲突的文件中获取project_name
		for _, fileConflict := range fileConflicts {
			if data, exists := allSourceData[fileConflict.FilePath]; exists {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v", row["project_code"], row["document_number"])
					if key == conflictKey {
						conflictDetail.ProjectName = fmt.Sprintf("%v", row["project_name"])
						break
					}
				}
			}
			conflictDetail.Conflict = append(conflictDetail.Conflict, fileConflict)
		}

		conflictInfo.Conflicts = append(conflictInfo.Conflicts, conflictDetail)
	}

	// 设置文件名列表
	conflictInfo.FileNames = make([]string, 0, len(fileNamesSet))
	for fileName := range fileNamesSet {
		conflictInfo.FileNames = append(conflictInfo.FileNames, fileName)
	}

	conflictInfo.ConflictCount = len(conflictInfo.Conflicts)
	conflictInfo.HasConflict = conflictInfo.ConflictCount > 0

	fmt.Printf("表3冲突检查完成 - 冲突数: %d, 非冲突数据数: %d\n", len(conflictInfo.Conflicts), len(nonConflictData))
	if len(nonConflictData) > 0 {
		fmt.Printf("表3非冲突数据示例: obj_id=%v, project_name=%v, project_code=%v\n",
			nonConflictData[0]["obj_id"], nonConflictData[0]["project_name"], nonConflictData[0]["project_code"])
	}

	return conflictInfo, nonConflictData
}

// checkAttachment2Conflicts 检查附件2冲突（省+市+县+年份）- 只在上传文件之间检查冲突
func (a *App) checkAttachment2Conflicts(sourceDbs []*db.Database, originalSourcePaths []string, areaConfig AreaConfig) (TableConflictInfo, []map[string]interface{}, error) {
	conflictInfo := TableConflictInfo{}
	var nonConflictData []map[string]interface{}

	// 存储所有源数据，用于文件间冲突检查
	allSourceData := make(map[string][]map[string]interface{}) // key: filePath, value: source data
	fileNamesSet := make(map[string]bool)
	
	// 1. 收集所有源数据
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
				allSourceData[filePath] = data
			}
		}
	}

	// 2. 检查文件间冲突
	conflictKeyMap := make(map[string]map[string]ConflictSourceInfo) // key: conflictKey, value: map[filePath]ConflictSourceInfo
	
	// 首先收集所有可能的冲突键
	allKeys := make(map[string][]string) // key: conflictKey, value: []filePath
	for filePath, data := range allSourceData {
		for _, row := range data {
			key := fmt.Sprintf("%v_%v_%v_%v", row["province_name"], row["city_name"], row["country_name"], row["stat_date"])
			if allKeys[key] == nil {
				allKeys[key] = []string{}
			}
			allKeys[key] = append(allKeys[key], filePath)

			areaValidation := a.validateAreaConsistency(getStringValue(row["province_name"]), getStringValue(row["city_name"]), getStringValue(row["country_name"]), areaConfig)
			if areaValidation != nil {
				return conflictInfo, nonConflictData, areaValidation
			}
		}
	}
	
	// 然后处理真正的冲突（同一个键出现在多个文件中）
	for conflictKey, filePaths := range allKeys {
		// 去重文件路径
		uniqueFilePaths := make(map[string]bool)
		for _, filePath := range filePaths {
			uniqueFilePaths[filePath] = true
		}
		
		// 只有当同一个键出现在多个文件中时才算冲突
		if len(uniqueFilePaths) > 1 {
			// 创建冲突映射
			conflictKeyMap[conflictKey] = make(map[string]ConflictSourceInfo)
			
			// 为每个包含此冲突键的文件创建冲突记录
			for filePath := range uniqueFilePaths {
				fileName := filepath.Base(filePath)
				var objIds []string
				
				// 收集该文件中所有匹配此冲突键的obj_id
				if data, exists := allSourceData[filePath]; exists {
					for _, row := range data {
						key := fmt.Sprintf("%v_%v_%v_%v", row["province_name"], row["city_name"], row["country_name"], row["stat_date"])
						if key == conflictKey {
							objIds = append(objIds, fmt.Sprintf("%v", row["obj_id"]))
						}
					}
				}
				
				conflictKeyMap[conflictKey][filePath] = ConflictSourceInfo{
					FilePath:  filePath,
					FileName:  fileName,
					TableType: "attachment2",
					ObjIds:    objIds,
				}
			}
		} else {
			// 没有冲突，添加到非冲突数据
			filePath := filePaths[0] // 只有一个文件包含此键
			if data, exists := allSourceData[filePath]; exists {
				for _, row := range data {
					key := fmt.Sprintf("%v_%v_%v_%v", row["province_name"], row["city_name"], row["country_name"], row["stat_date"])
					if key == conflictKey {
						nonConflictData = append(nonConflictData, row)
					}
				}
			}
		}
	}

	// 3. 构建冲突详情，按冲突键分组
	for conflictKey, fileConflicts := range conflictKeyMap {
		// 解析冲突键
		parts := strings.Split(conflictKey, "_")
		if len(parts) != 4 {
			continue
		}

		// 创建冲突详情（不再依赖目标数据库）
		conflictDetail := ConflictDetail{
			ObjId:        "", // 不再使用目标数据库的obj_id
			ProvinceName: parts[0],
			CityName:     parts[1],
			CountryName:  parts[2],
			StatDate:     parts[3],
			Conflict:     make([]ConflictSourceInfo, 0, len(fileConflicts)),
		}

		// 将文件冲突信息添加到冲突详情中
		for _, fileConflict := range fileConflicts {
			conflictDetail.Conflict = append(conflictDetail.Conflict, fileConflict)
		}

		conflictInfo.Conflicts = append(conflictInfo.Conflicts, conflictDetail)
	}

	// 设置文件名列表
	conflictInfo.FileNames = make([]string, 0, len(fileNamesSet))
	for fileName := range fileNamesSet {
		conflictInfo.FileNames = append(conflictInfo.FileNames, fileName)
	}

	conflictInfo.ConflictCount = len(conflictInfo.Conflicts)
	conflictInfo.HasConflict = conflictInfo.ConflictCount > 0

	return conflictInfo, nonConflictData, nil
}

// mergeTable1DataWithTx 使用事务合并表1数据
func (a *App) mergeTable1DataWithTx(tx *sql.Tx, nonConflictData []map[string]interface{}, sourceDbs []*db.Database, originalSourcePaths []string) MergeResult {
	result := MergeResult{}

	if len(nonConflictData) == 0 {
		fmt.Printf("表1没有数据需要插入\n")
		return result
	}

	// 1. 插入主表数据
	insertQuery := `INSERT INTO enterprise_coal_consumption_main (
		obj_id, unit_name, stat_date, sg_code, tel, credit_code,
		trade_a, trade_b, trade_c, province_code, province_name, city_code, city_name,
		country_code, country_name, annual_energy_equivalent_value, annual_energy_equivalent_cost,
		annual_raw_material_energy, annual_total_coal_consumption, annual_total_coal_products,
		annual_raw_coal, annual_raw_coal_consumption, annual_clean_coal_consumption,
		annual_other_coal_consumption, annual_coke_consumption, is_confirm, is_check, create_time, create_user
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	for _, row := range nonConflictData {
		// 使用源数据的obj_id、create_time、create_user，不生成新的

		// 插入数据
		_, err := tx.Exec(insertQuery,
			row["obj_id"], row["unit_name"], row["stat_date"], row["sg_code"], row["tel"], row["credit_code"],
			row["trade_a"], row["trade_b"], row["trade_c"], row["province_code"], row["province_name"], row["city_code"], row["city_name"],
			row["country_code"], row["country_name"], row["annual_energy_equivalent_value"], row["annual_energy_equivalent_cost"],
			row["annual_raw_material_energy"], row["annual_total_coal_consumption"], row["annual_total_coal_products"],
			row["annual_raw_coal"], row["annual_raw_coal_consumption"], row["annual_clean_coal_consumption"],
			row["annual_other_coal_consumption"], row["annual_coke_consumption"], row["is_confirm"], row["is_check"], row["create_time"], row["create_user"])

		if err != nil {
			result.ErrorCount++
			result.Message = "插入主表数据失败: " + err.Error()
			fmt.Printf("表1主表插入失败: %v\n", err)
		} else {
			result.SuccessCount++
		}
	}

	// 2. 插入扩展表数据
	// 2.1 插入主要用途情况表
	usageInsertQuery := `INSERT INTO enterprise_coal_consumption_usage (
		obj_id, fk_id, stat_date, create_time, main_usage, specific_usage, input_variety, input_unit,
		input_quantity, output_energy_types, output_quantity, measurement_unit, remarks, row_no, is_confirm, is_check
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 2.2 插入重点耗煤装置情况表
	equipInsertQuery := `INSERT INTO enterprise_coal_consumption_equip (
		obj_id, fk_id, stat_date, create_time, equip_type, equip_no, total_runtime, design_life,
		energy_efficiency, capacity_unit, capacity, coal_type, annual_coal_consumption, row_no, is_confirm, is_check
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// 遍历每个源数据库，查询并插入扩展表数据
	for _, sourceDb := range sourceDbs {
		// 根据主表的credit_code查询对应的扩展表数据
		for _, mainRow := range nonConflictData {

			// 查询主要用途情况表
			usageQuery := `SELECT * FROM enterprise_coal_consumption_usage WHERE fk_id = ?`
			usageResult, err := sourceDb.Query(usageQuery, mainRow["obj_id"])
			if err == nil && usageResult.Ok && usageResult.Data != nil {
				if data, ok := usageResult.Data.([]map[string]interface{}); ok {
					if len(data) > 0 {

						for _, row := range data {
							_, err := tx.Exec(usageInsertQuery,
								row["obj_id"], row["fk_id"], row["stat_date"], row["create_time"], row["main_usage"], row["specific_usage"],
								row["input_variety"], row["input_unit"], row["input_quantity"], row["output_energy_types"],
								row["output_quantity"], row["measurement_unit"], row["remarks"], row["row_no"], row["is_confirm"], row["is_check"])

							if err != nil {
								result.ErrorCount++
								result.Message = "插入主要用途情况表数据失败: " + err.Error()
								fmt.Printf("主要用途情况表插入失败: %v\n", err)
							} else {
								result.SuccessCount++
							}
						}
					}
				}
			}

			// 查询重点耗煤装置情况表
			equipQuery := `SELECT * FROM enterprise_coal_consumption_equip WHERE fk_id = ?`
			equipResult, err := sourceDb.Query(equipQuery, mainRow["obj_id"])
			if err == nil && equipResult.Ok && equipResult.Data != nil {
				if data, ok := equipResult.Data.([]map[string]interface{}); ok {
					if len(data) > 0 {
						for _, row := range data {
							_, err := tx.Exec(equipInsertQuery,
								row["obj_id"], row["fk_id"], row["stat_date"], row["create_time"], row["equip_type"], row["equip_no"],
								row["total_runtime"], row["design_life"], row["energy_efficiency"], row["capacity_unit"],
								row["capacity"], row["coal_type"], row["annual_coal_consumption"], row["row_no"], row["is_confirm"], row["is_check"])

							if err != nil {
								result.ErrorCount++
								result.Message = "插入重点耗煤装置情况表数据失败: " + err.Error()
								fmt.Printf("重点耗煤装置情况表插入失败: %v\n", err)
							} else {
								result.SuccessCount++
							}
						}
					}
				}
			}
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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

	for _, row := range nonConflictData {
		// 使用源数据的obj_id、create_time、create_user，不生成新的
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
		}
	}

	result.Ok = true
	return result
}

// 合并冲突数据
func (a *App) MergeConflictData(dbFilePath string, conflictData []ConflictData) db.QueryResult {
	// 使用包装函数来处理异常
	return a.mergeConflictDataWithRecover(dbFilePath, conflictData)
}

// mergeConflictDataWithRecover 带异常处理的合并冲突数据函数
func (a *App) mergeConflictDataWithRecover(dbFilePath string, conflictData []ConflictData) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MergeConflictData 发生异常: %v", r)
		}
	}()

	result := db.QueryResult{}

	if len(conflictData) == 0 {
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

	// 处理所有冲突数据
	for _, conflict := range conflictData {
		if len(conflict.Conditions) == 0 {
			continue
		}

		var successCount, errorCount int
		var tableErr error

		// 根据表类型处理冲突数据
		switch conflict.TableType {
		case "table1":
			successCount, errorCount, tableErr = a.mergeTable1ConflictDataNew(tx, conflict)
		case "table2":
			successCount, errorCount, tableErr = a.mergeTable2ConflictDataNew(tx, conflict)
		case "table3":
			successCount, errorCount, tableErr = a.mergeTable3ConflictDataNew(tx, conflict)
		case "attachment2":
			successCount, errorCount, tableErr = a.mergeAttachment2ConflictDataNew(tx, conflict)
		default:
			// 跳过不支持的表类型
			continue
		}

		// 记录每个表的结果
		tableResults[conflict.TableType] = map[string]interface{}{
			"successCount": successCount,
			"errorCount":   errorCount,
			"error":        tableErr,
		}

		totalSuccessCount += successCount
		totalErrorCount += errorCount

		// 如果某个表处理出错，记录错误但不中断整个事务
		if tableErr != nil {
			err = fmt.Errorf("表 %s 处理失败: %v", conflict.TableType, tableErr)
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

// mergeTable1ConflictDataNew 合并表1冲突数据（新版本）
func (a *App) mergeTable1ConflictDataNew(tx *sql.Tx, conflict ConflictData) (int, int, error) {
	successCount := 0
	errorCount := 0

	// 打开源数据库
	sourceDb, err := db.NewDatabase(conflict.FilePath, DB_PASSWORD)
	if err != nil {
		return 0, 0, fmt.Errorf("打开源数据库失败: %v", err)
	}
	defer sourceDb.Close()

	// 处理每个冲突条件
	for _, condition := range conflict.Conditions {
		// 先查询目标表中符合条件的所有obj_id
		targetObjIdsQuery := `SELECT obj_id FROM enterprise_coal_consumption_main WHERE credit_code = ? AND stat_date = ?`
		targetRows, err := tx.Query(targetObjIdsQuery, condition.CreditCode, condition.StatDate)
		if err != nil {
			errorCount++
			continue
		}
		defer targetRows.Close()

		var targetObjIds []string
		for targetRows.Next() {
			var targetObjId string
			err = targetRows.Scan(&targetObjId)
			if err != nil {
				fmt.Printf("扫描目标表obj_id失败: %v\n", err)
				errorCount++
				continue
			}
			targetObjIds = append(targetObjIds, targetObjId)
		}

		if len(targetObjIds) == 0 {
			fmt.Printf("未找到目标表对应的obj_id: credit_code=%s, stat_date=%s\n", condition.CreditCode, condition.StatDate)
			continue
		}

		// 根据查询到的obj_id删除扩展表数据
		for _, targetObjId := range targetObjIds {
			// 删除主要用途情况表数据
			deleteUsageQuery := `DELETE FROM enterprise_coal_consumption_usage WHERE fk_id = ?`
			_, err = tx.Exec(deleteUsageQuery, targetObjId)
			if err != nil {
				fmt.Printf("删除主要用途情况表数据失败: fk_id=%s, error=%v\n", targetObjId, err)
				errorCount++
				continue
			}

			// 删除重点耗煤装置情况表数据
			deleteEquipQuery := `DELETE FROM enterprise_coal_consumption_equip WHERE fk_id = ?`
			_, err = tx.Exec(deleteEquipQuery, targetObjId)
			if err != nil {
				fmt.Printf("删除重点耗煤装置情况表数据失败: fk_id=%s, error=%v\n", targetObjId, err)
				errorCount++
				continue
			}
		}

		// 删除主表数据
		deleteMainQuery := `DELETE FROM enterprise_coal_consumption_main WHERE credit_code = ? AND stat_date = ?`
		_, err = tx.Exec(deleteMainQuery, condition.CreditCode, condition.StatDate)
		if err != nil {
			fmt.Printf("删除主表记录失败: credit_code=%s, stat_date=%s, error=%v\n", condition.CreditCode, condition.StatDate, err)
			errorCount++
			continue
		}

		// 查询源数据并插入
		query := `SELECT * FROM enterprise_coal_consumption_main WHERE credit_code = ? AND stat_date = ?`
		result, err := sourceDb.Query(query, condition.CreditCode, condition.StatDate)
		if err != nil {
			fmt.Printf("查询源数据失败: credit_code=%s, stat_date=%s, error=%v\n", condition.CreditCode, condition.StatDate, err)
			errorCount++
			continue
		}

		if !result.Ok || result.Data == nil {
			fmt.Printf("查询源数据失败 mergeTable1ConflictDataNew1: credit_code=%s, stat_date=%s, result=%v\n", condition.CreditCode, condition.StatDate, result)
			continue
		}

		data, ok := result.Data.([]map[string]interface{})
		if !ok {
			continue
		}

		// 插入源数据到目标表
		for _, row := range data {
			// 完全按照源表数据插入，包括obj_id
			insertQuery := `INSERT INTO enterprise_coal_consumption_main (
				obj_id, unit_name, stat_date, sg_code, tel, credit_code,
				trade_a, trade_b, trade_c, province_code, province_name, city_code, city_name,
				country_code, country_name, annual_energy_equivalent_value, annual_energy_equivalent_cost,
				annual_raw_material_energy, annual_total_coal_consumption, annual_total_coal_products,
				annual_raw_coal, annual_raw_coal_consumption, annual_clean_coal_consumption,
				annual_other_coal_consumption, annual_coke_consumption, is_confirm, is_check, create_time, create_user
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

			_, execErr := tx.Exec(insertQuery,
				row["obj_id"], row["unit_name"], row["stat_date"], row["sg_code"], row["tel"], row["credit_code"],
				row["trade_a"], row["trade_b"], row["trade_c"], row["province_code"], row["province_name"], row["city_code"], row["city_name"],
				row["country_code"], row["country_name"], row["annual_energy_equivalent_value"], row["annual_energy_equivalent_cost"],
				row["annual_raw_material_energy"], row["annual_total_coal_consumption"], row["annual_total_coal_products"],
				row["annual_raw_coal"], row["annual_raw_coal_consumption"], row["annual_clean_coal_consumption"],
				row["annual_other_coal_consumption"], row["annual_coke_consumption"], row["is_confirm"], row["is_check"], row["create_time"], row["create_user"])

			if execErr != nil {
				fmt.Printf("插入主表数据失败: obj_id=%s, stat_date=%s, error=%v\n", row["obj_id"], row["stat_date"], execErr)
				errorCount++
			} else {
				successCount++
			}

			// 插入扩展表数据
			// 插入主要用途情况表数据
			usageQuery := `SELECT * FROM enterprise_coal_consumption_usage WHERE fk_id = ?`
			usageResult, err := sourceDb.Query(usageQuery, row["obj_id"])
			if err == nil && usageResult.Ok && usageResult.Data != nil {
				if usageData, ok := usageResult.Data.([]map[string]interface{}); ok {
					if len(usageData) > 0 {

						usageInsertQuery := `INSERT INTO enterprise_coal_consumption_usage (
							obj_id, fk_id, stat_date, create_time, main_usage, specific_usage, input_variety, input_unit,
							input_quantity, output_energy_types, output_quantity, measurement_unit, remarks, row_no, is_confirm, is_check
						) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

						for _, usageRow := range usageData {
							_, err := tx.Exec(usageInsertQuery,
								usageRow["obj_id"], row["obj_id"], usageRow["stat_date"], usageRow["create_time"], usageRow["main_usage"], usageRow["specific_usage"],
								usageRow["input_variety"], usageRow["input_unit"], usageRow["input_quantity"], usageRow["output_energy_types"],
								usageRow["output_quantity"], usageRow["measurement_unit"], usageRow["remarks"], usageRow["row_no"], usageRow["is_confirm"], usageRow["is_check"])

							if err != nil {
								fmt.Printf("插入主要用途情况表数据失败: %v\n", err)
								errorCount++
							} else {
								successCount++
							}
						}
					}
				}
			}

			// 插入重点耗煤装置情况表数据
			equipQuery := `SELECT * FROM enterprise_coal_consumption_equip WHERE fk_id = ?`
			equipResult, err := sourceDb.Query(equipQuery, row["obj_id"])
			if err == nil && equipResult.Ok && equipResult.Data != nil {
				if equipData, ok := equipResult.Data.([]map[string]interface{}); ok {
					if len(equipData) > 0 {

						equipInsertQuery := `INSERT INTO enterprise_coal_consumption_equip (
							obj_id, fk_id, stat_date, create_time, equip_type, equip_no, total_runtime, design_life,
							energy_efficiency, capacity_unit, capacity, coal_type, annual_coal_consumption, row_no, is_confirm, is_check
						) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

						for _, equipRow := range equipData {
							_, err := tx.Exec(equipInsertQuery,
								equipRow["obj_id"], row["obj_id"], equipRow["stat_date"], equipRow["create_time"], equipRow["equip_type"], equipRow["equip_no"],
								equipRow["total_runtime"], equipRow["design_life"], equipRow["energy_efficiency"], equipRow["capacity_unit"],
								equipRow["capacity"], equipRow["coal_type"], equipRow["annual_coal_consumption"], equipRow["row_no"], equipRow["is_confirm"], equipRow["is_check"])

							if err != nil {
								fmt.Printf("插入重点耗煤装置情况表数据失败: %v\n", err)
								errorCount++
							} else {
								successCount++
							}
						}
					}
				}
			}
		}
	}

	return successCount, errorCount, nil
}

// mergeTable2ConflictDataNew 合并表2冲突数据（新版本）
func (a *App) mergeTable2ConflictDataNew(tx *sql.Tx, conflict ConflictData) (int, int, error) {
	successCount := 0
	errorCount := 0

	// 打开源数据库
	sourceDb, err := db.NewDatabase(conflict.FilePath, DB_PASSWORD)
	if err != nil {
		return 0, 0, fmt.Errorf("打开源数据库失败: %v", err)
	}
	defer sourceDb.Close()

	// 处理每个冲突条件
	for _, condition := range conflict.Conditions {
		// 删除目标表中符合条件的所有记录
		deleteQuery := `DELETE FROM critical_coal_equipment_consumption WHERE credit_code = ? AND stat_date = ?`
		_, err := tx.Exec(deleteQuery, condition.CreditCode, condition.StatDate)
		if err != nil {
			fmt.Printf("删除目标表记录失败 mergeTable2ConflictDataNew1: credit_code=%s, stat_date=%s, error=%v\n", condition.CreditCode, condition.StatDate, err)
			errorCount++
			continue
		}

		// 查询源数据并插入
		query := `SELECT * FROM critical_coal_equipment_consumption WHERE credit_code = ? AND stat_date = ?`
		result, err := sourceDb.Query(query, condition.CreditCode, condition.StatDate)
		if err != nil {
			fmt.Printf("查询源数据失败 mergeTable2ConflictDataNew2: credit_code=%s, stat_date=%s, error=%v\n", condition.CreditCode, condition.StatDate, err)
			errorCount++
			continue
		}

		if !result.Ok || result.Data == nil {
			fmt.Printf("查询源数据失败 mergeTable2ConflictDataNew3: credit_code=%s, stat_date=%s, result=%v\n", condition.CreditCode, condition.StatDate, result)
			continue
		}

		data, ok := result.Data.([]map[string]interface{})
		if !ok {
			continue
		}

		// 插入源数据到目标表
		for _, row := range data {
			// 完全按照源表数据插入，包括obj_id
			insertQuery := `INSERT INTO critical_coal_equipment_consumption (
				obj_id, stat_date, create_time, sg_code, unit_name, credit_code, trade_a, trade_b, trade_c, trade_d,
				province_code, province_name, city_code, city_name, country_code, country_name, unit_addr,
				coal_type, coal_no, usage_time, design_life, enecrgy_efficienct_bmk, capacity_unit, capacity,
				use_info, status, annual_coal_consumption, row_no, create_user, is_confirm, is_check
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

			_, execErr := tx.Exec(insertQuery,
				row["obj_id"], row["stat_date"], row["create_time"], row["sg_code"], row["unit_name"], row["credit_code"],
				row["trade_a"], row["trade_b"], row["trade_c"], row["trade_d"], row["province_code"], row["province_name"],
				row["city_code"], row["city_name"], row["country_code"], row["country_name"], row["unit_addr"],
				row["coal_type"], row["coal_no"], row["usage_time"], row["design_life"], row["enecrgy_efficienct_bmk"],
				row["capacity_unit"], row["capacity"], row["use_info"], row["status"], row["annual_coal_consumption"],
				row["row_no"], row["create_user"], row["is_confirm"], row["is_check"])

			if execErr != nil {
				fmt.Printf("插入目标表记录失败 mergeTable2ConflictDataNew4: obj_id=%s, stat_date=%s, error=%v\n", row["obj_id"], row["stat_date"], execErr)
				errorCount++
			} else {
				successCount++
			}
		}
	}

	return successCount, errorCount, nil
}

// mergeTable3ConflictDataNew 合并表3冲突数据（新版本）
func (a *App) mergeTable3ConflictDataNew(tx *sql.Tx, conflict ConflictData) (int, int, error) {
	successCount := 0
	errorCount := 0

	// 打开源数据库
	sourceDb, err := db.NewDatabase(conflict.FilePath, DB_PASSWORD)
	if err != nil {
		return 0, 0, fmt.Errorf("打开源数据库失败: %v", err)
	}
	defer sourceDb.Close()

	// 处理每个冲突条件
	for _, condition := range conflict.Conditions {
		// 删除目标表中符合条件的所有记录
		deleteQuery := `DELETE FROM fixed_assets_investment_project WHERE project_code = ? AND document_number = ?`
		_, err := tx.Exec(deleteQuery, condition.ProjectCode, condition.DocumentNumber)
		if err != nil {
			fmt.Printf("删除目标表记录失败 mergeTable3ConflictDataNew1: project_code=%s, document_number=%s, error=%v\n", condition.ProjectCode, condition.DocumentNumber, err)
			errorCount++
			continue
		}

		// 查询源数据并插入
		query := `SELECT * FROM fixed_assets_investment_project WHERE project_code = ? AND document_number = ?`
		result, err := sourceDb.Query(query, condition.ProjectCode, condition.DocumentNumber)
		if err != nil {
			fmt.Printf("查询源数据失败 mergeTable3ConflictDataNew2: project_code=%s, document_number=%s, error=%v\n", condition.ProjectCode, condition.DocumentNumber, err)
			errorCount++
			continue
		}

		if !result.Ok || result.Data == nil {
			continue
		}
		data, ok := result.Data.([]map[string]interface{})
		if !ok {
			continue
		}

		// 插入源数据到目标表
		for _, row := range data {
			// 完全按照源表数据插入，包括obj_id
			insertQuery := `INSERT INTO fixed_assets_investment_project (
				obj_id, stat_date, sg_code, project_name, project_code, construction_unit, main_construction_content,
				unit_id, province_name, city_name, country_name, trade_a, trade_c, examination_approval_time,
				scheduled_time, actual_time, examination_authority, document_number, equivalent_value, equivalent_cost,
				pq_total_coal_consumption, pq_coal_consumption, pq_coke_consumption, pq_blue_coke_consumption,
				sce_total_coal_consumption, sce_coal_consumption, sce_coke_consumption, sce_blue_coke_consumption,
				is_substitution, substitution_source, substitution_quantity, pq_annual_coal_quantity, sce_annual_coal_quantity,
				create_time, create_user, is_confirm, is_check
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

			_, execErr := tx.Exec(insertQuery,
				row["obj_id"], row["stat_date"], row["sg_code"], row["project_name"], row["project_code"], row["construction_unit"], row["main_construction_content"],
				row["unit_id"], row["province_name"], row["city_name"], row["country_name"], row["trade_a"], row["trade_c"], row["examination_approval_time"],
				row["scheduled_time"], row["actual_time"], row["examination_authority"], row["document_number"], row["equivalent_value"], row["equivalent_cost"],
				row["pq_total_coal_consumption"], row["pq_coal_consumption"], row["pq_coke_consumption"], row["pq_blue_coke_consumption"],
				row["sce_total_coal_consumption"], row["sce_coal_consumption"], row["sce_coke_consumption"], row["sce_blue_coke_consumption"],
				row["is_substitution"], row["substitution_source"], row["substitution_quantity"], row["pq_annual_coal_quantity"], row["sce_annual_coal_quantity"],
				row["create_time"], row["create_user"], row["is_confirm"], row["is_check"])

			if execErr != nil {
				fmt.Printf("插入目标表记录失败 mergeTable3ConflictDataNew3: obj_id=%s, stat_date=%s, error=%v\n", row["obj_id"], row["stat_date"], execErr)
				errorCount++
			} else {
				successCount++
			}
		}
	}

	return successCount, errorCount, nil
}

// mergeAttachment2ConflictDataNew 合并附件2冲突数据（新版本）
func (a *App) mergeAttachment2ConflictDataNew(tx *sql.Tx, conflict ConflictData) (int, int, error) {
	successCount := 0
	errorCount := 0

	// 打开源数据库
	sourceDb, err := db.NewDatabase(conflict.FilePath, DB_PASSWORD)
	if err != nil {
		return 0, 0, fmt.Errorf("打开源数据库失败: %v", err)
	}
	defer sourceDb.Close()

	// 处理每个冲突条件
	for _, condition := range conflict.Conditions {
		// 删除目标表中符合条件的所有记录
		deleteQuery := `DELETE FROM coal_consumption_report WHERE province_name = ? AND city_name = ? AND country_name = ? AND stat_date = ?`
		_, err := tx.Exec(deleteQuery, condition.ProvinceName, condition.CityName, condition.CountryName, condition.StatDate)
		if err != nil {
			fmt.Printf("删除目标表记录失败 mergeAttachment2ConflictDataNew1: province_name=%s, city_name=%s, country_name=%s, stat_date=%s, error=%v\n",
				condition.ProvinceName, condition.CityName, condition.CountryName, condition.StatDate, err)
			errorCount++
			continue
		}

		// 查询源数据并插入
		query := `SELECT * FROM coal_consumption_report WHERE province_name = ? AND city_name = ? AND country_name = ? AND stat_date = ?`
		result, err := sourceDb.Query(query, condition.ProvinceName, condition.CityName, condition.CountryName, condition.StatDate)
		if err != nil {
			fmt.Printf("查询源数据失败 mergeAttachment2ConflictDataNew2: province_name=%s, city_name=%s, country_name=%s, stat_date=%s, error=%v\n",
				condition.ProvinceName, condition.CityName, condition.CountryName, condition.StatDate, err)
			errorCount++
			continue
		}

		if !result.Ok || result.Data == nil {
			continue
		}

		data, ok := result.Data.([]map[string]interface{})
		if !ok {
			continue
		}

		// 插入源数据到目标表
		for _, row := range data {
			// 完全按照源表数据插入，包括obj_id
			insertQuery := `INSERT INTO coal_consumption_report (
				obj_id, stat_date, sg_code, unit_id, unit_name, unit_level, province_name, city_name, country_name,
				total_coal, raw_coal, washed_coal, other_coal, power_generation, heating, coal_washing, coking,
				oil_refining, gas_production, industry, raw_materials, other_uses, coke, create_user, create_time, is_confirm, is_check
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

			_, execErr := tx.Exec(insertQuery,
				row["obj_id"], row["stat_date"], row["sg_code"], row["unit_id"], row["unit_name"], row["unit_level"],
				row["province_name"], row["city_name"], row["country_name"], row["total_coal"], row["raw_coal"],
				row["washed_coal"], row["other_coal"], row["power_generation"], row["heating"], row["coal_washing"],
				row["coking"], row["oil_refining"], row["gas_production"], row["industry"], row["raw_materials"],
				row["other_uses"], row["coke"], row["create_user"], row["create_time"], row["is_confirm"], row["is_check"])

			if execErr != nil {
				fmt.Printf("插入目标表记录失败 mergeAttachment2ConflictDataNew3: obj_id=%s, stat_date=%s, error=%v\n", row["obj_id"], row["stat_date"], execErr)
				errorCount++
			} else {
				successCount++
			}
		}
	}

	return successCount, errorCount, nil
}

// DBTranformExcel 数据库转换为Excel
func (a *App) DBTranformExcel(dbPath string) db.QueryResult {
	result := db.QueryResult{}

	// 打开源数据库
	sourceDb, err := db.NewDatabase(dbPath, DB_PASSWORD)
	if err != nil {
		result.Message = "打开源数据库失败: " + err.Error()
		return result
	}
	defer sourceDb.Close()

	// 查询附表1数据
	query := `SELECT stat_date, credit_code, unit_name, province_name, city_name, country_name,
	annual_energy_equivalent_value, annual_energy_equivalent_cost, annual_total_coal_consumption AS annual_coal_consumption
	 FROM enterprise_coal_consumption_main`
	table1Result, err := sourceDb.Query(query)
	if err != nil {
		result.Message = "查询数据失败: " + err.Error()
		return result
	}
	if !table1Result.Ok || table1Result.Data == nil {
		result.Message = "查询数据失败: " + err.Error()
		return result
	}

	dataMap := make(map[string][]map[string]interface{})
	for _, row := range table1Result.Data.([]map[string]interface{}) {
		statDate := getStringValue(row["stat_date"])
		creditCode := getStringValue(row["credit_code"])
		province_name := getStringValue(row["province_name"])
		city_name := getStringValue(row["city_name"])
		country_name := getStringValue(row["country_name"])
		unit_name := getStringValue(row["unit_name"])
		annual_energy_equivalent_value := getStringValue(row["annual_energy_equivalent_value"])
		annual_energy_equivalent_cost := getStringValue(row["annual_energy_equivalent_cost"])
		annual_coal_consumption := getStringValue(row["annual_coal_consumption"])
		key := fmt.Sprintf("%s_%s_%s_%s_%s", statDate, creditCode, province_name, city_name, country_name)

		annual_energy_equivalent_value, _ = SM4Decrypt(annual_energy_equivalent_value)
		annual_energy_equivalent_cost, _ = SM4Decrypt(annual_energy_equivalent_cost)
		annual_coal_consumption, _ = SM4Decrypt(annual_coal_consumption)

		row := map[string]interface{}{
			"stat_date":                      statDate,
			"credit_code":                    creditCode,
			"province_name":                  province_name,
			"city_name":                      city_name,
			"country_name":                   country_name,
			"unit_name":                      unit_name,
			"annual_energy_equivalent_value": annual_energy_equivalent_value,
			"annual_energy_equivalent_cost":  annual_energy_equivalent_cost,
			"annual_coal_consumption":        annual_coal_consumption,
			"unit_type":                      "规上企业",
		}

		if rows, ok := dataMap[key]; ok {
			rows = append(rows, row)
			dataMap[key] = rows
		} else {
			dataMap[key] = []map[string]interface{}{row}
		}
	}


	// 查询附表1 设备数据
	equipQuery := `SELECT b.stat_date, b.credit_code, b.unit_name, b.province_name, b.city_name, b.country_name, a.equip_type, a.equip_no, a.coal_type, a.annual_coal_consumption
		FROM enterprise_coal_consumption_equip a LEFT JOIN enterprise_coal_consumption_main b ON a.fk_id = b.obj_id`
	table1EquipResult, err := sourceDb.Query(equipQuery)
	if err != nil {
		result.Message = "查询数据失败: " + err.Error()
		return result
	}
	if !table1EquipResult.Ok || table1EquipResult.Data == nil {
		result.Message = "查询数据失败: " + err.Error()
		return result
	}

	// 装置清单列表
	equipList := make([]map[string]interface{}, 0)

	for _, row := range table1EquipResult.Data.([]map[string]interface{}) {
		statDate := getStringValue(row["stat_date"])
		creditCode := getStringValue(row["credit_code"])
		province_name := getStringValue(row["province_name"])
		city_name := getStringValue(row["city_name"])
		country_name := getStringValue(row["country_name"])
		unit_name := getStringValue(row["unit_name"])
		equip_type := getStringValue(row["equip_type"])
		equip_no := getStringValue(row["equip_no"])
		coal_type := getStringValue(row["coal_type"])
		annual_coal_consumption := getStringValue(row["annual_coal_consumption"])

		equipRow := map[string]interface{}{
			"stat_date":               statDate,
			"credit_code":             creditCode,
			"province_name":           province_name,
			"city_name":               city_name,
			"country_name":            country_name,
			"unit_name":               unit_name,
			"coal_type":               coal_type,
			"equip_type":              equip_type,
			"equip_no":                equip_no,
			"unit_type":               "规上企业",
		}

        if	annual_coal_consumption != "" {
            annual_coal_consumption_value, _ := SM4Decrypt(annual_coal_consumption)
            annual_coal_consumptionValue := parseFloat(getStringValue(annual_coal_consumption_value))
            equipRow["annual_coal_consumption"] = fmt.Sprintf("%.4f", annual_coal_consumptionValue/ 10000)
        }

		equipList = append(equipList, equipRow)
	}

	table2DataMap := make(map[string]map[string]interface{})

	// 查询附表2数据
	table2Query := `SELECT stat_date, credit_code, unit_name, province_name, city_name, country_name,
	annual_coal_consumption, coal_type AS equip_type, coal_no AS equip_no
	FROM critical_coal_equipment_consumption`
	table2Result, err := sourceDb.Query(table2Query)
	if err != nil {
		result.Message = "查询数据失败: " + err.Error()
		return result
	}
	if !table2Result.Ok || table2Result.Data == nil {
		result.Message = "查询数据失败: " + err.Error()
		return result
	}

	for _, row := range table2Result.Data.([]map[string]interface{}) {
		statDate := getStringValue(row["stat_date"])
		creditCode := getStringValue(row["credit_code"])
		province_name := getStringValue(row["province_name"])
		city_name := getStringValue(row["city_name"])
		country_name := getStringValue(row["country_name"])
		unit_name := getStringValue(row["unit_name"])
		annual_coal_consumption := getStringValue(row["annual_coal_consumption"])
		equip_type := getStringValue(row["equip_type"])
		equip_no := getStringValue(row["equip_no"])

		var annual_coal_consumptionValue float64
		if	annual_coal_consumption != "" {
			annual_coal_consumption, _ = SM4Decrypt(annual_coal_consumption)
			annual_coal_consumptionValue = parseFloat(getStringValue(annual_coal_consumption))
		}

		key := fmt.Sprintf("%s_%s_%s_%s_%s", statDate, creditCode, province_name, city_name, country_name)

		

		if existingItem, ok := table2DataMap[key]; ok {
			if	annual_coal_consumption != "" {
				itemAnnualTotal, _ := existingItem["annual_coal_consumption"].(float64)
				existingItem["annual_coal_consumption"] = addFloat64(itemAnnualTotal, annual_coal_consumptionValue)
			}
		} else {
			item := map[string]interface{}{
				"stat_date":               statDate,
				"credit_code":             creditCode,
				"province_name":           province_name,
				"city_name":               city_name,
				"country_name":            country_name,
				"unit_name":               unit_name,
				"equip_type":              equip_type,
				"equip_no":                equip_no,
				"annual_coal_consumption": annual_coal_consumptionValue,
				"unit_type":               "其他用能单位",
			}
			table2DataMap[key] = item
		}

		equipItem := map[string]interface{}{
			"stat_date":               statDate,
			"credit_code":             creditCode,
			"province_name":           province_name,
			"city_name":               city_name,
			"country_name":            country_name,
			"unit_name":               unit_name,
			"equip_type":              equip_type,
			"equip_no":                equip_no,
			"annual_coal_consumption": "",
			"unit_type":               "其他用能单位",
		}
		if	annual_coal_consumption != "" {
			equipItem["annual_coal_consumption"] = fmt.Sprintf("%.4f", annual_coal_consumptionValue/ 10000)
		}

		equipList = append(equipList, equipItem)
	}

	for key, item := range table2DataMap {
		itemAnnualTotal, _ := item["annual_coal_consumption"].(float64)
		// 单位万吨, 保留两位小数
		item["annual_coal_consumption"] = fmt.Sprintf("%.4f", itemAnnualTotal/ 10000)
		if rows, ok := dataMap[key]; ok {
			rows = append(rows, item)
			dataMap[key] = rows
		} else {
			dataMap[key] = []map[string]interface{}{item}
		}
	}

	sourceDb.Close()

	// 导出Excel文件
	excelResult := a.ExportDataToExcel(dataMap, equipList, dbPath)
	if !excelResult.Ok {
		log.Printf("导出Excel失败: %s", excelResult.Message)
		// Excel导出失败不影响合并结果，只记录日志
	} else {
		log.Printf("Excel导出成功: %s", excelResult.Data)
	}

	return excelResult
}

// ExportDataToExcel 导出数据到单个Excel文件，包含两个sheet页
func (a *App) ExportDataToExcel(dataMap map[string][]map[string]interface{}, equipList []map[string]interface{}, dbPath string) db.QueryResult {
	result := db.QueryResult{}

	// 创建新的Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("关闭Excel文件失败: %v", err)
		}
	}()

	// 1. 创建耗煤单位数据汇总表sheet
	unitSheetResult := a.createUnitDataSheet(f, dataMap)
	if !unitSheetResult.Ok {
		return unitSheetResult
	}

	// 2. 创建耗煤装置数据汇总表sheet
	equipSheetResult := a.createEquipDataSheet(f, equipList)
	if !equipSheetResult.Ok {
		return equipSheetResult
	}

	// 3. 根据dbPath生成Excel文件名
	dbDir := filepath.Dir(dbPath)
	dbBaseName := filepath.Base(dbPath)
	dbBaseName = strings.TrimSuffix(dbBaseName, filepath.Ext(dbBaseName))
	fileName := dbBaseName + ".xlsx"
	filePath := filepath.Join(dbDir, fileName)
	fmt.Printf("filePath: %s", filePath)
	fmt.Printf("fileName: %s", fileName)

	if err := f.SaveAs(filePath); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "used by another process") {
			result.Message = "保存Excel文件失败: 文件已被其他程序占用，请关闭文件后重试"
			return result
		}
		result.Message = "保存Excel文件失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Data = map[string]interface{}{
		"outputPath": filePath,
		"fileName":   fileName,
	}
	result.Message = "转换成功"
	return result
}

// createUnitDataSheet 在Excel文件中创建耗煤单位数据汇总表sheet
func (a *App) createUnitDataSheet(f *excelize.File, dataMap map[string][]map[string]interface{}) db.QueryResult {
	result := db.QueryResult{}

	// 设置工作表名称
	sheetName := "耗煤单位数据汇总表"
	f.SetSheetName("Sheet1", sheetName)

	// 设置大标题
	f.SetCellValue(sheetName, "A1", "耗煤单位数据汇总表")
	f.MergeCell(sheetName, "A1", "J1")

	// 设置列标题
	headers := []string{"年份", "省", "市", "县", "单位名称", "统一社会信用代码", "单位类别", "综合能耗量（当量值，万吨标准煤）", "综合能耗量（等价值，万吨标准煤）", "耗煤总量（实物量，万吨）"}

	// 设置列标题样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6E6FA"},
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
	if err != nil {
		result.Message = "创建样式失败: " + err.Error()
		return result
	}

	// 设置大标题样式
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		result.Message = "创建标题样式失败: " + err.Error()
		return result
	}

	// 应用标题样式
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	// 写入列标题
	for i, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheetName, cellName, header)
		f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
	}

	// 创建数据行样式（带黑色边框）
	dataStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		log.Printf("创建数据行样式失败: %v", err)
	}

	// 写入数据
	rowIndex := 3
	for _, rows := range dataMap {
		for _, row := range rows {
			// 年份
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), row["stat_date"])
			// 省
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), row["province_name"])
			// 市
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), row["city_name"])
			// 县
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), row["country_name"])
			// 单位名称
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIndex), row["unit_name"])
			// 统一社会信用代码
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIndex), row["credit_code"])
			// 单位类别
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowIndex), row["unit_type"])
			// 综合能耗量（当量值，万吨标准煤）
			if energyValue, ok := row["annual_energy_equivalent_value"]; ok {
				f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowIndex), energyValue)
			}
			// 综合能耗量（等价值，万吨标准煤）
			if energyCost, ok := row["annual_energy_equivalent_cost"]; ok {
				f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowIndex), energyCost)
			}
			// 耗煤总量（实物量，万吨）
			if coalTotal, ok := row["annual_coal_consumption"]; ok {
				f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowIndex), coalTotal)
			}

			// 为整行应用边框样式
			if dataStyle != 0 {
				for col := 1; col <= 10; col++ {
					cellName, _ := excelize.CoordinatesToCellName(col, rowIndex)
					f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
				}
			}

			rowIndex++
		}
	}

	// 添加注释行
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), "其他用能单位的耗煤总量由各装置年耗煤量加总得到，不一定与其整体耗煤量相等")
	f.MergeCell(sheetName, fmt.Sprintf("A%d", rowIndex), fmt.Sprintf("K%d", rowIndex))

	// 设置注释行样式
	commentStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Italic: true,
			Size:   10,
			Color:  "808080",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
	})
	if err == nil {
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowIndex), fmt.Sprintf("A%d", rowIndex), commentStyle)
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 12) // 年份
	f.SetColWidth(sheetName, "B", "B", 15) // 省
	f.SetColWidth(sheetName, "C", "C", 15) // 市
	f.SetColWidth(sheetName, "D", "D", 15) // 县
	f.SetColWidth(sheetName, "E", "E", 25) // 单位名称
	f.SetColWidth(sheetName, "F", "F", 25) // 统一社会信用代码
	f.SetColWidth(sheetName, "G", "G", 15) // 单位类别
	f.SetColWidth(sheetName, "H", "H", 25) // 综合能耗量（当量值）
	f.SetColWidth(sheetName, "I", "I", 25) // 综合能耗量（等价值）
	f.SetColWidth(sheetName, "J", "J", 30) // 耗煤总量

	result.Ok = true
	result.Message = "耗煤单位数据汇总表sheet创建成功"
	return result
}

// createEquipDataSheet 在Excel文件中创建耗煤装置数据汇总表sheet
func (a *App) createEquipDataSheet(f *excelize.File, equipList []map[string]interface{}) db.QueryResult {
	result := db.QueryResult{}

	// 创建新的工作表
	sheetName := "耗煤装置数据汇总表"
	f.NewSheet(sheetName)

	// 设置大标题
	f.SetCellValue(sheetName, "A1", "耗煤装置数据汇总表")
	f.MergeCell(sheetName, "A1", "K1")

	// 设置列标题
	headers := []string{"年份", "省", "市", "县", "单位名称", "统一社会信用代码", "单位类别", "装置类型", "装置编号", "耗煤品种", "年耗煤量（实物量，万吨）"}

	// 设置列标题样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6E6FA"},
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
	if err != nil {
		result.Message = "创建样式失败: " + err.Error()
		return result
	}

	// 设置大标题样式
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		result.Message = "创建标题样式失败: " + err.Error()
		return result
	}

	// 应用标题样式
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)

	// 写入列标题
	for i, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheetName, cellName, header)
		f.SetCellStyle(sheetName, cellName, cellName, headerStyle)
	}

	// 创建数据行样式（带黑色边框）
	dataStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		log.Printf("创建数据行样式失败: %v", err)
	}

	// 写入数据
	rowIndex := 3
	for _, row := range equipList {
		// 年份
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowIndex), row["stat_date"])
		// 省
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowIndex), row["province_name"])
		// 市
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowIndex), row["city_name"])
		// 县
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowIndex), row["country_name"])
		// 单位名称
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowIndex), row["unit_name"])
		// 统一社会信用代码
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowIndex), row["credit_code"])
		// 单位类别
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowIndex), row["unit_type"])
		// 装置类型
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowIndex), row["equip_type"])
		// 装置编号
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowIndex), row["equip_no"])
		// 耗煤品种
		if coal_type, ok := row["coal_type"]; ok {
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowIndex), coal_type)
		}

		// 耗煤总量（实物量，万吨）
		if coalTotal, ok := row["annual_coal_consumption"]; ok {
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", rowIndex), coalTotal)
		}
		// 为整行应用边框样式
		if dataStyle != 0 {
			for col := 1; col <= 11; col++ {
				cellName, _ := excelize.CoordinatesToCellName(col, rowIndex)
				f.SetCellStyle(sheetName, cellName, cellName, dataStyle)
			}
		}

		rowIndex++
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 12) // 年份
	f.SetColWidth(sheetName, "B", "B", 15) // 省
	f.SetColWidth(sheetName, "C", "C", 15) // 市
	f.SetColWidth(sheetName, "D", "D", 15) // 县
	f.SetColWidth(sheetName, "E", "E", 25) // 单位名称
	f.SetColWidth(sheetName, "F", "F", 25) // 统一社会信用代码
	f.SetColWidth(sheetName, "G", "G", 15) // 单位类别
	f.SetColWidth(sheetName, "H", "H", 20) // 装置类型
	f.SetColWidth(sheetName, "I", "I", 20) // 装置编号
	f.SetColWidth(sheetName, "J", "J", 15) // 耗煤品种
	f.SetColWidth(sheetName, "K", "K", 30) // 年耗煤量

	result.Ok = true
	result.Message = "耗煤装置数据汇总表sheet创建成功"
	return result
}
