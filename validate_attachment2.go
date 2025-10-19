package main
import (
	"fmt"
	"path/filepath"
	"github.com/xuri/excelize/v2"
)


var attachment2Names = []string{
		"", "stat_date", "province_name", "city_name", "country_name", "total_coal", "raw_coal", "washed_coal", "other_coal", "power_generation", "heating", "coal_washing", "coking", "oil_refining", "gas_production", "industry", "raw_materials", "other_uses", "coke", "state",
	}
var attachment2FieldMapping = CreateExcelFieldMapping(attachment2Names)


// ValidateAttachment2 验证附件2数据	
func (a *App) validateAttachment2(filePath string) ([]ValidateSheetResult, error) {

	fileName := filepath.Base(filePath)

	// 第二步: 文件是否可读取
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件%s失败: %v", fileName, err)
	}
	defer f.Close()

	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) < 1 {
		return nil, fmt.Errorf("与数据模板不匹配：文件%s没有足够的工作表", fileName)
	}

	 err = a.validateAttachment2Headers(f, sheets)
	 if err != nil {
		return nil, fmt.Errorf("与数据模板不匹配：文件%s:%s", fileName, err.Error())
	}

	mainData, err := a.parseAttachment2Excel(f, sheets)
	if err != nil {
		return nil, fmt.Errorf("解析文件%s失败：%s", fileName, err.Error())
	}

	errors := a.validateAttachment2Data(mainData)	

	return []ValidateSheetResult{
		{SheetName: sheets[0], Errors: errors, RowNumber: len(mainData) + 3, ColumnNumber: len(attachment2Names),},
	}, nil
}

// parseAttachment2Excel 解析附件2数据
func (a *App) parseAttachment2Excel(f *excelize.File, sheets []string) ([]map[string]interface{}, error) {
	mainData, err := a.parseTableSheet(f, sheets[0], attachment2Names, 3)
	if err != nil {
		return nil, fmt.Errorf("与数据模板不匹配")
	}
	return mainData, nil
}

// validateAttachment2Headers 验证附件2表头
func (a *App) validateAttachment2Headers(f *excelize.File, sheets []string) error {
	sheetName := sheets[0]
	rows1, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("%s没有数据", sheetName)
	}

	// 附表3煤炭消费主要信息

	expectedHeadersRow1 := []string{
		"序号", "年份", "单位所在省", "单位所在地市", "单位所在区县", "分品种煤炭消费摸底", "", "","", "分用途煤炭消费摸底", "","","","","","","","","焦炭消费摸底", "状态",
	}
	err = a.validateSheetHeaders(rows1, sheetName, 0, 0, expectedHeadersRow1)
	if err != nil {
		return err
	}
	
	// 第3行 从J3开始
	expectedHeadersRow2 := []string{ 
		"煤合计", "原煤", "洗精煤", "其他", "1.火力发电", "2.供热", "3.煤炭洗选", "4.炼焦", "5.炼油及煤制油", "6.制气", "1.工业", "#用作原料、材料", "2.其他用途", "焦炭", "状态",
	}
	err = a.validateSheetHeaders(rows1, sheetName, 2, 5, expectedHeadersRow2)
	if err != nil {
		return err
	}
	return nil
}

// validateAttachment2Data 验证附件2数据
func (a *App) validateAttachment2Data(mainData []map[string]interface{}) []ValidationError {
	var errors []ValidationError

	for _, row := range mainData {
		rowNumber := getRowNumber(row)
		// 所有表格不能为空，若无相关煤炭消费，则填0。
		a.validateAttachment2EmptyData(row, rowNumber, &errors)
		// 表格内数字单位错误。
		a.validateAttachment2UnitData(row, rowNumber, &errors)
	}

	return errors
}

// validateAttachment2EmptyData 验证所有表格不能为空，若无相关煤炭消费，则填0。
func (a *App) validateAttachment2EmptyData(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	fieldNames := []string{
		"stat_date", "province_name", "city_name", "country_name",
	}
	for _, fieldName := range fieldNames {
		value := getStringValue(row[fieldName])
		if isEmpty(value) {
			*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_TEXT_FORMAT_ERROR, Message: "不能为空”", Flag: 0, Cells: GetCellPosition(attachment2FieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		}
	}
}


// validateAttachment2UnitData 验证表格内数字单位错误。
func (a *App) validateAttachment2UnitData(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	fieldNames := []string{
		"total_coal", "raw_coal", "washed_coal", "other_coal", "power_generation", "heating", "coal_washing", "coking", "oil_refining", "gas_production", "industry", "raw_materials", "other_uses", "coke",
	}
	for _, fieldName := range fieldNames {
		value := getStringValue(row[fieldName])
		if isEmpty(value) {
			*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_TEXT_FORMAT_ERROR, Message: "不能为空,如不存在，请填写“0”", Flag: 0, Cells: GetCellPosition(attachment2FieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		} else if !isNumber(value) {
			*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_TEXT_FORMAT_ERROR, Message: "应为数字", Flag: 0, Cells: GetCellPosition(attachment2FieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		} else {
			a.validateEnergyUnit(attachment2FieldMapping, row, rowNumber, 0, []string{fieldName}, errors)
		}
	}
	
}


