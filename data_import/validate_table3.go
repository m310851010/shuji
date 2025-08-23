package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ValidateTable3File 校验附表3文件
func (s *App) ValidateTable3File(filePath string) QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		s.insertImportRecord(fileName, "附表3", "上传失败", fmt.Sprintf("读取Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 2. 解析Excel文件
	data, err := s.parseTable3Excel(f)
	if err != nil {
		s.insertImportRecord(fileName, "附表3", "上传失败", fmt.Sprintf("解析Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateTable3Data(data)
	if len(validationErrors) > 0 {
		s.insertImportRecord(fileName, "附表3", "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	// 4. 查询该表是否有数据
	hasData := s.checkTable3HasData()

	// 5. 返回结果
	return QueryResult{
		Ok:      true,
		Data:    hasData,
		Message: "校验通过",
	}
}

// validateTable3Data 校验附表3数据
func (s *App) validateTable3Data(data []map[string]interface{}) []string {
	errors := []string{}

	// 1. 检查年份和单位是否为空
	for i, row := range data {
		fieldErrors := s.validateRequiredFields(row, Table3RequiredFields, i)
		errors = append(errors, fieldErrors...)
	}

	// 2. 检查区域与当前单位是否相符
	regionErrors := s.validateTable3Region(data)
	errors = append(errors, regionErrors...)

	// 3. 检查固定资产投资项目重复数据
	duplicateErrors := s.validateTable3DuplicateData(data)
	errors = append(errors, duplicateErrors...)

	return errors
}

// validateTable3Region 检查附表3区域与当前单位是否相符
func (s *App) validateTable3Region(data []map[string]interface{}) []string {
	errors := []string{}

	// 获取当前单位信息
	result := s.GetAreaConfig()
	if !result.Ok {
		// 如果获取失败，跳过区域校验
		return errors
	}

	// 解析返回的数据
	var currentProvince, currentCity, currentCountry string
	if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
		row := data[0]
		currentProvince = s.getStringValue(row["province_name"])
		currentCity = s.getStringValue(row["city_name"])
		currentCountry = s.getStringValue(row["country_name"])
	} else {
		// 如果没有配置，跳过区域校验
		return errors
	}

	for i, row := range data {
		provinceName := s.getStringValue(row["province_name"])
		cityName := s.getStringValue(row["city_name"])
		countryName := s.getStringValue(row["country_name"])

		// 检查区域是否与当前单位相符
		if provinceName != "" && currentProvince != "" && provinceName != currentProvince {
			errors = append(errors, fmt.Sprintf("第%d行：上传的数据单位与当前单位不符", i+1))
			continue
		}

		if cityName != "" && currentCity != "" && cityName != currentCity {
			errors = append(errors, fmt.Sprintf("第%d行：上传的数据单位与当前单位不符", i+1))
			continue
		}

		if countryName != "" && currentCountry != "" && countryName != currentCountry {
			errors = append(errors, fmt.Sprintf("第%d行：上传的数据单位与当前单位不符", i+1))
			continue
		}
	}

	return errors
}

// validateTable3DuplicateData 检查附表3重复数据
func (s *App) validateTable3DuplicateData(data []map[string]interface{}) []string {
	errors := []string{}

	// 用于存储已检查的项目信息
	projectMap := make(map[string]int)

	for i, row := range data {
		projectName := s.getStringValue(row["project_name"])
		projectCode := s.getStringValue(row["project_code"])
		approvalNumber := s.getStringValue(row["approval_number"])

		// 生成唯一标识
		key := fmt.Sprintf("%s|%s|%s", projectName, projectCode, approvalNumber)

		if existingIndex, exists := projectMap[key]; exists {
			errors = append(errors, fmt.Sprintf("第%d行：[项目名称、项目代码、审查意见文号]数据重复（与第%d行重复）", i+1, existingIndex+1))
		} else {
			projectMap[key] = i
		}
	}

	return errors
}

// checkTable3HasData 检查附表3表是否有数据
func (s *App) checkTable3HasData() bool {
	return s.checkTableHasData(TableFixedAssetsInvestmentProject)
}
