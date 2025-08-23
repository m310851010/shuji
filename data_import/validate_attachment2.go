package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ValidateAttachment2File 校验附件2文件
func (s *App) ValidateAttachment2File(filePath string) QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		s.insertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("读取Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 2. 解析Excel文件
	data, err := s.parseAttachment2Excel(f)
	if err != nil {
		s.insertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("解析Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateAttachment2Data(data)
	if len(validationErrors) > 0 {
		s.insertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	// 4. 查询该表是否有数据
	hasData := s.checkAttachment2HasData()

	// 5. 返回结果
	return QueryResult{
		Ok:      true,
		Data:    hasData,
		Message: "校验通过",
	}
}

// validateAttachment2Data 校验附件2数据
func (s *App) validateAttachment2Data(data []map[string]interface{}) []string {
	errors := []string{}

	// 1. 检查年份和单位是否为空
	for i, row := range data {
		fieldErrors := s.validateRequiredFields(row, Attachment2RequiredFields, i)
		errors = append(errors, fieldErrors...)
	}

	// 2. 检查区域与当前单位是否相符
	regionErrors := s.validateAttachment2Region(data)
	errors = append(errors, regionErrors...)

	return errors
}

// validateAttachment2Region 检查附件2区域与当前单位是否相符
func (s *App) validateAttachment2Region(data []map[string]interface{}) []string {
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

// checkAttachment2HasData 检查附件2表是否有数据
func (s *App) checkAttachment2HasData() bool {
	return s.checkTableHasData(TableCoalConsumptionReport)
}
