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

	// 强制校验规则实现
	for i, row := range data {
		fieldErrors := s.validateRequiredFields(row, Attachment2RequiredFields, i)
		errors = append(errors, fieldErrors...)
	}

	return errors
}

// checkAttachment2HasData 检查附件2表是否有数据
func (s *App) checkAttachment2HasData() bool {
	return s.checkTableHasData(TableCoalConsumptionReport)
}
