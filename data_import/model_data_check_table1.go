package data_import

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"shuji/db"
	"slices"
	"strings"
	"time"

	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

// ModelDataCheckReportDownload 下载模型校验结果
func (s *DataImportService) ModelDataCheckReportDownload(tableType string) db.QueryResult {
	// 使用包装函数来处理异常
	return s.modelDataCheckReportDownloadWithRecover(tableType)
}

// modelDataCheckReportDownloadWithRecover 带异常处理的下载模型校验结果函数
func (s *DataImportService) modelDataCheckReportDownloadWithRecover(tableType string) db.QueryResult {
	var result db.QueryResult

	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ModelDataCheckReportDownload 发生异常: %v", r)
			// 设置错误结果
			result = db.QueryResult{
				Ok:      false,
				Message: fmt.Sprintf("函数执行异常: %v", r),
				Data:    nil,
			}
		}
	}()

	cacheDir := s.app.GetCachePath(tableType)
	// 创建ZIP文件
	zipFileName := fmt.Sprintf("%s模型报告.zip", s.getTableName(tableType))
	zipFilePath := filepath.Join(cacheDir, zipFileName)

	selectPath, err := runtime.SaveFileDialog(s.app.GetCtx(), runtime.SaveDialogOptions{
		Title:           "下载模型报告",
		DefaultFilename: zipFileName,
		Filters: []runtime.FileFilter{
			{
				Pattern: "*.zip",
			},
		},
	})

	if err != nil {
		result = db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("下载模型报告失败: %v", err),
		}
		return result
	}

	if selectPath == "" {
		result = db.QueryResult{
			Ok:      false,
			Message: "下载模型报告失败: 未选择文件",
		}
		return result
	}

	dstFile, err := os.OpenFile(selectPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		result = db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("下载模型校验结果失败: %v", err),
		}
		return result
	}
	defer dstFile.Close()

	srcFile, err := os.Open(zipFilePath)
	if err != nil {
		result = db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("下载模型校验结果失败: %v", err),
		}
		return result
	}
	defer srcFile.Close()

	// 这里的复制是覆盖复制
	_, err = io.Copy(dstFile, srcFile)

	if err != nil {
		result = db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("下载模型校验结果失败: %v", err),
		}
		return result
	}

	result = db.QueryResult{
		Ok:      true,
		Message: "下载模型校验结果成功",
	}
	return result

}

// ModelDataCoverTable1 覆盖附表1数据
func (s *DataImportService) ModelDataCoverTable1(filePaths []string) db.QueryResult {
	// 使用包装函数来处理异常
	return s.modelDataCoverTable1WithRecover(filePaths)
}

// modelDataCoverTable1WithRecover 带异常处理的覆盖附表1数据函数
func (s *DataImportService) modelDataCoverTable1WithRecover(filePaths []string) db.QueryResult {
	var result db.QueryResult

	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ModelDataCoverTable1 发生异常: %v", r)
			// 设置错误结果
			result = db.QueryResult{
				Ok:      false,
				Message: fmt.Sprintf("函数执行异常: %v", r),
				Data:    nil,
			}
		}
	}()

	cacheDir := s.app.GetCachePath(TableType1)
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

			mainData, usageData, equipData, err := s.parseTable1Excel(f, true)
			f.Close()
			os.Remove(filePath)

			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			err = s.coverTable1Data(mainData, usageData, equipData, file.Name())
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

// ModelDataCheckTable1 附表1模型校验函数
func (s *DataImportService) ModelDataCheckTable1() db.QueryResult {
	// 使用包装函数来处理异常
	return s.modelDataCheckTable1WithRecover()
}

// modelDataCheckTable1WithRecover 带异常处理的附表1模型校验函数
func (s *DataImportService) modelDataCheckTable1WithRecover() db.QueryResult {
	var result db.QueryResult

	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ModelDataCheckTable1 发生异常: %v", r)
			// 设置错误结果
			result = db.QueryResult{
				Ok:      false,
				Message: fmt.Sprintf("函数执行异常: %v", r),
				Data:    nil,
			}
		}
	}()

	// 1. 读取缓存目录指定表格类型下的所有Excel文件
	cacheDir := s.app.GetCachePath(TableType1)

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取缓存目录失败: %v", err),
		}
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

			mainData, usageData, equipData, err := s.parseTable1Excel(f, true)
			f.Close()

			if err != nil {
				systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			// 4. 调用校验函数,对每一行数据验证
			errors := s.validateTable1DataWithEnterpriseCheckForModel(mainData, usageData, equipData)
			if len(errors) > 0 {
				// 校验失败，在Excel文件中错误行最后添加错误信息
				err = s.addValidationErrorsToExcelTable1(filePath, errors)

				if err != nil {
					msg := err.Error()
					// 如果错误是文件名长度超出限制，则跳过
					if err == excelize.ErrMaxFilePathLength {
						msg = "软件存放的路径过长，建议将软件放在磁盘一级目录再操作。"
					}
					systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 添加错误信息失败: %s", file.Name(), msg)})
					failedFiles = append(failedFiles, filePath)
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
			if s.isTable1FileImported(mainData) {
				coverFiles = append(coverFiles, filePath)
				continue
			}

			// 6. 如果没有导入过,把数据保存到相应的数据库表中
			err = s.saveTable1Data(mainData, usageData, equipData)
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
		err = s.createValidationErrorZip(failedFiles, TableType1, TableName1)
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

// coverTable1Data 覆盖附表1数据
func (s *DataImportService) coverTable1Data(mainData, usageData, equipData []map[string]interface{}, fileName string) error {
	if len(mainData) == 0 {
		return fmt.Errorf("主表数据为空")
	}

	// 获取统一信用代码和年份
	creditCode := s.getStringValue(mainData[0]["credit_code"])
	statDate := s.getStringValue(mainData[0]["stat_date"])

	// 根据年份+统一信用代码删除表数据
	err := s.deleteTable1DataByCreditCodeAndYear(creditCode, statDate)
	if err != nil {
		return fmt.Errorf("删除旧数据失败: %v", err)
	}

	// 插入新数据
	return s.saveTable1Data(mainData, usageData, equipData)
}

// deleteTable1DataByCreditCodeAndYear 根据统一信用代码和年份删除附表1数据
func (s *DataImportService) deleteTable1DataByCreditCodeAndYear(creditCode, statDate string) error {
	// 先查出一条主表记录，获取obj_id，扩展表的fk_id就是obj_id
	var objID string
	query := "SELECT obj_id FROM enterprise_coal_consumption_main WHERE credit_code = ? AND stat_date = ? LIMIT 1"
	result, err := s.app.GetDB().QueryRow(query, creditCode, statDate)
	if err != nil || result.Data == nil {
		return err
	}

	objID = result.Data.(map[string]interface{})["obj_id"].(string)

	// 如果查不到obj_id，说明主表没有数据，直接返回
	if objID == "" {
		return nil
	}

	// 删除扩展表数据
	query = "DELETE FROM enterprise_coal_consumption_usage WHERE fk_id = ?"
	_, err = s.app.GetDB().Exec(query, objID)
	if err != nil {
		return err
	}

	query = "DELETE FROM enterprise_coal_consumption_equip WHERE fk_id = ?"
	_, err = s.app.GetDB().Exec(query, objID)
	if err != nil {
		return err
	}

	// 最后根据obj_id删除主表数据
	query = "DELETE FROM enterprise_coal_consumption_main WHERE obj_id = ?"
	_, err = s.app.GetDB().Exec(query, objID)
	if err != nil {
		return err
	}

	return nil
}

// isTable1FileImported 检查附表1文件是否已导入
func (s *DataImportService) isTable1FileImported(mainData []map[string]interface{}) bool {
	if len(mainData) == 0 {
		return false
	}

	creditCode := s.getStringValue(mainData[0]["credit_code"])
	statDate := s.getStringValue(mainData[0]["stat_date"])

	query := "SELECT COUNT(1) as count FROM enterprise_coal_consumption_main WHERE credit_code = ? AND stat_date = ?"
	result, err := s.app.GetDB().QueryRow(query, creditCode, statDate)
	if err != nil || result.Data == nil {
		return false
	}

	return result.Data.(map[string]interface{})["count"].(int64) > 0
}

// saveTable1Data 保存附表1数据到数据库
func (s *DataImportService) saveTable1Data(mainData, usageData, equipData []map[string]interface{}) error {
	if len(mainData) == 0 {
		return fmt.Errorf("主表数据为空")
	}

	objID := s.generateUUID()
	createTime := time.Now().UnixMilli()
	// 保存主表数据
	mainRecord := mainData[0]
	mainRecord["obj_id"] = objID
	mainRecord["create_time"] = createTime

	// 对数值字段进行SM4加密
	encryptedValues := s.encryptTable1MainNumericFields(mainRecord)

	query := `INSERT INTO enterprise_coal_consumption_main (
		obj_id, unit_name, stat_date, tel, credit_code, create_time, trade_a, trade_b, trade_c,
		province_name, city_name, country_name, annual_energy_equivalent_value, annual_energy_equivalent_cost,
		annual_raw_material_energy, annual_total_coal_consumption, annual_total_coal_products,
		annual_raw_coal, annual_raw_coal_consumption, annual_clean_coal_consumption,
		annual_other_coal_consumption, annual_coke_consumption, create_user, is_check
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.app.GetDB().Exec(query,
		mainRecord["obj_id"], mainRecord["unit_name"], mainRecord["stat_date"], mainRecord["tel"],
		mainRecord["credit_code"], mainRecord["create_time"], mainRecord["trade_a"], mainRecord["trade_b"],
		mainRecord["trade_c"], mainRecord["province_name"], mainRecord["city_name"], mainRecord["country_name"],
		encryptedValues["annual_energy_equivalent_value"], encryptedValues["annual_energy_equivalent_cost"],
		encryptedValues["annual_raw_material_energy"], encryptedValues["annual_total_coal_consumption"],
		encryptedValues["annual_total_coal_products"], encryptedValues["annual_raw_coal"], encryptedValues["annual_raw_coal_consumption"],
		encryptedValues["annual_clean_coal_consumption"], encryptedValues["annual_other_coal_consumption"],
		encryptedValues["annual_coke_consumption"], s.app.GetAreaStr(), EncryptedOne)
	if err != nil {
		return fmt.Errorf("保存主表数据失败: %v", err)
	}

	// 保存用途数据
	for _, usage := range usageData {
		usage["obj_id"] = s.generateUUID()
		usage["fk_id"] = objID
		usage["stat_date"] = mainRecord["stat_date"]
		usage["create_time"] = createTime

		// 对数值字段进行SM4加密
		encryptedUsageValues := s.encryptTable1UsageNumericFields(usage)

		query := `INSERT INTO enterprise_coal_consumption_usage (
			obj_id, fk_id, stat_date, create_time, main_usage, specific_usage, input_variety,
			input_unit, input_quantity, output_energy_types, output_quantity, measurement_unit, remarks, row_no, is_check
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err := s.app.GetDB().Exec(query,
			usage["obj_id"], usage["fk_id"], usage["stat_date"], usage["create_time"],
			usage["main_usage"], usage["specific_usage"], usage["input_variety"], usage["input_unit"],
			encryptedUsageValues["input_quantity"], usage["output_energy_types"], encryptedUsageValues["output_quantity"],
			usage["measurement_unit"], usage["remarks"], usage["row_no"], EncryptedOne)
		if err != nil {
			return fmt.Errorf("保存用途数据失败: %v", err)
		}
	}

	// 保存设备数据
	for _, equip := range equipData {
		equip["obj_id"] = s.generateUUID()
		equip["fk_id"] = objID
		equip["stat_date"] = mainRecord["stat_date"]
		equip["create_time"] = createTime

		// 对数值字段进行SM4加密
		encryptedEquipValues := s.encryptTable1EquipNumericFields(equip)

		query := `INSERT INTO enterprise_coal_consumption_equip (
			obj_id, fk_id, stat_date, create_time, equip_type, equip_no, total_runtime,
			design_life, energy_efficiency, capacity_unit, capacity, coal_type, annual_coal_consumption, row_no, is_check
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err := s.app.GetDB().Exec(query,
			equip["obj_id"], equip["fk_id"], equip["stat_date"], equip["create_time"],
			equip["equip_type"], equip["equip_no"], encryptedEquipValues["total_runtime"], encryptedEquipValues["design_life"],
			encryptedEquipValues["energy_efficiency"], equip["capacity_unit"], encryptedEquipValues["capacity"], equip["coal_type"],
			encryptedEquipValues["annual_coal_consumption"], equip["row_no"], EncryptedOne)
		if err != nil {
			return fmt.Errorf("保存设备数据失败: %v", err)
		}
	}

	return nil
}

// addValidationErrorsToExcel 在Excel文件中添加校验错误信息
func (s *DataImportService) addValidationErrorsToExcelTable1(filePath string, errors []ValidationError) error {
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

	maxCol := 10

	// 为每个错误行添加错误信息
	for rowNum, errorMsg := range errorMap {
		// 在最后一列添加错误信息
		errorCol := maxCol + 1
		errorCellName, err := excelize.CoordinatesToCellName(errorCol, rowNum)
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

// validateTable1DataWithEnterpriseCheckForModel 校验附表1数据（模型校验专用）
func (s *DataImportService) validateTable1DataWithEnterpriseCheckForModel(mainData, usageData, equipData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	// 校验主表数据
	for _, data := range mainData {
		// 获取记录的实际Excel行号,使用第二部分的excel行号
		excelRowNum := data["_excel_row2"].(int)

		// 校验基本信息表格的数值字段
		valueErrors := s.validateTable1MainNumericFields(data, excelRowNum)
		errors = append(errors, valueErrors...)

		// 校验综合能源消费量与煤炭消费量间的关系
		relationErrors := s.validateTable1EnergyCoalRelation(data, excelRowNum)
		errors = append(errors, relationErrors...)
	}

	// 校验用途数据
	for _, data := range usageData {
		// 获取记录的实际Excel行号
		excelRowNum := s.getExcelRowNumber(data)

		// 校验用途表格的数值字段
		valueErrors := s.validateTable1UsageNumericFields(data, excelRowNum)
		errors = append(errors, valueErrors...)
	}

	// 校验设备数据
	for _, data := range equipData {
		// 获取记录的实际Excel行号
		excelRowNum := s.getExcelRowNumber(data)

		// 校验设备表格的数值字段
		valueErrors := s.validateTable1EquipNumericFields(data, excelRowNum)
		errors = append(errors, valueErrors...)
	}

	return errors
}

// validateTable1MainNumericFields 校验附表1主表数值字段
func (s *DataImportService) validateTable1MainNumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 1. 年综合能耗当量值、年综合能耗等价值、年原料用能消费量校验
	annualEnergyEquivalentValue := s.parseFloat(s.getStringValue(data["annual_energy_equivalent_value"]))
	annualEnergyEquivalentCost := s.parseFloat(s.getStringValue(data["annual_energy_equivalent_cost"]))
	annualRawMaterialEnergy := s.parseFloat(s.getStringValue(data["annual_raw_material_energy"]))

	// ①≧0
	if s.isIntegerLessThan(annualEnergyEquivalentValue, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_energy_equivalent_value", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗当量值不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(annualEnergyEquivalentCost, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_energy_equivalent_cost", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗等价值不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(annualRawMaterialEnergy, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_raw_material_energy", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用能消费量不能为负数", Cells: cells})
	}

	// ②≦100000
	if s.isIntegerGreaterThan(annualEnergyEquivalentValue, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_energy_equivalent_value", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗当量值不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(annualEnergyEquivalentCost, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_energy_equivalent_cost", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗等价值不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(annualRawMaterialEnergy, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_raw_material_energy", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用能消费量不能大于100000", Cells: cells})
	}

	// ③年原料用能消费量≦年综合能耗当量值
	if s.isIntegerGreaterThan(annualRawMaterialEnergy, annualEnergyEquivalentValue) {
		cells := []string{
			s.getCellPosition(TableType1, "annual_raw_material_energy", rowNum),
			s.getCellPosition(TableType1, "annual_energy_equivalent_value", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用能消费量不能大于年综合能耗当量值", Cells: cells})
	}

	// ④年原料用能消费量≦年综合能耗等价值
	if s.isIntegerGreaterThan(annualRawMaterialEnergy, annualEnergyEquivalentCost) {
		cells := []string{
			s.getCellPosition(TableType1, "annual_raw_material_energy", rowNum),
			s.getCellPosition(TableType1, "annual_energy_equivalent_cost", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用能消费量不能大于年综合能耗等价值", Cells: cells})
	}

	// 2. 煤炭消费相关字段校验
	annualTotalCoalConsumption := s.parseFloat(s.getStringValue(data["annual_total_coal_consumption"]))
	annualTotalCoalProducts := s.parseFloat(s.getStringValue(data["annual_total_coal_products"]))
	annualRawCoal := s.parseFloat(s.getStringValue(data["annual_raw_coal"]))
	annualRawCoalConsumption := s.parseFloat(s.getStringValue(data["annual_raw_coal_consumption"]))
	annualCleanCoalConsumption := s.parseFloat(s.getStringValue(data["annual_clean_coal_consumption"]))
	annualOtherCoalConsumption := s.parseFloat(s.getStringValue(data["annual_other_coal_consumption"]))
	annualCokeConsumption := s.parseFloat(s.getStringValue(data["annual_coke_consumption"]))

	// ①≧0
	if s.isIntegerLessThan(annualTotalCoalConsumption, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_total_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（实物量）不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(annualTotalCoalProducts, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_total_coal_products", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（标准量）不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(annualRawCoal, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_raw_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原料用煤（实物量）不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(annualRawCoalConsumption, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_raw_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤消费（实物量）不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(annualCleanCoalConsumption, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_clean_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤消费（实物量）不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(annualOtherCoalConsumption, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_other_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他煤炭消费（实物量）不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(annualCokeConsumption, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费（实物量）不能为负数", Cells: cells})
	}

	// ②≦100000
	if s.isIntegerGreaterThan(annualTotalCoalConsumption, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_total_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（实物量）不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(annualTotalCoalProducts, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_total_coal_products", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（标准量）不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(annualRawCoal, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_raw_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原料用煤（实物量）不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(annualRawCoalConsumption, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_raw_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤消费（实物量）不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(annualCleanCoalConsumption, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_clean_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤消费（实物量）不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(annualOtherCoalConsumption, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_other_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他煤炭消费（实物量）不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(annualCokeConsumption, 100000) {
		cells := []string{s.getCellPosition(TableType1, "annual_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费（实物量）不能大于100000", Cells: cells})
	}

	// ③耗煤总量（实物量）≧耗煤总量（标准量）
	if s.isIntegerLessThan(annualTotalCoalConsumption, annualTotalCoalProducts) {
		// 获取涉及到的单元格位置
		cells := []string{
			s.getCellPosition(TableType1, "annual_total_coal_consumption", rowNum),
			s.getCellPosition(TableType1, "annual_total_coal_products", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "耗煤总量（实物量）不能小于耗煤总量（标准量）",
			Cells:     cells,
		})
	}

	// ④耗煤总量（实物量）≧原料用煤（实物量）
	if s.isIntegerLessThan(annualTotalCoalConsumption, annualRawCoal) {
		// 获取涉及到的单元格位置
		cells := []string{
			s.getCellPosition(TableType1, "annual_total_coal_consumption", rowNum),
			s.getCellPosition(TableType1, "annual_raw_coal", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "耗煤总量（实物量）不能小于原料用煤（实物量）",
			Cells:     cells,
		})
	}

	// ⑤耗煤总量（实物量）=原煤消费（实物量）+洗精煤消费（实物量）+其他煤炭消费（实物量）
	// 使用精度安全的浮点运算
	expectedTotal := s.sumFloat64(annualRawCoalConsumption, annualCleanCoalConsumption, annualOtherCoalConsumption)

	if !s.isIntegerEqual(annualTotalCoalConsumption, expectedTotal) {
		// 获取涉及到的单元格位置
		cells := []string{
			s.getCellPosition(TableType1, "annual_total_coal_consumption", rowNum),
			s.getCellPosition(TableType1, "annual_raw_coal_consumption", rowNum),
			s.getCellPosition(TableType1, "annual_clean_coal_consumption", rowNum),
			s.getCellPosition(TableType1, "annual_other_coal_consumption", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "耗煤总量（实物量）应等于原煤消费+洗精煤消费+其他煤炭消费",
			Cells:     cells,
		})
	}

	return errors
}

// validateTable1EnergyCoalRelation 校验附表1综合能源消费量与煤炭消费量间的关系
func (s *DataImportService) validateTable1EnergyCoalRelation(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 获取相关数值
	annualEnergyEquivalentValue := s.parseFloat(s.getStringValue(data["annual_energy_equivalent_value"]))
	annualEnergyEquivalentCost := s.parseFloat(s.getStringValue(data["annual_energy_equivalent_cost"]))
	annualTotalCoalProducts := s.parseFloat(s.getStringValue(data["annual_total_coal_products"]))

	// 年综合能耗当量值≧耗煤总量（标准量）
	if s.isIntegerLessThan(annualEnergyEquivalentValue, annualTotalCoalProducts) {
		// 获取涉及到的单元格位置
		cells := []string{
			s.getCellPosition(TableType1, "annual_energy_equivalent_value", rowNum),
			s.getCellPosition(TableType1, "annual_total_coal_products", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年综合能耗当量值应大于等于耗煤总量（标准量）",
			Cells:     cells,
		})
	}

	// 年综合能耗等价值≧耗煤总量（标准量）
	if s.isIntegerLessThan(annualEnergyEquivalentCost, annualTotalCoalProducts) {
		// 获取涉及到的单元格位置
		cells := []string{
			s.getCellPosition(TableType1, "annual_energy_equivalent_cost", rowNum),
			s.getCellPosition(TableType1, "annual_total_coal_products", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年综合能耗等价值应大于等于耗煤总量（标准量）",
			Cells:     cells,
		})
	}

	return errors
}

// validateTable1UsageNumericFields 校验附表1用途表数值字段
func (s *DataImportService) validateTable1UsageNumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 获取投入量和产出量
	inputQuantity := s.parseFloat(s.getStringValue(data["input_quantity"]))
	outputQuantity := s.parseFloat(s.getStringValue(data["output_quantity"]))

	// ①投入量≧0
	if s.isIntegerLessThan(inputQuantity, 0) {
		cells := []string{s.getCellPosition(TableType1, "input_quantity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "投入量不能为负数", Cells: cells})
	}

	// ②投入量≦100000
	if s.isIntegerGreaterThan(inputQuantity, 100000) {
		cells := []string{s.getCellPosition(TableType1, "input_quantity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "投入量不能大于100000", Cells: cells})
	}

	// 产出量≧0
	if s.isIntegerLessThan(outputQuantity, 0) {
		cells := []string{s.getCellPosition(TableType1, "output_quantity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "产出量不能为负数", Cells: cells})
	}

	return errors
}

// validateTable1EquipNumericFields 校验附表1设备表数值字段
func (s *DataImportService) validateTable1EquipNumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 获取设备相关数值
	totalRuntime := s.parseFloat(s.getStringValue(data["total_runtime"]))
	designLife := s.parseFloat(s.getStringValue(data["design_life"]))
	energyEfficiency := s.parseFloat(s.getStringValue(data["energy_efficiency"]))
	capacity := s.parseFloat(s.getStringValue(data["capacity"]))
	annualCoalConsumption := s.parseFloat(s.getStringValue(data["annual_coal_consumption"]))

	// 1. 累计使用时间、设计年限校验
	// 应为0-50（含0和50）间的整数
	if s.isIntegerLessThan(totalRuntime, 0) || s.isIntegerGreaterThan(totalRuntime, 50) {
		cells := []string{s.getCellPosition(TableType1, "total_runtime", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "累计使用时间应在0-50之间", Cells: cells})
	}
	// 检查是否为整数
	if !s.isIntegerInteger(totalRuntime) {
		cells := []string{s.getCellPosition(TableType1, "total_runtime", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "累计使用时间应为整数", Cells: cells})
	}

	if s.isIntegerLessThan(designLife, 0) || s.isIntegerGreaterThan(designLife, 50) {
		cells := []string{s.getCellPosition(TableType1, "design_life", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "设计年限应在0-50之间", Cells: cells})
	}
	// 检查是否为整数
	if !s.isIntegerInteger(designLife) {
		cells := []string{s.getCellPosition(TableType1, "design_life", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "设计年限应为整数", Cells: cells})
	}

	// 2. 容量校验
	// 应为正整数
	if s.isIntegerLessThan(capacity, 0) {
		cells := []string{s.getCellPosition(TableType1, "capacity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "容量不能为负数", Cells: cells})
	}
	// 检查是否为整数
	if !s.isIntegerInteger(capacity) {
		cells := []string{s.getCellPosition(TableType1, "capacity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "容量应为整数", Cells: cells})
	}

	// 3. 年耗煤量校验
	// ①≧0
	if s.isIntegerLessThan(annualCoalConsumption, 0) {
		cells := []string{s.getCellPosition(TableType1, "annual_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年耗煤量不能为负数", Cells: cells})
	}

	// ②≦1000000000
	if s.isIntegerGreaterThan(annualCoalConsumption, 1000000000) {
		cells := []string{s.getCellPosition(TableType1, "annual_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年耗煤量不能大于1000000000", Cells: cells})
	}

	// 能效水平校验（保持原有的非负校验）
	if s.isIntegerLessThan(energyEfficiency, 0) {
		cells := []string{s.getCellPosition(TableType1, "energy_efficiency", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "能效水平不能为负数", Cells: cells})
	}

	return errors
}

// encryptTable1MainNumericFields 加密附表1主表数值字段
func (s *DataImportService) encryptTable1MainNumericFields(record map[string]interface{}) map[string]interface{} {
	numericFields := []string{
		"annual_energy_equivalent_value",
		"annual_energy_equivalent_cost",
		"annual_raw_material_energy",
		"annual_total_coal_consumption",
		"annual_total_coal_products",
		"annual_raw_coal",
		"annual_raw_coal_consumption",
		"annual_clean_coal_consumption",
		"annual_other_coal_consumption",
		"annual_coke_consumption",
	}
	return s.encryptNumericFields(record, numericFields)
}

// encryptTable1UsageNumericFields 加密附表1用途表数值字段
func (s *DataImportService) encryptTable1UsageNumericFields(record map[string]interface{}) map[string]interface{} {
	numericFields := []string{
		"input_quantity", "output_quantity",
	}
	return s.encryptNumericFields(record, numericFields)
}

// encryptTable1EquipNumericFields 加密附表1设备表数值字段
func (s *DataImportService) encryptTable1EquipNumericFields(record map[string]interface{}) map[string]interface{} {
	numericFields := []string{
		"total_runtime", "design_life", "energy_efficiency", "capacity", "annual_coal_consumption",
	}
	return s.encryptNumericFields(record, numericFields)
}
