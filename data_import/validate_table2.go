package data_import

import (
	"fmt"
	"os"
	"path/filepath"
	"shuji/db"
	"strings"

	"github.com/xuri/excelize/v2"
)

// parseTable2Excel 解析附表2Excel文件
func (s *DataImportService) parseTable2Excel(f *excelize.File, skipValidate bool) ([]map[string]interface{}, error) {
	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	// 解析主表数据
	mainData, err := s.parseTable2MainSheet(f, sheets[0], skipValidate)
	if err != nil {
		return nil, fmt.Errorf("和%s模板不匹配,  %v", TableName2, err)
	}

	return mainData, nil
}

// parseTable2MainSheet 解析附表2主表数据
func (s *DataImportService) parseTable2MainSheet(f *excelize.File, sheetName string, skipValidate bool) ([]map[string]interface{}, error) {
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

	// 构建表头映射
	headerMap := make(map[int]string)

	if !skipValidate {
		// 检查表头一致性
		if len(headers) < len(expectedHeaders) {
			return nil, fmt.Errorf("表头列数不足，期望%d列，实际%d列", len(expectedHeaders), len(headers))
		}

		for i, expected := range expectedHeaders {
			if i >= len(headers) {
				return nil, fmt.Errorf("缺少表头：%s", expected)
			}

			actual := strings.TrimSpace(headers[i])
			if actual != expected {
				return nil, fmt.Errorf("第%d列表头不匹配，期望：%s，实际：%s", i+1, expected, actual)
			}
		}
	}

	// 构建表头映射（无论是否跳过校验都需要）
	for i, expected := range expectedHeaders {
		if i < len(headers) {
			headerMap[i] = s.mapTable2HeaderToField(expected)
		}
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
		dataRow["_excel_row"] = i + 1

		// 添加设备信息
		for j, cell := range row {
			if fieldName, exists := headerMap[j]; exists && fieldName != "" {
				cleanedValue := s.cleanCellValue(cell)
				dataRow[fieldName] = cleanedValue
			}
		}

		// 只添加有耗煤类型的数据行
		if equipType, ok := dataRow["coal_type"].(string); ok && equipType != "" {
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
		"序号":         "row_no",
		"类型":         "coal_type",
		"编号":         "coal_no",
		"累计使用时间":     "usage_time",
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
func (s *DataImportService) ValidateTable2File(filePath string, isCover bool) db.QueryResult {
	fileName := filepath.Base(filePath)

	// 第一步: 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		s.app.InsertImportRecord(fileName, TableType2, "上传失败", "文件不存在")
		return db.QueryResult{
			Ok:      false,
			Message: "文件不存在",
		}
	}

	// 第二步: 文件是否可读取
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		s.app.InsertImportRecord(fileName, TableType2, "上传失败", fmt.Sprintf("读取Excel文件失败: %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 第三步: 文件是否和模板文件匹配
	mainData, err := s.parseTable2Excel(f, false)
	if err != nil {
		s.app.InsertImportRecord(fileName, TableType2, "上传失败", fmt.Sprintf("解析Excel文件失败: %v", err))
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 第四步: 去缓存目录检查是否有同名的文件, 直接返回,需要前端确认
	if isCover {
		// 第四步: 去缓存目录检查是否有同名的文件, 直接返回,需要前端确认
		cacheResult := s.app.CacheFileExists(TableType2, fileName)
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
	validationErrors := s.validateTable2DataWithEnterpriseCheck(mainData)
	if len(validationErrors) > 0 {
		s.app.InsertImportRecord(fileName, TableType2, "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return db.QueryResult{
			Ok:      false,
			Data:    validationErrors,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	if len(validationErrors) == 0 {
		// 第六步: 复制文件到缓存目录（只有校验通过才复制）
		copyResult := s.app.CopyFileToCache(TableType2, filePath)
		if !copyResult.Ok {
			copyMessage := fmt.Sprintf("文件复制到缓存失败: %s", copyResult.Message)
			s.app.InsertImportRecord(fileName, TableType2, "上传失败", copyMessage)
			return db.QueryResult{
				Ok:      false,
				Data:    []string{copyMessage},
				Message: copyMessage,
			}
		}
		s.app.InsertImportRecord(fileName, TableType2, "上传成功", "校验通过")
	}

	return db.QueryResult{
		Ok:      true,
		Message: "校验通过",
	}
}

// validateTable2DataWithEnterpriseCheck 校验附表2数据（包含企业名称和统一信用代码校验）
func (s *DataImportService) validateTable2DataWithEnterpriseCheck(mainData []map[string]interface{}) []string {
	errors := []string{}

	// 在一个循环中完成所有验证
	for _, data := range mainData {
		// Excel中的实际行号：设备数据从第7行开始（表头第5行+1行说明+1行数据）
		excelRowNum := s.getExcelRowNumber(data)

		// 1. 检查必填字段
		fieldErrors := s.validateRequiredFields(data, Table2RequiredFields, excelRowNum)
		errors = append(errors, fieldErrors...)

		// 2. 年耗煤量特殊校验（状态为"停用"时可以为空，其他情况下不能为空）
		coalConsumptionErrors := s.validateTable2CoalConsumption(data, excelRowNum)
		errors = append(errors, coalConsumptionErrors...)

		// 3. 企业名称和统一信用代码校验
		enterpriseErrors := s.validateEnterpriseAndCreditCode(data, excelRowNum)
		errors = append(errors, enterpriseErrors...)

		// 4. 省市县和统一社会信用代码对应关系校验
		regionErrors := s.validateRegionCorrespondence(data, excelRowNum)
		errors = append(errors, regionErrors...)
	}

	return errors
}

// validateTable2CoalConsumption 校验附表2年耗煤量（状态为"停用"时可以为空，其他情况下不能为空）
func (s *DataImportService) validateTable2CoalConsumption(data map[string]interface{}, excelRowNum int) []string {
	errors := []string{}

	// 获取设备状态和年耗煤量
	status, _ := data["status"].(string)
	annualCoalConsumption, _ := data["annual_coal_consumption"].(string)

	// 如果状态为"停用"，年耗煤量可以为空
	if status == "停用" {
		return errors
	}

	// 其他情况下，年耗煤量不能为空
	if annualCoalConsumption == "" {
		errors = append(errors, fmt.Sprintf("第%d行：年耗煤量不能为空（状态为\"%s\"时）", excelRowNum, status))
	}

	return errors
}
