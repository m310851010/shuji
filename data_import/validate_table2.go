package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ValidateTable2File 校验附表2文件
func (s *App) ValidateTable2File(filePath string) QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		s.insertImportRecord(fileName, "附表2", "上传失败", fmt.Sprintf("读取Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 2. 解析Excel文件
	data, err := s.parseTable2Excel(f)
	if err != nil {
		s.insertImportRecord(fileName, "附表2", "上传失败", fmt.Sprintf("解析Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateTable2Data(data)
	if len(validationErrors) > 0 {
		s.insertImportRecord(fileName, "附表2", "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	// 4. 查询该表是否有数据
	hasData := s.checkTable2HasData()

	// 5. 返回结果
	return QueryResult{
		Ok:      true,
		Data:    hasData,
		Message: "校验通过",
	}
}

// validateTable2Data 校验附表2数据
func (s *App) validateTable2Data(data []map[string]interface{}) []string {
	errors := []string{}

	// 1. 检查年份和单位是否为空
	for i, row := range data {
		fieldErrors := s.validateRequiredFields(row, Table2RequiredFields, i)
		errors = append(errors, fieldErrors...)

		// 2. 检查企业是否在企业清单中（如果有清单的话）
		unitName := s.getStringValue(row["unit_name"])
		creditCode := s.getStringValue(row["credit_code"])
		enterpriseListErrors := s.validateEnterpriseNameCreditCodeCorrespondence(unitName, creditCode, i, true)
		errors = append(errors, enterpriseListErrors...)

		// 3. 检查企业名称和统一信用代码是否对应（如果有清单的话）
		correspondenceErrors := s.validateEnterpriseNameCreditCodeCorrespondence(unitName, creditCode, i, false)
		errors = append(errors, correspondenceErrors...)
	}

	return errors
}

// checkTable2HasData 检查附表2表是否有数据
func (s *App) checkTable2HasData() bool {
	return s.checkTableHasData(TableCriticalCoalEquipmentConsumption)
}
