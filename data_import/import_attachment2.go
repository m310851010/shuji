package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ImportAttachment2 导入附件2数据
func (s *App) ImportAttachment2(filePath string) QueryResult {
	fileName := filepath.Base(filePath)

	// 1. 读取Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		s.insertImportRecord(fileName, FileTypeAttachment2, ImportStateFailed, fmt.Sprintf("读取Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("读取Excel文件失败: %v", err),
		}
	}
	defer f.Close()

	// 2. 解析Excel文件
	data, err := s.parseAttachment2Excel(f)
	if err != nil {
		s.insertImportRecord(fileName, FileTypeAttachment2, ImportStateFailed, fmt.Sprintf("解析Excel文件失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("解析Excel文件失败: %v", err),
		}
	}

	// 3. 校验数据
	validationErrors := s.validateAttachment2Data(data)
	if len(validationErrors) > 0 {
		s.insertImportRecord(fileName, FileTypeAttachment2, ImportStateFailed, fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("数据校验失败: %s", strings.Join(validationErrors, "; ")),
		}
	}

	// 4. 清空原表数据
	err = s.clearAttachment2Data()
	if err != nil {
		s.insertImportRecord(fileName, FileTypeAttachment2, ImportStateFailed, fmt.Sprintf("清空原表数据失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("清空原表数据失败: %v", err),
		}
	}

	// 5. 导入数据到数据库
	err = s.importAttachment2Data(data)
	if err != nil {
		s.insertImportRecord(fileName, FileTypeAttachment2, ImportStateFailed, fmt.Sprintf("导入数据失败: %v", err))
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
	s.insertImportRecord(fileName, FileTypeAttachment2, ImportStateSuccess, fmt.Sprintf("成功导入 %d 条数据", len(data)))

	return QueryResult{
		Ok:      true,
		Data:    len(data),
		Message: "导入成功",
	}
}

// clearAttachment2Data 清空附件2表数据
func (s *App) clearAttachment2Data() error {
	return s.clearTableData(TableCoalConsumptionReport)
}

// importAttachment2Data 导入附件2数据到数据库
func (s *App) importAttachment2Data(data []map[string]interface{}) error {
	// TODO: 实现附件2数据导入逻辑
	// 导入到表：coal_consumption_report

	// 注意：敏感数据需要使用SM4加密存储
	// 根据main.sql文件描述，以下字段需要加密：
	// - total_coal (煤炭消费总量)
	// - raw_coal (原煤)
	// - washed_coal (洗精煤)
	// - other_coal (其他煤炭)
	// - power_generation (火力发电)
	// - heating (供热)
	// - coal_washing (煤炭洗选)
	// - coking (炼焦)
	// - oil_refining (炼油及煤制油)
	// - gas_production (制气)
	// - industry (工业)
	// - raw_materials (用作原材料)
	// - other_uses (其他用途)
	// - coke (焦炭)
	// - is_confirm (是否已确认)
	// - is_check (是否已校核)

	return fmt.Errorf("importAttachment2Data 方法待实现")
}

// parseAttachment2Excel 解析附件2 Excel文件
func (s *App) parseAttachment2Excel(f *excelize.File) ([]map[string]interface{}, error) {
	// TODO: 实现附件2 Excel解析逻辑
	// 附件2 XX省（自治区、直辖市）202X年煤炭消费状况表 -> coal_consumption_report
	return nil, fmt.Errorf("parseAttachment2Excel 方法待实现")
}
