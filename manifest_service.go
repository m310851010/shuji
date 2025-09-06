package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"shuji/db"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// ValidateEnterpriseListFile 校验企业清单文件并返回数据
func (a *App) ValidateEnterpriseListFile(filePath string) db.QueryResult {
	// 使用包装函数来处理异常
	return a.validateEnterpriseListFileWithRecover(filePath)
}

// validateEnterpriseListFileWithRecover 带异常处理的校验企业清单文件函数
func (a *App) validateEnterpriseListFileWithRecover(filePath string) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ValidateEnterpriseListFile 发生异常: %v", r)
		}
	}()

	result := db.QueryResult{
		Ok:      false,
		Data:    nil,
		Message: "",
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		result.Message = "文件" + filePath + "不存在"
		return result
	}

	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		result.Message = "无效的Excel文件: " + err.Error()
		return result
	}
	defer f.Close()

	fileName := "文件:" + filepath.Base(filePath) + " "

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		result.Message = fileName + "读取Excel数据失败: " + err.Error()
		return result
	}

	// 检查表头
	if len(rows) < 2 {
		result.Message = fileName + "数据不足，至少需要表头和一行数据"
		return result
	}

	// 验证表头
	headers := rows[0]
	expectedHeaders := []string{"省(自治区、直辖市)", "地(区、市、州、盟)", "县(区、市、旗)", "单位详细名称", "统一社会信用代码"}
	if !validateHeaders(headers, expectedHeaders) {
		result.Message = fileName + "表头与模板不匹配，请使用正确的企业清单模板"
		return result
	}

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 5 {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行数据不完整"
			return result
		}

		// 检查必填字段
		if strings.TrimSpace(row[0]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：省份名称不能为空"
			return result
		}
		if strings.TrimSpace(row[1]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：地区名称不能为空"
			return result
		}
		if strings.TrimSpace(row[3]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：单位详细名称不能为空"
			return result
		}
		if strings.TrimSpace(row[4]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：统一社会信用代码不能为空"
			return result
		}
	}

	isEnterpriseListExist = false
	ok, err := a.IsEnterpriseListExist()
	if err != nil {
		result.Message = "查询企业清单失败: " + err.Error()
		return result
	}
	result.Data = ok
	result.Ok = true
	return result
}

// ValidateKeyEquipmentListFile 校验装置清单文件
func (a *App) ValidateKeyEquipmentListFile(filePath string) db.QueryResult {
	// 使用包装函数来处理异常
	return a.validateKeyEquipmentListFileWithRecover(filePath)
}

// validateKeyEquipmentListFileWithRecover 带异常处理的校验装置清单文件函数
func (a *App) validateKeyEquipmentListFileWithRecover(filePath string) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ValidateKeyEquipmentListFile 发生异常: %v", r)
		}
	}()

	result := db.QueryResult{
		Ok:      false,
		Data:    nil,
		Message: "",
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		result.Message = "文件" + filePath + "不存在"
		return result
	}

	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		result.Message = "无效的Excel文件: " + err.Error()
		return result
	}
	defer f.Close()

	fileName := "文件:" + filepath.Base(filePath) + " "

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		result.Message = fileName + "读取Excel数据失败: " + err.Error()
		return result
	}

	// 检查表头
	if len(rows) < 2 {
		result.Message = fileName + "数据不足，至少需要表头和一行数据"
		return result
	}

	// 验证表头
	headers := rows[0]
	expectedHeaders := []string{"省(自治区、直辖市)", "地(区、市、州、盟)", "县(区、市、旗)", "使用单位名称", "使用单位统一社会信用代码", "设备类型", "设备型号", "设备编号"}
	if !validateHeaders(headers, expectedHeaders) {
		result.Message = fileName + "表头与模板不匹配，请使用正确的装置清单模板"
		return result
	}

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 5 {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行数据不完整"
			return result
		}

		// 检查必填字段
		if strings.TrimSpace(row[0]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：省份名称不能为空"
			return result
		}
		if strings.TrimSpace(row[1]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：地区名称不能为空"
			return result
		}
		if strings.TrimSpace(row[3]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：使用单位名称不能为空"
			return result
		}
		if strings.TrimSpace(row[4]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：使用单位统一社会信用代码不能为空"
			return result
		}
		if strings.TrimSpace(row[5]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：设备类型不能为空"
			return result
		}
		if strings.TrimSpace(row[6]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：设备型号不能为空"
			return result
		}
		if strings.TrimSpace(row[7]) == "" {
			result.Message = fileName + "第" + strconv.Itoa(i+1) + "行：设备编号不能为空"
			return result
		}
	}

	isEnterpriseListExist = false
	ok, err := a.IsEquipmentListExist()
	if err != nil {
		result.Message = "查询装置清单失败: " + err.Error()
		return result
	}
	result.Data = ok
	result.Ok = true
	return result
}

// ImportEnterpriseList 导入企业清单
func (a *App) ImportEnterpriseList(filePath string) db.QueryResult {
	// 使用包装函数来处理异常
	return a.importEnterpriseListWithRecover(filePath)
}

// importEnterpriseListWithRecover 带异常处理的导入企业清单函数
func (a *App) importEnterpriseListWithRecover(filePath string) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ImportEnterpriseList 发生异常: %v", r)
		}
	}()

	result := db.QueryResult{
		Ok:      false,
		Data:    nil,
		Message: "",
	}

	fileName := filepath.Base(filePath)
	fileNameTip := "文件:" + fileName + " "

	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		result.Message = fileNameTip + "无效的Excel文件: " + err.Error()
		return result
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		result.Message = fileNameTip + "读取Excel数据失败: " + err.Error()
		return result
	}

	// 开始事务
	tx, err := a.db.Begin()
	if err != nil {
		result.Message = "开始数据库事务失败: " + err.Error()
		return result
	}

	defer func() {
		if result.Ok {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// 全量数据导入：先清空表数据
	_, err = tx.Exec("DELETE FROM enterprise_list")
	if err != nil {
		result.Message = "清空企业清单数据失败: " + err.Error()
		return result
	}


	provinceName := ""
	cityName := ""
	countryName := ""

	areaConfigResult := a.GetAreaConfig()
	if areaConfigResult.Ok {
		areaConfigData, ok := areaConfigResult.Data.(map[string]interface{})
		if ok {
			provinceName = getStringValue(areaConfigData["province_name"])
			cityName = getStringValue(areaConfigData["city_name"])
			countryName = getStringValue(areaConfigData["country_name"])
		}
	}
	
	countyNameMap := make(map[string]bool, 0)

	if countryName == "" {
		targetLocation, _, _, err := a.getCurrentUserLocationData()
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
		if targetLocationMap, ok := targetLocation.(map[string]interface{}); ok {
			if children, exists := targetLocationMap["children"]; exists && children != nil {
				if childrenList, ok := children.([]interface{}); ok {
					for _, county := range childrenList {
						if countyMap, ok := county.(map[string]interface{}); ok {
							if name, exists := countyMap["name"]; exists && name != nil {
								countyNameMap[fmt.Sprintf("%v", name)] = true
							}
						}
					}
				}
			}
		}
	} else {
		countyNameMap[countryName] = true
	}


	count := 0

	// 处理数据行
	for i := 1; i < len(rows); i++ {
		row := rows[i]

		data := EnterpriseListData{
			ProvinceName: strings.TrimSpace(row[0]),
			CityName:     strings.TrimSpace(row[1]),
			CountryName:  strings.TrimSpace(row[2]),
			UnitName:     strings.TrimSpace(row[3]),
			CreditCode:   strings.TrimSpace(row[4]),
		}

		if data.ProvinceName != provinceName {
			result.Message = fmt.Sprintf("第%d行：企业清单区域和当前区域不一致", i+2)
			return result
		}
		if data.CityName != cityName {
			result.Message = fmt.Sprintf("第%d行：企业清单区域和当前区域不一致", i+2)
			return result
		}
		if !countyNameMap[data.CountryName] {
			result.Message = fmt.Sprintf("第%d行：企业清单区域和当前区域不一致", i+2)
			return result
		}

		// 全量数据导入：直接插入数据
		objID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO enterprise_list (obj_id, province_name, city_name, country_name, unit_name, credit_code)
			VALUES (?, ?, ?, ?, ?, ?)
		`, objID, data.ProvinceName, data.CityName, data.CountryName, data.UnitName, data.CreditCode)

		if err != nil {
			result.Message = "插入数据失败: " + err.Error()
			return result
		}
		count++
	}

	// 记录导入历史
	a.InsertImportRecord(fileName, "企业清单", "导入成功", fmt.Sprintf("成功导入%d条记录", count))

	result.Ok = true
	result.Message = fmt.Sprintf(fileNameTip+"企业清单导入完成：成功%d条", count)
	return result
}

// ImportKeyEquipmentList 导入装置清单
func (a *App) ImportKeyEquipmentList(filePath string) db.QueryResult {
	// 使用包装函数来处理异常
	return a.importKeyEquipmentListWithRecover(filePath)
}

// importKeyEquipmentListWithRecover 带异常处理的导入装置清单函数
func (a *App) importKeyEquipmentListWithRecover(filePath string) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ImportKeyEquipmentList 发生异常: %v", r)
		}
	}()

	var result db.QueryResult
	fileName := filepath.Base(filePath)
	fileNameTip := "文件:" + fileName + " "

	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		result.Message = fileNameTip + "无效的Excel文件: " + err.Error()
		return result
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		result.Message = fileNameTip + "读取Excel数据失败: " + err.Error()
		return result
	}

	// 开始事务
	tx, err := a.db.Begin()
	if err != nil {
		result.Message = "开始数据库事务失败: " + err.Error()
		return result
	}

	defer func() {
		if result.Ok {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// 全量数据导入：先清空表数据
	_, err = tx.Exec("DELETE FROM key_equipment_list")
	if err != nil {
		result.Message = "清空装置清单数据失败: " + err.Error()
		return result
	}


	provinceName := ""
	cityName := ""
	countryName := ""

	areaConfigResult := a.GetAreaConfig()
	if areaConfigResult.Ok {
		areaConfigData, ok := areaConfigResult.Data.(map[string]interface{})
		if ok {
			provinceName = getStringValue(areaConfigData["province_name"])
			cityName = getStringValue(areaConfigData["city_name"])
			countryName = getStringValue(areaConfigData["country_name"])
		}
	}
	
	countyNameMap := make(map[string]bool, 0)

	if countryName == "" {
		targetLocation, _, _, err := a.getCurrentUserLocationData()
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

		if targetLocationMap, ok := targetLocation.(map[string]interface{}); ok {
			if children, exists := targetLocationMap["children"]; exists && children != nil {
				if childrenList, ok := children.([]interface{}); ok {
					for _, county := range childrenList {
						if countyMap, ok := county.(map[string]interface{}); ok {
							if name, exists := countyMap["name"]; exists && name != nil {
								countyNameMap[fmt.Sprintf("%v", name)] = true
							}
						}
					}
				}
			}
		}
	} else {
		countyNameMap[countryName] = true
	}


	count := 0

	// 处理数据行
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		data := KeyEquipmentListData{
			ProvinceName:     strings.TrimSpace(row[0]),
			CityName:         strings.TrimSpace(row[1]),
			CountryName:      strings.TrimSpace(row[2]),
			UnitName:         strings.TrimSpace(row[3]),
			CreditCode:       strings.TrimSpace(row[4]),
			EquipType:        strings.TrimSpace(row[5]),
			EquipModelNumber: strings.TrimSpace(row[6]),
			EquipNo:          strings.TrimSpace(row[7]),
		}

		if data.ProvinceName != provinceName {
			result.Message = fmt.Sprintf("第%d行：装置清单区域和当前区域不一致", i+2)
			return result
		}
		if data.CityName != cityName {
			result.Message = fmt.Sprintf("第%d行：装置清单区域和当前区域不一致", i+2)
			return result
		}
		if !countyNameMap[data.CountryName] {
			result.Message = fmt.Sprintf("第%d行：装置清单区域和当前区域不一致", i+2)
			return result
		}
		

		// 全量数据导入：直接插入数据
		objID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO key_equipment_list (obj_id, province_name, city_name, country_name, unit_name, credit_code, equip_type, equip_model_number, equip_no)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, objID, data.ProvinceName, data.CityName, data.CountryName, data.UnitName, data.CreditCode, data.EquipType, data.EquipModelNumber, data.EquipNo)

		if err != nil {
			result.Message = "插入数据失败: " + err.Error()
			return result
		}
		count++
	}

	// 记录导入历史
	a.InsertImportRecord(fileName, "装置清单", "导入成功", fmt.Sprintf("成功导入%d条记录", count))

	result.Ok = true
	result.Message = fmt.Sprintf("装置清单导入完成：成功%d条", count)
	return result
}

// validateHeaders 验证表头
func validateHeaders(headers []string, expectedHeaders []string) bool {
	if len(headers) < len(expectedHeaders) {
		return false
	}

	for i, expected := range expectedHeaders {
		if i >= len(headers) || strings.TrimSpace(headers[i]) != expected {
			return false
		}
	}
	return true
}

// 缓存企业清单是否存在
var isEnterpriseListExist = false

// 检查企业清单是否存在
func (a *App) IsEnterpriseListExist() (bool, error) {
	if isEnterpriseListExist {
		return isEnterpriseListExist, nil
	}

	rows, err := a.db.QueryRow("SELECT COUNT(1) as count FROM enterprise_list")
	if err != nil {
		return false, fmt.Errorf("查询企业清单失败: %v", err)
	}

	isEnterpriseListExist = rows.Data.(map[string]interface{})["count"].(int64) > 0
	return isEnterpriseListExist, nil
}

// GetEnterpriseInfoByCreditCode 获取企业信息
func (a *App) GetEnterpriseInfoByCreditCode(creditCode string) db.QueryResult {
	rows, err := a.db.QueryRow("SELECT province_name, city_name, country_name, unit_name FROM enterprise_list WHERE credit_code = ?", creditCode)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Data:    nil,
			Message: "查询企业信息失败: " + err.Error(),
		}
	}

	return db.QueryResult{
		Ok:      true,
		Data:    rows.Data,
		Message: "查询企业信息成功",
	}
}

// 缓存装置清单是否存在
var isEquipmentListExist = false

// 检查装置清单是否存在
func (a *App) IsEquipmentListExist() (bool, error) {
	if isEquipmentListExist {
		return isEquipmentListExist, nil
	}
	rows, err := a.db.QueryRow("SELECT COUNT(1) as count FROM key_equipment_list")
	if err != nil {
		return false, fmt.Errorf("查询装置清单失败: %v", err)
	}

	isEquipmentListExist = rows.Data.(map[string]interface{})["count"].(int64) > 0
	return isEquipmentListExist, nil
}

// 获取装置清单信息By信用代码
func (a *App) GetEquipmentByCreditCode(creditCode string) db.QueryResult {
	rows, err := a.db.QueryRow("SELECT province_name, city_name, country_name, unit_name FROM key_equipment_list WHERE credit_code = ?", creditCode)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Data:    nil,
			Message: "查询装置清单失败: " + err.Error(),
		}
	}

	return db.QueryResult{
		Ok:      true,
		Data:    rows.Data,
		Message: "查询装置清单成功",
	}
}

// GetStateManifest 读取state.json文件中的manifest值
func (a *App) GetStateManifest() db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetStateManifest 发生异常: %v", r)
		}
	}()

	result := db.QueryResult{
		Ok:      false,
		Data:    nil,
		Message: "",
	}

	// 使用 getCachePath("") 获取缓存路径，然后在该路径下存放 state.json
	cachePath := a.GetCachePath("")
	cacheStateFilePath := filepath.Join(cachePath, "state.json")
	// 获取public目录下的state.json文件路径（作为模板）
	publicStateFilePath := filepath.Join("frontend", "public", "state.json")

	// 检查缓存目录下的state.json文件是否存在
	if _, err := os.Stat(cacheStateFilePath); os.IsNotExist(err) {
		// 如果缓存目录下的文件不存在，检查public目录下的模板文件是否存在
		if _, err := os.Stat(publicStateFilePath); os.IsNotExist(err) {
			result.Message = "模板文件" + publicStateFilePath + "不存在"
			return result
		}

		// 确保缓存目录存在
		cacheDir := filepath.Dir(cacheStateFilePath)
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			result.Message = "创建缓存目录失败: " + err.Error()
			return result
		}

		// 从public目录复制state.json到缓存目录
		publicContent, err := os.ReadFile(publicStateFilePath)
		if err != nil {
			result.Message = "读取模板文件失败: " + err.Error()
			return result
		}

		err = os.WriteFile(cacheStateFilePath, publicContent, 0644)
		if err != nil {
			result.Message = "复制state.json文件到缓存目录失败: " + err.Error()
			return result
		}
	}

	// 读取缓存目录下的state.json文件内容
	fileContent, err := os.ReadFile(cacheStateFilePath)
	if err != nil {
		result.Message = "读取缓存目录下的state.json文件失败: " + err.Error()
		return result
	}

	// 解析JSON
	var stateData map[string]interface{}
	err = json.Unmarshal(fileContent, &stateData)
	if err != nil {
		result.Message = "解析state.json文件失败: " + err.Error()
		return result
	}

	// 获取manifest值
	manifestValue := stateData["manifest"]

	result.Ok = true
	result.Data = manifestValue
	result.Message = "获取manifest值成功"
	return result
}

// UpdateStateManifest 更新state.json文件中的manifest值
func (a *App) UpdateStateManifest(manifestValue interface{}) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("UpdateStateManifest 发生异常: %v", r)
		}
	}()

	result := db.QueryResult{
		Ok:      false,
		Data:    nil,
		Message: "",
	}

	// 使用 getCachePath("") 获取缓存路径，然后在该路径下存放 state.json
	cachePath := a.GetCachePath("")
	cacheStateFilePath := filepath.Join(cachePath, "state.json")
	// 获取public目录下的state.json文件路径（作为模板）
	publicStateFilePath := filepath.Join("frontend", "public", "state.json")

	// 检查缓存目录下的state.json文件是否存在
	if _, err := os.Stat(cacheStateFilePath); os.IsNotExist(err) {
		// 如果缓存目录下的文件不存在，检查public目录下的模板文件是否存在
		if _, err := os.Stat(publicStateFilePath); os.IsNotExist(err) {
			// 如果模板文件也不存在，创建包含默认值的新文件
			stateData := map[string]interface{}{
				"manifest": manifestValue,
			}

			// 确保缓存目录存在
			cacheDir := filepath.Dir(cacheStateFilePath)
			if err := os.MkdirAll(cacheDir, 0755); err != nil {
				result.Message = "创建缓存目录失败: " + err.Error()
				return result
			}

			// 将数据转换为JSON
			jsonData, err := json.MarshalIndent(stateData, "", "    ")
			if err != nil {
				result.Message = "创建state.json文件失败: " + err.Error()
				return result
			}

			// 写入文件
			err = os.WriteFile(cacheStateFilePath, jsonData, 0644)
			if err != nil {
				result.Message = "写入state.json文件失败: " + err.Error()
				return result
			}

			result.Ok = true
			result.Data = manifestValue
			result.Message = "创建并更新manifest值成功"
			return result
		}

		// 确保缓存目录存在
		cacheDir := filepath.Dir(cacheStateFilePath)
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			result.Message = "创建缓存目录失败: " + err.Error()
			return result
		}

		// 从public目录复制state.json到缓存目录
		publicContent, err := os.ReadFile(publicStateFilePath)
		if err != nil {
			result.Message = "读取模板文件失败: " + err.Error()
			return result
		}

		// 解析模板文件内容
		var stateData map[string]interface{}
		err = json.Unmarshal(publicContent, &stateData)
		if err != nil {
			result.Message = "解析模板文件失败: " + err.Error()
			return result
		}

		// 更新manifest值
		stateData["manifest"] = manifestValue

		// 将更新后的数据转换为JSON
		jsonData, err := json.MarshalIndent(stateData, "", "    ")
		if err != nil {
			result.Message = "转换JSON数据失败: " + err.Error()
			return result
		}

		// 写入缓存目录下的文件
		err = os.WriteFile(cacheStateFilePath, jsonData, 0644)
		if err != nil {
			result.Message = "复制并更新state.json文件到缓存目录失败: " + err.Error()
			return result
		}

		result.Ok = true
		result.Data = manifestValue
		result.Message = "复制并更新manifest值成功"
		return result
	}

	// 读取缓存目录下的现有文件内容
	fileContent, err := os.ReadFile(cacheStateFilePath)
	if err != nil {
		result.Message = "读取缓存目录下的state.json文件失败: " + err.Error()
		return result
	}

	// 解析JSON
	var stateData map[string]interface{}
	err = json.Unmarshal(fileContent, &stateData)
	if err != nil {
		result.Message = "解析state.json文件失败: " + err.Error()
		return result
	}

	// 更新manifest值
	stateData["manifest"] = manifestValue

	// 将更新后的数据转换为JSON
	jsonData, err := json.MarshalIndent(stateData, "", "    ")
	if err != nil {
		result.Message = "转换JSON数据失败: " + err.Error()
		return result
	}

	// 写入缓存目录下的文件
	err = os.WriteFile(cacheStateFilePath, jsonData, 0644)
	if err != nil {
		result.Message = "写入缓存目录下的state.json文件失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Data = manifestValue
	result.Message = "更新manifest值成功"
	return result
}
