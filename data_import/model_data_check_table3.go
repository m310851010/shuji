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

// ModelDataCoverTable3 覆盖附表3数据
func (s *DataImportService) ModelDataCoverTable3(filePaths []string) db.QueryResult {
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

	return db.QueryResult{
		Ok:      true,
		Message: "覆盖完成",
		Data: map[string]interface{}{
			"failed_files": failedFiles,      // 失败的文件
			"errors":       validationErrors, // 错误信息
		},
	}
}

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

	if !hasExcelFile {
		return db.QueryResult{
			Ok:      false,
			Message: "没有待校验Excel文件，请先进行数据导入",
		}
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

	return db.QueryResult{
		Ok:      true,
		Message: message,
		Data: map[string]interface{}{
			"cover_files":    coverFiles,           // 覆盖的文件
			"hasFailedFiles": len(failedFiles) > 0, // 是否有失败的文件
		},
	}
}

// validateTable3DataForModel 校验附表3数据（模型校验专用）
func (s *DataImportService) validateTable3DataForModel(mainData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	// 1. 逐行校验数值字段
	for _, data := range mainData {
		// 获取记录的实际Excel行号
		excelRowNum := s.getExcelRowNumber(data)

		// 数值字段校验
		valueErrors := s.validateTable3NumericFields(data, excelRowNum)
		errors = append(errors, valueErrors...)

		// 新增校验规则
		newRuleErrors := s.validateTable3NewRules(data, excelRowNum)
		errors = append(errors, newRuleErrors...)
	}

	// 2. Excel文件内整体校验规则
	overallErrors := s.validateTable3OverallRules(mainData)
	errors = append(errors, overallErrors...)

	// 3. 数据库验证
	dbErrors := s.validateTable3DatabaseRules(mainData)
	errors = append(errors, dbErrors...)

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
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能源消费量当量值不能为负数"})
	}
	if equivalentValue > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能源消费量当量值不能大于100000"})
	}

	if equivalentCost < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能源消费量等价值不能为负数"})
	}
	if equivalentCost > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年综合能源消费量等价值不能大于100000"})
	}

	// 2. 年煤品消费量部分校验
	// 煤品消费总量（实物量）、煤炭消费量（实物量）、焦炭消费量（实物量）、兰炭消费量（实物量）
	// 煤品消费总量（折标量）、煤炭消费量（折标量）、焦炭消费量（折标量）、兰炭消费量（折标量）
	pqTotalCoalConsumption, _ := s.parseFloat(s.getStringValue(data["pq_total_coal_consumption"]))
	pqCoalConsumption, _ := s.parseFloat(s.getStringValue(data["pq_coal_consumption"]))
	pqCokeConsumption, _ := s.parseFloat(s.getStringValue(data["pq_coke_consumption"]))
	pqBlueCokeConsumption, _ := s.parseFloat(s.getStringValue(data["pq_blue_coke_consumption"]))

	sceTotalCoalConsumption, _ := s.parseFloat(s.getStringValue(data["sce_total_coal_consumption"]))
	sceCoalConsumption, _ := s.parseFloat(s.getStringValue(data["sce_coal_consumption"]))
	sceCokeConsumption, _ := s.parseFloat(s.getStringValue(data["sce_coke_consumption"]))
	sceBlueCokeConsumption, _ := s.parseFloat(s.getStringValue(data["sce_blue_coke_consumption"]))

	// ①≧0
	if pqTotalCoalConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（实物量）不能为负数"})
	}
	if pqCoalConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费量（实物量）不能为负数"})
	}
	if pqCokeConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量（实物量）不能为负数"})
	}
	if pqBlueCokeConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "兰炭消费量（实物量）不能为负数"})
	}

	if sceTotalCoalConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（折标量）不能为负数"})
	}
	if sceCoalConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费量（折标量）不能为负数"})
	}
	if sceCokeConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量（折标量）不能为负数"})
	}
	if sceBlueCokeConsumption < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "兰炭消费量（折标量）不能为负数"})
	}

	// ②≦100000
	if pqTotalCoalConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（实物量）不能大于100000"})
	}
	if pqCoalConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费量（实物量）不能大于100000"})
	}
	if pqCokeConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量（实物量）不能大于100000"})
	}
	if pqBlueCokeConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "兰炭消费量（实物量）不能大于100000"})
	}

	if sceTotalCoalConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（折标量）不能大于100000"})
	}
	if sceCoalConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费量（折标量）不能大于100000"})
	}
	if sceCokeConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量（折标量）不能大于100000"})
	}
	if sceBlueCokeConsumption > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "兰炭消费量（折标量）不能大于100000"})
	}

	// ③煤炭消费量（实物量）≧煤炭消费量（折标量）
	if pqCoalConsumption < sceCoalConsumption {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费量（实物量）应大于等于煤炭消费量（折标量）"})
	}

	// ④焦炭消费量（实物量）≧焦炭消费量（折标量）
	if pqCokeConsumption < sceCokeConsumption {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量（实物量）应大于等于焦炭消费量（折标量）"})
	}

	// ⑤兰炭消费量（实物量）≧兰炭消费量（折标量）
	if pqBlueCokeConsumption < sceBlueCokeConsumption {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "兰炭消费量（实物量）应大于等于兰炭消费量（折标量）"})
	}

	// ⑥煤品消费总量（实物量）=煤炭消费量（实物量）+焦炭消费量（实物量）+兰炭消费量（实物量）
	expectedPqTotal := pqCoalConsumption + pqCokeConsumption + pqBlueCokeConsumption
	if pqTotalCoalConsumption != expectedPqTotal {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（实物量）应等于煤炭消费量+焦炭消费量+兰炭消费量"})
	}

	// ⑦煤品消费总量（折标量）=煤炭消费量（折标量）+焦炭消费量（折标量）+兰炭消费量（折标量）
	expectedSceTotal := sceCoalConsumption + sceCokeConsumption + sceBlueCokeConsumption
	if sceTotalCoalConsumption != expectedSceTotal {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤品消费总量（折标量）应等于煤炭消费量+焦炭消费量+兰炭消费量"})
	}

	// 3. 煤炭消费替代情况部分校验
	// 煤炭消费替代量（实物量）规则：①≧0；②≦100000
	substitutionQuantity, _ := s.parseFloat(s.getStringValue(data["substitution_quantity"]))

	if substitutionQuantity < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费替代量（实物量）不能为负数"})
	}
	if substitutionQuantity > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费替代量（实物量）不能大于100000"})
	}

	// 4. 原料用煤部分校验
	// 年原料用煤量（实物量）、年原料用煤量（折标量）规则：①≧0；②≦100000；③年原料用煤量（实物量）≧年原料用煤量（折标量）
	pqAnnualCoalQuantity, _ := s.parseFloat(s.getStringValue(data["pq_annual_coal_quantity"]))
	sceAnnualCoalQuantity, _ := s.parseFloat(s.getStringValue(data["sce_annual_coal_quantity"]))

	if pqAnnualCoalQuantity < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（实物量）不能为负数"})
	}
	if pqAnnualCoalQuantity > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（实物量）不能大于100000"})
	}

	if sceAnnualCoalQuantity < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（折标量）不能为负数"})
	}
	if sceAnnualCoalQuantity > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（折标量）不能大于100000"})
	}

	if pqAnnualCoalQuantity < sceAnnualCoalQuantity {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "年原料用煤量（实物量）应大于等于年原料用煤量（折标量）"})
	}

	return errors
}

// validateTable3NewRules 校验附表3新增校验规则
func (s *DataImportService) validateTable3NewRules(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 1. 分品种煤炭消费摸底部分校验
	totalCoal, _ := s.parseFloat(s.getStringValue(data["total_coal"]))
	rawCoal, _ := s.parseFloat(s.getStringValue(data["raw_coal"]))
	washedCoal, _ := s.parseFloat(s.getStringValue(data["washed_coal"]))
	otherCoal, _ := s.parseFloat(s.getStringValue(data["other_coal"]))

	// ①≧0
	if totalCoal < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计不能为负数"})
	}
	if rawCoal < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤不能为负数"})
	}
	if washedCoal < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤不能为负数"})
	}
	if otherCoal < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他不能为负数"})
	}

	// ②≦200000
	if totalCoal > 200000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计不能大于200000"})
	}
	if rawCoal > 200000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤不能大于200000"})
	}
	if washedCoal > 200000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤不能大于200000"})
	}
	if otherCoal > 200000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他不能大于200000"})
	}

	// ③煤合计=原煤+洗精煤+其他
	expectedTotal := rawCoal + washedCoal + otherCoal
	if totalCoal != expectedTotal {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计应等于原煤+洗精煤+其他"})
	}

	// 2. 分用途煤炭消费摸底部分校验
	powerGeneration, _ := s.parseFloat(s.getStringValue(data["power_generation"]))
	heating, _ := s.parseFloat(s.getStringValue(data["heating"]))
	coalWashing, _ := s.parseFloat(s.getStringValue(data["coal_washing"]))
	coking, _ := s.parseFloat(s.getStringValue(data["coking"]))
	oilRefining, _ := s.parseFloat(s.getStringValue(data["oil_refining"]))
	gasProduction, _ := s.parseFloat(s.getStringValue(data["gas_production"]))
	industry, _ := s.parseFloat(s.getStringValue(data["industry"]))
	rawMaterials, _ := s.parseFloat(s.getStringValue(data["raw_materials"]))
	otherUses, _ := s.parseFloat(s.getStringValue(data["other_uses"]))

	// ①≧0
	if powerGeneration < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "火力发电不能为负数"})
	}
	if heating < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "供热不能为负数"})
	}
	if coalWashing < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭洗选不能为负数"})
	}
	if coking < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼焦不能为负数"})
	}
	if oilRefining < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼油及煤制油不能为负数"})
	}
	if gasProduction < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "制气不能为负数"})
	}
	if industry < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业不能为负数"})
	}
	if rawMaterials < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业（#用作原料、材料）不能为负数"})
	}
	if otherUses < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他用途不能为负数"})
	}

	// ②≦100000
	if powerGeneration > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "火力发电不能大于100000"})
	}
	if heating > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "供热不能大于100000"})
	}
	if coalWashing > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭洗选不能大于100000"})
	}
	if coking > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼焦不能大于100000"})
	}
	if oilRefining > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼油及煤制油不能大于100000"})
	}
	if gasProduction > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "制气不能大于100000"})
	}
	if industry > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业不能大于100000"})
	}
	if rawMaterials > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业（#用作原料、材料）不能大于100000"})
	}
	if otherUses > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他用途不能大于100000"})
	}

	// ③工业≧工业（#用作原料、材料）
	if industry < rawMaterials {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业应大于等于工业（#用作原料、材料）"})
	}

	// 3. 焦炭消费摸底部分校验
	coke, _ := s.parseFloat(s.getStringValue(data["coke"]))

	// ①≧0
	if coke < 0 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭不能为负数"})
	}

	// ②≦100000
	if coke > 100000 {
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭不能大于100000"})
	}

	return errors
}

// validateTable3OverallRules 校验附表3整体规则
func (s *DataImportService) validateTable3OverallRules(mainData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	if len(mainData) == 0 {
		return errors
	}

	// 计算整个Excel文件的总量
	var totalEquivalentValue float64
	var totalEquivalentCost float64
	var totalSceTotalCoalConsumption float64
	var totalPqTotalCoalConsumption float64
	var totalPqAnnualCoalQuantity float64
	var totalSceAnnualCoalQuantity float64

	for _, data := range mainData {
		equivalentValue, _ := s.parseFloat(s.getStringValue(data["equivalent_value"]))
		equivalentCost, _ := s.parseFloat(s.getStringValue(data["equivalent_cost"]))
		sceTotalCoalConsumption, _ := s.parseFloat(s.getStringValue(data["sce_total_coal_consumption"]))
		pqTotalCoalConsumption, _ := s.parseFloat(s.getStringValue(data["pq_total_coal_consumption"]))
		pqAnnualCoalQuantity, _ := s.parseFloat(s.getStringValue(data["pq_annual_coal_quantity"]))
		sceAnnualCoalQuantity, _ := s.parseFloat(s.getStringValue(data["sce_annual_coal_quantity"]))

		totalEquivalentValue += equivalentValue
		totalEquivalentCost += equivalentCost
		totalSceTotalCoalConsumption += sceTotalCoalConsumption
		totalPqTotalCoalConsumption += pqTotalCoalConsumption
		totalPqAnnualCoalQuantity += pqAnnualCoalQuantity
		totalSceAnnualCoalQuantity += sceAnnualCoalQuantity
	}

	// ①年综合能源消费量与年煤品消费量（折标量）
	// 年综合能源消费量（当量值）≧年煤品消费量（折标量）
	if totalEquivalentValue < totalSceTotalCoalConsumption {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "年综合能源消费量（当量值）应大于等于年煤品消费量（折标量）"})
	}

	// 年综合能源消费量（等价值）≧年煤品消费量（折标量）
	if totalEquivalentCost < totalSceTotalCoalConsumption {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "年综合能源消费量（等价值）应大于等于年煤品消费量（折标量）"})
	}

	// ②年煤品消费量与原料用煤情况
	// 煤品消费总量（实物量）≧年原料用煤量（实物量）
	if totalPqTotalCoalConsumption < totalPqAnnualCoalQuantity {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "煤品消费总量（实物量）应大于等于年原料用煤量（实物量）"})
	}

	// 煤品消费总量（折标量）≧年原料用煤量（折标量）
	if totalSceTotalCoalConsumption < totalSceAnnualCoalQuantity {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "煤品消费总量（折标量）应大于等于年原料用煤量（折标量）"})
	}

	return errors
}

// validateTable3DatabaseRules 校验附表3数据库验证规则
func (s *DataImportService) validateTable3DatabaseRules(mainData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	if len(mainData) == 0 {
		return errors
	}

	// 获取当前用户的省市县级别
	areaResult := s.app.GetAreaConfig()
	if !areaResult.Ok || areaResult.Data == nil {
		return errors
	}

	areaData, ok := areaResult.Data.([]map[string]interface{})
	if !ok || len(areaData) == 0 {
		return errors
	}

	area := areaData[0]
	provinceName := s.getStringValue(area["province_name"])
	cityName := s.getStringValue(area["city_name"])
	countryName := s.getStringValue(area["country_name"])

	// 构建查询条件
	var whereClause string
	var args []interface{}

	if countryName != "" {
		// 有县，查询该县下的所有数据
		whereClause = "WHERE province_name = ? AND city_name = ? AND country_name = ?"
		args = []interface{}{provinceName, cityName, countryName}
	} else if cityName != "" {
		// 有市无县，查询该市下的所有县的数据
		whereClause = "WHERE province_name = ? AND city_name = ?"
		args = []interface{}{provinceName, cityName}
	} else if provinceName != "" {
		// 只有省，查询该省下的所有市的数据
		whereClause = "WHERE province_name = ?"
		args = []interface{}{provinceName}
	} else {
		// 没有区域信息，跳过验证
		return errors
	}

	// 获取当前数据的年份
	statDate := s.getStringValue(mainData[0]["stat_date"])
	if statDate == "" {
		return errors
	}

	// 查询下级所有同一年份的数据
	query := fmt.Sprintf(`
		SELECT 
			total_coal, raw_coal, washed_coal, other_coal,
			power_generation, heating, coal_washing, coking, oil_refining, gas_production,
			industry, raw_materials, other_uses, coke
		FROM fixed_assets_investment_project 
		%s AND stat_date = ?
	`, whereClause)
	args = append(args, statDate)

	result, err := s.app.GetDB().Query(query, args...)
	if err != nil || result.Data == nil {
		return errors
	}

	subData, ok := result.Data.([]map[string]interface{})
	if !ok || len(subData) == 0 {
		return errors
	}

	// 计算下级数据总和
	var subTotalCoal, subRawCoal, subWashedCoal, subOtherCoal float64
	var subPowerGeneration, subHeating, subCoalWashing, subCoking, subOilRefining, subGasProduction float64
	var subIndustry, subRawMaterials, subOtherUses, subCoke float64

	for _, record := range subData {
		// 解密数值字段
		totalCoal, _ := s.parseFloat(s.decryptValue(record["total_coal"]))
		rawCoal, _ := s.parseFloat(s.decryptValue(record["raw_coal"]))
		washedCoal, _ := s.parseFloat(s.decryptValue(record["washed_coal"]))
		otherCoal, _ := s.parseFloat(s.decryptValue(record["other_coal"]))
		powerGeneration, _ := s.parseFloat(s.decryptValue(record["power_generation"]))
		heating, _ := s.parseFloat(s.decryptValue(record["heating"]))
		coalWashing, _ := s.parseFloat(s.decryptValue(record["coal_washing"]))
		coking, _ := s.parseFloat(s.decryptValue(record["coking"]))
		oilRefining, _ := s.parseFloat(s.decryptValue(record["oil_refining"]))
		gasProduction, _ := s.parseFloat(s.decryptValue(record["gas_production"]))
		industry, _ := s.parseFloat(s.decryptValue(record["industry"]))
		rawMaterials, _ := s.parseFloat(s.decryptValue(record["raw_materials"]))
		otherUses, _ := s.parseFloat(s.decryptValue(record["other_uses"]))
		coke, _ := s.parseFloat(s.decryptValue(record["coke"]))

		subTotalCoal += totalCoal
		subRawCoal += rawCoal
		subWashedCoal += washedCoal
		subOtherCoal += otherCoal
		subPowerGeneration += powerGeneration
		subHeating += heating
		subCoalWashing += coalWashing
		subCoking += coking
		subOilRefining += oilRefining
		subGasProduction += gasProduction
		subIndustry += industry
		subRawMaterials += rawMaterials
		subOtherUses += otherUses
		subCoke += coke
	}

	// 计算当前数据总和
	var currentTotalCoal, currentRawCoal, currentWashedCoal, currentOtherCoal float64
	var currentPowerGeneration, currentHeating, currentCoalWashing, currentCoking, currentOilRefining, currentGasProduction float64
	var currentIndustry, currentRawMaterials, currentOtherUses, currentCoke float64

	for _, record := range mainData {
		totalCoal, _ := s.parseFloat(s.getStringValue(record["total_coal"]))
		rawCoal, _ := s.parseFloat(s.getStringValue(record["raw_coal"]))
		washedCoal, _ := s.parseFloat(s.getStringValue(record["washed_coal"]))
		otherCoal, _ := s.parseFloat(s.getStringValue(record["other_coal"]))
		powerGeneration, _ := s.parseFloat(s.getStringValue(record["power_generation"]))
		heating, _ := s.parseFloat(s.getStringValue(record["heating"]))
		coalWashing, _ := s.parseFloat(s.getStringValue(record["coal_washing"]))
		coking, _ := s.parseFloat(s.getStringValue(record["coking"]))
		oilRefining, _ := s.parseFloat(s.getStringValue(record["oil_refining"]))
		gasProduction, _ := s.parseFloat(s.getStringValue(record["gas_production"]))
		industry, _ := s.parseFloat(s.getStringValue(record["industry"]))
		rawMaterials, _ := s.parseFloat(s.getStringValue(record["raw_materials"]))
		otherUses, _ := s.parseFloat(s.getStringValue(record["other_uses"]))
		coke, _ := s.parseFloat(s.getStringValue(record["coke"]))

		currentTotalCoal += totalCoal
		currentRawCoal += rawCoal
		currentWashedCoal += washedCoal
		currentOtherCoal += otherCoal
		currentPowerGeneration += powerGeneration
		currentHeating += heating
		currentCoalWashing += coalWashing
		currentCoking += coking
		currentOilRefining += oilRefining
		currentGasProduction += gasProduction
		currentIndustry += industry
		currentRawMaterials += rawMaterials
		currentOtherUses += otherUses
		currentCoke += coke
	}

	// 校验规则：同年份本单位所上传数值*120%应≥下级单位相加之和
	threshold := 1.2

	// 校验各个字段
	if currentTotalCoal*threshold < subTotalCoal {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "煤合计数值*120%应大于等于下级单位相加之和"})
	}
	if currentRawCoal*threshold < subRawCoal {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "原煤数值*120%应大于等于下级单位相加之和"})
	}
	if currentWashedCoal*threshold < subWashedCoal {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "洗精煤数值*120%应大于等于下级单位相加之和"})
	}
	if currentOtherCoal*threshold < subOtherCoal {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "其他数值*120%应大于等于下级单位相加之和"})
	}
	if currentPowerGeneration*threshold < subPowerGeneration {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "火力发电数值*120%应大于等于下级单位相加之和"})
	}
	if currentHeating*threshold < subHeating {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "供热数值*120%应大于等于下级单位相加之和"})
	}
	if currentCoalWashing*threshold < subCoalWashing {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "煤炭洗选数值*120%应大于等于下级单位相加之和"})
	}
	if currentCoking*threshold < subCoking {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "炼焦数值*120%应大于等于下级单位相加之和"})
	}
	if currentOilRefining*threshold < subOilRefining {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "炼油及煤制油数值*120%应大于等于下级单位相加之和"})
	}
	if currentGasProduction*threshold < subGasProduction {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "制气数值*120%应大于等于下级单位相加之和"})
	}
	if currentIndustry*threshold < subIndustry {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "工业数值*120%应大于等于下级单位相加之和"})
	}
	if currentRawMaterials*threshold < subRawMaterials {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "工业（#用作原料、材料）数值*120%应大于等于下级单位相加之和"})
	}
	if currentOtherUses*threshold < subOtherUses {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "其他用途数值*120%应大于等于下级单位相加之和"})
	}
	if currentCoke*threshold < subCoke {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "焦炭数值*120%应大于等于下级单位相加之和"})
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

		query := "SELECT COUNT(1) as count FROM fixed_assets_investment_project WHERE project_code = ? AND construction_unit = ?"
		result, err := s.app.GetDB().QueryRow(query, projectCode, constructionUnit)
		if err != nil || result.Data == nil {
			continue
		}

		if result.Data.(map[string]interface{})["count"].(int64) > 0 {
			return true // 检查到立即停止表示已导入
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
		equivalent_cost, pq_total_coal_consumption, pq_coal_consumption, pq_coke_consumption, pq_blue_coke_consumption,
		sce_total_coal_consumption, sce_coal_consumption, sce_coke_consumption, sce_blue_coke_consumption,
		is_substitution, substitution_source, substitution_quantity, pq_annual_coal_quantity, sce_annual_coal_quantity,
		total_coal, raw_coal, washed_coal, other_coal, power_generation, heating, coal_washing, coking,
		oil_refining, gas_production, industry, raw_materials, other_uses, coke, create_time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
		encryptedValues["total_coal"], encryptedValues["raw_coal"], encryptedValues["washed_coal"], encryptedValues["other_coal"],
		encryptedValues["power_generation"], encryptedValues["heating"], encryptedValues["coal_washing"], encryptedValues["coking"],
		encryptedValues["oil_refining"], encryptedValues["gas_production"], encryptedValues["industry"],
		encryptedValues["raw_materials"], encryptedValues["other_uses"], encryptedValues["coke"], record["create_time"])
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
		"equivalent_value", "equivalent_cost",
		"pq_total_coal_consumption", "pq_coal_consumption", "pq_coke_consumption", "pq_blue_coke_consumption",
		"sce_total_coal_consumption", "sce_coal_consumption", "sce_coke_consumption", "sce_blue_coke_consumption",
		"substitution_quantity", "pq_annual_coal_quantity", "sce_annual_coal_quantity",
		"total_coal", "raw_coal", "washed_coal", "other_coal",
		"power_generation", "heating", "coal_washing", "coking", "oil_refining", "gas_production",
		"industry", "raw_materials", "other_uses", "coke",
	}
	return s.encryptNumericFields(record, numericFields)
}


