package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// ValidateEnterpriseListFile 校验企业清单文件并返回数据
func (a *App) ValidateEnterpriseListFile(filePath string) QueryResult {
	result := QueryResult{
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

	ok, err := a.checkEnterpriseList()
	if err != nil {
		result.Message = "查询企业清单失败: " + err.Error()
		return result
	}
	result.Data = ok
	result.Ok = true
	return result
}

// ValidateKeyEquipmentListFile 校验装置清单文件
func (a *App) ValidateKeyEquipmentListFile(filePath string) QueryResult {
	result := QueryResult{
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

	ok, err := a.checkKeyEquipmentList()
	if err != nil {
		result.Message = "查询装置清单失败: " + err.Error()
		return result
	}
	result.Data = ok
	result.Ok = true
	return result
}

// ImportEnterpriseList 导入企业清单
func (a *App) ImportEnterpriseList(filePath string) QueryResult {

	result := QueryResult{
		Ok:      false,
		Data:    nil,
		Message: "",
	}

	fileName := "文件:" + filepath.Base(filePath) + " "

	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		result.Message = fileName + "无效的Excel文件: " + err.Error()
		return result
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		result.Message = fileName + "读取Excel数据失败: " + err.Error()
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

	newFileName, err := CopyCacheFile(filePath)
	if err != nil {
		result.Message = "复制缓存文件失败: " + err.Error()
		return result
	}

	// 记录导入历史
	record := &DataImportRecord{
		FileName:    newFileName,
		FileType:    "企业清单",
		ImportState: "上传成功",
		Describe:    fmt.Sprintf("成功导入%d条记录", count),
		CreateUser:  a.GetCurrentOSUser(),
	}

	importRecordResult := a.InsertImportRecord(record)
	if !importRecordResult.Ok {
		log.Printf("记录导入日志失败: %s", importRecordResult.Message)
	}

	result.Ok = true
	result.Message = fmt.Sprintf(fileName+"企业清单导入完成：成功%d条", count)
	return result
}

// ImportKeyEquipmentList 导入装置清单
func (a *App) ImportKeyEquipmentList(filePath string) QueryResult {
	var result QueryResult
	fileName := "文件:" + filepath.Base(filePath) + " "

	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		result.Message = fileName + "无效的Excel文件: " + err.Error()
		return result
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		result.Message = fileName + "读取Excel数据失败: " + err.Error()
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

	newFileName, err := CopyCacheFile(filePath)
	if err != nil {
		result.Message = "复制缓存文件失败: " + err.Error()
		return result
	}

	// 记录导入历史
	record := &DataImportRecord{
		FileName:    newFileName,
		FileType:    "装置清单",
		ImportState: "上传成功",
		Describe:    fmt.Sprintf("成功导入%d条记录", count),
		CreateUser:  a.GetCurrentOSUser(),
	}

	importRecordResult := a.InsertImportRecord(record)
	if !importRecordResult.Ok {
		log.Printf("记录导入日志失败: %s", importRecordResult.Message)
	}

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

// 检查企业清单是否存在
func (a *App) checkEnterpriseList() (bool, error) {
	rows, err := a.db.QueryRow("SELECT COUNT(1) as count FROM enterprise_list")
	if err != nil {
		return false, fmt.Errorf("查询企业清单失败: %v", err)
	}

	return rows.Data.(map[string]interface{})["count"].(int64) > 0, nil
}

// 检查装置清单是否存在
func (a *App) checkKeyEquipmentList() (bool, error) {
	rows, err := a.db.QueryRow("SELECT COUNT(1) as count FROM key_equipment_list")
	if err != nil {
		return false, fmt.Errorf("查询装置清单失败: %v", err)
	}

	return rows.Data.(map[string]interface{})["count"].(int64) > 0, nil
}
