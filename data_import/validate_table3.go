package data_import

import (
	"fmt"
	"path/filepath"
	"shuji/db"
	"strings"

	"github.com/xuri/excelize/v2"
)

// parseTable3Excel 解析附表3Excel文件
func (s *DataImportService) parseTable3Excel(f *excelize.File) ([]map[string]interface{}, error) {
	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	// 解析主表数据
	mainData, err := s.parseTable3MainSheet(f, sheets[0])
	if err != nil {
		return nil, fmt.Errorf("解析主表数据失败: %v", err)
	}

	return mainData, nil
}

// parseTable3MainSheet 解析附表3主表数据
func (s *DataImportService) parseTable3MainSheet(f *excelize.File, sheetName string) ([]map[string]interface{}, error) {
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
		"序号", "项目名称", "项目代码", "建设单位", "主要建设内容", "项目所在省、自治区、直辖市",
		"项目所在地市", "项目所在区县", "所属行业大类（2位代码）", "所属行业小类（4位代码）",
		"节能审查批复时间", "拟投产时间", "实际投产时间", "节能审查机关", "审查意见文号",
		"当量值", "等价值", "煤品消费总量", "煤炭消费量", "焦炭消费量", "兰炭消费量",
		"煤品消费总量", "煤炭消费量", "焦炭消费量", "兰炭消费量", "是否煤炭消费替代",
		"煤炭消费替代来源", "煤炭消费替代量（万吨，实物量）", "年原料用煤量（万吨，实物量）", "年原料用煤量（万吨标准煤，折标量）",
	}

	// 检查表头一致性
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("表头列数不足，期望%d列，实际%d列", len(expectedHeaders), len(headers))
	}

	// 构建表头映射（基于位置）
	headerMap := make(map[int]string)
	for i, expected := range expectedHeaders {
		if i >= len(headers) {
			return nil, fmt.Errorf("缺少表头：%s", expected)
		}

		actual := strings.TrimSpace(headers[i])
		if actual != expected {
			return nil, fmt.Errorf("第%d列表头不匹配，期望：%s，实际：%s", i+1, expected, actual)
		}
		headerMap[i] = s.mapTable3HeaderToFieldByPosition(expected, i)
	}

	// 解析数据行（从第7行开始，表头下第一行作为数据）
	for i := startRow + 1; i < len(rows); i++ {
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
				} else if strings.Contains(fieldName, "date") || strings.Contains(fieldName, "time") {
					// 日期字段
					dataRow[fieldName] = s.parseDateValue(cleanedValue)
				} else {
					// 文本字段
					dataRow[fieldName] = cleanedValue
				}
			}
		}

		// 只添加有项目名称的数据行
		if projectName, ok := dataRow["project_name"].(string); ok && projectName != "" {
			mainData = append(mainData, dataRow)
		}
	}

	return mainData, nil
}

// mapTable3HeaderToFieldByPosition 基于位置映射附表3表头到字段名
func (s *DataImportService) mapTable3HeaderToFieldByPosition(header string, position int) string {
	header = strings.TrimSpace(header)

	// 基础字段映射
	baseFieldMap := map[string]string{
		"序号":     "sequence_no",
		"项目名称":   "project_name",
		"项目代码":   "project_code",
		"建设单位":   "construction_unit",
		"主要建设内容": "main_construction_content",
		"项目所在省、自治区、直辖市":     "province_name",
		"项目所在地市":            "city_name",
		"项目所在区县":            "country_name",
		"所属行业大类（2位代码）":      "trade_a",
		"所属行业小类（4位代码）":      "trade_c",
		"节能审查批复时间":          "examination_approval_time",
		"拟投产时间":             "scheduled_time",
		"实际投产时间":            "actual_time",
		"节能审查机关":            "examination_authority",
		"审查意见文号":            "document_number",
		"当量值":               "equivalent_value",
		"等价值":               "equivalent_cost",
		"是否煤炭消费替代":          "is_substitution",
		"煤炭消费替代来源":          "substitution_source",
		"煤炭消费替代量（万吨，实物量）":   "substitution_quantity",
		"年原料用煤量（万吨，实物量）":    "pq_annual_coal_quantity",
		"年原料用煤量（万吨标准煤，折标量）": "sce_annual_coal_quantity",
	}

	// 检查是否是基础字段
	if fieldName, exists := baseFieldMap[header]; exists {
		return fieldName
	}

	// 处理重复字段（基于位置）
	switch header {
	case "煤品消费总量":
		if position == 17 { // 第一个煤品消费总量（实物量）
			return "pq_total_coal_consumption"
		} else if position == 22 { // 第二个煤品消费总量（折标量）
			return "sce_total_coal_consumption"
		}
	case "煤炭消费量":
		if position == 18 { // 第一个煤炭消费量（实物量）
			return "pq_coal_consumption"
		} else if position == 23 { // 第二个煤炭消费量（折标量）
			return "sce_coal_consumption"
		}
	case "焦炭消费量":
		if position == 19 { // 第一个焦炭消费量（实物量）
			return "pq_coke_consumption"
		} else if position == 24 { // 第二个焦炭消费量（折标量）
			return "sce_coke_consumption"
		}
	case "兰炭消费量":
		if position == 20 { // 第一个兰炭消费量（实物量）
			return "pq_blue_coke_consumption"
		} else if position == 25 { // 第二个兰炭消费量（折标量）
			return "sce_blue_coke_consumption"
		}
	}

	return ""
}

// ValidateTable3File 校验附表3文件
func (s *DataImportService) ValidateTable3File(filePath string) db.QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		// 插入导入记录
		s.insertImportRecord(fileName, "附表3", "上传失败", fmt.Sprintf("读取Excel文件失败: %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 2. 解析Excel文件
	mainData, err := s.parseTable3Excel(f)
	if err != nil {
		// 插入导入记录
		s.insertImportRecord(fileName, "附表3", "上传失败", fmt.Sprintf("解析Excel文件失败: %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateTable3Data(mainData)
	if len(validationErrors) > 0 {
		// 插入导入记录
		s.insertImportRecord(fileName, "附表3", "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	// 4. 查询该表是否有数据放进data属性
	hasData := s.checkTable3HasData()

	// 5. 返回QueryResult
	return db.QueryResult{
		Ok:      true,
		Data:    hasData,
		Message: "校验通过",
	}
}

// validateTable3Data 校验附表3数据
func (s *DataImportService) validateTable3Data(mainData []map[string]interface{}) []string {
	errors := []string{}

	// 1. 检查项目名称和项目代码是否为空
	for i, data := range mainData {
		fieldErrors := s.validateRequiredFields(data, Table3RequiredFields, i)
		errors = append(errors, fieldErrors...)
	}

	// 2. 检查区域与当前单位是否相符
	regionErrors := s.validateTable3Region(mainData)
	errors = append(errors, regionErrors...)

	// 3. 检查固定资产投资项目重复数据
	duplicateErrors := s.validateTable3DuplicateData(mainData)
	errors = append(errors, duplicateErrors...)

	return errors
}

// validateTable3Region 检查附表3区域与当前单位是否相符
func (s *DataImportService) validateTable3Region(data []map[string]interface{}) []string {
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

// validateTable3DuplicateData 检查附表3重复数据
func (s *DataImportService) validateTable3DuplicateData(data []map[string]interface{}) []string {
	errors := []string{}

	// 用于存储已检查的项目信息
	projectMap := make(map[string]int)

	for i, row := range data {
		projectName := s.getStringValue(row["project_name"])
		projectCode := s.getStringValue(row["project_code"])
		approvalNumber := s.getStringValue(row["document_number"])

		// 生成唯一标识
		key := fmt.Sprintf("%s|%s|%s", projectName, projectCode, approvalNumber)

		if existingIndex, exists := projectMap[key]; exists {
			errors = append(errors, fmt.Sprintf("第%d行：[项目名称、项目代码、审查意见文号]数据重复（与第%d行重复）", i+1, existingIndex+1))
		} else {
			projectMap[key] = i
		}
	}

	return errors
}

// checkTable3HasData 检查附表3相关表是否有数据
func (s *DataImportService) checkTable3HasData() bool {
	return s.checkTableHasData(TableFixedAssetsInvestmentProject)
}
