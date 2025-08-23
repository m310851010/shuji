package data_import

import (
	"fmt"
	"path/filepath"
	"shuji/db"
	"strings"

	"github.com/xuri/excelize/v2"
)

// parseTable2Excel 解析附表2Excel文件
func (s *DataImportService) parseTable2Excel(f *excelize.File) ([]map[string]interface{}, error) {
	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	// 解析主表数据
	mainData, err := s.parseTable2MainSheet(f, sheets[0])
	if err != nil {
		return nil, fmt.Errorf("和表2模板不匹配,  %v", err)
	}

	return mainData, nil
}

// parseTable2MainSheet 解析附表2主表数据
func (s *DataImportService) parseTable2MainSheet(f *excelize.File, sheetName string) ([]map[string]interface{}, error) {
	var mainData []map[string]interface{}

	// 读取表格数据
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	// 解析单位基本信息（第3-4行）
	unitInfo, err := s.parseTable2UnitInfo(rows)
	if err != nil {
		return nil, fmt.Errorf("解析单位基本信息失败: %v", err)
	}

	// 查找设备表格的开始位置（第5行是表头）
	startRow := 4 // 从第5行开始（0索引为4）
	if startRow >= len(rows) {
		return nil, fmt.Errorf("表格行数不足")
	}

	// 获取表头
	headers := rows[startRow]

	// 期望的表头
	expectedHeaders := []string{
		"序号", "类型", "编号", "累计使用时间", "设计年限", "能效水平",
		"容量单位", "容量", "用途", "状态", "年耗煤量（单位：吨）",
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
		headerMap[i] = s.mapTable2HeaderToField(expected)
	}

	// 解析数据行（跳过表头下的第一行说明行）
	for i := startRow + 2; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 2 || (len(row) > 0 && strings.TrimSpace(row[0]) == "") {
			continue // 跳过空行
		}

		// 构建数据行，先复制单位基本信息
		dataRow := make(map[string]interface{})
		for key, value := range unitInfo {
			dataRow[key] = value
		}

		// 添加设备信息
		for j, cell := range row {
			if fieldName, exists := headerMap[j]; exists && fieldName != "" {
				cleanedValue := s.cleanCellValue(cell)
				if strings.Contains(fieldName, "consumption") || strings.Contains(fieldName, "capacity") || strings.Contains(fieldName, "runtime") || strings.Contains(fieldName, "life") {
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
			mainData = append(mainData, dataRow)
		}
	}

	return mainData, nil
}

// parseTable2UnitInfo 解析附表2单位基本信息
func (s *DataImportService) parseTable2UnitInfo(rows [][]string) (map[string]interface{}, error) {
	unitInfo := make(map[string]interface{})

	// 第3行：单位名称和统一社会信用代码
	if len(rows) < 4 {
		return nil, fmt.Errorf("表格行数不足，无法解析单位基本信息")
	}

	row3 := rows[2] // 第3行（0索引为2）
	if len(row3) >= 6 {
		// 单位名称（第1列）
		unitName := s.cleanCellValue(row3[1])
		unitInfo["unit_name"] = unitName

		// 统一社会信用代码（第6列）
		creditCode := s.cleanCellValue(row3[6])
		unitInfo["credit_code"] = creditCode
	}

	// 第4行：单位地址、所属行业、数据年份
	if len(rows) < 4 {
		return nil, fmt.Errorf("表格行数不足，无法解析单位基本信息")
	}

	row4 := rows[3] // 第4行（0索引为3）
	if len(row4) >= 11 {
		// 单位地址：省（第1列）、市（第2列）、区县（第3列）
		province := s.cleanCellValue(row4[1])
		city := s.cleanCellValue(row4[2])
		country := s.cleanCellValue(row4[3])

		unitInfo["province_name"] = province
		unitInfo["city_name"] = city
		unitInfo["country_name"] = country

		// 所属行业：门类（第6列）、大类（第7列）、小类（第8列）
		if len(row4) >= 8 {
			industryDoor := s.cleanCellValue(row4[6])
			industryBig := s.cleanCellValue(row4[7])
			industrySmall := s.cleanCellValue(row4[8])

			unitInfo["trade_a"] = industryDoor
			unitInfo["trade_b"] = industryBig
			unitInfo["trade_c"] = industrySmall
		}

		// 数据年份（第10列）
		if len(row4) >= 10 {
			statDate := s.cleanCellValue(row4[10])
			unitInfo["stat_date"] = statDate
		}
	}

	return unitInfo, nil
}

// mapTable2HeaderToField 映射附表2表头到字段名
func (s *DataImportService) mapTable2HeaderToField(header string) string {
	header = strings.TrimSpace(header)

	// 字段映射
	fieldMap := map[string]string{
		"序号":         "sequence_no",
		"类型":         "equip_type",
		"编号":         "equip_no",
		"累计使用时间":     "total_runtime",
		"设计年限":       "design_life",
		"能效水平":       "enecrgy_efficienct_bmk",
		"容量单位":       "capacity_unit",
		"容量":         "capacity",
		"用途":         "use_info",
		"状态":         "status",
		"年耗煤量（单位：吨）": "annual_coal_consumption",
	}

	return fieldMap[header]
}

// ValidateTable2File 校验附表2文件
func (s *DataImportService) ValidateTable2File(filePath string) db.QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		// 插入导入记录
		s.app.InsertImportRecord(fileName, "附表2", "上传失败", fmt.Sprintf("读取Excel文件失败: %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 2. 解析Excel文件
	mainData, err := s.parseTable2Excel(f)
	if err != nil {
		// 插入导入记录
		s.app.InsertImportRecord(fileName, "附表2", "上传失败", fmt.Sprintf("文件%s, %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("文件%s, %v", fileName, err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateTable2Data(mainData)
	if len(validationErrors) > 0 {
		// 插入导入记录
		s.app.InsertImportRecord(fileName, "附表2", "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	// 4. 查询该表是否有数据放进data属性
	hasData := s.checkTable2HasData()

	// 5. 返回QueryResult
	return db.QueryResult{
		Ok:      true,
		Data:    hasData,
		Message: "校验通过",
	}
}

// validateTable2Data 校验附表2数据
func (s *DataImportService) validateTable2Data(mainData []map[string]interface{}) []string {
	errors := []string{}

	// 1. 检查设备类型和编号是否为空
	for i, data := range mainData {
		fieldErrors := s.validateRequiredFields(data, Table2RequiredFields, i)
		errors = append(errors, fieldErrors...)
	}

	// 2. 检查企业是否在企业清单中（如果有清单的话）
	for i, data := range mainData {
		unitName := s.getStringValue(data["unit_name"])
		creditCode := s.getStringValue(data["credit_code"])
		enterpriseListErrors := s.validateEnterpriseNameCreditCodeCorrespondence(unitName, creditCode, i, true)
		errors = append(errors, enterpriseListErrors...)

		// 3. 检查企业名称和统一信用代码是否对应（如果有清单的话）
		correspondenceErrors := s.validateEnterpriseNameCreditCodeCorrespondence(unitName, creditCode, i, false)
		errors = append(errors, correspondenceErrors...)
	}

	return errors
}

// checkTable2HasData 检查附表2相关表是否有数据
func (s *DataImportService) checkTable2HasData() bool {
	return s.checkTableHasData(TableCriticalCoalEquipmentConsumption)
}
