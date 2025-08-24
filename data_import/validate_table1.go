package data_import

import (
	"fmt"
	"path/filepath"
	"shuji/db"
	"strings"

	"os"

	"github.com/xuri/excelize/v2"
)

// parseTable1Excel 解析附表1Excel文件
func (s *DataImportService) parseTable1Excel(f *excelize.File) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, nil, fmt.Errorf("Excel文件没有工作表")
	}

	// 解析主表数据（企业基本信息）
	mainData, err := s.parseTable1MainSheet(f, sheets[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("和%s模板不匹配, 主表: %v", TableName1, err)
	}

	// 解析用途数据（煤炭消费主要用途情况）
	usageData, err := s.parseTable1UsageSheet(f, sheets[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("和%s模板不匹配, 用途数据: %v", TableName1, err)
	}

	// 解析设备数据（重点耗煤装置情况）
	equipData, err := s.parseTable1EquipSheet(f, sheets[0])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("和%s模板不匹配, 设备数据: %v", TableName1, err)
	}

	return mainData, usageData, equipData, nil
}

// parseTable1MainSheet 解析附表1主表数据
func (s *DataImportService) parseTable1MainSheet(f *excelize.File, sheetName string) ([]map[string]interface{}, error) {
	var mainData []map[string]interface{}

	// 读取企业基本信息表格
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	// 查找企业基本信息表格的开始位置（第5行是表头）
	startRow := 4 // 从第5行开始
	if startRow >= len(rows) {
		return nil, fmt.Errorf("表格行数不足")
	}

	// 获取表头
	headers := rows[startRow]

	// 企业基本信息表格表头
	expectedHeaders := []string{
		"年份", "单位名称", "统一社会信用代码", "行业门类", "行业大类", "行业中类",
		"单位所在省/市/区", "单位所在地市", "单位所在区县", "联系电话",
	}

	// 检查表头一致性
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("企业基本信息表头列数不足，模板需要%d列，上传文件%d列", len(expectedHeaders), len(headers))
	}

	// 构建表头映射
	headerMap := make(map[int]string)
	for i, expected := range expectedHeaders {
		if i >= len(headers) {
			return nil, fmt.Errorf("缺少表头：%s", expected)
		}

		actual := strings.TrimSpace(headers[i])
		if actual != expected {
			return nil, fmt.Errorf("第%d列表头不匹配，模板需要：%s，上传文件：%s", i+1, expected, actual)
		}
		headerMap[i] = s.mapTable1HeaderToField(expected)
	}

	// 解析企业基本信息数据行（主表只有一条数据）
	dataRow := make(map[string]interface{})
	if startRow+2 < len(rows) {
		row := rows[startRow+2]
		// 记录实际Excel行号（第7行）
		dataRow["_excel_row"] = startRow + 3 // 0索引转换为1索引

		for j, cell := range row {
			if fieldName, exists := headerMap[j]; exists && fieldName != "" {
				cleanedValue := s.cleanCellValue(cell)
				dataRow[fieldName] = cleanedValue
			}
		}
	}

	// 查找综合能源消费情况和煤炭消费情况表格的开始位置（它们在同一行）
	energyCoalStartRow := -1
	for i := startRow + 3; i < len(rows); i++ {
		row := rows[i]
		if len(row) > 0 && (strings.Contains(row[0], "综合能源消费情况") || strings.Contains(row[0], "煤炭消费情况")) {
			energyCoalStartRow = i + 1 // 表头在下一行
			break
		}
	}

	if energyCoalStartRow != -1 && energyCoalStartRow < len(rows) {
		// 获取表头
		headers := rows[energyCoalStartRow]

		// 期望的表头（综合能源消费情况和煤炭消费情况在同一行）
		expectedHeaders := []string{
			"年综合能耗当量值（万吨标准煤，含原料用能）", "年综合能耗等价值（万吨标准煤，含原料用能）", "年原料用能消费量（万吨标准煤）",
			"耗煤总量（实物量，万吨）", "耗煤总量（标准量，万吨标准煤）", "原料用煤（实物量，万吨）",
			"原煤消费（实物量，万吨）", "洗精煤消费（实物量，万吨）", "其他煤炭消费（实物量，万吨）", "焦炭消费（实物量，万吨）",
		}

		// 检查表头一致性
		if len(headers) >= len(expectedHeaders) {
			// 构建表头映射
			headerMap := make(map[int]string)
			for i, expected := range expectedHeaders {
				if i < len(headers) {
					actual := strings.TrimSpace(headers[i])
					if actual == expected {
						headerMap[i] = s.mapTable1HeaderToField(expected)
					}
				}
			}

			// 解析数据行（第3行是数据行，跳过第1行提示行和第2行说明行）
			if energyCoalStartRow+2 < len(rows) {
				rowData := rows[energyCoalStartRow+2] // 表头下的第3行是数据行
				// 记录第二部分表格的实际Excel行号
				dataRow["_excel_row2"] = energyCoalStartRow + 3 // 0索引转换为1索引

				for j, cell := range rowData {
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
			}
		}
	}

	// 只添加有企业名称的数据行
	if unitName, ok := dataRow["unit_name"].(string); ok && unitName != "" {
		mainData = append(mainData, dataRow)
	}

	return mainData, nil
}

// parseTable1UsageSheet 解析附表1用途数据
func (s *DataImportService) parseTable1UsageSheet(f *excelize.File, sheetName string) ([]map[string]interface{}, error) {
	var usageData []map[string]interface{}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	// 查找煤炭消费主要用途情况表格的开始位置
	startRow := -1
	for i, row := range rows {
		if len(row) > 0 && strings.Contains(row[0], "煤炭消费主要用途情况") {
			startRow = i + 1 // 表头在下一行
			break
		}
	}

	if startRow == -1 {
		return nil, fmt.Errorf("未找到煤炭消费主要用途情况表格")
	}

	// 获取表头
	headers := rows[startRow]

	// 期望的表头
	expectedHeaders := []string{
		"序号", "主要用途", "具体用途", "投入品种", "投入计量单位", "投入量",
		"产出品种品类", "产出计量单位", "产出量", "备注",
	}

	// 检查表头一致性
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("用途表表头列数不足，模板需要%d列，上传文件%d列", len(expectedHeaders), len(headers))
	}

	// 构建表头映射
	headerMap := make(map[int]string)
	for i, expected := range expectedHeaders {
		if i >= len(headers) {
			return nil, fmt.Errorf("缺少表头：%s", expected)
		}

		actual := strings.TrimSpace(headers[i])
		if actual != expected {
			return nil, fmt.Errorf("第%d列表头不匹配，模板需要：%s，上传文件：%s", i+1, expected, actual)
		}
		headerMap[i] = s.mapTable1UsageHeaderToField(expected)
	}

	// 解析数据行（跳过表头下的第一行提示行）
	dataRowIndex := 0
	for i := startRow + 2; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || (len(row) > 0 && strings.TrimSpace(row[0]) == "") {
			continue
		}

		// 检查是否到达下一个表格
		if len(row) > 0 && strings.Contains(row[0], "重点耗煤装置（设备)情况") {
			break
		}

		// 构建数据行
		dataRow := make(map[string]interface{})
		// 记录实际Excel行号
		dataRow["_excel_row"] = i + 1 // 0索引转换为1索引

		for j, cell := range row {
			if fieldName, exists := headerMap[j]; exists && fieldName != "" {
				cleanedValue := s.cleanCellValue(cell)
				if strings.Contains(fieldName, "quantity") {
					// 数值字段
					dataRow[fieldName] = s.parseNumericValue(cleanedValue)
				} else {
					// 文本字段
					dataRow[fieldName] = cleanedValue
				}
			}
		}

		// 只添加有主要用途的数据行
		if mainUsage, ok := dataRow["main_usage"].(string); ok && mainUsage != "" {
			usageData = append(usageData, dataRow)
		}
		dataRowIndex++
	}

	return usageData, nil
}

// parseTable1EquipSheet 解析附表1设备数据
func (s *DataImportService) parseTable1EquipSheet(f *excelize.File, sheetName string) ([]map[string]interface{}, error) {
	var equipData []map[string]interface{}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	// 查找重点耗煤装置情况表格的开始位置
	startRow := -1
	for i, row := range rows {
		if len(row) > 0 && strings.Contains(row[0], "重点耗煤装置（设备)情况") {
			startRow = i + 1 // 表头在下一行
			break
		}
	}

	if startRow == -1 {
		return nil, fmt.Errorf("未找到重点耗煤装置情况表格")
	}

	// 获取表头
	headers := rows[startRow]

	// 期望的表头
	expectedHeaders := []string{
		"序号", "类型", "编号", "累计使用时间", "设计年限", "能效水平",
		"容量单位", "容量", "耗煤品种", "年耗煤量（单位：吨）",
	}

	// 检查表头一致性
	if len(headers) < len(expectedHeaders) {
		return nil, fmt.Errorf("设备表表头列数不足，模板需要%d列，上传文件%d列", len(expectedHeaders), len(headers))
	}

	// 构建表头映射
	headerMap := make(map[int]string)
	for i, expected := range expectedHeaders {
		if i >= len(headers) {
			return nil, fmt.Errorf("缺少表头：%s", expected)
		}

		actual := strings.TrimSpace(headers[i])
		if actual != expected {
			return nil, fmt.Errorf("第%d列表头不匹配，模板需要：%s，上传文件：%s", i+1, expected, actual)
		}
		headerMap[i] = s.mapTable1EquipHeaderToField(expected)
	}

	// 解析数据行（跳过表头下的第一行提示行）
	for i := startRow + 2; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || (len(row) > 0 && strings.TrimSpace(row[0]) == "") {
			continue
		}

		// 构建数据行
		dataRow := make(map[string]interface{})
		// 记录实际Excel行号
		dataRow["_excel_row"] = i + 1 // 0索引转换为1索引

		for j, cell := range row {
			if fieldName, exists := headerMap[j]; exists && fieldName != "" {
				cleanedValue := s.cleanCellValue(cell)
				if strings.Contains(fieldName, "consumption") || strings.Contains(fieldName, "capacity") {
					// 数值字段
					dataRow[fieldName] = s.parseNumericValue(cleanedValue)
				} else {
					// 文本字段
					dataRow[fieldName] = cleanedValue
				}
			}
		}

		// 只添加有设备类型的数据行
		if equipType, ok := dataRow["equip_type"].(string); ok && equipType != "" {
			equipData = append(equipData, dataRow)
		}
	}

	return equipData, nil
}

// mapTable1HeaderToField 映射附表1主表表头到字段名
func (s *DataImportService) mapTable1HeaderToField(header string) string {
	header = strings.TrimSpace(header)

	// 主表字段映射
	fieldMap := map[string]string{
		"年份":        "stat_date",
		"单位名称":      "unit_name",
		"统一社会信用代码":  "credit_code",
		"行业门类":      "trade_a",
		"行业大类":      "trade_b",
		"行业中类":      "trade_c",
		"单位所在省/市/区": "province_name",
		"单位所在地市":    "city_name",
		"单位所在区县":    "country_name",
		"联系电话":      "tel",
		"年综合能耗当量值（万吨标准煤，含原料用能）": "annual_energy_equivalent_value",
		"年综合能耗等价值（万吨标准煤，含原料用能）": "annual_energy_equivalent_cost",
		"年原料用能消费量（万吨标准煤）":       "annual_raw_material_energy",
		"耗煤总量（实物量，万吨）":          "annual_total_coal_consumption",
		"耗煤总量（标准量，万吨标准煤）":       "annual_total_coal_products",
		"原料用煤（实物量，万吨）":          "annual_raw_coal",
		"原煤消费（实物量，万吨）":          "annual_raw_coal_consumption",
		"洗精煤消费（实物量，万吨）":         "annual_clean_coal_consumption",
		"其他煤炭消费（实物量，万吨）":        "annual_other_coal_consumption",
		"焦炭消费（实物量，万吨）":          "annual_coke_consumption",
	}

	return fieldMap[header]
}

// mapTable1UsageHeaderToField 映射附表1用途表表头到字段名
func (s *DataImportService) mapTable1UsageHeaderToField(header string) string {
	header = strings.TrimSpace(header)

	// 用途表字段映射
	fieldMap := map[string]string{
		"序号":     "sequence_no",
		"主要用途":   "main_usage",
		"具体用途":   "specific_usage",
		"投入品种":   "input_variety",
		"投入计量单位": "input_unit",
		"投入量":    "input_quantity",
		"产出品种品类": "output_energy_types",
		"产出计量单位": "measurement_unit",
		"产出量":    "output_quantity",
		"备注":     "remarks",
	}

	return fieldMap[header]
}

// mapTable1EquipHeaderToField 映射附表1设备表表头到字段名
func (s *DataImportService) mapTable1EquipHeaderToField(header string) string {
	header = strings.TrimSpace(header)

	// 设备表字段映射
	fieldMap := map[string]string{
		"序号":         "sequence_no",
		"类型":         "equip_type",
		"编号":         "equip_no",
		"累计使用时间":     "total_runtime",
		"设计年限":       "design_life",
		"能效水平":       "energy_efficiency",
		"容量单位":       "capacity_unit",
		"容量":         "capacity",
		"耗煤品种":       "coal_type",
		"年耗煤量（单位：吨）": "annual_coal_consumption",
	}

	return fieldMap[header]
}

// ValidateTable1File 校验附表1文件
func (s *DataImportService) ValidateTable1File(filePath string, isCover bool) db.QueryResult {
	fileName := filepath.Base(filePath)

	// 第一步: 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errorMessage := fmt.Sprintf("文件不存在: %v", err)
		fmt.Println(errorMessage)
		s.app.InsertImportRecord(fileName, TableType1, "上传失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    []string{errorMessage},
			Message: errorMessage,
		}
	}

	// 第二步: 文件是否可读取
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		errorMessage := fmt.Sprintf("读取文件失败: %v", err)
		fmt.Println(errorMessage)
		s.app.InsertImportRecord(fileName, TableType1, "上传失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    []string{errorMessage},
			Message: errorMessage,
		}
	}
	defer f.Close()

	// 第三步: 文件是否和模板文件匹配
	mainData, usageData, equipData, err := s.parseTable1Excel(f)
	if err != nil {
		errorMessage := fmt.Sprintf("解析文件失败: %v", err)
		fmt.Println(errorMessage)
		s.app.InsertImportRecord(fileName, TableType1, "上传失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    []string{errorMessage},
			Message: errorMessage,
		}
	}

	if isCover {
		// 第四步: 去缓存目录检查是否有同名的文件, 直接返回,需要前端确认
		cacheResult := s.app.CacheFileExists(TableType1, fileName)
		if cacheResult.Ok {
			// 文件已存在，直接返回，需要前端确认
			return db.QueryResult{
				Ok:      false,
				Message: "文件已存在，需要确认是否覆盖",
				Data:    "FILE_EXISTS",
			}
		}
	}

	// 第五步: 按行读取文件数据并校验
	validationErrors := s.validateTable1DataWithEnterpriseCheck(mainData, usageData, equipData)
	if len(validationErrors) > 0 {
		errorMessage := fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; "))
		fmt.Println(errorMessage)
		s.app.InsertImportRecord(fileName, TableType1, "上传失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    validationErrors,
			Message: errorMessage,
		}
	}

	// 第六步: 复制文件到缓存目录（只有校验通过才复制）
	if len(validationErrors) == 0 {

		copyResult := s.app.CopyFileToCache(TableType1, filePath)
		if !copyResult.Ok {
			errorMessage := fmt.Sprintf("文件复制到缓存失败: %s", copyResult.Message)
			fmt.Println(errorMessage)
			s.app.InsertImportRecord(fileName, TableType1, "上传失败", errorMessage)
			return db.QueryResult{
				Ok:      false,
				Data:    []string{errorMessage},
				Message: errorMessage,
			}
		}

		s.app.InsertImportRecord(fileName, TableType1, "上传成功", "数据校验通过")
	}

	return db.QueryResult{
		Ok:      true,
		Message: "校验通过",
	}
}

// validateTable1DataWithEnterpriseCheck 校验附表1数据（包含企业名称和统一信用代码校验）
func (s *DataImportService) validateTable1DataWithEnterpriseCheck(mainData, usageData, equipData []map[string]interface{}) []string {
	errors := []string{}

	// 0. 检查主表数据条数
	if len(mainData) == 0 {
		errors = append(errors, "单位基本信息不能为空，请核对并重新上传数据")
		return errors
	}
	if len(mainData) > 1 {
		errors = append(errors, fmt.Sprintf("单位基本信息表格数据条数错误，模板需要1条，上传文件%d条", len(mainData)))
		return errors
	}

	// 1. 在一个循环中完成所有验证
	for _, data := range mainData {
		// 使用记录的实际Excel行号
		excelRowNum := 7 // 企业基本信息固定在第7行
		if rowNum, ok := data["_excel_row"].(int); ok {
			excelRowNum = rowNum
		}

		// 获取第二部分表格的行号
		excelRowNum2 := excelRowNum // 默认使用第一部分的行号
		if rowNum2, ok := data["_excel_row2"].(int); ok {
			excelRowNum2 = rowNum2
		}

		// 1.1 校验必填字段
		fieldErrors := s.validateTable1RequiredFieldsWithRowNumbers(data, excelRowNum, excelRowNum2)
		errors = append(errors, fieldErrors...)

		// 1.2 企业名称和统一信用代码校验
		enterpriseErrors := s.validateEnterpriseAndCreditCode(data, excelRowNum)
		errors = append(errors, enterpriseErrors...)

		// 1.3 省市县和统一社会信用代码对应关系校验
		regionErrors := s.validateRegionCorrespondence(data, excelRowNum)
		errors = append(errors, regionErrors...)
	}

	return errors
}

// validateTable1RequiredFieldsWithRowNumbers 校验附表1必填字段（支持两个不同行号）
func (s *DataImportService) validateTable1RequiredFieldsWithRowNumbers(data map[string]interface{}, rowNum1, rowNum2 int) []string {
	errors := []string{}

	// 第一部分表格的字段（企业基本信息）
	part1Fields := map[string]string{
		"stat_date":     "年份",
		"unit_name":     "单位名称",
		"credit_code":   "统一社会信用代码",
		"trade_a":       "行业门类",
		"trade_b":       "行业大类",
		"trade_c":       "行业中类",
		"province_name": "单位所在省/市/区",
		"city_name":     "单位所在地市",
		"country_name":  "单位所在区县",
		"tel":           "联系电话",
	}

	// 第二部分表格的字段（综合能源消费情况）
	part2Fields := map[string]string{
		"annual_energy_equivalent_value": "年综合能耗当量值（万吨标准煤，含原料用能）",
		"annual_energy_equivalent_cost":  "年综合能耗等价值（万吨标准煤，含原料用能）",
	}

	// 校验第一部分字段
	for fieldName, displayName := range part1Fields {
		if value, ok := data[fieldName].(string); !ok || value == "" {
			errors = append(errors, fmt.Sprintf("第%d行：%s不能为空，请核对并重新上传数据", rowNum1, displayName))
		}
	}

	// 校验第二部分字段
	for fieldName, displayName := range part2Fields {
		if value, ok := data[fieldName].(string); !ok || value == "" {
			errors = append(errors, fmt.Sprintf("第%d行：%s不能为空，请核对并重新上传数据", rowNum2, displayName))
		}
	}

	return errors
}
