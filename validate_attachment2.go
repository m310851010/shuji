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
func (a *App) validateAttachment2(filePath string) QueryResult {

	fileName := filepath.Base(filePath)

	// 第二步: 文件是否可读取
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		errorMessage := fmt.Sprintf("读取文件%s失败: %v", fileName, err)
		fmt.Println(errorMessage)
		
		return QueryResult{
			Ok:      false,
			Data:    nil,
			Message: errorMessage,
		}
	}
	defer f.Close()

	// 获取所有工作表
	sheets := f.GetSheetList()
	if len(sheets) < 1 {
		errorMessage := fmt.Sprintf("与数据模板不匹配：%s文件中没有足够的工作表", fileName)
		fmt.Println(errorMessage)
		return QueryResult{
			Ok:      false,
			Data:    nil,
			Message: errorMessage,
		}
	}

	 err = a.validateAttachment2Headers(f, sheets)
	 if err != nil {
		errorMessage := fmt.Sprintf("与数据模板不匹配：%s", err.Error())
		fmt.Println(errorMessage)
		return QueryResult{
			Ok:      false,
			Data:    nil,
			Message: errorMessage,
		}
	}

	mainData, err := a.parseTable3Excel(f, sheets)
	if err != nil {
		errorMessage := fmt.Sprintf("解析%s文件失败：%s", fileName, err.Error())
		fmt.Println(errorMessage)
		return QueryResult{
			Ok:      false,
			Data:    nil,
			Message: errorMessage,
		}
	}

	errors := a.validateAttachment2Data(mainData)	

	if len(errors) > 0 {
		return QueryResult{
			Ok:      false,
			Data:    false,
		}
	}
	
	return QueryResult{
		Ok:      true,
		Data:    true,
		Message: "验证通过",
	}
}

// parseAttachment2Excel 解析附件2数据
func (a *App) parseAttachment2Excel(f *excelize.File, sheets []string) ([]map[string]interface{}, error) {
	mainData, err := a.parseTableSheet(f, sheets[0], table3Names, 1)
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
		"序号", "项目名称", "项目代码", "建设单位", "主要建设内容", "项目所在省、自治区、直辖市", "项目所在地市", "项目所在区县", "所属行业大类（2位代码）", "所属行业小类", "节能审查批复时间", "拟投产时间", "实际投产时间", "节能审查机关", "审查意见文号", "年综合能源消费量（万吨标准煤，含原料用能和可再生能源）",
		 "",  "年煤品消费量（万吨，实物量）", "", "", "", "年煤品消费量（万吨标准煤，折标量）", "", "","", "煤炭消费替代情况", "", "","原料用煤情况", "", "状态",
	}
	err = a.validateSheetHeaders(rows1, sheetName, 0, 0, expectedHeadersRow1)
	if err != nil {
		return err
	}
	
	// 第二行 从PL2开始
	expectedHeadersRow2 := []string{ 
		"当量值", "等价值", "煤品消费总量", "煤炭消费量", "焦炭消费量", "兰炭消费量", "煤品消费总量", "煤炭消费量", "焦炭消费量", "兰炭消费量", "是否煤炭消费替代", "煤炭消费替代来源", "煤炭消费替代量（万吨，实物量）", "年原料用煤量（万吨，实物量）", "年原料用煤量（万吨标准煤，折标量）", "",
	}
	err = a.validateSheetHeaders(rows1, sheetName, 1, 14, expectedHeadersRow2)
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
			*errors = append(*errors, ValidationError{ Type: "required", Message: "不能为空”", Flag: 0, Cells: GetCellPosition(attachment2FieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
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
			*errors = append(*errors, ValidationError{ Type: "required", Message: "不能为空,如不存在，请填写“0”", Flag: 0, Cells: GetCellPosition(attachment2FieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		} else if !isNumber(value) {
			*errors = append(*errors, ValidationError{ Type: "required", Message: "应为数字", Flag: 0, Cells: GetCellPosition(attachment2FieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		} else {
			a.validateEnergyUnit(attachment2FieldMapping, row, rowNumber, 0, []string{fieldName}, errors)
		}
	}
	
}


