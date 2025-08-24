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

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/xuri/excelize/v2"
)

// ModelDataCheckReportDownload 下载模型校验结果
func (s *DataImportService) ModelDataCheckReportDownload(tableType string) db.QueryResult {
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
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("下载模型报告失败: %v", err),
		}
	}

	srcFile, err := os.Open(zipFilePath)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("下载模型校验结果失败: %v", err),
		}
	}
	defer srcFile.Close()
	dstFile, err := os.Create(selectPath)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("下载模型校验结果失败: %v", err),
		}
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)

	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("下载模型校验结果失败: %v", err),
		}
	}

	return db.QueryResult{
		Ok:      true,
		Message: "下载模型校验结果成功",
	}

}

// ModelDataCoverTable1 覆盖附表1数据
func (s *DataImportService) ModelDataCoverTable1(filePaths []string) db.QueryResult {
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

	return db.QueryResult{
		Ok:      true,
		Message: "覆盖完成",
		Data: map[string]interface{}{
			"failed_files": failedFiles,      // 失败的文件
			"errors":       validationErrors, // 错误信息
		},
	}
}

// ModelDataCheckTable1 附表1模型校验函数
func (s *DataImportService) ModelDataCheckTable1() db.QueryResult {
	// 1. 读取缓存目录指定表格类型下的所有Excel文件
	cacheDir := s.app.GetCachePath(TableType1)

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取缓存目录失败: %v", err),
		}
	}

	var validationErrors []ValidationError = []ValidationError{} // 错误信息
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
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 读取失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			mainData, usageData, equipData, err := s.parseTable1Excel(f, true)
			f.Close()

			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			// 4. 调用校验函数,对每一行数据验证
			errors := s.validateTable1DataWithEnterpriseCheckForModel(mainData, usageData, equipData)
			if len(errors) > 0 {
				// 校验失败，在Excel文件中错误行最后添加错误信息
				err = s.addValidationErrorsToExcel(filePath, errors, mainData, usageData, equipData)
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
			if s.isTable1FileImported(mainData) {
				coverFiles = append(coverFiles, filePath)
				continue
			}

			// 6. 如果没有导入过,把数据保存到相应的数据库表中
			err = s.saveTable1Data(mainData, usageData, equipData)
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

	if !hasExcelFile {
		return db.QueryResult{
			Ok:      false,
			Message: "没有待校验Excel文件，请先进行数据导入",
		}
	}

	// 7. 把所有的模型验证失败的文件打个zip包
	if len(failedFiles) > 0 {
		err = s.createValidationErrorZip(failedFiles, TableType1, TableName1)
		if err != nil {
			validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("创建错误报告失败: %v", err)})
		}

		// 删除失败文件
		for _, filePath := range failedFiles {
			os.Remove(filePath)
		}
	}

	// 8. 返回结果
	message := fmt.Sprintf("处理完成。成功导入: %d 个文件，失败: %d 个文件", len(importedFiles), len(failedFiles))
	if len(validationErrors) > 0 {
		message += "。详细错误信息请查看生成的错误报告。"
	}

	return db.QueryResult{
		Ok:      true,
		Message: message,
		Data: map[string]interface{}{
			"cover_files":    coverFiles,           // 覆盖的文件
			"hasFailedFiles": len(failedFiles) > 0, // 是否有失败的文件
		},
	}
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
	createTime := time.Now().Format("2006-01-02 15:04:05")
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
		annual_other_coal_consumption, annual_coke_consumption, create_user
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.app.GetDB().Exec(query,
		mainRecord["obj_id"], mainRecord["unit_name"], mainRecord["stat_date"], mainRecord["tel"],
		mainRecord["credit_code"], mainRecord["create_time"], mainRecord["trade_a"], mainRecord["trade_b"],
		mainRecord["trade_c"], mainRecord["province_name"], mainRecord["city_name"], mainRecord["country_name"],
		encryptedValues["annual_energy_equivalent_value"], encryptedValues["annual_energy_equivalent_cost"],
		encryptedValues["annual_raw_material_energy"], encryptedValues["annual_total_coal_consumption"],
		encryptedValues["annual_total_coal_products"], encryptedValues["annual_raw_coal"], encryptedValues["annual_raw_coal_consumption"],
		encryptedValues["annual_clean_coal_consumption"], encryptedValues["annual_other_coal_consumption"],
		encryptedValues["annual_coke_consumption"], s.app.GetCurrentOSUser())
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
			input_unit, input_quantity, output_energy_types, output_quantity, measurement_unit, remarks, row_no
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err := s.app.GetDB().Exec(query,
			usage["obj_id"], usage["fk_id"], usage["stat_date"], usage["create_time"],
			usage["main_usage"], usage["specific_usage"], usage["input_variety"], usage["input_unit"],
			encryptedUsageValues["input_quantity"], usage["output_energy_types"], encryptedUsageValues["output_quantity"],
			usage["measurement_unit"], usage["remarks"], usage["row_no"])
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
			design_life, energy_efficiency, capacity_unit, capacity, coal_type, annual_coal_consumption,row_no
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err := s.app.GetDB().Exec(query,
			equip["obj_id"], equip["fk_id"], equip["stat_date"], equip["create_time"],
			equip["equip_type"], equip["equip_no"], encryptedEquipValues["total_runtime"], encryptedEquipValues["design_life"],
			encryptedEquipValues["energy_efficiency"], equip["capacity_unit"], encryptedEquipValues["capacity"], equip["coal_type"],
			encryptedEquipValues["annual_coal_consumption"], equip["row_no"])
		if err != nil {
			return fmt.Errorf("保存设备数据失败: %v", err)
		}
	}

	return nil
}

// addValidationErrorsToExcel 在Excel文件中添加校验错误信息
func (s *DataImportService) addValidationErrorsToExcel(filePath string, errors []ValidationError, mainData, usageData, equipData []map[string]interface{}) error {
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
	annualEnergyEquivalentValue, _ := s.parseFloat(s.getStringValue(data["annual_energy_equivalent_value"]))
	annualEnergyEquivalentCost, _ := s.parseFloat(s.getStringValue(data["annual_energy_equivalent_cost"]))
	annualRawMaterialEnergy, _ := s.parseFloat(s.getStringValue(data["annual_raw_material_energy"]))

	// ①≧0
	if annualEnergyEquivalentValue < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗当量值不能为负数"})
	}
	if annualEnergyEquivalentCost < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗等价值不能为负数"})
	}
	if annualRawMaterialEnergy < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用能消费量不能为负数"})
	}

	// ②≦100000
	if annualEnergyEquivalentValue > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗当量值不能大于100000"})
	}
	if annualEnergyEquivalentCost > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗等价值不能大于100000"})
	}
	if annualRawMaterialEnergy > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用能消费量不能大于100000"})
	}

	// ③年原料用能消费量≦年综合能耗当量值
	if annualRawMaterialEnergy > annualEnergyEquivalentValue {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用能消费量不能大于年综合能耗当量值"})
	}

	// ④年原料用能消费量≦年综合能耗等价值
	if annualRawMaterialEnergy > annualEnergyEquivalentCost {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用能消费量不能大于年综合能耗等价值"})
	}

	// 2. 煤炭消费相关字段校验
	annualTotalCoalConsumption, _ := s.parseFloat(s.getStringValue(data["annual_total_coal_consumption"]))
	annualTotalCoalProducts, _ := s.parseFloat(s.getStringValue(data["annual_total_coal_products"]))
	annualRawCoal, _ := s.parseFloat(s.getStringValue(data["annual_raw_coal"]))
	annualRawCoalConsumption, _ := s.parseFloat(s.getStringValue(data["annual_raw_coal_consumption"]))
	annualCleanCoalConsumption, _ := s.parseFloat(s.getStringValue(data["annual_clean_coal_consumption"]))
	annualOtherCoalConsumption, _ := s.parseFloat(s.getStringValue(data["annual_other_coal_consumption"]))
	annualCokeConsumption, _ := s.parseFloat(s.getStringValue(data["annual_coke_consumption"]))

	// ①≧0
	if annualTotalCoalConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（实物量）不能为负数"})
	}
	if annualTotalCoalProducts < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（标准量）不能为负数"})
	}
	if annualRawCoal < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原料用煤（实物量）不能为负数"})
	}
	if annualRawCoalConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤消费（实物量）不能为负数"})
	}
	if annualCleanCoalConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤消费（实物量）不能为负数"})
	}
	if annualOtherCoalConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他煤炭消费（实物量）不能为负数"})
	}
	if annualCokeConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费（实物量）不能为负数"})
	}

	// ②≦100000
	if annualTotalCoalConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（实物量）不能大于100000"})
	}
	if annualTotalCoalProducts > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（标准量）不能大于100000"})
	}
	if annualRawCoal > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原料用煤（实物量）不能大于100000"})
	}
	if annualRawCoalConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤消费（实物量）不能大于100000"})
	}
	if annualCleanCoalConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤消费（实物量）不能大于100000"})
	}
	if annualOtherCoalConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他煤炭消费（实物量）不能大于100000"})
	}
	if annualCokeConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费（实物量）不能大于100000"})
	}

	// ③耗煤总量（实物量）≧耗煤总量（标准量）
	if annualTotalCoalConsumption < annualTotalCoalProducts {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（实物量）不能小于耗煤总量（标准量）"})
	}

	// ④耗煤总量（实物量）≧原料用煤（实物量）
	if annualTotalCoalConsumption < annualRawCoal {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（实物量）不能小于原料用煤（实物量）"})
	}

	// ⑤耗煤总量（实物量）=原煤消费（实物量）+洗精煤消费（实物量）+其他煤炭消费（实物量）
	expectedTotal := annualRawCoalConsumption + annualCleanCoalConsumption + annualOtherCoalConsumption
	if annualTotalCoalConsumption != expectedTotal {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "耗煤总量（实物量）应等于原煤消费+洗精煤消费+其他煤炭消费"})
	}

	return errors
}

// validateTable1EnergyCoalRelation 校验附表1综合能源消费量与煤炭消费量间的关系
func (s *DataImportService) validateTable1EnergyCoalRelation(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 获取相关数值
	annualEnergyEquivalentValue, _ := s.parseFloat(s.getStringValue(data["annual_energy_equivalent_value"]))
	annualEnergyEquivalentCost, _ := s.parseFloat(s.getStringValue(data["annual_energy_equivalent_cost"]))
	annualTotalCoalProducts, _ := s.parseFloat(s.getStringValue(data["annual_total_coal_products"]))

	// 年综合能耗当量值≧耗煤总量（标准量）
	if annualEnergyEquivalentValue < annualTotalCoalProducts {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗当量值应大于等于耗煤总量（标准量）"})
	}

	// 年综合能耗等价值≧耗煤总量（标准量）
	if annualEnergyEquivalentCost < annualTotalCoalProducts {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能耗等价值应大于等于耗煤总量（标准量）"})
	}

	return errors
}

// validateTable1UsageNumericFields 校验附表1用途表数值字段
func (s *DataImportService) validateTable1UsageNumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 获取投入量和产出量
	inputQuantity, _ := s.parseFloat(s.getStringValue(data["input_quantity"]))
	outputQuantity, _ := s.parseFloat(s.getStringValue(data["output_quantity"]))

	// ①投入量≧0
	if inputQuantity < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "投入量不能为负数"})
	}

	// ②投入量≦100000
	if inputQuantity > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "投入量不能大于100000"})
	}

	// 产出量≧0
	if outputQuantity < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "产出量不能为负数"})
	}

	return errors
}

// validateTable1EquipNumericFields 校验附表1设备表数值字段
func (s *DataImportService) validateTable1EquipNumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 获取设备相关数值
	totalRuntime, _ := s.parseFloat(s.getStringValue(data["total_runtime"]))
	designLife, _ := s.parseFloat(s.getStringValue(data["design_life"]))
	energyEfficiency, _ := s.parseFloat(s.getStringValue(data["energy_efficiency"]))
	capacity, _ := s.parseFloat(s.getStringValue(data["capacity"]))
	annualCoalConsumption, _ := s.parseFloat(s.getStringValue(data["annual_coal_consumption"]))

	// 1. 累计使用时间、设计年限校验
	// 应为0-50（含0和50）间的整数
	if totalRuntime < 0 || totalRuntime > 50 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "累计使用时间应在0-50之间"})
	}
	if totalRuntime != float64(int(totalRuntime)) {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "累计使用时间应为整数"})
	}

	if designLife < 0 || designLife > 50 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "设计年限应在0-50之间"})
	}
	if designLife != float64(int(designLife)) {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "设计年限应为整数"})
	}

	// 2. 容量校验
	// 应为正整数
	if capacity < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "容量不能为负数"})
	}
	if capacity != float64(int(capacity)) {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "容量应为整数"})
	}

	// 3. 年耗煤量校验
	// ①≧0
	if annualCoalConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年耗煤量不能为负数"})
	}

	// ②≦1000000000
	if annualCoalConsumption > 1000000000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年耗煤量不能大于1000000000"})
	}

	// 能效水平校验（保持原有的非负校验）
	if energyEfficiency < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "能效水平不能为负数"})
	}

	return errors
}

// encryptTable1MainNumericFields 加密附表1主表数值字段
func (s *DataImportService) encryptTable1MainNumericFields(record map[string]interface{}) map[string]interface{} {
	numericFields := []string{
		"annual_energy_equivalent_value", "annual_energy_equivalent_cost", "annual_raw_material_energy",
		"annual_total_coal_consumption", "annual_total_coal_products", "annual_raw_coal",
		"annual_raw_coal_consumption", "annual_clean_coal_consumption", "annual_other_coal_consumption",
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
