package data_import

import (
	"fmt"
	"os"
	"path/filepath"
	"shuji/db"
	"slices"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ModelDataCheckTable3 附表3模型校验函数
func (s *DataImportService) ModelDataCheckTable3() db.QueryResult {
	// 1. 读取缓存目录指定表格类型下的所有Excel文件
	cacheDir := s.app.GetCachePath(TableType3)

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取缓存目录失败: %v", err),
		}
	}

	var validationErrors []ValidationError
	var importedFiles []string
	var failedFiles []string

	// 2. 循环调用对应的解析Excel函数
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".xlsx") {
			filePath := filepath.Join(cacheDir, file.Name())

			// 解析Excel文件 (skipValidate=true)
			f, err := excelize.OpenFile(filePath)
			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 读取失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			mainData, err := s.parseTable3Excel(f, true)
			f.Close()

			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			// 4. 调用校验函数,对每一行数据验证
			errors := s.validateTable3DataForModel(mainData)
			if len(errors) > 0 {
				// 校验失败，在Excel文件中错误行最后添加错误信息
				err = s.addValidationErrorsToExcelTable3(filePath, errors, mainData)
				if err != nil {
					validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 添加错误信息失败: %v", file.Name(), err)})
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
			if s.isTable3FileImported(mainData) {
				importedFiles = append(importedFiles, file.Name())
				continue
			}

			// 6. 如果没有导入过,把数据保存到相应的数据库表中
			err = s.saveTable3Data(mainData)
			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 保存数据失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
			} else {
				// 删除该Excel文件
				os.Remove(filePath)
				importedFiles = append(importedFiles, file.Name())
			}
		}
	}

	// 7. 把所有的模型验证失败的文件打个zip包
	if len(failedFiles) > 0 {
		err = s.createValidationErrorZip(failedFiles, TableType3, TableName3)
		if err != nil {
			validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("创建错误报告失败: %v", err)})
		}
	}

	// 8. 返回结果
	message := fmt.Sprintf("处理完成。成功导入: %d 个文件，失败: %d 个文件", len(importedFiles), len(failedFiles))
	if len(validationErrors) > 0 {
		message += "。详细错误信息请查看生成的错误报告。"
	}

	return db.QueryResult{
		Ok:      len(failedFiles) == 0,
		Message: message,
		Data: map[string]interface{}{
			"imported_files": importedFiles,
			"failed_files":   failedFiles,
			"errors":         validationErrors,
		},
	}
}

// validateTable3DataForModel 校验附表3数据（模型校验专用）
func (s *DataImportService) validateTable3DataForModel(mainData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	for _, data := range mainData {
		// 获取记录的实际Excel行号
		excelRowNum := s.getExcelRowNumber(data)

		// 数值字段校验
		valueErrors := s.validateTable3NumericFields(data, excelRowNum)
		errors = append(errors, valueErrors...)
	}

	return errors
}

// validateTable3NumericFields 校验附表3数值字段
func (s *DataImportService) validateTable3NumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 校验当量值范围 (>= 0)
	if equivalentValue, ok := data["equivalent_value"].(string); ok && equivalentValue != "" {
		if value, err := s.parseFloat(equivalentValue); err == nil {
			if value < 0 {
				errors = append(errors, ValidationError{RowNumber: rowNum, Message: "当量值不能为负数"})
			}
		}
	}

	// 校验等价值范围 (>= 0)
	if equivalentCost, ok := data["equivalent_cost"].(string); ok && equivalentCost != "" {
		if value, err := s.parseFloat(equivalentCost); err == nil {
			if value < 0 {
				errors = append(errors, ValidationError{RowNumber: rowNum, Message: "等价值不能为负数"})
			}
		}
	}

	return errors
}

// coverTable3Data 覆盖附表3数据
func (s *DataImportService) coverTable3Data(mainData []map[string]interface{}, fileName string) error {
	if len(mainData) == 0 {
		return fmt.Errorf("数据为空")
	}

	// 逐行检查，根据项目代码+建设单位检查是否已导入
	for _, record := range mainData {
		projectCode := s.getStringValue(record["project_code"])
		constructionUnit := s.getStringValue(record["construction_unit"])

		// 根据项目代码+建设单位做where条件更新数据
		err := s.updateTable3DataByProjectCodeAndUnit(projectCode, constructionUnit, record)
		if err != nil {
			return fmt.Errorf("更新数据失败: %v", err)
		}
	}

	return nil
}

// updateTable3DataByProjectCodeAndUnit 根据项目代码和建设单位更新附表3数据
func (s *DataImportService) updateTable3DataByProjectCodeAndUnit(projectCode, constructionUnit string, record map[string]interface{}) error {
	// 先删除旧数据
	query := "DELETE FROM fixed_assets_investment_project WHERE project_code = ? AND construction_unit = ?"
	_, err := s.app.GetDB().Exec(query, projectCode, constructionUnit)
	if err != nil {
		return err
	}

	// 插入新数据
	return s.insertTable3Data(record)
}

// isTable3FileImported 检查附表3文件是否已导入
func (s *DataImportService) isTable3FileImported(mainData []map[string]interface{}) bool {
	// 按Excel数据逐行检查，根据项目代码+建设单位检查是否已导入
	for _, record := range mainData {
		projectCode := s.getStringValue(record["project_code"])
		constructionUnit := s.getStringValue(record["construction_unit"])

		query := "SELECT COUNT(*) as count FROM fixed_assets_investment_project WHERE project_code = ? AND construction_unit = ?"
		result, err := s.app.GetDB().Query(query, projectCode, constructionUnit)
		if err != nil {
			continue
		}

		if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
			if count, ok := data[0]["count"].(int64); ok {
				if count > 0 {
					return true // 检查到立即停止表示已导入
				}
			}
		}
	}

	return false
}

// saveTable3Data 保存附表3数据到数据库
func (s *DataImportService) saveTable3Data(mainData []map[string]interface{}) error {
	for _, record := range mainData {
		err := s.insertTable3Data(record)
		if err != nil {
			return err
		}
	}

	return nil
}

// insertTable3Data 插入附表3数据
func (s *DataImportService) insertTable3Data(record map[string]interface{}) error {
	record["obj_id"] = s.generateUUID()
	record["create_time"] = time.Now().Format("2006-01-02 15:04:05")

	// 对数值字段进行SM4加密
	encryptedValues := s.encryptTable3NumericFields(record)

	query := `INSERT INTO fixed_assets_investment_project (
		obj_id, stat_date, project_name, project_code, construction_unit, main_construction_content,
		province_name, city_name, country_name, trade_a, trade_c, examination_approval_time,
		scheduled_time, actual_time, examination_authority, document_number, equivalent_value,
		equivalent_cost, create_time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.app.GetDB().Exec(query,
		record["obj_id"], record["stat_date"], record["project_name"], record["project_code"],
		record["construction_unit"], record["main_construction_content"], record["province_name"],
		record["city_name"], record["country_name"], record["trade_a"], record["trade_c"],
		record["examination_approval_time"], record["scheduled_time"], record["actual_time"],
		record["examination_authority"], record["document_number"], encryptedValues["equivalent_value"],
		encryptedValues["equivalent_cost"], record["create_time"])
	if err != nil {
		return fmt.Errorf("保存数据失败: %v", err)
	}

	return nil
}

// addValidationErrorsToExcelTable3 在附表3Excel文件中添加校验错误信息
func (s *DataImportService) addValidationErrorsToExcelTable3(filePath string, errors []ValidationError, mainData []map[string]interface{}) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 创建错误信息映射
	errorMap := make(map[int]string)
	for _, err := range errors {
		// 如果该行已有错误信息，则追加
		if existing, exists := errorMap[err.RowNumber]; exists {
			errorMap[err.RowNumber] = existing + "; " + err.Message
		} else {
			errorMap[err.RowNumber] = err.Message
		}
	}

	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("Excel文件没有工作表")
	}

	// 处理第一个工作表
	sheetName := sheets[0]

	// 获取最大列数
	cols, err := f.GetCols(sheetName)
	if err != nil {
		return err
	}

	maxCol := len(cols)
	if maxCol == 0 {
		return fmt.Errorf("工作表为空")
	}

	// 为每个错误行添加错误信息
	for rowNum, errorMsg := range errorMap {
		// Excel行号从1开始，且需要加上标题行偏移
		excelRow := rowNum + 6 // 附表3通常从第7行开始有数据

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

// encryptTable3NumericFields 加密附表3数值字段
func (s *DataImportService) encryptTable3NumericFields(record map[string]interface{}) map[string]interface{} {
	numericFields := []string{
		"equivalent_value", "equivalent_cost",
	}
	return s.encryptNumericFields(record, numericFields)
}

// ModelDataCoverTable3 覆盖附表3数据
func (s *DataImportService) ModelDataCoverTable3(fileNames []string) db.QueryResult {
	cacheDir := s.app.GetCachePath(TableType3)
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
			if !slices.Contains(fileNames, file.Name()) {
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

			mainData, err := s.parseTable3Excel(f, true)
			f.Close()
			os.Remove(filePath)

			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			err = s.coverTable3Data(mainData, file.Name())
			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 覆盖数据失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
			}
		}
	}

	return db.QueryResult{
		Ok:      true,
		Message: "覆盖完成",
		Data: map[string]interface{}{
			"failed_files": failedFiles,      // 失败的文件
			"errors":       validationErrors, // 错误信息
		},
	}
}
