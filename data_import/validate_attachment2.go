package data_import

import (
	"fmt"
	"os"
	"path/filepath"
	"shuji/db"
	"strings"

	"github.com/xuri/excelize/v2"
)

// parseAttachment2Excel 解析附件2Excel文件
func (s *DataImportService) parseAttachment2Excel(f *excelize.File, skipValidate bool) ([]map[string]interface{}, error) {
	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}

	// 解析主表数据
	mainData, err := s.parseAttachment2MainSheet(f, sheets[0], skipValidate)
	if err != nil {
		return nil, fmt.Errorf("解析%s数据失败: %v", TableTypeAttachment2, err)
	}

	return mainData, nil
}

// parseAttachment2MainSheet 解析附件2主表数据
func (s *DataImportService) parseAttachment2MainSheet(f *excelize.File, sheetName string, skipValidate bool) ([]map[string]interface{}, error) {
	var mainData []map[string]interface{}

	// 读取表格数据
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	// 查找表格的开始位置（第4行是表头）
	startDataRow := 7
	if startDataRow >= len(rows) {
		return nil, fmt.Errorf("和%s模板不匹配，表格行数不足", TableTypeAttachment2)
	}

	// 解析制表单位（第3行）
	var reportUnit string
	var row3FirstCell string
	row3 := rows[2] // 第3行（0索引为2）

	if len(row3) <= 2 {
		return nil, fmt.Errorf("和%s模板不匹配，第3行列数不足", TableAttachment2)
	}

	row3FirstCell = s.cleanCellValue(row3[0]) // 第1列：制表单位：
	if row3FirstCell != "制表单位：" {
		return nil, fmt.Errorf("和%s模板不匹配，模板要求第3行第1列为：制表单位：，上传数据为：%s", TableAttachment2, row3FirstCell)
	}

	reportUnit = s.cleanCellValue(row3[1]) // 第2列：制表单位值

	// 期望的表头（第4行的主要表头，基础信息, 包含合并单元格）
	expectedHeaders4 := []string{
		"省（市、区）", "地市（州）", "县（区）", "年份", "分品种煤炭消费摸底", "", "", "", "分用途煤炭消费摸底", "", "", "", "", "", "", "", "", "焦炭消费摸底",
	}

	// 获取表头（第4行）
	headers := rows[3]
	expectedHeadersCount := len(expectedHeaders4)
	if !skipValidate {
		// 检查表头一致性
		if len(headers) < expectedHeadersCount {
			return nil, fmt.Errorf("和%s模板不匹配，第4行列数不足，模板要求%d列，上传数据为:%d列", TableAttachment2, expectedHeadersCount, len(headers))
		}

		for i, expected := range expectedHeaders4 {
			if i >= len(headers) {
				return nil, fmt.Errorf("缺少表头：%s", expected)
			}

			actual := strings.TrimSpace(headers[i])
			if actual != expected {
				return nil, fmt.Errorf("和%s模板不匹配，第4行第%d列表头不匹配， 模板要求：%s，上传数据为：%s", TableAttachment2, i+1, expected, actual)
			}
		}

		// 校验第5行表头（分品种煤炭消费摸底）
		row5 := rows[4] // 第5行（0索引为4）
		if len(row5) < 8 {
			return nil, fmt.Errorf("和%s模板不匹配，第5行列数不足，需要至少8列", TableAttachment2)
		}

		// 校验第5行关键表头字段
		expectedHeaders5 := map[int]string{
			4: "煤合计",
			5: "原煤",
			6: "洗精煤",
			7: "其他",
		}

		for colIndex, expectedHeader := range expectedHeaders5 {
			if colIndex < len(row5) {
				actualHeader := s.cleanCellValue(row5[colIndex])
				if actualHeader != expectedHeader {
					return nil, fmt.Errorf("和%s模板不匹配，第5行第%d列表头错误，期望：%s，实际：%s", TableAttachment2, colIndex+1, expectedHeader, actualHeader)
				}
			}
		}

		// 校验第6行表头（能源加工转换和终端消费）
		row6 := rows[5] // 第6行（0索引为5）
		if len(row6) < 15 {
			return nil, fmt.Errorf("和%s模板不匹配，第6行列数不足，需要至少15列", TableAttachment2)
		}

		// 校验第6行关键表头字段
		expectedHeaders6 := map[int]string{
			8:  "1.火力发电",
			9:  "2.供热",
			10: "3.煤炭洗选",
			11: "4.炼焦",
			12: "5.炼油及煤制油",
			13: "6.制气",
			14: "1.工业",
		}

		for colIndex, expectedHeader := range expectedHeaders6 {
			if colIndex < len(row6) {
				actualHeader := s.cleanCellValue(row6[colIndex])
				if actualHeader != expectedHeader {
					return nil, fmt.Errorf("和%s模板不匹配，第6行第%d列表头错误，模板要求：%s，上传数据为：%s", TableAttachment2, colIndex+1, expectedHeader, actualHeader)
				}
			}
		}
	}

	// 构建表头映射（基于位置）
	headerMap := make(map[int]string)

	// 处理复杂的合并单元格表头结构（基于HTML模板的实际结构）
	// 前4列：基础信息（每列占4行）
	headerMap[0] = "province_name" // 省（市、区）
	headerMap[1] = "city_name"     // 地市（州）
	headerMap[2] = "country_name"  // 县（区）
	headerMap[3] = "stat_date"     // 年份

	// 第5-8列：分品种煤炭消费摸底（每列占3行）
	headerMap[4] = "total_coal"  // 煤合计
	headerMap[5] = "raw_coal"    // 原煤
	headerMap[6] = "washed_coal" // 洗精煤
	headerMap[7] = "other_coal"  // 其他

	// 第9-14列：能源加工转换（每列占2行）
	headerMap[8] = "power_generation" // 1.火力发电
	headerMap[9] = "heating"          // 2.供热
	headerMap[10] = "coal_washing"    // 3.煤炭洗选
	headerMap[11] = "coking"          // 4.炼焦
	headerMap[12] = "oil_refining"    // 5.炼油及煤制油
	headerMap[13] = "gas_production"  // 6.制气

	// 第15-17列：终端消费
	headerMap[14] = "industry"      // 1.工业（占2行）
	headerMap[15] = "raw_materials" // #用作原料、材料（占1行）
	headerMap[16] = "other_uses"    // 2.其他用途（占1行）

	// 第18列：焦炭消费摸底（占2行）
	headerMap[17] = "coke" // 焦炭

	// 解析数据行（从第8行开始，跳过表头、子表头等）
	for i := startDataRow; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 2 || (len(row) > 0 && strings.TrimSpace(row[0]) == "") {
			continue // 跳过空行
		}

		// 构建数据行
		dataRow := make(map[string]interface{})

		// 添加制表单位信息
		dataRow["report_unit"] = reportUnit

		for j, cell := range row {
			if fieldName, exists := headerMap[j]; exists && fieldName != "" {
				cleanedValue := s.cleanCellValue(cell)
				dataRow[fieldName] = cleanedValue
			}
		}

		// 只添加有省份名称的数据行
		if provinceName, ok := dataRow["province_name"].(string); ok && provinceName != "" {
			mainData = append(mainData, dataRow)
		}
	}

	return mainData, nil
}

// ValidateAttachment2File 校验附件2文件
func (s *DataImportService) ValidateAttachment2File(filePath string, isCover bool) db.QueryResult {
	fileName := filepath.Base(filePath)

	// 第一步: 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		errorMessage := fmt.Sprintf("文件不存在: %v", err)
		s.app.InsertImportRecord(fileName, TableTypeAttachment2, "上传失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Message: errorMessage,
			Data:    []string{errorMessage},
		}
	}

	// 第二步: 文件是否可读取
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		errorMessage := fmt.Sprintf("读取Excel文件失败: %v", err)
		s.app.InsertImportRecord(fileName, TableTypeAttachment2, "上传失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Message: errorMessage,
			Data:    []string{errorMessage},
		}
	}
	defer f.Close()

	// 第三步: 文件是否和模板文件匹配
	mainData, err := s.parseAttachment2Excel(f, false)
	if err != nil {
		errorMessage := err.Error()
		s.app.InsertImportRecord(fileName, TableTypeAttachment2, "上传失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Message: errorMessage,
			Data:    []string{errorMessage},
		}
	}

	if isCover {
		// 第四步: 去缓存目录检查是否有同名的文件, 直接返回,需要前端确认
		cacheResult := s.app.CacheFileExists(TableTypeAttachment2, fileName)
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
	validationErrors := s.validateAttachment2Data(mainData)
	if len(validationErrors) > 0 {
		errorMessage := fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; "))
		s.app.InsertImportRecord(fileName, TableTypeAttachment2, "上传失败", errorMessage)
		return db.QueryResult{
			Ok:      false,
			Message: errorMessage,
			Data:    validationErrors,
		}
	}

	if len(validationErrors) == 0 {
		// 第六步: 复制文件到缓存目录（只有校验通过才复制）
		copyResult := s.app.CopyFileToCache(TableTypeAttachment2, filePath)
		if !copyResult.Ok {
			errorMessage := fmt.Sprintf("文件复制到缓存失败: %s", copyResult.Message)
			s.app.InsertImportRecord(fileName, TableTypeAttachment2, "上传失败", errorMessage)
			return db.QueryResult{
				Ok:      false,
				Message: errorMessage,
				Data:    []string{errorMessage},
			}
		}

		s.app.InsertImportRecord(fileName, TableTypeAttachment2, "上传成功", "校验通过")
	}

	return db.QueryResult{
		Ok:      true,
		Message: "校验通过",
	}
}

// validateAttachment2Data 校验附件2数据
func (s *DataImportService) validateAttachment2Data(mainData []map[string]interface{}) []string {
	errors := []string{}

	// 在一个循环中完成所有验证
	for i, data := range mainData {
		// Excel中的实际行号：数据从第8行开始（表头第4行+3行说明+1行数据）
		excelRowNum := 8 + i

		// 1. 检查必填字段
		fieldErrors := s.validateRequiredFields(data, Attachment2RequiredFields, excelRowNum)
		errors = append(errors, fieldErrors...)

		// 2. 检查区域与当前单位是否相符
		regionErrors := s.validateRegionOnly(data, excelRowNum)
		errors = append(errors, regionErrors...)
	}

	return errors
}
