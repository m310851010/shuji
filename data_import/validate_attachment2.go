package data_import

import (
	"fmt"
	"log"
	"path/filepath"
	"shuji/db"
	"strings"

	"github.com/xuri/excelize/v2"
)

// parseAttachment2Excel 解析附件2Excel文件
func (s *DataImportService) parseAttachment2Excel(f *excelize.File) ([]map[string]interface{}, error) {
	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	// 解析主表数据
	mainData, err := s.parseAttachment2MainSheet(f, sheets[0])
	if err != nil {
		return nil, fmt.Errorf("解析主表数据失败: %v", err)
	}

	return mainData, nil
}

// parseAttachment2MainSheet 解析附件2主表数据
func (s *DataImportService) parseAttachment2MainSheet(f *excelize.File, sheetName string) ([]map[string]interface{}, error) {
	var mainData []map[string]interface{}

	// 读取表格数据
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	// 解析制表单位（第3行）
	var reportUnit string
	if len(rows) >= 3 {
		row3 := rows[2] // 第3行（0索引为2）
		if len(row3) >= 3 {
			reportUnit = s.cleanCellValue(row3[2]) // 制表单位在第3列
		}
	}

	// 查找表格的开始位置（第4行是表头）
	startRow := 3 // 从第4行开始（0索引为3）
	if startRow >= len(rows) {
		return nil, fmt.Errorf("表格行数不足")
	}

	// 获取表头（第4行）
	headers := rows[startRow]

	// 期望的表头（第4行的主要表头，包含合并单元格）
	expectedHeaders := []string{
		"省（市、区）", "地市（州）", "县（区）", "年份", "分品种煤炭消费摸底", "", "", "", "分用途煤炭消费摸底", "", "", "", "", "", "", "", "", "焦炭消费摸底",
	}

	expectedHeadersCount := 18
	// 检查表头一致性
	if len(headers) < expectedHeadersCount {
		return nil, fmt.Errorf("表头列数不足，模板要求%d列，实际%d列", expectedHeadersCount, len(headers))
	}

	// 构建表头映射（基于位置）
	headerMap := make(map[int]string)
	for i, expected := range expectedHeaders {
		if i >= len(headers) {
			return nil, fmt.Errorf("缺少表头：%s", expected)
		}

		actual := strings.TrimSpace(headers[i])
		if actual != expected {
			return nil, fmt.Errorf("第%d列表头不匹配，模板要求：%s，实际：%s", i+1, expected, actual)
		}
		headerMap[i] = s.mapAttachment2HeaderToFieldByPosition(expected, i)
	}

	// 解析数据行（从第8行开始，跳过表头、子表头等）
	for i := startRow + 4; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 2 || (len(row) > 0 && strings.TrimSpace(row[0]) == "") {
			continue // 跳过空行
		}

		// 构建数据行
		dataRow := make(map[string]interface{})

		// 添加制表单位信息
		if reportUnit != "" {
			dataRow["report_unit"] = reportUnit
		}

		for j, cell := range row {
			if fieldName, exists := headerMap[j]; exists && fieldName != "" {
				cleanedValue := s.cleanCellValue(cell)
				if strings.Contains(fieldName, "consumption") || strings.Contains(fieldName, "coal") || strings.Contains(fieldName, "coke") {
					// 数值字段
					dataRow[fieldName] = s.parseNumericValue(cleanedValue)
				} else {
					// 文本字段
					dataRow[fieldName] = cleanedValue
				}
			}
		}

		// 只添加有省份名称的数据行
		if provinceName, ok := dataRow["province_name"].(string); ok && provinceName != "" {
			mainData = append(mainData, dataRow)
		}
	}

	return mainData, nil
}

// mapAttachment2HeaderToFieldByPosition 基于位置映射附件2表头到字段名
func (s *DataImportService) mapAttachment2HeaderToFieldByPosition(header string, position int) string {
	header = strings.TrimSpace(header)

	// 基础字段映射
	baseFieldMap := map[string]string{
		"省（市、区）": "province_name",
		"地市（州）":  "city_name",
		"县（区）":   "country_name",
		"年份":     "stat_date",
		"焦炭消费摸底": "coke_consumption",
	}

	// 检查是否是基础字段
	if fieldName, exists := baseFieldMap[header]; exists {
		return fieldName
	}

	// 基于位置的字段映射
	positionFieldMap := map[int]string{
		4:  "total_coal",       // 第5列：煤合计
		5:  "raw_coal",         // 第6列：原煤
		6:  "washed_coal",      // 第7列：洗精煤
		7:  "other_coal",       // 第8列：其他
		8:  "power_generation", // 第9列：1.火力发电
		9:  "heating",          // 第10列：2.供热
		10: "coal_washing",     // 第11列：3.煤炭洗选
		11: "coking",           // 第12列：4.炼焦
		12: "oil_refining",     // 第13列：5.炼油及煤制油
		13: "gas_production",   // 第14列：6.制气
		14: "industrial",       // 第15列：1.工业
		15: "raw_material",     // 第16列：#用作原料、材料
		16: "other_use",        // 第17列：2.其他用途
	}

	// 检查基于位置的字段映射
	if fieldName, exists := positionFieldMap[position]; exists {
		return fieldName
	}

	return ""
}

// ValidateAttachment2File 校验附件2文件
func (s *DataImportService) ValidateAttachment2File(filePath string) db.QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		// 插入导入记录
		s.app.InsertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("读取Excel文件失败: %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 2. 解析Excel文件
	mainData, err := s.parseAttachment2Excel(f)
	log.Println("mainData", mainData)
	if err != nil {
		// 插入导入记录
		s.app.InsertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("解析Excel文件失败: %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateAttachment2Data(mainData)
	if len(validationErrors) > 0 {
		// 插入导入记录
		s.app.InsertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	// 4. 查询该表是否有数据放进data属性
	hasData := s.checkAttachment2HasData()

	// 5. 返回QueryResult
	return db.QueryResult{
		Ok:      true,
		Data:    hasData,
		Message: "校验通过",
	}
}

// validateAttachment2Data 校验附件2数据
func (s *DataImportService) validateAttachment2Data(mainData []map[string]interface{}) []string {
	errors := []string{}

	// 1. 检查省份和年份是否为空
	for i, data := range mainData {
		fieldErrors := s.validateRequiredFields(data, Attachment2RequiredFields, i)
		errors = append(errors, fieldErrors...)
	}

	// 2. 检查区域与当前单位是否相符
	regionErrors := s.validateAttachment2Region(mainData)
	errors = append(errors, regionErrors...)

	return errors
}

// validateAttachment2Region 检查附件2区域与当前单位是否相符
func (s *DataImportService) validateAttachment2Region(data []map[string]interface{}) []string {
	errors := []string{}

	// 获取当前单位信息
	result := s.GetAreaConfig()
	if !result.Ok {
		// 如果获取失败，跳过区域校验
		return errors
	}

	// 解析返回的数据
	var currentProvince, currentCity, currentCountry string
	if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
		row := data[0]
		currentProvince = s.getStringValue(row["province_name"])
		currentCity = s.getStringValue(row["city_name"])
		currentCountry = s.getStringValue(row["country_name"])
	} else {
		// 如果没有配置，跳过区域校验
		return errors
	}

	for i, row := range data {
		provinceName := s.getStringValue(row["province_name"])
		cityName := s.getStringValue(row["city_name"])
		countryName := s.getStringValue(row["country_name"])

		// 检查区域是否与当前单位相符
		if provinceName != "" && currentProvince != "" && provinceName != currentProvince {
			errors = append(errors, fmt.Sprintf("第%d行：上传的数据单位与当前单位不符", i+1))
			continue
		}

		if cityName != "" && currentCity != "" && cityName != currentCity {
			errors = append(errors, fmt.Sprintf("第%d行：上传的数据单位与当前单位不符", i+1))
			continue
		}

		if countryName != "" && currentCountry != "" && countryName != currentCountry {
			errors = append(errors, fmt.Sprintf("第%d行：上传的数据单位与当前单位不符", i+1))
			continue
		}
	}

	return errors
}

// checkAttachment2HasData 检查附件2相关表是否有数据
func (s *DataImportService) checkAttachment2HasData() bool {
	return s.checkTableHasData(TableCoalConsumptionReport)
}
