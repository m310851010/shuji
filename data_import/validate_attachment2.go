package data_import

import (
	"fmt"
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

	// 查找表格的开始位置（第6行是表头）
	startRow := 5 // 从第6行开始（0索引为5）
	if startRow >= len(rows) {
		return nil, fmt.Errorf("表格行数不足")
	}

	// 获取表头
	headers := rows[startRow]

	// 期望的表头
	expectedHeaders := []string{
		"省（市、区）", "地市（州）", "县（区）", "年份", "分品种煤炭消费摸底", "分用途煤炭消费摸底", "焦炭消费摸底",
	}

	// 检查表头一致性
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("表头列数不足，期望%d列，实际%d列", len(expectedHeaders), len(headers))
	}

	// 构建表头映射
	headerMap := make(map[int]string)
	for i, expected := range expectedHeaders {
		if i >= len(headers) {
			return nil, fmt.Errorf("缺少表头：%s", expected)
		}

		actual := strings.TrimSpace(headers[i])
		if actual != expected {
			return nil, fmt.Errorf("第%d列表头不匹配，期望：%s，实际：%s", i+1, expected, actual)
		}
		headerMap[i] = s.mapAttachment2HeaderToField(expected)
	}

	// 解析数据行（从第7行开始，跳过表头下的第一行）
	for i := startRow + 2; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || (len(row) > 0 && strings.TrimSpace(row[0]) == "") {
			continue // 跳过空行
		}

		// 构建数据行
		dataRow := make(map[string]interface{})
		for j, cell := range row {
			if fieldName, exists := headerMap[j]; exists && fieldName != "" {
				cleanedValue := s.cleanCellValue(cell)
				if strings.Contains(fieldName, "consumption") || strings.Contains(fieldName, "value") || strings.Contains(fieldName, "cost") {
					// 数值字段
					dataRow[fieldName] = s.parseNumericValue(cleanedValue)
				} else {
					// 文本字段
					dataRow[fieldName] = cleanedValue
				}
			}
		}

		// 只添加有省份的数据行
		if province, ok := dataRow["province_name"].(string); ok && province != "" {
			mainData = append(mainData, dataRow)
		}
	}

	return mainData, nil
}

// mapAttachment2HeaderToField 映射附件2表头到字段名
func (s *DataImportService) mapAttachment2HeaderToField(header string) string {
	header = strings.TrimSpace(header)

	// 字段映射
	fieldMap := map[string]string{
		"省（市、区）":    "province_name",
		"地市（州）":     "city_name",
		"县（区）":      "country_name",
		"年份":        "stat_date",
		"分品种煤炭消费摸底": "total_coal",
		"分用途煤炭消费摸底": "total_coal",
		"焦炭消费摸底":    "coke",
	}

	return fieldMap[header]
}

// ValidateAttachment2File 校验附件2文件
func (s *DataImportService) ValidateAttachment2File(filePath string) db.QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		// 插入导入记录
		s.insertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("读取Excel文件失败: %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 2. 解析Excel文件
	mainData, err := s.parseAttachment2Excel(f)
	if err != nil {
		// 插入导入记录
		s.insertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("解析Excel文件失败: %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateAttachment2Data(mainData)
	if len(validationErrors) > 0 {
		// 插入导入记录
		s.insertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
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
