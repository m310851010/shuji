package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// DataImportService 数据导入服务
type DataImportService struct {
	db *db.Database
}

// NewDataImportService 创建数据导入服务实例
func NewDataImportService(db *db.Database) *DataImportService {
	return &DataImportService{
		db: db,
	}
}

// ValidateTable1File 校验附表1文件
func (s *DataImportService) ValidateTable1File(filePath string) QueryResult {
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
		Ok:   true,
		Data: hasData,
		Message: "校验通过",
	}
}

// ValidateTable2File 校验附表2文件
func (s *DataImportService) ValidateTable2File(filePath string) QueryResult {
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
		Ok:   true,
		Data: hasData,
		Message: "校验通过",
	}
}

// ValidateTable3File 校验附表3文件
func (s *DataImportService) ValidateTable3File(filePath string) QueryResult {
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
		},
		}
	}

	// 4. 查询该表是否有数据
	hasData := s.checkTable3HasData()

	// 5. 返回结果
	return QueryResult{
		Ok:   true,
		Data: hasData,
		Message: "校验通过",
	}
}

// ValidateAttachment2File 校验附件2文件
func (s *DataImportService) ValidateAttachment2File(filePath string) QueryResult {
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
		Ok:   true,
		Data: hasData,
		Message: "校验通过",
	}
}

// ImportTable1 导入附表1数据
func (s *DataImportService) ImportTable1(filePath string) QueryResult {
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
		Ok:   true,
		Data: map[string]interface{}{
			"mainCount":  len(mainData),
			"usageCount": len(usageData),
			"equipCount": len(equipData),
		},
		Message: "导入成功",
	}
}

// ImportTable2 导入附表2数据
func (s *DataImportService) ImportTable2(filePath string) QueryResult {
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
		Ok:   true,
		Data: len(data),
		Message: "导入成功",
	}
}

// ImportTable3 导入附表3数据
func (s *DataImportService) ImportTable3(filePath string) QueryResult {
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
		},
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
		Ok:   true,
		Data: len(data),
		Message: "导入成功",
	}
}

// ImportAttachment2 导入附件2数据
func (s *DataImportService) ImportAttachment2(filePath string) QueryResult {
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

	// 4. 清空原表数据
	err = s.clearAttachment2Data()
	if err != nil {
		s.insertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("清空原表数据失败: %v", err))
		return QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("清空原表数据失败: %v", err),
		}
	}

	// 5. 导入数据到数据库
	err = s.importAttachment2Data(data)
	if err != nil {
		s.insertImportRecord(fileName, "附件2", "上传失败", fmt.Sprintf("导入数据失败: %v", err))
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
	s.insertImportRecord(fileName, "附件2", "上传成功", fmt.Sprintf("成功导入 %d 条数据", len(data)))

	return QueryResult{
		Ok:   true,
		Data: len(data),
		Message: "导入成功",
	}
}

// insertImportRecord 插入导入记录
func (s *DataImportService) insertImportRecord(fileName, fileType, importState, describe string) {
	record := &DataImportRecord{
		FileName:    fileName,
		FileType:    fileType,
		ImportTime:  time.Now(),
		ImportState: importState,
		Describe:    describe,
		CreateUser:  GetCurrentOSUser(),
	}

	recordService := NewDataImportRecordService(s.db)
	err := recordService.InsertImportRecord(record)
	if err != nil {
		fmt.Printf("插入导入记录失败: %v", err)
	}
}

// 以下是具体的解析、校验和导入方法的实现
// 由于篇幅限制，这里只展示框架，具体实现需要根据Excel文件格式和校验规则来完成

func (s *DataImportService) parseTable1Excel(f *excelize.File) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error) {
	// TODO: 实现附表1 Excel解析逻辑
	// 需要解析三个表格区域：
	// 1. 综合能源消费情况和规模以上企业煤炭消费信息表 -> enterprise_coal_consumption_main
	// 2. 煤炭消费主要用途情况 -> enterprise_coal_consumption_usage  
	// 3. 重点耗煤装置情况 -> enterprise_coal_consumption_equip
	return nil, nil, nil, fmt.Errorf("parseTable1Excel 方法待实现")
}

func (s *DataImportService) parseTable2Excel(f *excelize.File) ([]map[string]interface{}, error) {
	// TODO: 实现附表2 Excel解析逻辑
	// 附表2 202X年其他耗煤单位重点耗煤装置（设备）煤炭消耗信息表 -> critical_coal_equipment_consumption
	return nil, fmt.Errorf("parseTable2Excel 方法待实现")
}

func (s *DataImportService) parseTable3Excel(f *excelize.File) ([]map[string]interface{}, error) {
	// TODO: 实现附表3 Excel解析逻辑
	// 附表3 固定资产投资项目节能审查煤炭消费情况汇总表 -> fixed_assets_investment_project
	return nil, fmt.Errorf("parseTable3Excel 方法待实现")
}

func (s *DataImportService) parseAttachment2Excel(f *excelize.File) ([]map[string]interface{}, error) {
	// TODO: 实现附件2 Excel解析逻辑
	// 附件2 XX省（自治区、直辖市）202X年煤炭消费状况表 -> coal_consumption_report
	return nil, fmt.Errorf("parseAttachment2Excel 方法待实现")
}

func (s *DataImportService) validateTable1Data(mainData, usageData, equipData []map[string]interface{}) []string {
	// TODO: 实现附表1数据校验逻辑
	// 根据校验提示词_V5_0821.mhtml中的强制校验规则进行校验
	errors := []string{}
	
	// 1. 检查年份和单位是否为空
	// 2. 检查企业是否在企业清单中（如果有清单的话）
	// 3. 检查企业名称和统一信用代码是否对应（如果有清单的话）
	// 4. 检查数据单位与当前单位是否相符
	// 5. 检查文件格式是否正确
	
	return errors
}

func (s *DataImportService) validateTable2Data(data []map[string]interface{}) []string {
	// TODO: 实现附表2数据校验逻辑
	errors := []string{}
	return errors
}

func (s *DataImportService) validateTable3Data(data []map[string]interface{}) []string {
	// TODO: 实现附表3数据校验逻辑
	errors := []string{}
	return errors
}

func (s *DataImportService) validateAttachment2Data(data []map[string]interface{}) []string {
	// TODO: 实现附件2数据校验逻辑
	errors := []string{}
	return errors
}

func (s *DataImportService) checkTable1HasData() bool {
	// 检查附表1相关表是否有数据
	query := "SELECT COUNT(*) as count FROM enterprise_coal_consumption_main"
	result, err := s.db.Query(query)
	if err != nil {
		return false
	}
	
	if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
		if count, ok := data[0]["count"].(int64); ok {
			return count > 0
		}
	}
	return false
}

func (s *DataImportService) checkTable2HasData() bool {
	query := "SELECT COUNT(*) as count FROM critical_coal_equipment_consumption"
	result, err := s.db.Query(query)
	if err != nil {
		return false
	}
	
	if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
		if count, ok := data[0]["count"].(int64); ok {
			return count > 0
		}
	}
	return false
}

func (s *DataImportService) checkTable3HasData() bool {
	query := "SELECT COUNT(*) as count FROM fixed_assets_investment_project"
	result, err := s.db.Query(query)
	if err != nil {
		return false
	}
	
	if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
		if count, ok := data[0]["count"].(int64); ok {
			return count > 0
		}
	}
	return false
}

func (s *DataImportService) checkAttachment2HasData() bool {
	query := "SELECT COUNT(*) as count FROM coal_consumption_report"
	result, err := s.db.Query(query)
	if err != nil {
		return false
	}
	
	if data, ok := result.Data.([]map[string]interface{}); ok && len(data) > 0 {
		if count, ok := data[0]["count"].(int64); ok {
			return count > 0
		}
	}
	return false
}

func (s *DataImportService) clearTable1Data() error {
	// 清空附表1相关表数据
	queries := []string{
		"DELETE FROM enterprise_coal_consumption_main",
		"DELETE FROM enterprise_coal_consumption_usage", 
		"DELETE FROM enterprise_coal_consumption_equip",
	}
	
	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return fmt.Errorf("清空表失败 %s: %v", query, err)
		}
	}
	return nil
}

func (s *DataImportService) clearTable2Data() error {
	query := "DELETE FROM critical_coal_equipment_consumption"
	_, err := s.db.Exec(query)
	return err
}

func (s *DataImportService) clearTable3Data() error {
	query := "DELETE FROM fixed_assets_investment_project"
	_, err := s.db.Exec(query)
	return err
}

func (s *DataImportService) clearAttachment2Data() error {
	query := "DELETE FROM coal_consumption_report"
	_, err := s.db.Exec(query)
	return err
}

func (s *DataImportService) importTable1Data(mainData, usageData, equipData []map[string]interface{}) error {
	// TODO: 实现附表1数据导入逻辑
	// 需要导入到三个表：
	// 1. enterprise_coal_consumption_main
	// 2. enterprise_coal_consumption_usage
	// 3. enterprise_coal_consumption_equip
	return fmt.Errorf("importTable1Data 方法待实现")
}

func (s *DataImportService) importTable2Data(data []map[string]interface{}) error {
	// TODO: 实现附表2数据导入逻辑
	return fmt.Errorf("importTable2Data 方法待实现")
}

func (s *DataImportService) importTable3Data(data []map[string]interface{}) error {
	// TODO: 实现附表3数据导入逻辑
	return fmt.Errorf("importTable3Data 方法待实现")
}

func (s *DataImportService) importAttachment2Data(data []map[string]interface{}) error {
	// TODO: 实现附件2数据导入逻辑
	return fmt.Errorf("importAttachment2Data 方法待实现")
}
