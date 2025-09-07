package data_import

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"shuji/db"
	"slices"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ModelDataCoverTable2 覆盖附表2数据
func (s *DataImportService) ModelDataCoverTable2(filePaths []string) db.QueryResult {
	// 使用包装函数来处理异常
	return s.modelDataCoverTable2WithRecover(filePaths)
}

// modelDataCoverTable2WithRecover 带异常处理的覆盖附表2数据函数
func (s *DataImportService) modelDataCoverTable2WithRecover(filePaths []string) db.QueryResult {
	var result db.QueryResult

	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ModelDataCoverTable2 发生异常: %v", r)
			// 设置错误结果
			result = db.QueryResult{
				Ok:      false,
				Message: fmt.Sprintf("函数执行异常: %v", r),
				Data:    nil,
			}
		}
	}()

	cacheDir := s.app.GetCachePath(TableType2)
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取缓存目录失败: %v", err),
		}
	}

	var validationErrors []ValidationError
	var failedFiles []string

	for _, file := range files {
		// 检查是否xlsx或者xls文件
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".xlsx") || strings.HasSuffix(file.Name(), ".xls")) {
			filePath := filepath.Join(cacheDir, file.Name())
			if !slices.Contains(filePaths, filePath) {
				// 删除该Excel文件
				os.Remove(filePath)
				continue
			}

			f, err := excelize.OpenFile(filePath)
			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 读取失败: %v", file.Name(), err)})
				os.Remove(filePath)
				continue
			}

			_, mainData, err := s.parseTable2Excel(f, true)
			f.Close()
			os.Remove(filePath)

			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			err = s.coverTable2Data(mainData, file.Name())
			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 覆盖数据失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
			}
		}
	}

	result = db.QueryResult{
		Ok:      true,
		Message: "覆盖完成",
		Data: map[string]interface{}{
			"failed_files": failedFiles,      // 失败的文件
			"errors":       validationErrors, // 错误信息
		},
	}
	return result
}

// ModelDataCheckTable2 附表2模型校验函数
func (s *DataImportService) ModelDataCheckTable2() db.QueryResult {
	// 使用包装函数来处理异常
	return s.modelDataCheckTable2WithRecover()
}

// modelDataCheckTable2WithRecover 带异常处理的附表2模型校验函数
func (s *DataImportService) modelDataCheckTable2WithRecover() db.QueryResult {
	var result db.QueryResult

	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ModelDataCheckTable2 发生异常: %v", r)
			// 设置错误结果
			result = db.QueryResult{
				Ok:      false,
				Message: fmt.Sprintf("函数执行异常: %v", r),
				Data:    nil,
			}
		}
	}()

	// 1. 读取缓存目录指定表格类型下的所有Excel文件
	cacheDir := s.app.GetCachePath(TableType2)

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		result = db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取缓存目录失败: %v", err),
		}
		return result
	}

	var validationErrors []ValidationError = []ValidationError{} // 验证错误信息
	var systemErrors []ValidationError = []ValidationError{}     // 系统错误信息
	var importedFiles []string = []string{}                      // 导入的文件
	var coverFiles []string = []string{}                         // 覆盖的文件
	var failedFiles []string = []string{}                        // 失败的文件
	var hasExcelFile bool = false                                // 是否有Excel文件

	// 2. 循环调用对应的解析Excel函数
	for _, file := range files {
		// 检查是否xlsx或者xls文件
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".xlsx") || strings.HasSuffix(file.Name(), ".xls")) {
			hasExcelFile = true
			filePath := filepath.Join(cacheDir, file.Name())

			// 解析Excel文件 (skipValidate=true)
			f, err := excelize.OpenFile(filePath)
			if err != nil {
				systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 读取失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			_, mainData, err := s.parseTable2Excel(f, true)
			f.Close()

			if err != nil {
				systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			// 4. 调用校验函数,对每一行数据验证
			errors := s.validateTable2DataForModel(mainData)
			if len(errors) > 0 {
				// 校验失败，在Excel文件中错误行最后添加错误信息
				err = s.addValidationErrorsToExcelTable2(filePath, errors)

				if err != nil {
					msg := err.Error()
					// 如果错误是文件名长度超出限制，则跳过
					if err == excelize.ErrMaxFilePathLength {
						msg = "文件存放的路径过长，建议将软件放在磁盘一级目录再操作。"
					}
					systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 添加错误信息失败: %s", file.Name(), msg)})
					continue
				}
				failedFiles = append(failedFiles, filePath)
				// 将验证错误转换为字符串用于显示
				var errorMessages []string
				for _, err := range errors {
					errorMessages = append(errorMessages, err.Message)
				}
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s: %s", file.Name(), strings.Join(errorMessages, "; "))})
				continue
			}

			// 5. 校验通过后,检查文件是否已导入
			if s.isTable2FileImported(mainData) {
				coverFiles = append(coverFiles, filePath)
				continue
			}

			// 6. 如果没有导入过,把数据保存到相应的数据库表中
			err = s.saveTable2Data(mainData)
			if err != nil {
				systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 保存数据失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
			} else {
				// 删除该Excel文件
				os.Remove(filePath)
				importedFiles = append(importedFiles, file.Name())
			}
		}
	}

	if !hasExcelFile {
		result = db.QueryResult{
			Ok:      false,
			Message: "没有待校验Excel文件，请先进行数据导入",
		}
		return result
	}

	// 7. 把所有的模型验证失败的文件打个zip包
	if len(failedFiles) > 0 {
		err = s.createValidationErrorZip(failedFiles, TableType2, TableName2)
		if err != nil {
			systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("创建错误报告失败: %v", err)})
		}

		// 删除失败文件
		for _, filePath := range failedFiles {
			os.Remove(filePath)
		}
	}

	// 8. 返回结果
	message := fmt.Sprintf("处理完成。成功导入: %d 个文件，失败: %d 个文件", len(importedFiles), len(failedFiles))
	if len(systemErrors) > 0 {
		// 将验证错误转换为字符串用于显示
		var errorMessages []string
		for _, err := range systemErrors {
			errorMessages = append(errorMessages, err.Message)
		}
		message += "。错误信息如下：\n\n" + strings.Join(errorMessages, ";\n\n")
	} else if len(validationErrors) > 0 {
		message += "。详细错误信息请查看生成的错误报告。"
	}

	result = db.QueryResult{
		Ok:      true,
		Message: message,
		Data: map[string]interface{}{
			"cover_files":     coverFiles,                // 覆盖的文件
			"hasExportReport": len(validationErrors) > 0, // 是否有导出报告
			"hasFailedFiles":  len(failedFiles) > 0,      // 是否有失败的文件
		},
	}
	return result
}

// validateTable2DataForModel 校验附表2数据（模型校验专用）
func (s *DataImportService) validateTable2DataForModel(mainData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	// 逐行校验数值字段
	for _, data := range mainData {
		// 获取记录的实际Excel行号
		excelRowNum := s.getExcelRowNumber(data)

		// 数值字段校验
		valueErrors := s.validateTable2NumericFieldsForModel(data, excelRowNum)
		errors = append(errors, valueErrors...)
	}

	return errors
}

// validateTable2NumericFieldsForModel 校验附表2数值字段（模型校验专用）
func (s *DataImportService) validateTable2NumericFieldsForModel(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 1. 累计使用时间、设计年限校验
	// 应为0-50（含0和50）间的整数
	totalRuntime := s.parseFloat(s.getStringValue(data["usage_time"]))
	designLife := s.parseFloat(s.getStringValue(data["design_life"]))

	if s.isIntegerLessThan(totalRuntime, 0) || s.isIntegerGreaterThan(totalRuntime, 50) {
		cells := []string{s.getCellPosition(TableType2, "usage_time", rowNum)}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "累计使用时间应在0-50之间",
			Cells:     cells,
		})
	}
	if !s.isIntegerInteger(totalRuntime) {
		cells := []string{s.getCellPosition(TableType2, "usage_time", rowNum)}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "累计使用时间应为整数",
			Cells:     cells,
		})
	}

	if s.isIntegerLessThan(designLife, 0) || s.isIntegerGreaterThan(designLife, 50) {
		cells := []string{s.getCellPosition(TableType2, "design_life", rowNum)}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "设计年限应在0-50之间",
			Cells:     cells,
		})
	}
	if !s.isIntegerInteger(designLife) {
		cells := []string{s.getCellPosition(TableType2, "design_life", rowNum)}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "设计年限应为整数",
			Cells:     cells,
		})
	}

	// 2. 容量校验
	// 应为正整数
	capacity := s.parseFloat(s.getStringValue(data["capacity"]))
	if s.isIntegerLessThan(capacity, 0) {
		cells := []string{s.getCellPosition(TableType2, "capacity", rowNum)}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "容量不能为负数",
			Cells:     cells,
		})
	}
	if !s.isIntegerInteger(capacity) {
		cells := []string{s.getCellPosition(TableType2, "capacity", rowNum)}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "容量应为整数",
			Cells:     cells,
		})
	}

	// 3. 年耗煤量校验
	// ≧0且≦1000000000
	annualCoalConsumption := s.parseFloat(s.getStringValue(data["annual_coal_consumption"]))
	if s.isIntegerLessThan(annualCoalConsumption, 0) {
		cells := []string{s.getCellPosition(TableType2, "annual_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年耗煤量不能为负数",
			Cells:     cells,
		})
	}
	if s.isIntegerGreaterThan(annualCoalConsumption, 1000000000) {
		cells := []string{s.getCellPosition(TableType2, "annual_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年耗煤量不能大于1000000000",
			Cells:     cells,
		})
	}

	return errors
}

// coverTable2Data 覆盖附表2数据
func (s *DataImportService) coverTable2Data(mainData []map[string]interface{}, fileName string) error {
	if len(mainData) == 0 {
		return fmt.Errorf("数据为空")
	}

	// 获取统一信用代码和年份
	creditCode := s.getStringValue(mainData[0]["credit_code"])
	statDate := s.getStringValue(mainData[0]["stat_date"])

	// 根据年份+统一信用代码删除表数据
	err := s.deleteTable2DataByCreditCodeAndYear(creditCode, statDate)
	if err != nil {
		return fmt.Errorf("删除旧数据失败: %v", err)
	}

	// 插入新数据
	return s.saveTable2Data(mainData)
}

// deleteTable2DataByCreditCodeAndYear 根据统一信用代码和年份删除附表2数据
func (s *DataImportService) deleteTable2DataByCreditCodeAndYear(creditCode, statDate string) error {
	query := "DELETE FROM critical_coal_equipment_consumption WHERE credit_code = ? AND stat_date = ?"
	_, err := s.app.GetDB().Exec(query, creditCode, statDate)
	return err
}

// isTable2FileImported 检查附表2文件是否已导入
func (s *DataImportService) isTable2FileImported(mainData []map[string]interface{}) bool {
	if len(mainData) == 0 {
		return false
	}

	creditCode := s.getStringValue(mainData[0]["credit_code"])
	statDate := s.getStringValue(mainData[0]["stat_date"])

	query := "SELECT COUNT(1) as count FROM critical_coal_equipment_consumption WHERE credit_code = ? AND stat_date = ?"
	result, err := s.app.GetDB().QueryRow(query, creditCode, statDate)
	if err != nil || result.Data == nil {
		return false
	}

	return result.Data.(map[string]interface{})["count"].(int64) > 0
}

// saveTable2Data 保存附表2数据到数据库
func (s *DataImportService) saveTable2Data(mainData []map[string]interface{}) error {
	for _, record := range mainData {
		record["obj_id"] = s.generateUUID()
		record["create_time"] = time.Now().UnixMilli()

		// 对数值字段进行SM4加密
		encryptedValues := s.encryptTable2NumericFields(record)

		query := `INSERT INTO critical_coal_equipment_consumption (
			obj_id, stat_date, create_time, unit_name, credit_code, trade_a, trade_b, trade_c,
			province_name, city_name, country_name, coal_type, coal_no, usage_time, design_life,
			enecrgy_efficienct_bmk, capacity_unit, capacity, use_info, status, annual_coal_consumption, create_user, row_no, is_check
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err := s.app.GetDB().Exec(query,
			record["obj_id"], record["stat_date"], record["create_time"], record["unit_name"],
			record["credit_code"], record["trade_a"], record["trade_b"], record["trade_c"],
			record["province_name"], record["city_name"], record["country_name"], record["coal_type"],
			record["coal_no"], record["usage_time"], encryptedValues["design_life"], record["enecrgy_efficienct_bmk"],
			record["capacity_unit"], encryptedValues["capacity"], record["use_info"], record["status"],
			encryptedValues["annual_coal_consumption"], s.app.GetAreaStr(), record["row_no"], EncryptedOne)
		if err != nil {
			return fmt.Errorf("保存数据失败: %v", err)
		}
	}

	return nil
}

// addValidationErrorsToExcelTable2 在附表2Excel文件中添加校验错误信息
func (s *DataImportService) addValidationErrorsToExcelTable2(filePath string, errors []ValidationError) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 创建错误信息映射
	errorMap := make(map[int]string)
	// 收集所有需要高亮的单元格
	var allCells []string

	for _, err := range errors {
		// 如果该行已有错误信息，则追加
		if existing, exists := errorMap[err.RowNumber]; exists {
			errorMap[err.RowNumber] = existing + "; " + err.Message
		} else {
			errorMap[err.RowNumber] = err.Message
		}

		// 收集涉及到的单元格
		if err.Cells != nil {
			allCells = append(allCells, err.Cells...)
		}
	}

	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("Excel文件没有工作表")
	}

	// 处理第一个工作表
	sheetName := sheets[0]

	// 高亮涉及到的单元格
	if len(allCells) > 0 {
		err = s.highlightCellsInExcel(f, sheetName, allCells)
		if err != nil {
			fmt.Printf("高亮单元格失败: %v\n", err)
		}
	}

	maxCol := 11

	// 为每个错误行添加错误信息
	for excelRow, errorMsg := range errorMap {

		// 在最后一列添加错误信息
		errorCol := maxCol + 1
		errorCellName, err := excelize.CoordinatesToCellName(errorCol, excelRow)
		if err != nil {
			continue
		}

		// 格式化错误信息：每条错误使用序号标识并换行
		formattedErrorMsg := formatErrorMessages(errorMsg)
		f.SetCellValue(sheetName, errorCellName, formattedErrorMsg)

		style, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFF00"}, Pattern: 1},
			Alignment: &excelize.Alignment{
				Vertical: "center",
			},
		})
		if err != nil {
			fmt.Println(err)
		}

		f.SetCellStyle(sheetName, errorCellName, errorCellName, style)

		// 设置错误信息列的宽度为50
		colName, _ := excelize.ColumnNumberToName(errorCol)
		f.SetColWidth(sheetName, colName, colName, 50)
	}

	// 保存文件
	return f.Save()
}

// encryptTable2NumericFields 加密附表2数值字段
func (s *DataImportService) encryptTable2NumericFields(record map[string]interface{}) map[string]interface{} {
	numericFields := []string{"annual_coal_consumption", "design_life", "capacity"}
	return s.encryptNumericFields(record, numericFields)
}
