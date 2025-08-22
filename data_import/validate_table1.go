package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ValidateTable1File 校验附表1文件
func (s *App) ValidateTable1File(filePath string) QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		// 插入导入记录
		s.insertImportRecord(fileName, "附表1", "上传失败", fmt.Sprintf("读取Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 2. 解析Excel文件
	mainData, usageData, equipData, err := s.parseTable1Excel(f)
	if err != nil {
		// 插入导入记录
		s.insertImportRecord(fileName, "附表1", "上传失败", fmt.Sprintf("解析Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateTable1Data(mainData, usageData, equipData)
	if len(validationErrors) > 0 {
		// 插入导入记录
		s.insertImportRecord(fileName, "附表1", "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	// 4. 查询该表是否有数据放进data属性
	hasData := s.checkTable1HasData()

	// 5. 返回QueryResult
	return QueryResult{
		Ok:      true,
		Data:    hasData,
		Message: "校验通过",
	}
}

// validateTable1Data 校验附表1数据
func (s *App) validateTable1Data(mainData, usageData, equipData []map[string]interface{}) []string {
	errors := []string{}

	// 强制校验规则实现
	// 1. 检查年份和单位是否为空
	for i, data := range mainData {
		fieldErrors := s.validateRequiredFields(data, Table1RequiredFields, i)
		errors = append(errors, fieldErrors...)
	}

	// 2. 检查企业是否在企业清单中（如果有清单的话）
	// TODO: 实现企业清单检查逻辑

	// 3. 检查企业名称和统一信用代码是否对应（如果有清单的话）
	// TODO: 实现企业名称和信用代码对应关系检查

	// 4. 检查数据单位与当前单位是否相符
	// TODO: 实现数据单位检查逻辑

	// 5. 检查文件格式是否正确
	// TODO: 实现文件格式检查逻辑

	return errors
}

// checkTable1HasData 检查附表1相关表是否有数据
func (s *App) checkTable1HasData() bool {
	return s.checkTableHasData(TableEnterpriseCoalConsumptionMain)
}
