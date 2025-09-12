package data_import

import (
	"fmt"
	"os"
	"path/filepath"
	"shuji/db"
	"strings"
	"shuji/slices"

	"github.com/xuri/excelize/v2"
)

// parseTable3Excel 解析附表3Excel文件
func (s *DataImportService) parseTable3Excel(f *excelize.File, skipValidate bool) ([]map[string]interface{}, error) {
	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	// 解析主表数据
	mainData, err := s.parseTable3MainSheet(f, sheets[0], skipValidate)
	if err != nil {
		return nil, fmt.Errorf("解析主表数据失败: %v", err)
	}

	return mainData, nil
}

// parseTable3MainSheet 解析附表3主表数据
func (s *DataImportService) parseTable3MainSheet(f *excelize.File, sheetName string, skipValidate bool) ([]map[string]interface{}, error) {
	var mainData []map[string]interface{}

	// 读取表格数据
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	// 查找表格的开始位置
	startRow := 2
	if startRow >= len(rows) {
		return nil, fmt.Errorf("表格行数不足")
	}

	// 获取表头
	headers := rows[startRow]

	// 期望的表头
	expectedHeaders := []string{
		"序号",
		"项目名称",
		"项目代码",
		"建设单位",
		"主要建设内容",
		"项目所在省、自治区、直辖市",
		"项目所在地市",
		"项目所在区县",
		"所属行业大类（2位代码）",
		"所属行业小类",
		"节能审查批复时间",
		"拟投产时间",
		"实际投产时间",
		"节能审查机关",
		"审查意见文号",
		"年综合能源消费量（万吨标准煤，含原料用能和可再生能源）",
		"",
		"年煤品消费量（万吨，实物量）",
		"",
		"",
		"",
		"年煤品消费量（万吨标准煤，折标量）",
		"",
		"",
		"",
		"煤炭消费替代情况",
		"",
		"",
		"原料用煤情况",
	}

	if !skipValidate {
		// 检查表头一致性
		if len(headers) < len(expectedHeaders) {
			return nil, fmt.Errorf("与数据模板不匹配")
		}

		for i, expected := range expectedHeaders {
			if i >= len(headers) {
				return nil, fmt.Errorf("缺少表头：%s", expected)
			}

			actual := strings.TrimSpace(headers[i])
			if actual != expected {
				return nil, fmt.Errorf("与数据模板不匹配")
			}
		}
	}

	// 构建表头映射
	headerArr := []string{
		"row_no",
		"project_name",
		"project_code",
		"construction_unit",
		"main_construction_content",
		"province_name",
		"city_name",
		"country_name",
		"trade_a",
		"trade_c",
		"examination_approval_time",
		"scheduled_time",
		"actual_time",
		"examination_authority",
		"document_number",
		"equivalent_value",
		"equivalent_cost",
		"pq_total_coal_consumption",
		"pq_coal_consumption",
		"pq_coke_consumption",
		"pq_blue_coke_consumption",
		"sce_total_coal_consumption",
		"sce_coal_consumption",
		"sce_coke_consumption",
		"sce_blue_coke_consumption",
		"is_substitution",
		"substitution_source",
		"substitution_quantity",
		"pq_annual_coal_quantity",
		"sce_annual_coal_quantity",
	}
	
	// 时间字段
	timeFields := []string{
		"examination_approval_time",
		"scheduled_time",
		"actual_time",
	}

	// 解析数据行（跳过表头下的第一行）
	for i := startRow + 2; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 2 {
			continue // 跳过空行
		}

		// 构建数据行
		dataRow := make(map[string]interface{})
		hasData := false
		for j, cell := range row {
			if j < len(headerArr) && headerArr[j] != "" {
				fieldName := headerArr[j]
				if slices.Contains(timeFields, fieldName) {
					dataRow[fieldName] =s.parseDateValueToString(cell, "")
					fmt.Println(fieldName, dataRow[fieldName])
				} else {
					dataRow[fieldName] = s.cleanCellValue(cell)
				}
				hasData = true
			}
		}

		dataRow["_excel_row"] = i + 1

		// 只添加有项目名称的数据行
		if hasData {
			mainData = append(mainData, dataRow)
		}
	}

	return mainData, nil
}

// ValidateTable3File 校验附表3文件
func (s *DataImportService) ValidateTable3File(filePath string, isCover bool) db.QueryResult {
	fileName := filepath.Base(filePath)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errorMessage := fmt.Sprintf("文件不存在: %v", err)
		s.app.InsertImportRecord(fileName, TableType3, "导入失败", errorMessage)
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
		s.app.InsertImportRecord(fileName, TableType3, "导入失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    []string{errorMessage},
			Message: errorMessage,
		}
	}
	defer f.Close()

	// 文件是否和模板文件匹配
	mainData, err := s.parseTable3Excel(f, false)
	if err != nil {
		errorMessage := err.Error()
		s.app.InsertImportRecord(fileName, TableType3, "导入失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    []string{errorMessage},
			Message: errorMessage,
		}
	}

	// 按行读取文件数据并校验
	validationErrors := s.validateTable3Data(mainData)
	if len(validationErrors) > 0 {
		errorMessage := fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; "))
		s.app.InsertImportRecord(fileName, TableType3, "导入失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Data:    validationErrors,
			Message: errorMessage,
		}
	}

	// 复制文件到缓存目录（只有校验通过才复制）
	if len(validationErrors) == 0 {
		// 去缓存目录检查是否有同名的文件, 直接返回,需要前端确认
		if isCover {
			// 去缓存目录检查是否有同名的文件, 直接返回,需要前端确认
			cacheResult := s.app.CacheFileExists(TableType3, fileName)
			if cacheResult.Ok {
				// 文件已存在，直接返回，需要前端确认
				return db.QueryResult{
					Ok:      false,
					Message: "文件已存在，需要确认是否覆盖",
					Data:    "FILE_EXISTS",
				}
			}
		}

		copyResult := s.app.CopyFileToCache(TableType3, filePath)
		if !copyResult.Ok {
			errorMessage := fmt.Sprintf("文件复制到缓存失败: %s", copyResult.Message)
			s.app.InsertImportRecord(fileName, TableType3, "导入失败", errorMessage)
			return db.QueryResult{
				Ok:      false,
				Data:    []string{errorMessage},
				Message: errorMessage,
			}
		} else {
			s.UnprotecFile(copyResult.Data.(string))
		}
		s.app.InsertImportRecord(fileName, TableType3, "导入成功", "校验通过")
	}

	return db.QueryResult{
		Ok:      true,
		Message: "校验通过",
	}
}

// validateTable3Data 校验附表3数据
func (s *DataImportService) validateTable3Data(mainData []map[string]interface{}) []string {

	errors := []string{}

	// 用于存储已检查的项目信息（用于重复数据检查）
	projectNameMap := make(map[string]int)
	projectCodeMap := make(map[string]int)
	approvalNumberMap := make(map[string]int)

	// 在一个循环中完成所有验证
	for _, data := range mainData {
		// Excel中的实际行号：数据从第4行开始（表头第3行+1行数据）
		excelRowNum := s.getExcelRowNumber(data)

		// 1. 检查必填字段
		fieldErrors := s.validateRequiredFields(data, Table3RequiredFields, excelRowNum)
		errors = append(errors, fieldErrors...)

		// 2. 检查时间字段（拟投产时间和实际投产时间至少选择其一）
		timeErrors := s.validateTable3TimeFields(data, excelRowNum)
		errors = append(errors, timeErrors...)

		// 3. 检查区域与当前单位是否相符
		regionErrors := s.validateRegionOnly(data, excelRowNum)
		errors = append(errors, regionErrors...)

		// 4. 检查固定资产投资项目重复数据
		projectName := s.getStringValue(data["project_name"])
		projectCode := s.getStringValue(data["project_code"])
		approvalNumber := s.getStringValue(data["document_number"])

		// 分别检查项目名称、项目代码、审查意见文号的唯一性
		if projectName != "" {
			if existingIndex, exists := projectNameMap[projectName]; exists {
				existingExcelRowNum := existingIndex
				errors = append(errors, fmt.Sprintf("第%d行：项目名称重复（与第%d行重复）", excelRowNum, existingExcelRowNum))
			} else {
				projectNameMap[projectName] = excelRowNum
			}
		}

		if projectCode != "" {
			if existingIndex, exists := projectCodeMap[projectCode]; exists {
				existingExcelRowNum := existingIndex
				errors = append(errors, fmt.Sprintf("第%d行：项目代码重复（与第%d行重复）", excelRowNum, existingExcelRowNum))
			} else {
				projectCodeMap[projectCode] = excelRowNum
			}
		}

		if approvalNumber != "" {
			if existingIndex, exists := approvalNumberMap[approvalNumber]; exists {
				existingExcelRowNum := existingIndex
				errors = append(errors, fmt.Sprintf("第%d行：审查意见文号重复（与第%d行重复）", excelRowNum, existingExcelRowNum))
			} else {
				approvalNumberMap[approvalNumber] = excelRowNum
			}
		}
	}

	return errors
}
