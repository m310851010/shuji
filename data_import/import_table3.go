package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ImportTable3 导入附表3数据
func (s *App) ImportTable3(filePath string) QueryResult {
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

	// 4. 清空原表数据
	err = s.clearTable3Data()
	if err != nil {
		s.insertImportRecord(fileName, "附表3", "上传失败", fmt.Sprintf("清空原表数据失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("清空原表数据失败: %v", err),
		}
	}

	// 5. 导入数据到数据库
	err = s.importTable3Data(data)
	if err != nil {
		s.insertImportRecord(fileName, "附表3", "上传失败", fmt.Sprintf("导入数据失败: %v", err))
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
	s.insertImportRecord(fileName, "附表3", "上传成功", fmt.Sprintf("成功导入 %d 条数据", len(data)))

	return QueryResult{
		Ok:      true,
		Data:    len(data),
		Message: "导入成功",
	}
}

// clearTable3Data 清空附表3表数据
func (s *App) clearTable3Data() error {
	return s.clearTableData(TableFixedAssetsInvestmentProject)
}

// importTable3Data 导入附表3数据到数据库
func (s *App) importTable3Data(data []map[string]interface{}) error {
	// TODO: 实现附表3数据导入逻辑
	// 导入到表：fixed_assets_investment_project

	// 注意：敏感数据需要使用SM4加密存储
	// 根据main.sql文件描述，以下字段需要加密：
	// - 各种煤炭消费量字段
	// - 项目名称等敏感信息

	return fmt.Errorf("importTable3Data 方法待实现")
}

// parseTable3Excel 解析附表3 Excel文件
func (s *App) parseTable3Excel(f *excelize.File) ([]map[string]interface{}, error) {
	// TODO: 实现附表3 Excel解析逻辑
	// 附表3 固定资产投资项目节能审查煤炭消费情况汇总表 -> fixed_assets_investment_project
	return nil, fmt.Errorf("parseTable3Excel 方法待实现")
}
