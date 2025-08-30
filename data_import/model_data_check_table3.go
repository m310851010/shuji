package data_import

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"shuji/db"
	"slices"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ModelDataCoverTable3 覆盖附表3数据
func (s *DataImportService) ModelDataCoverTable3(filePaths []string) db.QueryResult {
	// 使用包装函数来处理异常
	return s.modelDataCoverTable3WithRecover(filePaths)
}

// modelDataCoverTable3WithRecover 带异常处理的覆盖附表3数据函数
func (s *DataImportService) modelDataCoverTable3WithRecover(filePaths []string) db.QueryResult {
	var result db.QueryResult

	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ModelDataCoverTable3 发生异常: %v", r)
			// 设置错误结果
			result = db.QueryResult{
				Ok:      false,
				Message: fmt.Sprintf("函数执行异常: %v", r),
				Data:    nil,
			}
		}
	}()

	cacheDir := s.app.GetCachePath(TableType3)
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		result = db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取缓存目录失败: %v", err),
		}
		return result
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

// ModelDataCheckTable3 附表3模型校验函数
func (s *DataImportService) ModelDataCheckTable3() db.QueryResult {
	// 使用包装函数来处理异常
	return s.modelDataCheckTable3WithRecover()
}

// modelDataCheckTable3WithRecover 带异常处理的附表3模型校验函数
func (s *DataImportService) modelDataCheckTable3WithRecover() db.QueryResult {
	var result db.QueryResult

	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ModelDataCheckTable3 发生异常: %v", r)
			// 设置错误结果
			result = db.QueryResult{
				Ok:      false,
				Message: fmt.Sprintf("函数执行异常: %v", r),
				Data:    nil,
			}
		}
	}()

	// 1. 读取缓存目录指定表格类型下的所有Excel文件
	cacheDir := s.app.GetCachePath(TableType3)

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		result = db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取缓存目录失败: %v", err),
		}
		return result
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

			// 解析Excel文件
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
				coverFiles = append(coverFiles, filePath)
				continue
			}

			// 6. 如果没有导入过,把数据保存到相应的数据库表中
			err = s.saveTable3DataForModel(mainData)
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
		result = db.QueryResult{
			Ok:      false,
			Message: "没有待校验Excel文件，请先进行数据导入",
		}
		return result
	}

	// 7. 把所有的模型验证失败的文件打个zip包
	if len(failedFiles) > 0 {
		err = s.createValidationErrorZip(failedFiles, TableType3, TableName3)
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

	result = db.QueryResult{
		Ok:      true,
		Message: message,
		Data: map[string]interface{}{
			"cover_files":    coverFiles,           // 覆盖的文件
			"hasFailedFiles": len(failedFiles) > 0, // 是否有失败的文件
		},
	}
	return result
}

// validateTable3DataForModel 校验附表3数据（模型校验专用）
func (s *DataImportService) validateTable3DataForModel(mainData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	// 逐行校验数值字段、新增校验规则和整体规则（行内字段间逻辑关系）
	for _, data := range mainData {
		// 获取记录的实际Excel行号
		excelRowNum := s.getExcelRowNumber(data)

		// 数值字段校验
		valueErrors := s.validateTable3NumericFields(data, excelRowNum)
		errors = append(errors, valueErrors...)

		// 整体规则校验（行内字段间逻辑关系）
		overallErrors := s.validateTable3OverallRulesForRow(data, excelRowNum)
		errors = append(errors, overallErrors...)
	}

	return errors
}

// validateTable3NumericFields 校验附表3数值字段
func (s *DataImportService) validateTable3NumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 1. 年综合能源消费量部分校验
	// 当量值、等价值校验规则：①≧0；②≦100000
	equivalentValue, _ := s.parseFloat(s.getStringValue(data["equivalent_value"]))
	equivalentCost, _ := s.parseFloat(s.getStringValue(data["equivalent_cost"]))

	if equivalentValue < 0 {
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年综合能源消费量当量值不能为负数",
			Cells:     []string{s.getCellPosition(TableType3, "equivalent_value", rowNum)},
		})
	}
	if equivalentValue > 100000 {
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年综合能源消费量当量值不能大于100000",
			Cells:     []string{s.getCellPosition(TableType3, "equivalent_value", rowNum)},
		})
	}

	if equivalentCost < 0 {
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年综合能源消费量等价值不能为负数",
			Cells:     []string{s.getCellPosition(TableType3, "equivalent_cost", rowNum)},
		})
	}
	if equivalentCost > 100000 {
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年综合能源消费量等价值不能大于100000",
			Cells:     []string{s.getCellPosition(TableType3, "equivalent_cost", rowNum)},
		})
	}

	// 2. 年煤品消费量部分校验
	// 煤品消费总量（实物量）、煤炭消费量（实物量）、焦炭消费量（实物量）、兰炭消费量（实物量）
	// 煤品消费总量（折标量）、煤炭消费量（折标量）、焦炭消费量（折标量）、兰炭消费量（折标量）
	pqTotalCoalConsumption := s.parseBigFloat(s.getStringValue(data["pq_total_coal_consumption"]))
	pqCoalConsumption := s.parseBigFloat(s.getStringValue(data["pq_coal_consumption"]))
	pqCokeConsumption := s.parseBigFloat(s.getStringValue(data["pq_coke_consumption"]))
	pqBlueCokeConsumption := s.parseBigFloat(s.getStringValue(data["pq_blue_coke_consumption"]))

	sceTotalCoalConsumption := s.parseBigFloat(s.getStringValue(data["sce_total_coal_consumption"]))
	sceCoalConsumption := s.parseBigFloat(s.getStringValue(data["sce_coal_consumption"]))
	sceCokeConsumption := s.parseBigFloat(s.getStringValue(data["sce_coke_consumption"]))
	sceBlueCokeConsumption := s.parseBigFloat(s.getStringValue(data["sce_blue_coke_consumption"]))

	// ①≧0
	if pqTotalCoalConsumption.Cmp(big.NewFloat(0)) < 0 {
		cells := []string{s.getCellPosition(TableType3, "pq_total_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（实物量）不能为负数", Cells: cells})
	}
	if pqCoalConsumption.Cmp(big.NewFloat(0)) < 0 {
		cells := []string{s.getCellPosition(TableType3, "pq_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费量（实物量）不能为负数", Cells: cells})
	}
	if pqCokeConsumption.Cmp(big.NewFloat(0)) < 0 {
		cells := []string{s.getCellPosition(TableType3, "pq_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量（实物量）不能为负数", Cells: cells})
	}
	if pqBlueCokeConsumption.Cmp(big.NewFloat(0)) < 0 {
		cells := []string{s.getCellPosition(TableType3, "pq_blue_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "兰炭消费量（实物量）不能为负数", Cells: cells})
	}

	if sceTotalCoalConsumption.Cmp(big.NewFloat(0)) < 0 {
		cells := []string{s.getCellPosition(TableType3, "sce_total_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（折标量）不能为负数", Cells: cells})
	}
	if sceCoalConsumption.Cmp(big.NewFloat(0)) < 0 {
		cells := []string{s.getCellPosition(TableType3, "sce_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费量（折标量）不能为负数", Cells: cells})
	}
	if sceCokeConsumption.Cmp(big.NewFloat(0)) < 0 {
		cells := []string{s.getCellPosition(TableType3, "sce_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量（折标量）不能为负数", Cells: cells})
	}
	if sceBlueCokeConsumption.Cmp(big.NewFloat(0)) < 0 {
		cells := []string{s.getCellPosition(TableType3, "sce_blue_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "兰炭消费量（折标量）不能为负数", Cells: cells})
	}

	// ②≦100000
	if pqTotalCoalConsumption.Cmp(big.NewFloat(100000)) > 0 {
		cells := []string{s.getCellPosition(TableType3, "pq_total_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（实物量）不能大于100000", Cells: cells})
	}
	if pqCoalConsumption.Cmp(big.NewFloat(100000)) > 0 {
		cells := []string{s.getCellPosition(TableType3, "pq_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费量（实物量）不能大于100000", Cells: cells})
	}
	if pqCokeConsumption.Cmp(big.NewFloat(100000)) > 0 {
		cells := []string{s.getCellPosition(TableType3, "pq_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量（实物量）不能大于100000", Cells: cells})
	}
	if pqBlueCokeConsumption.Cmp(big.NewFloat(100000)) > 0 {
		cells := []string{s.getCellPosition(TableType3, "pq_blue_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "兰炭消费量（实物量）不能大于100000", Cells: cells})
	}

	if sceTotalCoalConsumption.Cmp(big.NewFloat(100000)) > 0 {
		cells := []string{s.getCellPosition(TableType3, "sce_total_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（折标量）不能大于100000", Cells: cells})
	}
	if sceCoalConsumption.Cmp(big.NewFloat(100000)) > 0 {
		cells := []string{s.getCellPosition(TableType3, "sce_coal_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费量（折标量）不能大于100000", Cells: cells})
	}
	if sceCokeConsumption.Cmp(big.NewFloat(100000)) > 0 {
		cells := []string{s.getCellPosition(TableType3, "sce_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量（折标量）不能大于100000", Cells: cells})
	}
	if sceBlueCokeConsumption.Cmp(big.NewFloat(100000)) > 0 {
		cells := []string{s.getCellPosition(TableType3, "sce_blue_coke_consumption", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "兰炭消费量（折标量）不能大于100000", Cells: cells})
	}

	// ③煤炭消费量（实物量）≧煤炭消费量（折标量）
	if pqCoalConsumption.Cmp(sceCoalConsumption) < 0 {
		cells := []string{
			s.getCellPosition(TableType3, "pq_coal_consumption", rowNum),
			s.getCellPosition(TableType3, "sce_coal_consumption", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "煤炭消费量（实物量）应大于等于煤炭消费量（折标量）",
			Cells:     cells,
		})
	}

	// ④焦炭消费量（实物量）≧焦炭消费量（折标量）
	if pqCokeConsumption.Cmp(sceCokeConsumption) < 0 {
		cells := []string{
			s.getCellPosition(TableType3, "pq_coke_consumption", rowNum),
			s.getCellPosition(TableType3, "sce_coke_consumption", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "焦炭消费量（实物量）应大于等于焦炭消费量（折标量）",
			Cells:     cells,
		})
	}

	// ⑤兰炭消费量（实物量）≧兰炭消费量（折标量）
	if pqBlueCokeConsumption.Cmp(sceBlueCokeConsumption) < 0 {
		cells := []string{
			s.getCellPosition(TableType3, "pq_blue_coke_consumption", rowNum),
			s.getCellPosition(TableType3, "sce_blue_coke_consumption", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "兰炭消费量（实物量）应大于等于兰炭消费量（折标量）",
			Cells:     cells,
		})
	}

	// ⑥煤品消费总量（实物量）=煤炭消费量（实物量）+焦炭消费量（实物量）+兰炭消费量（实物量）
	expectedPqTotal := new(big.Float).Add(pqCoalConsumption, pqCokeConsumption)
	expectedPqTotal.Add(expectedPqTotal, pqBlueCokeConsumption)
	if pqTotalCoalConsumption.Cmp(expectedPqTotal) != 0 {
		cells := []string{
			s.getCellPosition(TableType3, "pq_total_coal_consumption", rowNum),
			s.getCellPosition(TableType3, "pq_coal_consumption", rowNum),
			s.getCellPosition(TableType3, "pq_coke_consumption", rowNum),
			s.getCellPosition(TableType3, "pq_blue_coke_consumption", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "煤品消费总量（实物量）应等于煤炭消费量+焦炭消费量+兰炭消费量",
			Cells:     cells,
		})
	}

	// ⑦煤品消费总量（折标量）=煤炭消费量（折标量）+焦炭消费量（折标量）+兰炭消费量（折标量）
	expectedSceTotal := new(big.Float).Add(sceCoalConsumption, sceCokeConsumption)
	expectedSceTotal.Add(expectedSceTotal, sceBlueCokeConsumption)
	if sceTotalCoalConsumption.Cmp(expectedSceTotal) != 0 {
		cells := []string{
			s.getCellPosition(TableType3, "sce_total_coal_consumption", rowNum),
			s.getCellPosition(TableType3, "sce_coal_consumption", rowNum),
			s.getCellPosition(TableType3, "sce_coke_consumption", rowNum),
			s.getCellPosition(TableType3, "sce_blue_coke_consumption", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "煤品消费总量（折标量）应等于煤炭消费量+焦炭消费量+兰炭消费量",
			Cells:     cells,
		})
	}

	// 3. 煤炭消费替代情况部分校验
	// 煤炭消费替代量（实物量）规则：①≧0；②≦100000
	substitutionQuantity, _ := s.parseFloat(s.getStringValue(data["substitution_quantity"]))

	if substitutionQuantity < 0 {
		cells := []string{s.getCellPosition(TableType3, "substitution_quantity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费替代量（实物量）不能为负数", Cells: cells})
	}
	if substitutionQuantity > 100000 {
		cells := []string{s.getCellPosition(TableType3, "substitution_quantity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费替代量（实物量）不能大于100000", Cells: cells})
	}

	// 4. 原料用煤部分校验
	// 年原料用煤量（实物量）、年原料用煤量（折标量）规则：①≧0；②≦100000；③年原料用煤量（实物量）≧年原料用煤量（折标量）
	pqAnnualCoalQuantity, _ := s.parseFloat(s.getStringValue(data["pq_annual_coal_quantity"]))
	sceAnnualCoalQuantity, _ := s.parseFloat(s.getStringValue(data["sce_annual_coal_quantity"]))

	if pqAnnualCoalQuantity < 0 {
		cells := []string{s.getCellPosition(TableType3, "pq_annual_coal_quantity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（实物量）不能为负数", Cells: cells})
	}
	if pqAnnualCoalQuantity > 100000 {
		cells := []string{s.getCellPosition(TableType3, "pq_annual_coal_quantity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（实物量）不能大于100000", Cells: cells})
	}

	if sceAnnualCoalQuantity < 0 {
		cells := []string{s.getCellPosition(TableType3, "sce_annual_coal_quantity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（折标量）不能为负数", Cells: cells})
	}
	if sceAnnualCoalQuantity > 100000 {
		cells := []string{s.getCellPosition(TableType3, "sce_annual_coal_quantity", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（折标量）不能大于100000", Cells: cells})
	}

	if pqAnnualCoalQuantity < sceAnnualCoalQuantity {
		cells := []string{
			s.getCellPosition(TableType3, "pq_annual_coal_quantity", rowNum),
			s.getCellPosition(TableType3, "sce_annual_coal_quantity", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（实物量）应大于等于年原料用煤量（折标量）", Cells: cells})
	}

	return errors
}

// validateTable3OverallRulesForRow 校验附表3单行整体规则（行内字段间逻辑关系）
func (s *DataImportService) validateTable3OverallRulesForRow(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 获取当前行的数值
	equivalentValue, _ := s.parseFloat(s.getStringValue(data["equivalent_value"]))
	equivalentCost, _ := s.parseFloat(s.getStringValue(data["equivalent_cost"]))
	sceTotalCoalConsumption, _ := s.parseFloat(s.getStringValue(data["sce_total_coal_consumption"]))
	pqTotalCoalConsumption, _ := s.parseFloat(s.getStringValue(data["pq_total_coal_consumption"]))
	pqAnnualCoalQuantity, _ := s.parseFloat(s.getStringValue(data["pq_annual_coal_quantity"]))
	sceAnnualCoalQuantity, _ := s.parseFloat(s.getStringValue(data["sce_annual_coal_quantity"]))

	// ①年综合能源消费量与年煤品消费量（折标量）的逻辑关系
	// 年综合能源消费量（当量值）≧年煤品消费量（折标量）
	if equivalentValue < sceTotalCoalConsumption {
		cells := []string{
			s.getCellPosition(TableType3, "equivalent_value", rowNum),
			s.getCellPosition(TableType3, "sce_total_coal_consumption", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年综合能源消费量（当量值）应大于等于年煤品消费量（折标量）",
			Cells:     cells,
		})
	}

	// 年综合能源消费量（等价值）≧年煤品消费量（折标量）
	if equivalentCost < sceTotalCoalConsumption {
		cells := []string{
			s.getCellPosition(TableType3, "equivalent_cost", rowNum),
			s.getCellPosition(TableType3, "sce_total_coal_consumption", rowNum),
		}
		errors = append(errors, ValidationError{
			RowNumber: rowNum,
			Message:   "年综合能源消费量（等价值）应大于等于年煤品消费量（折标量）",
			Cells:     cells,
		})
	}

	// ②年煤品消费量与原料用煤情况的逻辑关系
	// 煤品消费总量（实物量）≧年原料用煤量（实物量）
	if pqTotalCoalConsumption < pqAnnualCoalQuantity {
		cells := []string{
			s.getCellPosition(TableType3, "pq_total_coal_consumption", rowNum),
			s.getCellPosition(TableType3, "pq_annual_coal_quantity", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（实物量）应大于等于年原料用煤量（实物量）", Cells: cells})
	}

	// 煤品消费总量（折标量）≧年原料用煤量（折标量）
	if sceTotalCoalConsumption < sceAnnualCoalQuantity {
		cells := []string{
			s.getCellPosition(TableType3, "sce_total_coal_consumption", rowNum),
			s.getCellPosition(TableType3, "sce_annual_coal_quantity", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（折标量）应大于等于年原料用煤量（折标量）", Cells: cells})
	}

	return errors
}

// coverTable3Data 覆盖附表3数据
func (s *DataImportService) coverTable3Data(mainData []map[string]interface{}, fileName string) error {
	if len(mainData) == 0 {
		return fmt.Errorf("数据为空")
	}

	// 逐行检查，根据项目代码+审查意见文号检查是否已导入
	for _, record := range mainData {
		projectCode := s.getStringValue(record["project_code"])
		documentNumber := s.getStringValue(record["document_number"])

		// 先尝试更新，通过受影响行数判断是否存在
		affectedRows, err := s.updateTable3DataByProjectCodeAndDocumentNumber(projectCode, documentNumber, record)
		if err != nil {
			return fmt.Errorf("更新数据失败: %v", err)
		}

		// 如果受影响行数为0，说明数据不存在，执行插入
		if affectedRows == 0 {
			err = s.insertTable3Data(record)
			if err != nil {
				return fmt.Errorf("插入数据失败: %v", err)
			}
		}
	}

	return nil
}

// updateTable3DataByProjectCodeAndDocumentNumber 根据项目代码和审查意见文号更新附表3数据
func (s *DataImportService) updateTable3DataByProjectCodeAndDocumentNumber(projectCode, documentNumber string, record map[string]interface{}) (int64, error) {
	// 对数值字段进行SM4加密
	encryptedValues := s.encryptTable3NumericFields(record)

	query := `UPDATE fixed_assets_investment_project SET 
		stat_date = ?, project_name = ?, project_code = ?, construction_unit = ?, main_construction_content = ?,
		province_name = ?, city_name = ?, country_name = ?, trade_a = ?, trade_c = ?, 
		examination_approval_time = ?, scheduled_time = ?, actual_time = ?,
		examination_authority = ?, document_number = ?, equivalent_value = ?,
		equivalent_cost = ?, pq_total_coal_consumption = ?, pq_coal_consumption = ?, 
		pq_coke_consumption = ?, pq_blue_coke_consumption = ?, sce_total_coal_consumption = ?,
		sce_coal_consumption = ?, sce_coke_consumption = ?, sce_blue_coke_consumption = ?,
		is_substitution = ?, substitution_source = ?, substitution_quantity = ?, 
		pq_annual_coal_quantity = ?, sce_annual_coal_quantity = ?
		WHERE project_code = ? AND document_number = ?`

	result, err := s.app.GetDB().Exec(query,
		record["stat_date"], record["project_name"], record["project_code"], record["construction_unit"], record["main_construction_content"],
		record["province_name"], record["city_name"], record["country_name"],
		record["trade_a"], record["trade_c"], record["examination_approval_time"],
		record["scheduled_time"], record["actual_time"], record["examination_authority"],
		record["document_number"], encryptedValues["equivalent_value"], encryptedValues["equivalent_cost"],
		encryptedValues["pq_total_coal_consumption"], encryptedValues["pq_coal_consumption"],
		encryptedValues["pq_coke_consumption"], encryptedValues["pq_blue_coke_consumption"],
		encryptedValues["sce_total_coal_consumption"], encryptedValues["sce_coal_consumption"],
		encryptedValues["sce_coke_consumption"], encryptedValues["sce_blue_coke_consumption"],
		record["is_substitution"], record["substitution_source"], encryptedValues["substitution_quantity"],
		encryptedValues["pq_annual_coal_quantity"], encryptedValues["sce_annual_coal_quantity"],
		projectCode, documentNumber)

	if err != nil {
		return 0, err
	}

	// 从QueryResult中获取受影响的行数
	if result.Ok && result.Data != nil {
		if dataMap, ok := result.Data.(map[string]interface{}); ok {
			if rowsAffected, exists := dataMap["rowsAffected"]; exists {
				if affectedRows, ok := rowsAffected.(int64); ok {
					return affectedRows, nil
				}
			}
		}
	}

	return 0, nil
}

// isTable3FileImported 检查附表3文件是否已导入
func (s *DataImportService) isTable3FileImported(mainData []map[string]interface{}) bool {
	// 按Excel数据逐行检查，根据项目代码+审查意见文号检查是否已导入
	for _, record := range mainData {
		projectCode := s.getStringValue(record["project_code"])
		documentNumber := s.getStringValue(record["document_number"])

		query := "SELECT COUNT(1) as count FROM fixed_assets_investment_project WHERE project_code = ? AND document_number = ?"
		result, err := s.app.GetDB().QueryRow(query, projectCode, documentNumber)
		if err != nil || result.Data == nil {
			continue
		}

		if result.Data.(map[string]interface{})["count"].(int64) > 0 {
			return true // 检查到立即停止表示已导入
		}
	}

	return false
}

// saveTable3DataForModel 模型校验专用保存附表3数据到数据库（只使用INSERT）
func (s *DataImportService) saveTable3DataForModel(mainData []map[string]interface{}) error {
	for _, record := range mainData {
		// 直接执行插入操作，不检查数据是否已存在
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
		equivalent_cost, pq_total_coal_consumption, pq_coal_consumption, pq_coke_consumption, pq_blue_coke_consumption,
		sce_total_coal_consumption, sce_coal_consumption, sce_coke_consumption, sce_blue_coke_consumption,
		is_substitution, substitution_source, substitution_quantity, pq_annual_coal_quantity, sce_annual_coal_quantity,
		create_time, create_user, is_check
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.app.GetDB().Exec(query,
		record["obj_id"], record["stat_date"], record["project_name"], record["project_code"],
		record["construction_unit"], record["main_construction_content"], record["province_name"],
		record["city_name"], record["country_name"], record["trade_a"], record["trade_c"],
		record["examination_approval_time"], record["scheduled_time"], record["actual_time"],
		record["examination_authority"], record["document_number"], encryptedValues["equivalent_value"],
		encryptedValues["equivalent_cost"], encryptedValues["pq_total_coal_consumption"], encryptedValues["pq_coal_consumption"],
		encryptedValues["pq_coke_consumption"], encryptedValues["pq_blue_coke_consumption"], encryptedValues["sce_total_coal_consumption"],
		encryptedValues["sce_coal_consumption"], encryptedValues["sce_coke_consumption"], encryptedValues["sce_blue_coke_consumption"],
		record["is_substitution"], record["substitution_source"], encryptedValues["substitution_quantity"],
		encryptedValues["pq_annual_coal_quantity"], encryptedValues["sce_annual_coal_quantity"],
		record["create_time"], s.app.GetCurrentOSUser(), EncryptedOne)
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

// encryptTable3NumericFields 加密附表3数值字段
func (s *DataImportService) encryptTable3NumericFields(record map[string]interface{}) map[string]interface{} {
	numericFields := []string{
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
		"substitution_quantity",
		"pq_annual_coal_quantity",
		"sce_annual_coal_quantity",
	}
	return s.encryptNumericFields(record, numericFields)
}
