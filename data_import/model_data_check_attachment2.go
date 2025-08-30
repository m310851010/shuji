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

// ModelDataCoverAttachment2 覆盖附件2数据
func (s *DataImportService) ModelDataCoverAttachment2(filePaths []string) db.QueryResult {
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

	s.initAttachment2CacheManager()
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
				coverFiles = append(coverFiles, filePath)
				continue
			}

			// 6. 如果没有导入过,把数据保存到相应的数据库表中
			err = s.saveAttachment2DataForModel(mainData)

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
		err = s.createValidationErrorZip(failedFiles, TableTypeAttachment2, TableAttachment2)
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

// validateAttachment2DataForModel 校验附件2数据（模型校验专用）
func (s *DataImportService) validateAttachment2DataForModel(mainData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	// 逐行校验数值字段、数据一致性和整体规则（行内字段间逻辑关系）
	for _, data := range mainData {
		// 获取记录的实际Excel行号
		excelRowNum := s.getExcelRowNumber(data)

		// 数值字段校验
		valueErrors := s.validateAttachment2NumericFields(data, excelRowNum)
		errors = append(errors, valueErrors...)

		// 数据一致性校验
		consistencyErrors := s.validateAttachment2DataConsistency(data, excelRowNum)
		errors = append(errors, consistencyErrors...)

		// 整体规则校验（行内字段间逻辑关系）
		overallErrors := s.validateAttachment2OverallRulesForRow(data, excelRowNum)
		errors = append(errors, overallErrors...)
	}

	// 数据库验证
	dbErrors := s.validateAttachment2DatabaseRules(mainData)
	errors = append(errors, dbErrors...)

	return errors
}

// validateAttachment2NumericFields 校验附件2数值字段
func (s *DataImportService) validateAttachment2NumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 1. 分品种煤炭消费摸底部分校验
	totalCoal, _ := s.parseFloat(s.getStringValue(data["total_coal"]))
	rawCoal, _ := s.parseFloat(s.getStringValue(data["raw_coal"]))
	washedCoal, _ := s.parseFloat(s.getStringValue(data["washed_coal"]))
	otherCoal, _ := s.parseFloat(s.getStringValue(data["other_coal"]))

	// ①≧0
	if totalCoal < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "total_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计不能为负数", Cells: cells})
	}
	if rawCoal < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "raw_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤不能为负数", Cells: cells})
	}
	if washedCoal < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "washed_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤不能为负数", Cells: cells})
	}
	if otherCoal < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "other_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他不能为负数", Cells: cells})
	}

	// ②≦200000
	if totalCoal > 200000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "total_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计不能大于200000", Cells: cells})
	}
	if rawCoal > 200000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "raw_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤不能大于200000", Cells: cells})
	}
	if washedCoal > 200000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "washed_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤不能大于200000", Cells: cells})
	}
	if otherCoal > 200000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "other_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他不能大于200000", Cells: cells})
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
		cells := []string{s.getCellPosition(TableTypeAttachment2, "power_generation", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "火力发电不能为负数", Cells: cells})
	}
	if heating < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "heating", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "供热不能为负数", Cells: cells})
	}
	if coalWashing < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coal_washing", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭洗选不能为负数", Cells: cells})
	}
	if coking < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coking", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼焦不能为负数", Cells: cells})
	}
	if oilRefining < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "oil_refining", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼油及煤制油不能为负数", Cells: cells})
	}
	if gasProduction < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "gas_production", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "制气不能为负数", Cells: cells})
	}
	if industry < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "industry", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业不能为负数", Cells: cells})
	}
	if rawMaterials < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "raw_materials", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业（#用作原料、材料）不能为负数", Cells: cells})
	}
	if otherUses < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "other_uses", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他用途不能为负数", Cells: cells})
	}

	// ②≦100000
	if powerGeneration > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "power_generation", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "火力发电不能大于100000", Cells: cells})
	}
	if heating > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "heating", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "供热不能大于100000", Cells: cells})
	}
	if coalWashing > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coal_washing", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭洗选不能大于100000", Cells: cells})
	}
	if coking > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coking", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼焦不能大于100000", Cells: cells})
	}
	if oilRefining > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "oil_refining", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼油及煤制油不能大于100000", Cells: cells})
	}
	if gasProduction > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "gas_production", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "制气不能大于100000", Cells: cells})
	}
	if industry > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "industry", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业不能大于100000", Cells: cells})
	}
	if rawMaterials > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "raw_materials", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业（#用作原料、材料）不能大于100000", Cells: cells})
	}
	if otherUses > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "other_uses", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他用途不能大于100000", Cells: cells})
	}

	// 3. 焦炭消费摸底部分校验
	coke, _ := s.parseFloat(s.getStringValue(data["coke"]))

	// ①≧0
	if coke < 0 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coke", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭不能为负数", Cells: cells})
	}

	// ②≦100000
	if coke > 100000 {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coke", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭不能大于100000", Cells: cells})
	}

	return errors
}

// validateAttachment2DataConsistency 校验附件2数据一致性
func (s *DataImportService) validateAttachment2DataConsistency(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 1. 分品种煤炭消费摸底部分
	// ③煤合计=原煤+洗精煤+其他
	totalCoal, _ := s.parseFloat(s.getStringValue(data["total_coal"]))
	rawCoal, _ := s.parseFloat(s.getStringValue(data["raw_coal"]))
	washedCoal, _ := s.parseFloat(s.getStringValue(data["washed_coal"]))
	otherCoal, _ := s.parseFloat(s.getStringValue(data["other_coal"]))

	expectedTotal := rawCoal + washedCoal + otherCoal
	if totalCoal != expectedTotal {
		cells := []string{
			s.getCellPosition(TableTypeAttachment2, "total_coal", rowNum),
			s.getCellPosition(TableTypeAttachment2, "raw_coal", rowNum),
			s.getCellPosition(TableTypeAttachment2, "washed_coal", rowNum),
			s.getCellPosition(TableTypeAttachment2, "other_coal", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计应等于原煤+洗精煤+其他", Cells: cells})
	}

	// 2. 分用途煤炭消费摸底部分
	// ③工业≧工业（#用作原料、材料）
	industry, _ := s.parseFloat(s.getStringValue(data["industry"]))
	rawMaterials, _ := s.parseFloat(s.getStringValue(data["raw_materials"]))

	if industry < rawMaterials {
		cells := []string{
			s.getCellPosition(TableTypeAttachment2, "industry", rowNum),
			s.getCellPosition(TableTypeAttachment2, "raw_materials", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业应大于等于工业（#用作原料、材料）", Cells: cells})
	}

	// 3. 文件内整体校验
	// ①分品种煤炭消费摸底与分用途煤炭消费摸底
	// 煤合计≧能源加工转换+终端消费
	powerGeneration, _ := s.parseFloat(s.getStringValue(data["power_generation"]))
	heating, _ := s.parseFloat(s.getStringValue(data["heating"]))
	coalWashing, _ := s.parseFloat(s.getStringValue(data["coal_washing"]))
	coking, _ := s.parseFloat(s.getStringValue(data["coking"]))
	oilRefining, _ := s.parseFloat(s.getStringValue(data["oil_refining"]))
	gasProduction, _ := s.parseFloat(s.getStringValue(data["gas_production"]))
	otherUses, _ := s.parseFloat(s.getStringValue(data["other_uses"]))

	// 能源加工转换 = 火力发电 + 供热 + 煤炭洗选 + 炼焦 + 炼油及煤制油 + 制气
	energyConversion := powerGeneration + heating + coalWashing + coking + oilRefining + gasProduction
	// 终端消费 = 工业 + 其他用途
	terminalConsumption := industry + otherUses

	if totalCoal < energyConversion+terminalConsumption {
		cells := []string{
			s.getCellPosition(TableTypeAttachment2, "total_coal", rowNum),
			s.getCellPosition(TableTypeAttachment2, "power_generation", rowNum),
			s.getCellPosition(TableTypeAttachment2, "heating", rowNum),
			s.getCellPosition(TableTypeAttachment2, "coal_washing", rowNum),
			s.getCellPosition(TableTypeAttachment2, "coking", rowNum),
			s.getCellPosition(TableTypeAttachment2, "oil_refining", rowNum),
			s.getCellPosition(TableTypeAttachment2, "gas_production", rowNum),
			s.getCellPosition(TableTypeAttachment2, "industry", rowNum),
			s.getCellPosition(TableTypeAttachment2, "other_uses", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计应大于等于能源加工转换+终端消费", Cells: cells})
	}

	return errors
}

// validateAttachment2OverallRulesForRow 校验附件2单行整体规则（行内字段间逻辑关系）
func (s *DataImportService) validateAttachment2OverallRulesForRow(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 获取当前行的数值
	totalCoal, _ := s.parseFloat(s.getStringValue(data["total_coal"]))
	powerGeneration, _ := s.parseFloat(s.getStringValue(data["power_generation"]))
	heating, _ := s.parseFloat(s.getStringValue(data["heating"]))
	coalWashing, _ := s.parseFloat(s.getStringValue(data["coal_washing"]))
	coking, _ := s.parseFloat(s.getStringValue(data["coking"]))
	oilRefining, _ := s.parseFloat(s.getStringValue(data["oil_refining"]))
	gasProduction, _ := s.parseFloat(s.getStringValue(data["gas_production"]))
	industry, _ := s.parseFloat(s.getStringValue(data["industry"]))
	otherUses, _ := s.parseFloat(s.getStringValue(data["other_uses"]))
	coke, _ := s.parseFloat(s.getStringValue(data["coke"]))

	// ①煤炭消费总量与各用途消费量的逻辑关系
	// 煤炭消费总量应大于等于各用途消费量之和
	totalUsage := powerGeneration + heating + coalWashing + coking + oilRefining + gasProduction + industry + otherUses
	if totalCoal < totalUsage {
		cells := []string{
			s.getCellPosition(TableTypeAttachment2, "total_coal", rowNum),
			s.getCellPosition(TableTypeAttachment2, "power_generation", rowNum),
			s.getCellPosition(TableTypeAttachment2, "heating", rowNum),
			s.getCellPosition(TableTypeAttachment2, "coal_washing", rowNum),
			s.getCellPosition(TableTypeAttachment2, "coking", rowNum),
			s.getCellPosition(TableTypeAttachment2, "oil_refining", rowNum),
			s.getCellPosition(TableTypeAttachment2, "gas_production", rowNum),
			s.getCellPosition(TableTypeAttachment2, "industry", rowNum),
			s.getCellPosition(TableTypeAttachment2, "other_uses", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭消费总量应大于等于各用途消费量之和", Cells: cells})
	}

	// ②焦炭消费量与煤炭消费量的逻辑关系
	// 焦炭消费量应小于等于煤炭消费总量（焦炭是煤炭的加工产品）
	if coke > totalCoal {
		cells := []string{
			s.getCellPosition(TableTypeAttachment2, "coke", rowNum),
			s.getCellPosition(TableTypeAttachment2, "total_coal", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭消费量应小于等于煤炭消费总量", Cells: cells})
	}

	// ③能源加工转换与终端消费的逻辑关系
	// 能源加工转换量应大于等于终端消费量（加工转换会产生损耗）
	energyConversion := powerGeneration + heating + coalWashing + coking + oilRefining + gasProduction
	terminalConsumption := industry + otherUses
	if energyConversion < terminalConsumption {
		cells := []string{
			s.getCellPosition(TableTypeAttachment2, "power_generation", rowNum),
			s.getCellPosition(TableTypeAttachment2, "heating", rowNum),
			s.getCellPosition(TableTypeAttachment2, "coal_washing", rowNum),
			s.getCellPosition(TableTypeAttachment2, "coking", rowNum),
			s.getCellPosition(TableTypeAttachment2, "oil_refining", rowNum),
			s.getCellPosition(TableTypeAttachment2, "gas_production", rowNum),
			s.getCellPosition(TableTypeAttachment2, "industry", rowNum),
			s.getCellPosition(TableTypeAttachment2, "other_uses", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "能源加工转换量应大于等于终端消费量", Cells: cells})
	}

	// ④火力发电与其他能源转换的逻辑关系
	// 火力发电量应大于等于供热、制气等其他能源转换量
	otherEnergyConversion := heating + coalWashing + coking + oilRefining + gasProduction
	if powerGeneration < otherEnergyConversion {
		cells := []string{
			s.getCellPosition(TableTypeAttachment2, "power_generation", rowNum),
			s.getCellPosition(TableTypeAttachment2, "heating", rowNum),
			s.getCellPosition(TableTypeAttachment2, "coal_washing", rowNum),
			s.getCellPosition(TableTypeAttachment2, "coking", rowNum),
			s.getCellPosition(TableTypeAttachment2, "oil_refining", rowNum),
			s.getCellPosition(TableTypeAttachment2, "gas_production", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "火力发电量应大于等于其他能源转换量", Cells: cells})
	}

	return errors
}

// attachment2CacheManager 附件2缓存管理器实例
var attachment2CacheManager *Attachment2CacheManager

// initAttachment2CacheManager 初始化附件2缓存管理器
func (s *DataImportService) initAttachment2CacheManager() {
	if attachment2CacheManager == nil {
		attachment2CacheManager = NewAttachment2CacheManager(s)
		// 预加载数据库缓存（只在第一次调用时从数据库加载，后续直接使用缓存）
		attachment2CacheManager.PreloadDatabaseCache()
	}
}

// validateAttachment2DatabaseRules 校验附件2数据库验证规则（优化版本）
func (s *DataImportService) validateAttachment2DatabaseRules(mainData []map[string]interface{}) []ValidationError {
	errors := []ValidationError{}

	if len(mainData) == 0 {
		return errors
	}

	// 获取当前用户的省市县级别
	areaResult := s.app.GetAreaConfig()
	if !areaResult.Ok || areaResult.Data == nil {
		return errors
	}

	areaData, ok := areaResult.Data.(map[string]interface{})
	if !ok || areaData == nil {
		return errors
	}

	provinceName := s.getStringValue(areaData["province_name"])
	cityName := s.getStringValue(areaData["city_name"])
	countryName := s.getStringValue(areaData["country_name"])

	// 获取当前数据的年份
	statDate := s.getStringValue(mainData[0]["stat_date"])
	if statDate == "" {
		return errors
	}

	// 从缓存获取数据
	cacheKey := attachment2CacheManager.GetDatabaseCacheKey(provinceName, cityName, countryName, statDate)
	cache, exists := attachment2CacheManager.GetDatabaseCache(cacheKey)
	if !exists {
		// 如果缓存中不存在，返回空缓存
		cache = &Attachment2DatabaseCache{
			CacheKey: cacheKey,
		}
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

	// 校验规则：同年份本单位所导入数值*120%应≥下级单位相加之和
	threshold := 1.2

	// 校验各个字段（使用缓存数据）
	if currentTotalCoal*threshold < cache.TotalCoal {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "煤合计数值*120%应大于等于下级单位相加之和"})
	}
	if currentRawCoal*threshold < cache.RawCoal {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "原煤数值*120%应大于等于下级单位相加之和"})
	}
	if currentWashedCoal*threshold < cache.WashedCoal {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "洗精煤数值*120%应大于等于下级单位相加之和"})
	}
	if currentOtherCoal*threshold < cache.OtherCoal {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "其他数值*120%应大于等于下级单位相加之和"})
	}
	if currentPowerGeneration*threshold < cache.PowerGen {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "火力发电数值*120%应大于等于下级单位相加之和"})
	}
	if currentHeating*threshold < cache.Heating {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "供热数值*120%应大于等于下级单位相加之和"})
	}
	if currentCoalWashing*threshold < cache.CoalWashing {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "煤炭洗选数值*120%应大于等于下级单位相加之和"})
	}
	if currentCoking*threshold < cache.Coking {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "炼焦数值*120%应大于等于下级单位相加之和"})
	}
	if currentOilRefining*threshold < cache.OilRefining {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "炼油及煤制油数值*120%应大于等于下级单位相加之和"})
	}
	if currentGasProduction*threshold < cache.GasProd {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "制气数值*120%应大于等于下级单位相加之和"})
	}
	if currentIndustry*threshold < cache.Industry {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "工业数值*120%应大于等于下级单位相加之和"})
	}
	if currentRawMaterials*threshold < cache.RawMaterials {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "工业（#用作原料、材料）数值*120%应大于等于下级单位相加之和"})
	}
	if currentOtherUses*threshold < cache.OtherUses {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "其他用途数值*120%应大于等于下级单位相加之和"})
	}
	if currentCoke*threshold < cache.Coke {
		errors = append(errors, ValidationError{RowNumber: 0, Message: "焦炭数值*120%应大于等于下级单位相加之和"})
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

		// 先尝试更新，通过受影响行数判断是否存在
		affectedRows, err := s.updateAttachment2DataByRegionAndYear(statDate, provinceName, cityName, countryName, record)
		if err != nil {
			return fmt.Errorf("更新数据失败: %v", err)
		}

		// 如果受影响行数为0，说明数据不存在，执行插入
		if affectedRows == 0 {
			err = s.insertAttachment2Data(record)
			if err != nil {
				return fmt.Errorf("插入数据失败: %v", err)
			}
		}

		// 更新内存缓存
		attachment2CacheManager.UpdateDatabaseCache(statDate, provinceName, cityName, countryName, record)
	}

	return nil
}

// updateAttachment2DataByRegionAndYear 根据地区和时间更新附件2数据
func (s *DataImportService) updateAttachment2DataByRegionAndYear(statDate, provinceName, cityName, countryName string, record map[string]interface{}) (int64, error) {
	// 先获取旧数据用于缓存更新
	oldData, err := s.getAttachment2DataByRegionAndYear(statDate, provinceName, cityName, countryName)

	// 对数值字段进行SM4加密
	encryptedValues := s.encryptAttachment2NumericFields(record)

	query := `UPDATE coal_consumption_report SET
		stat_date = ?, province_name = ?, city_name = ?, country_name = ?, unit_level = ?,
		total_coal = ?, raw_coal = ?, washed_coal = ?, other_coal = ?,
		power_generation = ?, heating = ?, coal_washing = ?, coking = ?,
		oil_refining = ?, gas_production = ?, industry = ?, raw_materials = ?,
		other_uses = ?, coke = ?
		WHERE stat_date = ? AND province_name = ? AND city_name = ? AND country_name = ?`

	// 计算unit_level
	unitLevel := s.calculateUnitLevel(provinceName, cityName, countryName)

	result, err := s.app.GetDB().Exec(query,
		statDate, provinceName, cityName, countryName, unitLevel,
		encryptedValues["total_coal"], encryptedValues["raw_coal"], encryptedValues["washed_coal"],
		encryptedValues["other_coal"], encryptedValues["power_generation"], encryptedValues["heating"],
		encryptedValues["coal_washing"], encryptedValues["coking"], encryptedValues["oil_refining"],
		encryptedValues["gas_production"], encryptedValues["industry"], encryptedValues["raw_materials"],
		encryptedValues["other_uses"], encryptedValues["coke"], statDate, provinceName, cityName, countryName)

	if err != nil {
		return 0, err
	}

	// 从QueryResult中获取受影响的行数
	var affectedRows int64 = 0
	if result.Ok && result.Data != nil {
		if dataMap, ok := result.Data.(map[string]interface{}); ok {
			if rowsAffected, exists := dataMap["rowsAffected"]; exists {
				if affected, ok := rowsAffected.(int64); ok {
					affectedRows = affected
				}
			}
		}
	}

	// 如果更新成功且有旧数据，更新缓存
	if affectedRows > 0 && oldData != nil {
		s.updateAttachment2DatabaseCacheForUpdate(statDate, provinceName, cityName, countryName, oldData, record)
		// 更新数据存在性缓存
		attachment2CacheManager.CacheDataExists(statDate, provinceName, cityName, countryName, true)
	}

	return affectedRows, nil
}

// getAttachment2DataByRegionAndYear 根据地区和时间获取附件2数据
func (s *DataImportService) getAttachment2DataByRegionAndYear(statDate, provinceName, cityName, countryName string) (map[string]interface{}, error) {
	query := `SELECT
		total_coal, raw_coal, washed_coal, other_coal,
		power_generation, heating, coal_washing, coking, oil_refining, gas_production,
		industry, raw_materials, other_uses, coke
	FROM coal_consumption_report
	WHERE stat_date = ? AND province_name = ? AND city_name = ? AND country_name = ?`

	result, err := s.app.GetDB().QueryRow(query, statDate, provinceName, cityName, countryName)
	if err != nil || result.Data == nil {
		return nil, err
	}

	// 解密数值字段
	record := result.Data.(map[string]interface{})
	decryptedRecord := make(map[string]interface{})

	decryptedRecord["total_coal"] = s.decryptValue(record["total_coal"])
	decryptedRecord["raw_coal"] = s.decryptValue(record["raw_coal"])
	decryptedRecord["washed_coal"] = s.decryptValue(record["washed_coal"])
	decryptedRecord["other_coal"] = s.decryptValue(record["other_coal"])
	decryptedRecord["power_generation"] = s.decryptValue(record["power_generation"])
	decryptedRecord["heating"] = s.decryptValue(record["heating"])
	decryptedRecord["coal_washing"] = s.decryptValue(record["coal_washing"])
	decryptedRecord["coking"] = s.decryptValue(record["coking"])
	decryptedRecord["oil_refining"] = s.decryptValue(record["oil_refining"])
	decryptedRecord["gas_production"] = s.decryptValue(record["gas_production"])
	decryptedRecord["industry"] = s.decryptValue(record["industry"])
	decryptedRecord["raw_materials"] = s.decryptValue(record["raw_materials"])
	decryptedRecord["other_uses"] = s.decryptValue(record["other_uses"])
	decryptedRecord["coke"] = s.decryptValue(record["coke"])

	return decryptedRecord, nil
}

// updateAttachment2DatabaseCacheForUpdate 更新附件2数据库缓存（用于UPDATE操作）
func (s *DataImportService) updateAttachment2DatabaseCacheForUpdate(statDate, provinceName, cityName, countryName string, oldRecord, newRecord map[string]interface{}) {
	// 使用缓存管理器更新缓存
	attachment2CacheManager.UpdateDatabaseCacheForUpdate(statDate, provinceName, cityName, countryName, oldRecord, newRecord)
}

// calculateUnitLevel 计算单位等级
// unit_level为单位等级：01 国家 02-省 03-市 04-县
// 如果县不为空为04,市不为空为03,省不为空为02, 省为空则为01
func (s *DataImportService) calculateUnitLevel(provinceName, cityName, countryName string) string {
	if countryName != "" {
		return "04" // 县
	}
	if cityName != "" {
		return "03" // 市
	}
	if provinceName != "" {
		return "02" // 省
	}
	return "01" // 国家
}

// isAttachment2FileImported 检查附件2文件是否已导入
func (s *DataImportService) isAttachment2FileImported(mainData []map[string]interface{}) bool {
	// 按Excel数据逐行检查，根据年份+省+市+县检查是否已导入
	for _, record := range mainData {
		statDate := s.getStringValue(record["stat_date"])
		provinceName := s.getStringValue(record["province_name"])
		cityName := s.getStringValue(record["city_name"])
		countryName := s.getStringValue(record["country_name"])

		// 直接检查内存缓存，内存中没有就是不存在
		if attachment2CacheManager.IsDataExistsInCache(statDate, provinceName, cityName, countryName) {
			return true // 检查到立即停止表示已导入
		}
	}
	return false
}

// saveAttachment2Data 保存附件2数据到数据库
func (s *DataImportService) saveAttachment2Data(mainData []map[string]interface{}) error {
	for _, record := range mainData {
		statDate := s.getStringValue(record["stat_date"])
		provinceName := s.getStringValue(record["province_name"])
		cityName := s.getStringValue(record["city_name"])
		countryName := s.getStringValue(record["country_name"])

		// 直接检查内存缓存，内存中没有就是不存在
		if attachment2CacheManager.IsDataExistsInCache(statDate, provinceName, cityName, countryName) {
			// 已存在数据，执行更新
			_, err := s.updateAttachment2DataByRegionAndYear(statDate, provinceName, cityName, countryName, record)
			if err != nil {
				return err
			}
		} else {

			// 不存在数据，执行插入
			err := s.insertAttachment2Data(record)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// saveAttachment2DataForModel 模型校验专用保存附件2数据到数据库（只使用INSERT）
func (s *DataImportService) saveAttachment2DataForModel(mainData []map[string]interface{}) error {
	for _, record := range mainData {

		// 直接执行插入操作，不检查数据是否已存在
		err := s.insertAttachment2Data(record)

		if err != nil {
			return err
		}

		// 更新内存缓存
		statDate := s.getStringValue(record["stat_date"])
		provinceName := s.getStringValue(record["province_name"])
		cityName := s.getStringValue(record["city_name"])
		countryName := s.getStringValue(record["country_name"])

		attachment2CacheManager.UpdateDatabaseCache(statDate, provinceName, cityName, countryName, record)

	}

	return nil
}

// insertAttachment2Data 插入附件2数据
func (s *DataImportService) insertAttachment2Data(record map[string]interface{}) error {

	record["obj_id"] = s.generateUUID()
	record["create_time"] = time.Now().Format("2006-01-02 15:04:05")

	// 对数值字段进行SM4加密
	encryptedValues := s.encryptAttachment2NumericFields(record)

	// 计算unit_level
	unitLevel := s.calculateUnitLevel(s.getStringValue(record["province_name"]), s.getStringValue(record["city_name"]), s.getStringValue(record["country_name"]))

	query := `INSERT INTO coal_consumption_report (
		obj_id, stat_date, province_name, city_name, country_name, unit_level, total_coal, raw_coal,
		washed_coal, other_coal, power_generation, heating, coal_washing, coking,
		oil_refining, gas_production, industry, raw_materials, other_uses, coke, create_time, create_user, is_check
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.app.GetDB().Exec(query,
		record["obj_id"], record["stat_date"], record["province_name"], record["city_name"],
		record["country_name"], unitLevel, encryptedValues["total_coal"], encryptedValues["raw_coal"], encryptedValues["washed_coal"],
		encryptedValues["other_coal"], encryptedValues["power_generation"], encryptedValues["heating"], encryptedValues["coal_washing"],
		encryptedValues["coking"], encryptedValues["oil_refining"], encryptedValues["gas_production"], encryptedValues["industry"],
		encryptedValues["raw_materials"], encryptedValues["other_uses"], encryptedValues["coke"], record["create_time"], s.app.GetCurrentOSUser(), EncryptedOne)
	if err != nil {
		return fmt.Errorf("保存数据失败: %v", err)
	}

	// 更新数据存在性缓存
	statDate := s.getStringValue(record["stat_date"])
	provinceName := s.getStringValue(record["province_name"])
	cityName := s.getStringValue(record["city_name"])
	countryName := s.getStringValue(record["country_name"])
	attachment2CacheManager.CacheDataExists(statDate, provinceName, cityName, countryName, true)

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
	maxCol := 18

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

// encryptAttachment2NumericFields 加密附件2数值字段
func (s *DataImportService) encryptAttachment2NumericFields(record map[string]interface{}) map[string]interface{} {
	numericFields := []string{
		"total_coal",
		"raw_coal",
		"washed_coal",
		"other_coal",
		"power_generation",
		"heating",
		"coal_washing",
		"coking",
		"oil_refining",
		"gas_production",
		"industry",
		"raw_materials",
		"other_uses",
		"coke",
	}
	return s.encryptNumericFields(record, numericFields)
}
