package main
import (
	"fmt"
	"path/filepath"
	"github.com/xuri/excelize/v2"
)


var table3Names = []string{
		"", "project_name",	 "project_code", "construction_unit",	"main_construction_content",	"province_name",	"city_name",	"country_name",	"trade_a",	"trade_c",	"examination_approval_time",	"scheduled_time",	"actual_time",	"examination_authority",	"document_number",	"equivalent_value",	"equivalent_cost",	"pq_total_coal_consumption",	"pq_coal_consumption",	"pq_coke_consumption",	"pq_blue_coke_consumption",	"sce_total_coal_consumption",	"sce_coal_consumption",	"sce_coke_consumption",	"sce_blue_coke_consumption",	"is_substitution",	"substitution_source",	"substitution_quantity",	"pq_annual_coal_quantity",	"sce_annual_coal_quantity",	"state",
	}
var table3FieldMapping = CreateExcelFieldMapping(table3Names)


// ValidateTable3 验证表3数据	
func (a *App) validateTable3(filePath string) ([]ValidateSheetResult, error) {

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

	 err = a.validateTable3Headers(f, sheets)
	 if err != nil {
		return nil, fmt.Errorf("与数据模板不匹配：文件%s:%s", fileName, err.Error())
	}

	mainData, err := a.parseTable3Excel(f, sheets)
	if err != nil {
		return nil, fmt.Errorf("解析文件%s失败：%s", fileName, err.Error())
	}

	errors := a.validateTable3Data(mainData)	

	return []ValidateSheetResult{
		{SheetName: sheets[0], Errors: errors, RowNumber: len(mainData) + 2, ColumnNumber: len(table3Names),},
	}, nil
}

// parseTable3Excel 解析表3数据
func (a *App) parseTable3Excel(f *excelize.File, sheets []string) ([]map[string]interface{}, error) {
	mainData, err := a.parseTableSheet(f, sheets[0], table3Names, 1)
	if err != nil {
		return nil, fmt.Errorf("与数据模板不匹配")
	}
	return mainData, nil
}

// validateTable3Headers 验证表3表头
func (a *App) validateTable3Headers(f *excelize.File, sheets []string) error {
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

	// 第二行 从P2开始
	expectedHeadersRow2 := []string{ 
		"当量值", "等价值", "煤品消费总量", "#煤炭消费量", "#焦炭消费量", "#兰炭消费量", "煤品消费总量", "#煤炭消费量", "#焦炭消费量", "#兰炭消费量", "是否煤炭消费替代", "煤炭消费替代来源", "煤炭消费替代量（万吨，实物量）", "年原料用煤量（万吨，实物量）", "年原料用煤量（万吨标准煤，折标量）",
	}
	err = a.validateSheetHeaders(rows1, sheetName, 1, 15, expectedHeadersRow2)
	if err != nil {
		return err
	}
	return nil
}

// validateTable2Data 验证表3数据
func (a *App) validateTable3Data(mainData []map[string]interface{}) []ValidationError {
	var errors []ValidationError

	for _, row := range mainData {
		rowNumber := getRowNumber(row)
		// 单位所在省/市/区、单位所在地市、单位所在区县不一致。
		a.validateAddress(table3FieldMapping, row, rowNumber, &errors)
		// 所属行业大类、所属行业小类填写不规范，如“C32”、“黑色金属冶炼和压延加工业（31）”、“C2613无机盐制造”等。
		a.validateTrade3(row, rowNumber, &errors)
		// 是否煤炭消费替代填写不规范，如空、“0”等。填写“是”、“否”。
		a.validateIsSubstitution(row, rowNumber, &errors)
	}

	return errors
}

// validateTrade3 验证所属行业大类、所属行业小类填写不规范
func (a *App)  validateTrade3(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	trade_a := getStringValue(row["trade_a"])
	trade_c := getStringValue(row["trade_c"])

	if len(trade_a) != 2 || !isInteger(trade_a) {
		*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_TEXT_FORMAT_ERROR, Message: "应为2位代码", Flag: 0, Cells: GetCellPosition(table3FieldMapping, rowNumber, "trade_a"), RowNumber: rowNumber,})
	}

	if len(trade_c) != 4 || !isInteger(trade_c) {
		*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_TEXT_FORMAT_ERROR, Message: "应为4位代码", Flag: 0, Cells: GetCellPosition(table3FieldMapping, rowNumber, "trade_c"), RowNumber: rowNumber,})
	}
}

// validateIsSubstitution 验证是否煤炭消费替代填写不规范
func (a *App) validateIsSubstitution(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	is_substitution := getStringValue(row["is_substitution"])

	if is_substitution != "是" && is_substitution != "否" {
		*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_TEXT_FORMAT_ERROR, Message: "应为“是”或“否”", Flag: 0, Cells: GetCellPosition(table3FieldMapping, rowNumber, "is_substitution"), RowNumber: rowNumber,})
	}
}
