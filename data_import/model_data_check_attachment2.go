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

// ModelDataCheckAttachment2 附件2模型校验函数
func (s *DataImportService) ModelDataCheckAttachment2() db.QueryResult {
	// 1. 读取缓存目录指定表格类型下的所有Excel文件
	cacheDir := s.app.GetCachePath(TableTypeAttachment2)

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

			mainData, err := s.parseAttachment2Excel(f, true)
			f.Close()

			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			// 4. 调用校验函数,对每一行数据验证
			errors := s.validateAttachment2DataForModel(mainData)
			if len(errors) > 0 {
				// 校验失败，在Excel文件中错误行最后添加错误信息
				err = s.addValidationErrorsToExcelAttachment2(filePath, errors, mainData)
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
			if s.isAttachment2FileImported(mainData) {
				importedFiles = append(importedFiles, file.Name())
				continue
			}

			// 6. 如果没有导入过,把数据保存到相应的数据库表中
			err = s.saveAttachment2Data(mainData)
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
		err = s.createValidationErrorZip(failedFiles, TableTypeAttachment2, TableAttachment2)
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

// validateAttachment2DataForModel 校验附件2数据（模型校验专用）
func (s *DataImportService) validateAttachment2DataForModel(mainData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	for _, data := range mainData {
		// 获取记录的实际Excel行号
		excelRowNum := s.getExcelRowNumber(data)

		// 数值字段校验
		valueErrors := s.validateAttachment2NumericFields(data, excelRowNum)
		errors = append(errors, valueErrors...)

		// 数据一致性校验
		consistencyErrors := s.validateAttachment2DataConsistency(data, excelRowNum)
		errors = append(errors, consistencyErrors...)
	}

	return errors
}

// validateAttachment2NumericFields 校验附件2数值字段
func (s *DataImportService) validateAttachment2NumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 校验所有数值字段不能为负数
	numericFields := []string{
		"total_coal", "raw_coal", "washed_coal", "other_coal", "power_generation",
		"heating", "coal_washing", "coking", "oil_refining", "gas_production",
		"industry", "raw_materials", "other_uses", "coke",
	}

	for _, fieldName := range numericFields {
		if value, ok := data[fieldName].(string); ok && value != "" {
			if numValue, err := s.parseFloat(value); err == nil {
				if numValue < 0 {
					errors = append(errors, ValidationError{RowNumber: rowNum, Message: fmt.Sprintf("%s不能为负数", fieldName)})
				}
			}
		}
	}

	return errors
}

// validateAttachment2DataConsistency 校验附件2数据一致性
func (s *DataImportService) validateAttachment2DataConsistency(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 校验煤合计 = 原煤 + 洗精煤 + 其他
	totalCoal, _ := s.parseFloat(s.getStringValue(data["total_coal"]))
	rawCoal, _ := s.parseFloat(s.getStringValue(data["raw_coal"]))
	washedCoal, _ := s.parseFloat(s.getStringValue(data["washed_coal"]))
	otherCoal, _ := s.parseFloat(s.getStringValue(data["other_coal"]))

	if totalCoal != rawCoal+washedCoal+otherCoal {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计应等于原煤+洗精煤+其他"})
	}

	// 校验用途合计 = 火力发电 + 供热 + 煤炭洗选 + 炼焦 + 炼油及煤制油 + 制气 + 工业 + 用作原料、材料 + 其他用途
	powerGen, _ := s.parseFloat(s.getStringValue(data["power_generation"]))
	heating, _ := s.parseFloat(s.getStringValue(data["heating"]))
	coalWashing, _ := s.parseFloat(s.getStringValue(data["coal_washing"]))
	coking, _ := s.parseFloat(s.getStringValue(data["coking"]))
	oilRefining, _ := s.parseFloat(s.getStringValue(data["oil_refining"]))
	gasProduction, _ := s.parseFloat(s.getStringValue(data["gas_production"]))
	industry, _ := s.parseFloat(s.getStringValue(data["industry"]))
	rawMaterials, _ := s.parseFloat(s.getStringValue(data["raw_materials"]))
	otherUses, _ := s.parseFloat(s.getStringValue(data["other_uses"]))

	usageTotal := powerGen + heating + coalWashing + coking + oilRefining + gasProduction + industry + rawMaterials + otherUses
	if totalCoal != usageTotal {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计应等于各用途消费量之和"})
	}

	return errors
}

// coverAttachment2Data 覆盖附件2数据
func (s *DataImportService) coverAttachment2Data(mainData []map[string]interface{}, fileName string) error {
	if len(mainData) == 0 {
		return fmt.Errorf("数据为空")
	}

	// 逐行检查，根据年份+省+市+县检查是否已导入
	for _, record := range mainData {
		statDate := s.getStringValue(record["stat_date"])
		provinceName := s.getStringValue(record["province_name"])
		cityName := s.getStringValue(record["city_name"])
		countryName := s.getStringValue(record["country_name"])

		// 根据年份+省+市+县做where条件更新数据
		err := s.updateAttachment2DataByRegionAndYear(statDate, provinceName, cityName, countryName, record)
		if err != nil {
			return fmt.Errorf("更新数据失败: %v", err)
		}
	}

	return nil
}

// updateAttachment2DataByRegionAndYear 根据地区和时间更新附件2数据
func (s *DataImportService) updateAttachment2DataByRegionAndYear(statDate, provinceName, cityName, countryName string, record map[string]interface{}) error {
	// 先删除旧数据
	query := "DELETE FROM coal_consumption_report WHERE stat_date = ? AND province_name = ? AND city_name = ? AND country_name = ?"
	_, err := s.app.GetDB().Exec(query, statDate, provinceName, cityName, countryName)
	if err != nil {
		return err
	}

	// 插入新数据
	return s.insertAttachment2Data(record)
}

// isAttachment2FileImported 检查附件2文件是否已导入
func (s *DataImportService) isAttachment2FileImported(mainData []map[string]interface{}) bool {
	// 按Excel数据逐行检查，根据年份+省+市+县检查是否已导入
	for _, record := range mainData {
		statDate := s.getStringValue(record["stat_date"])
		provinceName := s.getStringValue(record["province_name"])
		cityName := s.getStringValue(record["city_name"])
		countryName := s.getStringValue(record["country_name"])

		query := "SELECT COUNT(*) as count FROM coal_consumption_report WHERE stat_date = ? AND province_name = ? AND city_name = ? AND country_name = ?"
		result, err := s.app.GetDB().Query(query, statDate, provinceName, cityName, countryName)
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

// saveAttachment2Data 保存附件2数据到数据库
func (s *DataImportService) saveAttachment2Data(mainData []map[string]interface{}) error {
	for _, record := range mainData {
		err := s.insertAttachment2Data(record)
		if err != nil {
			return err
		}
	}

	return nil
}

// insertAttachment2Data 插入附件2数据
func (s *DataImportService) insertAttachment2Data(record map[string]interface{}) error {
	record["obj_id"] = s.generateUUID()
	record["create_time"] = time.Now().Format("2006-01-02 15:04:05")

	// 对数值字段进行SM4加密
	encryptedValues := s.encryptAttachment2NumericFields(record)

	query := `INSERT INTO coal_consumption_report (
		obj_id, stat_date, province_name, city_name, country_name, total_coal, raw_coal,
		washed_coal, other_coal, power_generation, heating, coal_washing, coking,
		oil_refining, gas_production, industry, raw_materials, other_uses, coke, create_time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.app.GetDB().Exec(query,
		record["obj_id"], record["stat_date"], record["province_name"], record["city_name"],
		record["country_name"], encryptedValues["total_coal"], encryptedValues["raw_coal"], encryptedValues["washed_coal"],
		encryptedValues["other_coal"], encryptedValues["power_generation"], encryptedValues["heating"], encryptedValues["coal_washing"],
		encryptedValues["coking"], encryptedValues["oil_refining"], encryptedValues["gas_production"], encryptedValues["industry"],
		encryptedValues["raw_materials"], encryptedValues["other_uses"], encryptedValues["coke"], record["create_time"])
	if err != nil {
		return fmt.Errorf("保存数据失败: %v", err)
	}

	return nil
}

// addValidationErrorsToExcelAttachment2 在附件2Excel文件中添加校验错误信息
func (s *DataImportService) addValidationErrorsToExcelAttachment2(filePath string, errors []ValidationError, mainData []map[string]interface{}) error {
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
		excelRow := rowNum + 6 // 附件2通常从第7行开始有数据

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

// encryptAttachment2NumericFields 加密附件2数值字段
func (s *DataImportService) encryptAttachment2NumericFields(record map[string]interface{}) map[string]interface{} {
	numericFields := []string{
		"total_coal", "raw_coal", "washed_coal", "other_coal", "power_generation",
		"heating", "coal_washing", "coking", "oil_refining", "gas_production",
		"industry", "raw_materials", "other_uses", "coke",
	}
	return s.encryptNumericFields(record, numericFields)
}

// ModelDataCoverAttachment2 覆盖附件2数据
func (s *DataImportService) ModelDataCoverAttachment2(fileNames []string) db.QueryResult {
	cacheDir := s.app.GetCachePath(TableTypeAttachment2)
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

			mainData, err := s.parseAttachment2Excel(f, true)
			f.Close()
			os.Remove(filePath)

			if err != nil {
				validationErrors = append(validationErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			err = s.coverAttachment2Data(mainData, file.Name())
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
