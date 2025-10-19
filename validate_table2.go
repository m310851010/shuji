package main
import (
	"fmt"
	"path/filepath"
	"github.com/xuri/excelize/v2"
)


var table2Names = []string{
		"", "unit_name","credit_code","stat_date","trade_a","trade_b","trade_c","province_name","city_name","country_name","coal_type","coal_no","usage_time","design_life","enecrgy_efficienct_bmk","capacity","capacity_unit","use_info","status","annual_coal_consumption","state",
	}
var table2FieldMapping = CreateExcelFieldMapping(table2Names)

// table1EquipTypeMap 附表2设备类型映射
var table1EquipTypeMap2 = map[string]bool{
	"锅炉": true,
	"窑炉": true,
	"其他": true,
}

// table1UseInfoMap 附表2用途映射
var table1UseInfoMap = map[string]bool{
	"农林牧渔": true,
	"工业": true,
	"服务业": true,
	"居民生活": true,
	"其他": true,
}

// table2StatusMap 附表2状态映射
var table2StatusMap = map[string]bool{
	"运行": true,
	"停用": true,
}



// ValidateTable2 验证表2数据
func (a *App) validateTable2(filePath string) ([]ValidateSheetResult, error) {

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

	 err = a.validateTable2Headers(f, sheets)
	 if err != nil {
		return nil, fmt.Errorf("与数据模板不匹配：文件%s:%s", fileName, err.Error())
	}

	mainData, err := a.parseTable2Excel(f, sheets)
	if err != nil {
		return nil, fmt.Errorf("解析文件%s失败：%s", fileName, err.Error())
	}

	errors := a.validateTable2Data(mainData)

	return []ValidateSheetResult{
		{SheetName: sheets[0], Errors: errors, RowNumber: len(mainData) + 1, ColumnNumber: len(table2Names),},
	}, nil
}

// parseTable2Excel 解析表2数据
func (a *App) parseTable2Excel(f *excelize.File, sheets []string) ([]map[string]interface{}, error) {
	mainData, err := a.parseTableSheet(f, sheets[0], table2Names, 1)
	if err != nil {
		return nil, fmt.Errorf("与数据模板不匹配")
	}
	return mainData, nil
}

// validateTable2Headers 验证表2表头
func (a *App) validateTable2Headers(f *excelize.File, sheets []string) error {
	sheetName := sheets[0]
	rows1, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("%s没有数据", sheetName)
	}

	// 附表1煤炭消费主要信息
	expectedHeadersRow := []string{
		"序号","单位名称","统一社会信用代码","年份","行业门类","行业大类","行业中类","单位所在省","单位所在地市","单位所在区县","类型","编号","累计使用时间","设计年限","能效水平","容量","容量单位","用途","状态","年耗煤量(吨)","状态",
	}

	err = a.validateSheetHeaders(rows1, sheetName, 0, 0, expectedHeadersRow)
	if err != nil {
		return err
	}
	return nil
}

// validateTable2Data 验证表2数据
func (a *App) validateTable2Data(mainData []map[string]interface{}) []ValidationError {
	var errors []ValidationError

	for _, row := range mainData {
		rowNumber := getRowNumber(row)
		// 年份无法正确识别，部分填写为“2023年”“2024年”，部分填写数字为宽体字符。
		a.validateStatDate(table2FieldMapping, row, rowNumber, &errors)
		// 统一社会信用代码填写错误。
		a.validateCreditCode(table2FieldMapping, row, rowNumber, &errors)
		// 行业门类、行业大类、行业中类不一致。
		a.validateTrade(table2FieldMapping, row, rowNumber, &errors)
		// 单位所在省/市/区、单位所在地市、单位所在区县不一致。
		a.validateAddress(table2FieldMapping, row, rowNumber, &errors)
		// 重点耗煤装置（设备）中累计使用时间及设计年限未按照要求填写整数数字。
		a.validateEnergyDeviceTime(table2FieldMapping, row, rowNumber, "usage_time", &errors)
		// 类型、能效水平、单位容量、用途、状态未按照下拉菜单选项填写。
		a.validateEnergyDevice2(row, rowNumber, &errors)
		// 编号存在多个设备合并到1行中填写的情况。
		a.validateEnergyDeviceNumber(table2FieldMapping, row, rowNumber, "coal_no", &errors)
	}

	return errors
}


// validateEnergyDevice 验证其他耗煤单位重点耗煤装置（设备）中类型、能效水平、容量单位、用途、状态未按照下拉菜单选项填写。
func (a *App) validateEnergyDevice2(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	// 装置（设备）中类型
	equipType := getStringValue(row["coal_type"]) // 装置（设备）中类型
	if !table1EquipTypeMap2[equipType] {
		*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_OPTION_ERROR, Message: "不在选项中", Flag: 0, Cells: GetCellPosition(table2FieldMapping, rowNumber, "coal_type"), RowNumber: rowNumber,})
	}
	// 能效水平
	energyEfficiency := getStringValue(row["enecrgy_efficienct_bmk"]) // 能效水平
	if !table1EnergyEfficiencyMap[energyEfficiency] {
		*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_OPTION_ERROR, Message: "不在选项中", Flag: 0, Cells: GetCellPosition(table2FieldMapping, rowNumber, "enecrgy_efficienct_bmk"), RowNumber: rowNumber,})
	}
	// 容量单位
	capacity_unit := getStringValue(row["capacity_unit"]) // 容量单位
	if !table1CapacityUnitMap[capacity_unit] {
		*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_OPTION_ERROR, Message: "不在选项中", Flag: 0, Cells: GetCellPosition(table2FieldMapping, rowNumber, "capacity_unit"), RowNumber: rowNumber,})	
	}
	// 用途
	useInfo := getStringValue(row["use_info"]) // 用途
	if !table1UseInfoMap[useInfo] {
		*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_OPTION_ERROR, Message: "不在选项中", Flag: 0, Cells: GetCellPosition(table2FieldMapping, rowNumber, "use_info"), RowNumber: rowNumber,})	
	}
	// 状态
	status := getStringValue(row["status"]) // 状态
	if !table2StatusMap[status] {
		*errors = append(*errors, ValidationError{ Type: ERROR_TYPE_OPTION_ERROR, Message: "不在选项中", Flag: 0, Cells: GetCellPosition(table2FieldMapping, rowNumber, "status"), RowNumber: rowNumber,})	
	}
}