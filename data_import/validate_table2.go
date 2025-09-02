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
func (s *DataImportService) parseTable2Excel(f *excelize.File, skipValidate bool) (map[string]interface{}, []map[string]interface{}, error) {
	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, fmt.Errorf("Excel文件没有工作表")
	}

	// 解析主表数据
	unitInfo, mainData, err := s.parseTable2MainSheet(f, sheets[0], skipValidate)
	if err != nil {
		return nil, nil, fmt.Errorf("和%s模板不匹配,  %v", TableName2, err)
	}

	if len(mainData) == 0 {
		return nil, nil, fmt.Errorf("导入文件没有检测到数据, 请检查文件是否正确")
	}

	return unitInfo, mainData, nil
}

// parseTable2MainSheet 解析附表2主表数据
func (s *DataImportService) parseTable2MainSheet(f *excelize.File, sheetName string, skipValidate bool) (map[string]interface{}, []map[string]interface{}, error) {
	var mainData []map[string]interface{}

	// 读取表格数据
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}

	// 解析单位基本信息（第3-4行）
	unitInfo, err := s.parseTable2UnitInfo(rows)
	if err != nil {
		return nil, nil, fmt.Errorf("解析单位基本信息失败: %v", err)
	}

	// 查找设备表格的开始位置（第5行是表头）
	startRow := 4 // 从第5行开始（0索引为4）
	if startRow >= len(rows) {
		return unitInfo, nil, fmt.Errorf("表格行数不足")
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
		for i, expected := range expectedHeaders {
			if i >= len(headers) {
				return unitInfo, nil, fmt.Errorf("缺少表头：%s", expected)
			}

			actual := strings.TrimSpace(headers[i])
			if actual != expected {
				return unitInfo, nil, fmt.Errorf("第%d列表头不匹配，期望：%s，实际：%s", i+1, expected, actual)
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
		if len(row) < 2 {
			continue // 跳过空行
		}

		// 构建数据行，先复制单位基本信息
		dataRow := make(map[string]interface{})
		for key, value := range unitInfo {
			dataRow[key] = value
		}
		dataRow["_excel_row"] = i + 1 // 0索引转换为1索引

		hasData := false
		// 添加设备信息
		for j, cell := range row {
			if fieldName, exists := headerMap[j]; exists && fieldName != "" {
				cleanedValue := s.cleanCellValue(cell)
				dataRow[fieldName] = cleanedValue
				hasData = true
			}
		}

		// 只添加有数据的行
		if hasData {
			mainData = append(mainData, dataRow)
		}
	}

	return unitInfo, mainData, nil
}

// parseTable2UnitInfo 解析附表2单位基本信息
func (s *DataImportService) parseTable2UnitInfo(rows [][]string) (map[string]interface{}, error) {
	unitInfo := make(map[string]interface{})

	// 第3行：单位名称和统一社会信用代码
	if len(rows) < 4 {
		return nil, fmt.Errorf("表格行数不足，无法解析单位基本信息")
	}

	row3 := rows[2] // 第3行（0索引为2）

	// 单位名称（第2列）
	unitInfo["unit_name"] = s.GetCellValueByRow(row3, 1)
	//  统一社会信用代码（第7列）
	unitInfo["credit_code"] = s.GetCellValueByRow(row3, 6)

	row4 := rows[3] // 第4行（0索引为3）

	unitInfo["province_name"] = s.GetCellValueByRow(row4, 1)
	unitInfo["city_name"] = s.GetCellValueByRow(row4, 2)
	unitInfo["country_name"] = s.GetCellValueByRow(row4, 3)

	// 所属行业：门类（第6列）、大类（第7列）、小类（第8列）
	unitInfo["trade_a"] = s.GetCellValueByRow(row4, 6)
	unitInfo["trade_b"] = s.GetCellValueByRow(row4, 7)
	unitInfo["trade_c"] = s.GetCellValueByRow(row4, 8)

	// 数据年份（第10列）
	unitInfo["stat_date"] = s.GetCellValueByRow(row4, 10)

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

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errorMessage := "文件不存在"
		s.app.InsertImportRecord(fileName, TableType2, "导入失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    []string{errorMessage},
			Message: errorMessage,
		}
	}

	// 文件是否可读取
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		errorMessage := fmt.Sprintf("读取Excel文件失败: %v", err)
		s.app.InsertImportRecord(fileName, TableType2, "导入失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    []string{errorMessage},
			Message: errorMessage,
		}
	}
	defer f.Close()

	// 文件是否和模板文件匹配
	unitInfo, mainData, err := s.parseTable2Excel(f, false)
	if err != nil {
		errorMessage := err.Error()
		s.app.InsertImportRecord(fileName, TableType2, "导入失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    []string{errorMessage},
			Message: errorMessage,
		}
	}

	// 按行读取文件数据并校验
	validationErrors := s.validateTable2DataWithEnterpriseCheck(unitInfo, mainData)
	if len(validationErrors) > 0 {
		errorMessage := fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; "))
		s.app.InsertImportRecord(fileName, TableType2, "导入失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    validationErrors,
			Message: errorMessage,
		}
	}

	// 去缓存目录检查是否有同名的文件, 直接返回,需要前端确认
	if isCover {
		// 去缓存目录检查是否有同名的文件, 直接返回,需要前端确认
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

	// 复制文件到缓存目录（只有校验通过才复制）
	copyResult := s.app.CopyFileToCache(TableType2, filePath)
	if !copyResult.Ok {
		copyMessage := fmt.Sprintf("文件复制到缓存失败: %s", copyResult.Message)
		s.app.InsertImportRecord(fileName, TableType2, "导入失败", copyMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    []string{copyMessage},
			Message: copyMessage,
		}
	} else {
		s.UnprotecFile(copyResult.Data.(string))
	}

	s.app.InsertImportRecord(fileName, TableType2, "导入成功", "校验通过")

	return db.QueryResult{
		Ok:      true,
		Message: "校验通过",
	}
}

// validateTable2DataWithEnterpriseCheck 校验附表2数据（包含企业名称和统一信用代码校验）
func (s *DataImportService) validateTable2DataWithEnterpriseCheck(unitInfo map[string]interface{}, mainData []map[string]interface{}) []string {
	errors := []string{}

	unitInfoRequiredFields := map[string]string{
		"unit_name":   "单位名称",
		"credit_code": "统一社会信用代码",
	}
	// 检查基本信息必填字段
	unitInfoFieldErrors := s.validateRequiredFields(unitInfo, unitInfoRequiredFields, 3)
	errors = append(errors, unitInfoFieldErrors...)

	regionRequiredFields := map[string]string{
		"province_name": "单位所在省/市/区",
		"city_name":     "单位所在地市",
		"country_name":  "单位所在区县",
		"stat_date":     "年份",
		"trade_a":       "所属行业门类",
		"trade_b":       "所属行业大类",
		"trade_c":       "所属行业中类",
	}

	regionFieldErrors := s.validateRequiredFields(unitInfo, regionRequiredFields, 4)
	errors = append(errors, regionFieldErrors...)

	// 企业名称和统一信用代码校验
	enterpriseErrors := s.validateEnterpriseAndCreditCode(unitInfo, 3, 4)
	errors = append(errors, enterpriseErrors...)

	fmt.Println("主数据==", mainData)
	// 在一个循环中完成所有验证
	for _, data := range mainData {
		excelRowNum := s.getExcelRowNumber(data)

		// 检查必填字段
		fieldErrors := s.validateRequiredFields(data, Table2RequiredFields, excelRowNum)
		errors = append(errors, fieldErrors...)

		// 年耗煤量特殊校验（状态为"停用"时可以为空，其他情况下不能为空）
		coalConsumptionErrors := s.validateTable2CoalConsumption(data, excelRowNum)
		errors = append(errors, coalConsumptionErrors...)
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
	if status == "停用" || status == "" {
		return errors
	}

	// 其他情况下，年耗煤量不能为空
	if status != "" && annualCoalConsumption == "" {
		errors = append(errors, fmt.Sprintf("第%d行：年耗煤量不能为空（状态为\"%s\"时）", excelRowNum, status))
	}

	return errors
}
