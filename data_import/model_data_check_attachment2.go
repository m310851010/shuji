package data_import

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"shuji/db"
	"shuji/slices"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// AreaInfo 区域信息结构
type AreaInfo struct {
	Code string `json:"code"` // 区域代码
	Name string `json:"name"` // 区域名称
}

// EnhancedAreaConfig 增强的区域配置结构
type EnhancedAreaConfig struct {
	ObjID            string     `json:"obj_id"`            // 主键
	ProvinceName     string     `json:"province_name"`     // 省级名称
	CityName         string     `json:"city_name"`         // 市级名称
	CountryName      string     `json:"country_name"`      // 县级名称
	ProvinceCode     string     `json:"province_code"`     // 省级代码
	CityCode         string     `json:"city_code"`         // 市级代码
	CountryCode      string     `json:"country_code"`      // 县级代码
	DataLevel        int        `json:"data_level"`        // 数据级别：1-省级，2-市级，3-县级
	SubordinateAreas []AreaInfo `json:"subordinate_areas"` // 下级区域列表
}

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

	areaConfig := s.GetAreaConfig()

	var validationErrors []ValidationError
	var failedFiles []string

	for _, file := range files {
		// 检查是否xlsx或者xls文件
		if !file.IsDir() && s.isExcelFile(file.Name()) {
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

			err = s.coverAttachment2Data(mainData, file.Name(), areaConfig)
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
	// 使用包装函数来处理异常
	return s.modelDataCheckAttachment2WithRecover()
}

// modelDataCheckAttachment2WithRecover 带异常处理的附件2模型校验函数
func (s *DataImportService) modelDataCheckAttachment2WithRecover() db.QueryResult {
	var result db.QueryResult

	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ModelDataCheckAttachment2 发生异常: %v", r)
			// 设置错误结果
			result = db.QueryResult{
				Ok:      false,
				Message: fmt.Sprintf("函数执行异常: %v", r),
				Data:    nil,
			}
		}
	}()

	// 1. 读取缓存目录指定表格类型下的所有Excel文件

	cacheDir := s.app.GetCachePath(TableTypeAttachment2)

	files, err := os.ReadDir(cacheDir)
	if err != nil {
		errorMessage := fmt.Sprintf("读取缓存目录失败: %v", err)
		result = db.QueryResult{
			Ok:      false,
			Data:    []string{errorMessage},
			Message: errorMessage,
		}
		return result
	}

	s.initAttachment2CacheManager()

	areaConfig := s.GetAreaConfig()


	var validationErrors = []ValidationError{} // 错误信息
	var systemErrors = []ValidationError{}     // 系统错误信息
	var importedFiles = []string{}             // 导入的文件
	var coverFiles = []string{}                // 覆盖的文件
	var failedFiles = []string{}               // 失败的文件
	var hasExcelFile = false                   // 是否有Excel文件

	// 2. 循环调用对应的解析Excel函数
	for _, file := range files {
		// 检查是否xlsx或者xls文件
		if !file.IsDir() && s.isExcelFile(file.Name()) {
			hasExcelFile = true
			filePath := filepath.Join(cacheDir, file.Name())

			// 解析Excel文件 (skipValidate=true)
			f, err := excelize.OpenFile(filePath)
			if err != nil {
				systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 读取失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			mainData, err := s.parseAttachment2Excel(f, true)
			f.Close()

			if err != nil {
				systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 解析失败: %v", file.Name(), err)})
				failedFiles = append(failedFiles, filePath)
				continue
			}

			// 4. 调用校验函数,对每一行数据验证
			errors, hasDBError := s.validateAttachment2DataForModel(mainData, areaConfig)

			if len(errors) > 0 {
				lastRowNumber := 0
				if hasDBError && areaConfig.CountryName == "" {
					lastRowNumber = s.getExcelRowNumber(mainData[len(mainData) - 1]) + 2
				}
				// 校验失败，在Excel文件中错误行最后添加错误信息
				err = s.addValidationErrorsToExcelAttachment2(filePath, errors, lastRowNumber)

				if err != nil {
					msg := err.Error()
					// 如果错误是文件名长度超出限制，则跳过
					if err == excelize.ErrMaxFilePathLength {
						msg = "文件存放的路径过长，建议将文件放在磁盘一级目录再操作。"
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
			if s.isAttachment2FileImported(mainData) {
				coverFiles = append(coverFiles, filePath)
				continue
			}

			// 6. 如果没有导入过,把数据保存到相应的数据库表中
			err = s.saveAttachment2DataForModel(mainData, areaConfig)

			if err != nil {
				systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 保存数据失败: %v", file.Name(), err)})
			} else {
				// 删除该Excel文件
				err = os.Remove(filePath)
				if err != nil {
					systemErrors = append(systemErrors, ValidationError{RowNumber: 0, Message: fmt.Sprintf("文件 %s 删除失败: %v", file.Name(), err)})
				}
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
		err = s.createValidationErrorZip(failedFiles, TableTypeAttachment2, TableAttachment2)
		// 删除失败文件
		for _, filePath := range failedFiles {
			os.Remove(filePath)
		}

		if err != nil {
			result = db.QueryResult{
				Ok:      false,
				Message: fmt.Sprintf("创建错误报告失败: %v", err),
			}
			return result
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

// validateAttachment2DataForModel 校验附件2数据（模型校验专用）
func (s *DataImportService) validateAttachment2DataForModel(mainData []map[string]interface{}, areaConfig *EnhancedAreaConfig)( []ValidationError, bool) {
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
	}

	// 数据库验证
	dbErrors := s.validateAttachment2DatabaseRules(mainData, areaConfig)
	errors = append(errors, dbErrors...)

	return errors, len(dbErrors) > 0
}

// validateAttachment2NumericFields 校验附件2数值字段
func (s *DataImportService) validateAttachment2NumericFields(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 1. 分品种煤炭消费摸底部分校验
	totalCoal := s.parseFloat(s.getStringValue(data["total_coal"]))
	rawCoal := s.parseFloat(s.getStringValue(data["raw_coal"]))
	washedCoal := s.parseFloat(s.getStringValue(data["washed_coal"]))
	otherCoal := s.parseFloat(s.getStringValue(data["other_coal"]))

	// ①≧0
	if s.isIntegerLessThan(totalCoal, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "total_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(rawCoal, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "raw_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(washedCoal, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "washed_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(otherCoal, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "other_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他不能为负数", Cells: cells})
	}

	// ②≦200000
	if s.isIntegerGreaterThan(totalCoal, 200000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "total_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤合计不能大于200000", Cells: cells})
	}
	if s.isIntegerGreaterThan(rawCoal, 200000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "raw_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "原煤不能大于200000", Cells: cells})
	}
	if s.isIntegerGreaterThan(washedCoal, 200000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "washed_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "洗精煤不能大于200000", Cells: cells})
	}
	if s.isIntegerGreaterThan(otherCoal, 200000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "other_coal", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他不能大于200000", Cells: cells})
	}

	// 2. 分用途煤炭消费摸底部分校验
	powerGeneration := s.parseFloat(s.getStringValue(data["power_generation"]))
	heating := s.parseFloat(s.getStringValue(data["heating"]))
	coalWashing := s.parseFloat(s.getStringValue(data["coal_washing"]))
	coking := s.parseFloat(s.getStringValue(data["coking"]))
	oilRefining := s.parseFloat(s.getStringValue(data["oil_refining"]))
	gasProduction := s.parseFloat(s.getStringValue(data["gas_production"]))
	industry := s.parseFloat(s.getStringValue(data["industry"]))
	rawMaterials := s.parseFloat(s.getStringValue(data["raw_materials"]))
	otherUses := s.parseFloat(s.getStringValue(data["other_uses"]))

	// ①≧0
	if s.isIntegerLessThan(powerGeneration, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "power_generation", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "火力发电不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(heating, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "heating", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "供热不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(coalWashing, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coal_washing", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭洗选不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(coking, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coking", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼焦不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(oilRefining, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "oil_refining", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼油及煤制油不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(gasProduction, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "gas_production", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "制气不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(industry, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "industry", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(rawMaterials, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "raw_materials", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业（#用作原料、材料）不能为负数", Cells: cells})
	}
	if s.isIntegerLessThan(otherUses, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "other_uses", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他用途不能为负数", Cells: cells})
	}

	// ②≦100000
	if s.isIntegerGreaterThan(powerGeneration, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "power_generation", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "火力发电不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(heating, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "heating", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "供热不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(coalWashing, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coal_washing", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "煤炭洗选不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(coking, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coking", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼焦不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(oilRefining, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "oil_refining", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "炼油及煤制油不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(gasProduction, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "gas_production", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "制气不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(industry, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "industry", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(rawMaterials, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "raw_materials", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业（#用作原料、材料）不能大于100000", Cells: cells})
	}
	if s.isIntegerGreaterThan(otherUses, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "other_uses", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "其他用途不能大于100000", Cells: cells})
	}

	// 3. 焦炭消费摸底部分校验
	coke := s.parseFloat(s.getStringValue(data["coke"]))

	// ①≧0
	if s.isIntegerLessThan(coke, 0) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coke", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭不能为负数", Cells: cells})
	}

	// ②≦100000
	if s.isIntegerGreaterThan(coke, 100000) {
		cells := []string{s.getCellPosition(TableTypeAttachment2, "coke", rowNum)}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "焦炭不能大于100000", Cells: cells})
	}

	return errors
}

// validateAttachment2DataConsistency 校验附件2数据一致性（使用定点数运算）
func (s *DataImportService) validateAttachment2DataConsistency(data map[string]interface{}, rowNum int) []ValidationError {
	errors := []ValidationError{}

	// 2. 分用途煤炭消费摸底部分
	// ③工业≧工业（#用作原料、材料）
	industry := s.parseFloat(s.getStringValue(data["industry"]))
	rawMaterials := s.parseFloat(s.getStringValue(data["raw_materials"]))

	if s.isIntegerLessThan(industry, rawMaterials) {
		cells := []string{
			s.getCellPosition(TableTypeAttachment2, "industry", rowNum),
			s.getCellPosition(TableTypeAttachment2, "raw_materials", rowNum),
		}
		errors = append(errors, ValidationError{RowNumber: rowNum, Message: "工业应大于等于工业（#用作原料、材料）", Cells: cells})
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
		attachment2CacheManager.PreloadOptimizedCache()
	}
}

// validateAttachment2DatabaseRules 校验附件2数据库验证规则
func (s *DataImportService) validateAttachment2DatabaseRules(mainData []map[string]interface{}, areaConfig *EnhancedAreaConfig) []ValidationError {
	errors := []ValidationError{}

	if len(mainData) == 0 {
		return errors
	}

	// 获取当前数据的年份
	statDate := s.getStringValue(mainData[0]["stat_date"])
	if statDate == "" {
		return errors
	}

	// 初始化下辖区县累加数据
	subordinateData := &YearlyAggregatedData{
		StatDate:     statDate,
		TotalCoal:    0.0,
		RawCoal:      0.0,
		WashedCoal:   0.0,
		OtherCoal:    0.0,
		PowerGen:     0.0,
		Heating:      0.0,
		CoalWashing:  0.0,
		Coking:       0.0,
		OilRefining:  0.0,
		GasProd:      0.0,
		Industry:     0.0,
		RawMaterials: 0.0,
		OtherUses:    0.0,
		Coke:         0.0,
	}

	// 1. 从年份累计缓存获取数据（数据库中已存在的下辖县区数据）
	yearlyData, exists := attachment2CacheManager.GetYearlyAggregatedData(statDate)
	if exists {
		// 累加年份累计数据
		subordinateData.TotalCoal = yearlyData.TotalCoal
		subordinateData.RawCoal = yearlyData.RawCoal
		subordinateData.WashedCoal = yearlyData.WashedCoal
		subordinateData.OtherCoal = yearlyData.OtherCoal
		subordinateData.PowerGen = yearlyData.PowerGen
		subordinateData.Heating = yearlyData.Heating
		subordinateData.CoalWashing = yearlyData.CoalWashing
		subordinateData.Coking = yearlyData.Coking
		subordinateData.OilRefining = yearlyData.OilRefining
		subordinateData.GasProd = yearlyData.GasProd
		subordinateData.Industry = yearlyData.Industry
		subordinateData.RawMaterials = yearlyData.RawMaterials
		subordinateData.OtherUses = yearlyData.OtherUses
		subordinateData.Coke = yearlyData.Coke
	}

	// 计算当前本市数据总和（countryName为空的数据）
	currentTotalCoal := 0.0
	currentRawCoal := 0.0
	currentWashedCoal := 0.0
	currentOtherCoal := 0.0
	currentPowerGeneration := 0.0
	currentHeating := 0.0
	currentCoalWashing := 0.0
	currentCoking := 0.0
	currentOilRefining := 0.0
	currentGasProduction := 0.0
	currentIndustry := 0.0
	currentRawMaterials := 0.0
	currentOtherUses := 0.0
	currentCoke := 0.0

	cityData, exists := attachment2CacheManager.GetCityData(areaConfig.ProvinceName, areaConfig.CityName, statDate)
	if exists {
		currentTotalCoal = cityData.TotalCoal
		currentRawCoal = cityData.RawCoal
		currentWashedCoal = cityData.WashedCoal
		currentOtherCoal = cityData.OtherCoal
		currentPowerGeneration = cityData.PowerGen
		currentHeating = cityData.Heating
		currentCoalWashing = cityData.CoalWashing
		currentCoking = cityData.Coking
		currentOilRefining = cityData.OilRefining
		currentGasProduction = cityData.GasProd
		currentIndustry = cityData.Industry
		currentRawMaterials = cityData.RawMaterials
		currentOtherUses = cityData.OtherUses
		currentCoke = cityData.Coke
	}

	var rowNumber = 8
	for _, record := range mainData {
		recordCountryName := s.getStringValue(record["country_name"])
		recordProvinceName := s.getStringValue(record["province_name"])
		recordCityName := s.getStringValue(record["city_name"])

		totalCoal := s.parseFloat(s.getStringValue(record["total_coal"]))
		rawCoal := s.parseFloat(s.getStringValue(record["raw_coal"]))
		washedCoal := s.parseFloat(s.getStringValue(record["washed_coal"]))
		otherCoal := s.parseFloat(s.getStringValue(record["other_coal"]))
		powerGeneration := s.parseFloat(s.getStringValue(record["power_generation"]))
		heating := s.parseFloat(s.getStringValue(record["heating"]))
		coalWashing := s.parseFloat(s.getStringValue(record["coal_washing"]))
		coking := s.parseFloat(s.getStringValue(record["coking"]))
		oilRefining := s.parseFloat(s.getStringValue(record["oil_refining"]))
		gasProduction := s.parseFloat(s.getStringValue(record["gas_production"]))
		industry := s.parseFloat(s.getStringValue(record["industry"]))
		rawMaterials := s.parseFloat(s.getStringValue(record["raw_materials"]))
		otherUses := s.parseFloat(s.getStringValue(record["other_uses"]))
		coke := s.parseFloat(s.getStringValue(record["coke"]))

		// 检查是否已存在相同区县的数据，如果存在则先减去旧数据
		if existingData, exists := attachment2CacheManager.GetImportedData(statDate, recordProvinceName, recordCityName, recordCountryName); exists {
			// 从当前累加数据中减去已存在的旧数据
			if recordCountryName == "" {
				// 本市数据
				currentTotalCoal = s.subtractFloat64(currentTotalCoal, existingData.TotalCoal)
				currentRawCoal = s.subtractFloat64(currentRawCoal, existingData.RawCoal)
				currentWashedCoal = s.subtractFloat64(currentWashedCoal, existingData.WashedCoal)
				currentOtherCoal = s.subtractFloat64(currentOtherCoal, existingData.OtherCoal)
				currentPowerGeneration = s.subtractFloat64(currentPowerGeneration, existingData.PowerGen)
				currentHeating = s.subtractFloat64(currentHeating, existingData.Heating)
				currentCoalWashing = s.subtractFloat64(currentCoalWashing, existingData.CoalWashing)
				currentCoking = s.subtractFloat64(currentCoking, existingData.Coking)
				currentOilRefining = s.subtractFloat64(currentOilRefining, existingData.OilRefining)
				currentGasProduction = s.subtractFloat64(currentGasProduction, existingData.GasProd)
				currentIndustry = s.subtractFloat64(currentIndustry, existingData.Industry)
				currentRawMaterials = s.subtractFloat64(currentRawMaterials, existingData.RawMaterials)
				currentOtherUses = s.subtractFloat64(currentOtherUses, existingData.OtherUses)
				currentCoke = s.subtractFloat64(currentCoke, existingData.Coke)
			} else {
				// 下辖区县数据
				subordinateData.TotalCoal = s.subtractFloat64(subordinateData.TotalCoal, existingData.TotalCoal)
				subordinateData.RawCoal = s.subtractFloat64(subordinateData.RawCoal, existingData.RawCoal)
				subordinateData.WashedCoal = s.subtractFloat64(subordinateData.WashedCoal, existingData.WashedCoal)
				subordinateData.OtherCoal = s.subtractFloat64(subordinateData.OtherCoal, existingData.OtherCoal)
				subordinateData.PowerGen = s.subtractFloat64(subordinateData.PowerGen, existingData.PowerGen)
				subordinateData.Heating = s.subtractFloat64(subordinateData.Heating, existingData.Heating)
				subordinateData.CoalWashing = s.subtractFloat64(subordinateData.CoalWashing, existingData.CoalWashing)
				subordinateData.Coking = s.subtractFloat64(subordinateData.Coking, existingData.Coking)
				subordinateData.OilRefining = s.subtractFloat64(subordinateData.OilRefining, existingData.OilRefining)
				subordinateData.GasProd = s.subtractFloat64(subordinateData.GasProd, existingData.GasProd)
				subordinateData.Industry = s.subtractFloat64(subordinateData.Industry, existingData.Industry)
				subordinateData.RawMaterials = s.subtractFloat64(subordinateData.RawMaterials, existingData.RawMaterials)
				subordinateData.OtherUses = s.subtractFloat64(subordinateData.OtherUses, existingData.OtherUses)
				subordinateData.Coke = s.subtractFloat64(subordinateData.Coke, existingData.Coke)
			}
		}

		// 累加新数据
		if recordCountryName == "" {
			rowNumber = s.getExcelRowNumber(record)
			currentTotalCoal = s.addFloat64(currentTotalCoal, totalCoal)
			currentRawCoal = s.addFloat64(currentRawCoal, rawCoal)
			currentWashedCoal = s.addFloat64(currentWashedCoal, washedCoal)
			currentOtherCoal = s.addFloat64(currentOtherCoal, otherCoal)
			currentPowerGeneration = s.addFloat64(currentPowerGeneration, powerGeneration)
			currentHeating = s.addFloat64(currentHeating, heating)
			currentCoalWashing = s.addFloat64(currentCoalWashing, coalWashing)
			currentCoking = s.addFloat64(currentCoking, coking)
			currentOilRefining = s.addFloat64(currentOilRefining, oilRefining)
			currentGasProduction = s.addFloat64(currentGasProduction, gasProduction)
			currentIndustry = s.addFloat64(currentIndustry, industry)
			currentRawMaterials = s.addFloat64(currentRawMaterials, rawMaterials)
			currentOtherUses = s.addFloat64(currentOtherUses, otherUses)
			currentCoke = s.addFloat64(currentCoke, coke)
		} else {
			subordinateData.TotalCoal = s.addFloat64(subordinateData.TotalCoal, totalCoal)
			subordinateData.RawCoal = s.addFloat64(subordinateData.RawCoal, rawCoal)
			subordinateData.WashedCoal = s.addFloat64(subordinateData.WashedCoal, washedCoal)
			subordinateData.OtherCoal = s.addFloat64(subordinateData.OtherCoal, otherCoal)
			subordinateData.PowerGen = s.addFloat64(subordinateData.PowerGen, powerGeneration)
			subordinateData.Heating = s.addFloat64(subordinateData.Heating, heating)
			subordinateData.CoalWashing = s.addFloat64(subordinateData.CoalWashing, coalWashing)
			subordinateData.Coking = s.addFloat64(subordinateData.Coking, coking)
			subordinateData.OilRefining = s.addFloat64(subordinateData.OilRefining, oilRefining)
			subordinateData.GasProd = s.addFloat64(subordinateData.GasProd, gasProduction)
			subordinateData.Industry = s.addFloat64(subordinateData.Industry, industry)
			subordinateData.RawMaterials = s.addFloat64(subordinateData.RawMaterials, rawMaterials)
			subordinateData.OtherUses = s.addFloat64(subordinateData.OtherUses, otherUses)
			subordinateData.Coke = s.addFloat64(subordinateData.Coke, coke)
		}
	}

	if currentTotalCoal == 0 && currentRawCoal == 0{
		return errors
	}
	

	// 校验规则：同年份本单位所导入数值*120%应≥下级单位相加之和
	threshold := 1.2


	// 校验各个字段（使用下辖区县累加数据）
	// 计算 currentTotalCoal * threshold
	currentTotalCoalThreshold := s.multiplyFloat64(currentTotalCoal, threshold)
	if s.isIntegerLessThan(currentTotalCoalThreshold, subordinateData.TotalCoal) {
		cells := s.generateAllRelatedCells("total_coal", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "煤合计数值*120%应大于等于下级单位相加之和"})
	}

	currentRawCoalThreshold := s.multiplyFloat64(currentRawCoal, threshold)
	if s.isIntegerLessThan(currentRawCoalThreshold, subordinateData.RawCoal) {
		cells := s.generateAllRelatedCells("raw_coal", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "原煤数值*120%应大于等于下级单位相加之和"})
	}

	currentWashedCoalThreshold := s.multiplyFloat64(currentWashedCoal, threshold)
	if s.isIntegerLessThan(currentWashedCoalThreshold, subordinateData.WashedCoal) {
		cells := s.generateAllRelatedCells("washed_coal", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "洗精煤数值*120%应大于等于下级单位相加之和"})
	}

	currentOtherCoalThreshold := s.multiplyFloat64(currentOtherCoal, threshold)
	if s.isIntegerLessThan(currentOtherCoalThreshold, subordinateData.OtherCoal) {
		cells := s.generateAllRelatedCells("other_coal", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "其他数值*120%应大于等于下级单位相加之和"})
	}

	currentPowerGenerationThreshold := s.multiplyFloat64(currentPowerGeneration, threshold)
	if s.isIntegerLessThan(currentPowerGenerationThreshold, subordinateData.PowerGen) {
		cells := s.generateAllRelatedCells("power_generation", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "火力发电数值*120%应大于等于下级单位相加之和"})
	}

	currentHeatingThreshold := s.multiplyFloat64(currentHeating, threshold)
	if s.isIntegerLessThan(currentHeatingThreshold, subordinateData.Heating) {
		cells := s.generateAllRelatedCells("heating", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "供热数值*120%应大于等于下级单位相加之和"})
	}

	currentCoalWashingThreshold := s.multiplyFloat64(currentCoalWashing, threshold)
	if s.isIntegerLessThan(currentCoalWashingThreshold, subordinateData.CoalWashing) {
		cells := s.generateAllRelatedCells("coal_washing", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "煤炭洗选数值*120%应大于等于下级单位相加之和"})
	}

	currentCokingThreshold := s.multiplyFloat64(currentCoking, threshold)
	if s.isIntegerLessThan(currentCokingThreshold, subordinateData.Coking) {
		cells := s.generateAllRelatedCells("coking", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "炼焦数值*120%应大于等于下级单位相加之和"})
	}

	currentOilRefiningThreshold := s.multiplyFloat64(currentOilRefining, threshold)
	if s.isIntegerLessThan(currentOilRefiningThreshold, subordinateData.OilRefining) {
		cells := s.generateAllRelatedCells("oil_refining", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "炼油及煤制油数值*120%应大于等于下级单位相加之和"})
	}

	currentGasProductionThreshold := s.multiplyFloat64(currentGasProduction, threshold)
	if s.isIntegerLessThan(currentGasProductionThreshold, subordinateData.GasProd) {
		cells := s.generateAllRelatedCells("gas_production", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "制气数值*120%应大于等于下级单位相加之和"})
	}

	currentIndustryThreshold := s.multiplyFloat64(currentIndustry, threshold)
	if s.isIntegerLessThan(currentIndustryThreshold, subordinateData.Industry) {
		cells := s.generateAllRelatedCells("industry", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "工业数值*120%应大于等于下级单位相加之和"})
	}

	currentRawMaterialsThreshold := s.multiplyFloat64(currentRawMaterials, threshold)
	if s.isIntegerLessThan(currentRawMaterialsThreshold, subordinateData.RawMaterials) {
		cells := s.generateAllRelatedCells("raw_materials", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "工业（#用作原料、材料）数值*120%应大于等于下级单位相加之和"})
	}

	currentOtherUsesThreshold := s.multiplyFloat64(currentOtherUses, threshold)
	if s.isIntegerLessThan(currentOtherUsesThreshold, subordinateData.OtherUses) {
		cells := s.generateAllRelatedCells("other_uses", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "其他用途数值*120%应大于等于下级单位相加之和"})
	}

	currentCokeThreshold := s.multiplyFloat64(currentCoke, threshold)
	if s.isIntegerLessThan(currentCokeThreshold, subordinateData.Coke) {
		cells := s.generateAllRelatedCells("coke", mainData)
		errors = append(errors, ValidationError{RowNumber: rowNumber, Cells: cells, Message: "焦炭数值*120%应大于等于下级单位相加之和"})
	}

	return errors
}

// coverAttachment2Data 覆盖附件2数据
func (s *DataImportService) coverAttachment2Data(mainData []map[string]interface{}, fileName string, areaConfig *EnhancedAreaConfig) error {
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
			// 插入新数据后，更新缓存
			s.updateCacheWithNewData(statDate, provinceName, cityName, countryName, record)
		}
	}

	return nil
}

// updateAttachment2DataByRegionAndYear 根据地区和时间更新附件2数据
func (s *DataImportService) updateAttachment2DataByRegionAndYear(statDate, provinceName, cityName, countryName string, record map[string]interface{}) (int64, error) {
	// 对数值字段进行SM4加密
	encryptedValues := s.encryptAttachment2NumericFields(record)

	query := `UPDATE coal_consumption_report SET
		stat_date = ?, province_name = ?, city_name = ?, country_name = ?, unit_level = ?,
		total_coal = ?, raw_coal = ?, washed_coal = ?, other_coal = ?,
		power_generation = ?, heating = ?, coal_washing = ?, coking = ?,
		oil_refining = ?, gas_production = ?, industry = ?, raw_materials = ?,
		other_uses = ?, coke = ?, is_confirm = ?
		WHERE stat_date = ? AND province_name = ? AND city_name = ? AND country_name = ?`

	// 计算unit_level
	unitLevel := s.calculateUnitLevel(provinceName, cityName, countryName)

	result, err := s.app.GetDB().Exec(query,
		statDate, provinceName, cityName, countryName, unitLevel,
		encryptedValues["total_coal"], encryptedValues["raw_coal"], encryptedValues["washed_coal"],
		encryptedValues["other_coal"], encryptedValues["power_generation"], encryptedValues["heating"],
		encryptedValues["coal_washing"], encryptedValues["coking"], encryptedValues["oil_refining"],
		encryptedValues["gas_production"], encryptedValues["industry"], encryptedValues["raw_materials"],
		encryptedValues["other_uses"], encryptedValues["coke"], EncryptedZero, statDate, provinceName, cityName, countryName)

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

	// 如果更新成功，更新缓存（从缓存中获取旧数据）
	if affectedRows > 0 {
		s.updateAttachment2DatabaseCacheForUpdate(statDate, provinceName, cityName, countryName, nil, record)
	}

	return affectedRows, nil
}


// updateAttachment2DatabaseCacheForUpdate 更新附件2数据库缓存（用于UPDATE操作）
func (s *DataImportService) updateAttachment2DatabaseCacheForUpdate(statDate, provinceName, cityName, countryName string, oldRecord, newRecord map[string]interface{}) {
	// 从缓存中获取旧数据，而不是从数据库读取
	oldData, exists := attachment2CacheManager.GetImportedData(statDate, provinceName, cityName, countryName)
	if !exists {
		// 如果缓存中没有旧数据，说明这是新数据，直接更新缓存
		s.updateCacheWithNewData(statDate, provinceName, cityName, countryName, newRecord)
		return
	}

	// 解析新数据
	newTotalCoal := s.parseFloat(s.getStringValue(newRecord["total_coal"]))
	newRawCoal := s.parseFloat(s.getStringValue(newRecord["raw_coal"]))
	newWashedCoal := s.parseFloat(s.getStringValue(newRecord["washed_coal"]))
	newOtherCoal := s.parseFloat(s.getStringValue(newRecord["other_coal"]))
	newPowerGeneration := s.parseFloat(s.getStringValue(newRecord["power_generation"]))
	newHeating := s.parseFloat(s.getStringValue(newRecord["heating"]))
	newCoalWashing := s.parseFloat(s.getStringValue(newRecord["coal_washing"]))
	newCoking := s.parseFloat(s.getStringValue(newRecord["coking"]))
	newOilRefining := s.parseFloat(s.getStringValue(newRecord["oil_refining"]))
	newGasProduction := s.parseFloat(s.getStringValue(newRecord["gas_production"]))
	newIndustry := s.parseFloat(s.getStringValue(newRecord["industry"]))
	newRawMaterials := s.parseFloat(s.getStringValue(newRecord["raw_materials"]))
	newOtherUses := s.parseFloat(s.getStringValue(newRecord["other_uses"]))
	newCoke := s.parseFloat(s.getStringValue(newRecord["coke"]))

	// 计算差值（新数据 - 旧数据）
	diffTotalCoal := s.subtractFloat64(newTotalCoal, oldData.TotalCoal)
	diffRawCoal := s.subtractFloat64(newRawCoal, oldData.RawCoal)
	diffWashedCoal := s.subtractFloat64(newWashedCoal, oldData.WashedCoal)
	diffOtherCoal := s.subtractFloat64(newOtherCoal, oldData.OtherCoal)
	diffPowerGeneration := s.subtractFloat64(newPowerGeneration, oldData.PowerGen)
	diffHeating := s.subtractFloat64(newHeating, oldData.Heating)
	diffCoalWashing := s.subtractFloat64(newCoalWashing, oldData.CoalWashing)
	diffCoking := s.subtractFloat64(newCoking, oldData.Coking)
	diffOilRefining := s.subtractFloat64(newOilRefining, oldData.OilRefining)
	diffGasProduction := s.subtractFloat64(newGasProduction, oldData.GasProd)
	diffIndustry := s.subtractFloat64(newIndustry, oldData.Industry)
	diffRawMaterials := s.subtractFloat64(newRawMaterials, oldData.RawMaterials)
	diffOtherUses := s.subtractFloat64(newOtherUses, oldData.OtherUses)
	diffCoke := s.subtractFloat64(newCoke, oldData.Coke)

	// 更新年份累计缓存（如果是下辖县区数据）
	if countryName != "" {
		yearlyData, exists := attachment2CacheManager.GetYearlyAggregatedData(statDate)
		if exists {
			// 减去旧数据，加上新数据（即加上差值）
			yearlyData.TotalCoal = s.addFloat64(yearlyData.TotalCoal, diffTotalCoal)
			yearlyData.RawCoal = s.addFloat64(yearlyData.RawCoal, diffRawCoal)
			yearlyData.WashedCoal = s.addFloat64(yearlyData.WashedCoal, diffWashedCoal)
			yearlyData.OtherCoal = s.addFloat64(yearlyData.OtherCoal, diffOtherCoal)
			yearlyData.PowerGen = s.addFloat64(yearlyData.PowerGen, diffPowerGeneration)
			yearlyData.Heating = s.addFloat64(yearlyData.Heating, diffHeating)
			yearlyData.CoalWashing = s.addFloat64(yearlyData.CoalWashing, diffCoalWashing)
			yearlyData.Coking = s.addFloat64(yearlyData.Coking, diffCoking)
			yearlyData.OilRefining = s.addFloat64(yearlyData.OilRefining, diffOilRefining)
			yearlyData.GasProd = s.addFloat64(yearlyData.GasProd, diffGasProduction)
			yearlyData.Industry = s.addFloat64(yearlyData.Industry, diffIndustry)
			yearlyData.RawMaterials = s.addFloat64(yearlyData.RawMaterials, diffRawMaterials)
			yearlyData.OtherUses = s.addFloat64(yearlyData.OtherUses, diffOtherUses)
			yearlyData.Coke = s.addFloat64(yearlyData.Coke, diffCoke)
		}
	}

	// 更新市数据缓存（如果是本市数据）
	if countryName == "" {
		cityData, exists := attachment2CacheManager.GetCityData(provinceName, cityName, statDate)
		if exists {
			// 减去旧数据，加上新数据（即加上差值）
			cityData.TotalCoal = s.addFloat64(cityData.TotalCoal, diffTotalCoal)
			cityData.RawCoal = s.addFloat64(cityData.RawCoal, diffRawCoal)
			cityData.WashedCoal = s.addFloat64(cityData.WashedCoal, diffWashedCoal)
			cityData.OtherCoal = s.addFloat64(cityData.OtherCoal, diffOtherCoal)
			cityData.PowerGen = s.addFloat64(cityData.PowerGen, diffPowerGeneration)
			cityData.Heating = s.addFloat64(cityData.Heating, diffHeating)
			cityData.CoalWashing = s.addFloat64(cityData.CoalWashing, diffCoalWashing)
			cityData.Coking = s.addFloat64(cityData.Coking, diffCoking)
			cityData.OilRefining = s.addFloat64(cityData.OilRefining, diffOilRefining)
			cityData.GasProd = s.addFloat64(cityData.GasProd, diffGasProduction)
			cityData.Industry = s.addFloat64(cityData.Industry, diffIndustry)
			cityData.RawMaterials = s.addFloat64(cityData.RawMaterials, diffRawMaterials)
			cityData.OtherUses = s.addFloat64(cityData.OtherUses, diffOtherUses)
			cityData.Coke = s.addFloat64(cityData.Coke, diffCoke)
		}
	}

	// 更新已导入数据缓存
	newIndicatorData := &YearlyAggregatedData{
		StatDate:     statDate,
		TotalCoal:    newTotalCoal,
		RawCoal:      newRawCoal,
		WashedCoal:   newWashedCoal,
		OtherCoal:    newOtherCoal,
		PowerGen:     newPowerGeneration,
		Heating:      newHeating,
		CoalWashing:  newCoalWashing,
		Coking:       newCoking,
		OilRefining:  newOilRefining,
		GasProd:      newGasProduction,
		Industry:     newIndustry,
		RawMaterials: newRawMaterials,
		OtherUses:    newOtherUses,
		Coke:         newCoke,
	}
	attachment2CacheManager.SetImportedData(statDate, provinceName, cityName, countryName, newIndicatorData)
}

// updateCacheWithNewData 更新缓存中的新数据（用于INSERT操作）
func (s *DataImportService) updateCacheWithNewData(statDate, provinceName, cityName, countryName string, record map[string]interface{}) {
	// 解析新数据
	newTotalCoal := s.parseFloat(s.getStringValue(record["total_coal"]))
	newRawCoal := s.parseFloat(s.getStringValue(record["raw_coal"]))
	newWashedCoal := s.parseFloat(s.getStringValue(record["washed_coal"]))
	newOtherCoal := s.parseFloat(s.getStringValue(record["other_coal"]))
	newPowerGeneration := s.parseFloat(s.getStringValue(record["power_generation"]))
	newHeating := s.parseFloat(s.getStringValue(record["heating"]))
	newCoalWashing := s.parseFloat(s.getStringValue(record["coal_washing"]))
	newCoking := s.parseFloat(s.getStringValue(record["coking"]))
	newOilRefining := s.parseFloat(s.getStringValue(record["oil_refining"]))
	newGasProduction := s.parseFloat(s.getStringValue(record["gas_production"]))
	newIndustry := s.parseFloat(s.getStringValue(record["industry"]))
	newRawMaterials := s.parseFloat(s.getStringValue(record["raw_materials"]))
	newOtherUses := s.parseFloat(s.getStringValue(record["other_uses"]))
	newCoke := s.parseFloat(s.getStringValue(record["coke"]))

	// 更新年份累计缓存（如果是下辖县区数据）
	if countryName != "" {
		yearlyData, exists := attachment2CacheManager.GetYearlyAggregatedData(statDate)
		if exists {
			// 累加新数据
			yearlyData.TotalCoal = s.addFloat64(yearlyData.TotalCoal, newTotalCoal)
			yearlyData.RawCoal = s.addFloat64(yearlyData.RawCoal, newRawCoal)
			yearlyData.WashedCoal = s.addFloat64(yearlyData.WashedCoal, newWashedCoal)
			yearlyData.OtherCoal = s.addFloat64(yearlyData.OtherCoal, newOtherCoal)
			yearlyData.PowerGen = s.addFloat64(yearlyData.PowerGen, newPowerGeneration)
			yearlyData.Heating = s.addFloat64(yearlyData.Heating, newHeating)
			yearlyData.CoalWashing = s.addFloat64(yearlyData.CoalWashing, newCoalWashing)
			yearlyData.Coking = s.addFloat64(yearlyData.Coking, newCoking)
			yearlyData.OilRefining = s.addFloat64(yearlyData.OilRefining, newOilRefining)
			yearlyData.GasProd = s.addFloat64(yearlyData.GasProd, newGasProduction)
			yearlyData.Industry = s.addFloat64(yearlyData.Industry, newIndustry)
			yearlyData.RawMaterials = s.addFloat64(yearlyData.RawMaterials, newRawMaterials)
			yearlyData.OtherUses = s.addFloat64(yearlyData.OtherUses, newOtherUses)
			yearlyData.Coke = s.addFloat64(yearlyData.Coke, newCoke)
		}
	}

	// 更新市数据缓存（如果是本市数据）
	if countryName == "" {
		cityData, exists := attachment2CacheManager.GetCityData(provinceName, cityName, statDate)
		if exists {
			// 累加新数据
			cityData.TotalCoal = s.addFloat64(cityData.TotalCoal, newTotalCoal)
			cityData.RawCoal = s.addFloat64(cityData.RawCoal, newRawCoal)
			cityData.WashedCoal = s.addFloat64(cityData.WashedCoal, newWashedCoal)
			cityData.OtherCoal = s.addFloat64(cityData.OtherCoal, newOtherCoal)
			cityData.PowerGen = s.addFloat64(cityData.PowerGen, newPowerGeneration)
			cityData.Heating = s.addFloat64(cityData.Heating, newHeating)
			cityData.CoalWashing = s.addFloat64(cityData.CoalWashing, newCoalWashing)
			cityData.Coking = s.addFloat64(cityData.Coking, newCoking)
			cityData.OilRefining = s.addFloat64(cityData.OilRefining, newOilRefining)
			cityData.GasProd = s.addFloat64(cityData.GasProd, newGasProduction)
			cityData.Industry = s.addFloat64(cityData.Industry, newIndustry)
			cityData.RawMaterials = s.addFloat64(cityData.RawMaterials, newRawMaterials)
			cityData.OtherUses = s.addFloat64(cityData.OtherUses, newOtherUses)
			cityData.Coke = s.addFloat64(cityData.Coke, newCoke)
		}
	}

	// 存储新数据到已导入缓存
	newIndicatorData := &YearlyAggregatedData{
		StatDate:     statDate,
		TotalCoal:    newTotalCoal,
		RawCoal:      newRawCoal,
		WashedCoal:   newWashedCoal,
		OtherCoal:    newOtherCoal,
		PowerGen:     newPowerGeneration,
		Heating:      newHeating,
		CoalWashing:  newCoalWashing,
		Coking:       newCoking,
		OilRefining:  newOilRefining,
		GasProd:      newGasProduction,
		Industry:     newIndustry,
		RawMaterials: newRawMaterials,
		OtherUses:    newOtherUses,
		Coke:         newCoke,
	}
	attachment2CacheManager.SetImportedData(statDate, provinceName, cityName, countryName, newIndicatorData)
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

		// 使用新的优化缓存结构检查数据是否存在
		if attachment2CacheManager.IsDataExistsInOptimizedCache(statDate, provinceName, cityName, countryName) {
			return true // 检查到立即停止表示已导入
		}
	}
	return false
}


// saveAttachment2DataForModel 模型校验专用保存附件2数据到数据库（只使用INSERT）
func (s *DataImportService) saveAttachment2DataForModel(mainData []map[string]interface{}, areaConfig *EnhancedAreaConfig) error {
	for _, record := range mainData {

		// 直接执行插入操作，不检查数据是否已存在
		err := s.insertAttachment2Data(record)

		if err != nil {
			return err
		}

	}

	// 获取区域配置并更新优化缓存
	statDate := s.getStringValue(mainData[0]["stat_date"])
	s.UpdateOptimizedCacheAfterUpload(areaConfig, statDate, mainData)

	return nil
}

// insertAttachment2Data 插入附件2数据
func (s *DataImportService) insertAttachment2Data(record map[string]interface{}) error {

	record["obj_id"] = s.generateUUID()
	record["create_time"] = time.Now().UnixMilli()

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
		encryptedValues["raw_materials"], encryptedValues["other_uses"], encryptedValues["coke"], record["create_time"], s.app.GetAreaStr(), EncryptedOne)
	if err != nil {
		return fmt.Errorf("保存数据失败: %v", err)
	}

	return nil
}

// addValidationErrorsToExcelAttachment2 在附件2Excel文件中添加校验错误信息
func (s *DataImportService) addValidationErrorsToExcelAttachment2(filePath string, errors []ValidationError, lastRowNumber int) error {
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


	style, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFF00"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
	if err != nil {
		fmt.Println(err)
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
		
		f.SetCellStyle(sheetName, errorCellName, errorCellName, style)
		// 设置错误信息列的宽度为50
		colName, _ := excelize.ColumnNumberToName(errorCol)
		f.SetColWidth(sheetName, colName, colName, 50)
	}

	// 在最后一行添加备注
	if lastRowNumber > 0 {
		cellName, err := excelize.CoordinatesToCellName(1, lastRowNumber)
		if err != nil {
			return err
		}
		maxCellName, err := excelize.CoordinatesToCellName(maxCol, lastRowNumber)
		if err != nil {
			return err
		}
		f.MergeCell(sheetName, cellName, maxCellName)
		f.SetCellValue(sheetName, cellName, "本错误无法精准定位出数据不合理的区域，建议将全市及各县区数据全部进行再次检查，确认无误后再次整体导入")
		f.SetCellStyle(sheetName, cellName, cellName, style)
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

// UpdateOptimizedCacheAfterUpload 上传成功后更新优化缓存
func (s *DataImportService) UpdateOptimizedCacheAfterUpload(areaConfig *EnhancedAreaConfig, statDate string, mainData []map[string]interface{}) error {

	// 1. 更新年份累计数据（下辖县区数据累加）
	yearlyData := &YearlyAggregatedData{
		StatDate:     statDate,
		TotalCoal:    0.0,
		RawCoal:      0.0,
		WashedCoal:   0.0,
		OtherCoal:    0.0,
		PowerGen:     0.0,
		Heating:      0.0,
		CoalWashing:  0.0,
		Coking:       0.0,
		OilRefining:  0.0,
		GasProd:      0.0,
		Industry:     0.0,
		RawMaterials: 0.0,
		OtherUses:    0.0,
		Coke:         0.0,
	}

	// 2. 更新当前市数据缓存（本市数据）
	cityData := &CityData{
		ProvinceName: areaConfig.ProvinceName,
		CityName:     areaConfig.CityName,
		StatDate:     statDate,
		TotalCoal:    0.0,
		RawCoal:      0.0,
		WashedCoal:   0.0,
		OtherCoal:    0.0,
		PowerGen:     0.0,
		Heating:      0.0,
		CoalWashing:  0.0,
		Coking:       0.0,
		OilRefining:  0.0,
		GasProd:      0.0,
		Industry:     0.0,
		RawMaterials: 0.0,
		OtherUses:    0.0,
		Coke:         0.0,
	}

	// 处理上传的数据
	for _, record := range mainData {
		recordStatDate := s.getStringValue(record["stat_date"])
		recordProvinceName := s.getStringValue(record["province_name"])
		recordCityName := s.getStringValue(record["city_name"])
		recordCountryName := s.getStringValue(record["country_name"])

		// 存储指标数据到已导入缓存
		indicatorData := &YearlyAggregatedData{
			StatDate:     recordStatDate,
			TotalCoal:    s.parseFloat(s.getStringValue(record["total_coal"])),
			RawCoal:      s.parseFloat(s.getStringValue(record["raw_coal"])),
			WashedCoal:   s.parseFloat(s.getStringValue(record["washed_coal"])),
			OtherCoal:    s.parseFloat(s.getStringValue(record["other_coal"])),
			PowerGen:     s.parseFloat(s.getStringValue(record["power_generation"])),
			Heating:      s.parseFloat(s.getStringValue(record["heating"])),
			CoalWashing:  s.parseFloat(s.getStringValue(record["coal_washing"])),
			Coking:       s.parseFloat(s.getStringValue(record["coking"])),
			OilRefining:  s.parseFloat(s.getStringValue(record["oil_refining"])),
			GasProd:      s.parseFloat(s.getStringValue(record["gas_production"])),
			Industry:     s.parseFloat(s.getStringValue(record["industry"])),
			RawMaterials: s.parseFloat(s.getStringValue(record["raw_materials"])),
			OtherUses:    s.parseFloat(s.getStringValue(record["other_uses"])),
			Coke:         s.parseFloat(s.getStringValue(record["coke"])),
		}
		attachment2CacheManager.SetImportedData(recordStatDate, recordProvinceName, recordCityName, recordCountryName, indicatorData)

		// 判断是否为下辖县区数据（countryName不为空）
		if recordCountryName != "" {
			yearlyData.TotalCoal = s.addFloat64(yearlyData.TotalCoal, s.parseFloat(s.getStringValue(record["total_coal"])))
			yearlyData.RawCoal = s.addFloat64(yearlyData.RawCoal, s.parseFloat(s.getStringValue(record["raw_coal"])))
			yearlyData.WashedCoal = s.addFloat64(yearlyData.WashedCoal, s.parseFloat(s.getStringValue(record["washed_coal"])))
			yearlyData.OtherCoal = s.addFloat64(yearlyData.OtherCoal, s.parseFloat(s.getStringValue(record["other_coal"])))
			yearlyData.PowerGen = s.addFloat64(yearlyData.PowerGen, s.parseFloat(s.getStringValue(record["power_generation"])))
			yearlyData.Heating = s.addFloat64(yearlyData.Heating, s.parseFloat(s.getStringValue(record["heating"])))
			yearlyData.CoalWashing = s.addFloat64(yearlyData.CoalWashing, s.parseFloat(s.getStringValue(record["coal_washing"])))
			yearlyData.Coking = s.addFloat64(yearlyData.Coking, s.parseFloat(s.getStringValue(record["coking"])))
			yearlyData.OilRefining = s.addFloat64(yearlyData.OilRefining, s.parseFloat(s.getStringValue(record["oil_refining"])))
			yearlyData.GasProd = s.addFloat64(yearlyData.GasProd, s.parseFloat(s.getStringValue(record["gas_production"])))
			yearlyData.Industry = s.addFloat64(yearlyData.Industry, s.parseFloat(s.getStringValue(record["industry"])))
			yearlyData.RawMaterials = s.addFloat64(yearlyData.RawMaterials, s.parseFloat(s.getStringValue(record["raw_materials"])))
			yearlyData.OtherUses = s.addFloat64(yearlyData.OtherUses, s.parseFloat(s.getStringValue(record["other_uses"])))
			yearlyData.Coke = s.addFloat64(yearlyData.Coke, s.parseFloat(s.getStringValue(record["coke"])))
		}

		// 判断是否为当前市的数据（countryName为空）
		if recordCountryName == ""  {
			cityData.TotalCoal = s.addFloat64(cityData.TotalCoal, s.parseFloat(s.getStringValue(record["total_coal"])))
			cityData.RawCoal = s.addFloat64(cityData.RawCoal, s.parseFloat(s.getStringValue(record["raw_coal"])))
			cityData.WashedCoal = s.addFloat64(cityData.WashedCoal, s.parseFloat(s.getStringValue(record["washed_coal"])))
			cityData.OtherCoal = s.addFloat64(cityData.OtherCoal, s.parseFloat(s.getStringValue(record["other_coal"])))
			cityData.PowerGen = s.addFloat64(cityData.PowerGen, s.parseFloat(s.getStringValue(record["power_generation"])))
			cityData.Heating = s.addFloat64(cityData.Heating, s.parseFloat(s.getStringValue(record["heating"])))
			cityData.CoalWashing = s.addFloat64(cityData.CoalWashing, s.parseFloat(s.getStringValue(record["coal_washing"])))
			cityData.Coking = s.addFloat64(cityData.Coking, s.parseFloat(s.getStringValue(record["coking"])))
			cityData.OilRefining = s.addFloat64(cityData.OilRefining, s.parseFloat(s.getStringValue(record["oil_refining"])))
			cityData.GasProd = s.addFloat64(cityData.GasProd, s.parseFloat(s.getStringValue(record["gas_production"])))
			cityData.Industry = s.addFloat64(cityData.Industry, s.parseFloat(s.getStringValue(record["industry"])))
			cityData.RawMaterials = s.addFloat64(cityData.RawMaterials, s.parseFloat(s.getStringValue(record["raw_materials"])))
			cityData.OtherUses = s.addFloat64(cityData.OtherUses, s.parseFloat(s.getStringValue(record["other_uses"])))
			cityData.Coke = s.addFloat64(cityData.Coke, s.parseFloat(s.getStringValue(record["coke"])))
		}
	}

	// 更新年份累计缓存
	attachment2CacheManager.UpdateYearlyAggregatedData(statDate, yearlyData)

	// 设置当前市数据缓存
	attachment2CacheManager.SetCityData(areaConfig.ProvinceName, areaConfig.CityName, statDate, cityData)

	return nil
}

// generateAllRelatedCells 生成所有相关单元格位置（包括本市数据和下辖区县数据）
func (s *DataImportService) generateAllRelatedCells(fieldName string, mainData []map[string]interface{}) []string {
	var cells []string
	
	// 添加本市数据的单元格
	for _, rowData := range mainData {
		cell := s.getCellPosition(TableTypeAttachment2, fieldName, s.getExcelRowNumber(rowData))
		cells = append(cells, cell)
	}
	
	return cells
}
