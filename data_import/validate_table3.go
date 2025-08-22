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

	// 强制校验规则实现
	for i, row := range data {
		fieldErrors := s.validateRequiredFields(row, Table3RequiredFields, i)
		errors = append(errors, fieldErrors...)
	}

	return errors
}

// checkTable3HasData 检查附表3表是否有数据
func (s *App) checkTable3HasData() bool {
	return s.checkTableHasData(TableFixedAssetsInvestmentProject)
}
