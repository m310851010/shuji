package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ImportTable2 导入附表2数据
func (s *App) ImportTable2(filePath string) QueryResult {
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

	// 4. 清空原表数据
	err = s.clearTable2Data()
	if err != nil {
		s.insertImportRecord(fileName, "附表2", "上传失败", fmt.Sprintf("清空原表数据失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("清空原表数据失败: %v", err),
		}
	}

	// 5. 导入数据到数据库
	err = s.importTable2Data(data)
	if err != nil {
		s.insertImportRecord(fileName, "附表2", "上传失败", fmt.Sprintf("导入数据失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("导入数据失败: %v", err),
		}
	}

	// 6. 复制缓存文件
	_, err = CopyCacheFile(filePath)
	if err != nil {
		fmt.Printf("复制缓存文件失败: %v", err)
	}

	// 7. 插入成功记录
	s.insertImportRecord(fileName, "附表2", "上传成功", fmt.Sprintf("成功导入 %d 条数据", len(data)))

	return QueryResult{
		Ok:      true,
		Data:    len(data),
		Message: "导入成功",
	}
}

// clearTable2Data 清空附表2表数据
func (s *App) clearTable2Data() error {
	return s.clearTableData(TableCriticalCoalEquipmentConsumption)
}

// importTable2Data 导入附表2数据到数据库
func (s *App) importTable2Data(data []map[string]interface{}) error {
	// TODO: 实现附表2数据导入逻辑
	// 导入到表：critical_coal_equipment_consumption

	// 注意：敏感数据需要使用SM4加密存储
	// 根据main.sql文件描述，以下字段需要加密：
	// - annual_coal_consumption (年耗煤量)
	// - 其他敏感信息字段

	return fmt.Errorf("importTable2Data 方法待实现")
}

// parseTable2Excel 解析附表2 Excel文件
func (s *App) parseTable2Excel(f *excelize.File) ([]map[string]interface{}, error) {
	// TODO: 实现附表2 Excel解析逻辑
	// 附表2 202X年其他耗煤单位重点耗煤装置（设备）煤炭消耗信息表 -> critical_coal_equipment_consumption
	return nil, fmt.Errorf("parseTable2Excel 方法待实现")
}
