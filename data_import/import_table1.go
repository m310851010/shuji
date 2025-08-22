package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ImportTable1 导入附表1数据
func (s *App) ImportTable1(filePath string) QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
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
		s.insertImportRecord(fileName, "附表1", "上传失败", fmt.Sprintf("解析Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateTable1Data(mainData, usageData, equipData)
	if len(validationErrors) > 0 {
		s.insertImportRecord(fileName, "附表1", "上传失败", fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	// 4. 清空原表数据
	err = s.clearTable1Data()
	if err != nil {
		s.insertImportRecord(fileName, "附表1", "上传失败", fmt.Sprintf("清空原表数据失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("清空原表数据失败: %v", err),
		}
	}

	// 5. 导入数据到数据库
	err = s.importTable1Data(mainData, usageData, equipData)
	if err != nil {
		s.insertImportRecord(fileName, "附表1", "上传失败", fmt.Sprintf("导入数据失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("导入数据失败: %v", err),
		}
	}

	// 6. 复制缓存文件
	_, err = CopyCacheFile(filePath)
	if err != nil {
		// 复制失败不影响导入结果，只记录日志
		fmt.Printf("复制缓存文件失败: %v", err)
	}

	// 7. 插入成功记录
	s.insertImportRecord(fileName, "附表1", "上传成功", fmt.Sprintf("成功导入 %d 条主表数据, %d 条用途数据, %d 条设备数据", len(mainData), len(usageData), len(equipData)))

	return QueryResult{
		Ok: true,
		Data: map[string]interface{}{
			"mainCount":  len(mainData),
			"usageCount": len(usageData),
			"equipCount": len(equipData),
		},
		Message: "导入成功",
	}
}

// clearTable1Data 清空附表1相关表数据
func (s *App) clearTable1Data() error {
	// 清空附表1相关表数据
	tables := []string{
		TableEnterpriseCoalConsumptionMain,
		TableEnterpriseCoalConsumptionUsage,
		TableEnterpriseCoalConsumptionEquip,
	}

	for _, table := range tables {
		err := s.clearTableData(table)
		if err != nil {
			return fmt.Errorf("清空表失败 %s: %v", table, err)
		}
	}
	return nil
}

// importTable1Data 导入附表1数据到数据库
func (s *App) importTable1Data(mainData, usageData, equipData []map[string]interface{}) error {
	// TODO: 实现附表1数据导入逻辑
	// 需要导入到三个表：
	// 1. enterprise_coal_consumption_main
	// 2. enterprise_coal_consumption_usage
	// 3. enterprise_coal_consumption_equip

	// 注意：敏感数据需要使用SM4加密存储
	// 根据main.sql文件描述，以下字段需要加密：
	// - 各种煤炭消费量字段
	// - 企业名称等敏感信息

	return fmt.Errorf("importTable1Data 方法待实现")
}

// parseTable1Excel 解析附表1 Excel文件
func (s *App) parseTable1Excel(f *excelize.File) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	// TODO: 实现附表1 Excel解析逻辑
	// 需要解析三个表格区域：
	// 1. 综合能源消费情况和规模以上企业煤炭消费信息表 -> enterprise_coal_consumption_main
	// 2. 煤炭消费主要用途情况 -> enterprise_coal_consumption_usage
	// 3. 重点耗煤装置情况 -> enterprise_coal_consumption_equip
	return nil, nil, nil, fmt.Errorf("parseTable1Excel 方法待实现")
}
