package main
import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/xuri/excelize/v2"
)

// 解析主表数据（企业基本信息）
var mainDataNames = []string{
		"", "stat_date","unit_name","credit_code",	"trade_a",	"trade_b",	"trade_c",	"province_name","city_name","country_name",	"tel", "annual_energy_equivalent_value", "annual_energy_equivalent_cost",
		"annual_raw_material_energy",	"annual_total_coal_consumption",	"annual_total_coal_products",	"annual_raw_coal",	"annual_raw_coal_consumption",	"annual_clean_coal_consumption",	"annual_other_coal_consumption",	"annual_coke_consumption",
	}
var mainFieldMapping = CreateExcelFieldMapping(mainDataNames)
// 解析用途数据（煤炭消费主要用途情况）
var usageDataNames = []string{
		"", "stat_date","unit_name","credit_code","main_usage","specific_usage","input_variety","input_unit","input_quantity","output_energy_types","measurement_unit","output_quantity","remarks",
	}
var usageFieldMapping = CreateExcelFieldMapping(usageDataNames)
// 解析设备数据（重点耗煤装置情况）
var equipDataNames = []string{
		"", "stat_date","unit_name","credit_code","equip_type","equip_no","total_runtime","design_life","energy_efficiency","capacity_unit","capacity","coal_type","annual_coal_consumption",
	}
var equipFieldMapping = CreateExcelFieldMapping(equipDataNames)

// 主用途映射 -> 具体用途 -> 产出品种品类 ->单位
var mainUsageMapping = map[string]map[string]map[string]map[string]bool{
	"加工转换": {
		"发电": {"电力": {"万千瓦时": true},},
		"供热": {"热力": {"百万千焦": true},},
		"洗选": {"洗精煤": {"万吨": true}, "其他洗煤": {"万吨": true},},
		"焦化": {"焦炭": {"万吨": true},},
		"煤制油": {"煤制油": {"万吨": true},},
		"煤制气": {"煤制气": {"万立方米": true},},
	},
	"原料": {
		"原料": {"煤制合成氨": {"万吨": true}, "煤制甲醇": {"万吨": true}, "其他": {"万千瓦时": true, "百万千焦": true, "万吨": true, "万立方米": true, "其他": true,},},
	},
	"燃料": {
		"燃烧": {"其他": {"万千瓦时": true, "百万千焦": true, "万吨": true, "万立方米": true, "其他": true,},},
	},
}

// table1EquipTypeMap 附表1设备类型映射
var table1EquipTypeMap = map[string]bool{
	"锅炉": true,
	"窑炉": true,
	"气化炉": true,
	"炼铁高炉": true,
	"焦化炉": true,
	"矿热炉": true,
	"其他": true,
}

// table1EnergyEfficiencyMap 附表1能效水平映射
var table1EnergyEfficiencyMap = map[string]bool{
	"优于先进水平": true,
	"先进水平至节能水平之间": true,
	"节能水平至准入水平之间": true,
	"无能效标准": true,
}

// table1CapacityUnitMap 附表1容量单位映射
var table1CapacityUnitMap = map[string]bool{
	"蒸吨/小时": true,
	"立方米/小时": true,
	"吨/小时": true,
	"千伏安": true,
	"立方米": true,
	"兆瓦": true,
	"其他": true,
}

// table1CoalProductMap 附表1耗煤品种映射
var table1CoalProductMap = map[string]bool{
	"原煤": true,
	"洗精煤": true,
	"其他煤炭": true,
	"焦炭": true,
	"其他": true,
}

// creditCodeStatDateMapping 企业信用代码+统计日期映射
var creditCodeStatDateMapping = make(map[string]CreditCacheInfo)

// getCreditCodeStatDateKey 生成企业信用代码+统计日期的唯一键
func getCreditCodeStatDateKey(row map[string]interface{}) string {
	return getStringValue(row["credit_code"]) + getStringValue(row["stat_date"])
}

// ValidateTable1 验证表1数据
func (a *App) validateTable1(filePath string) QueryResult {
	creditCodeStatDateMapping = make(map[string]CreditCacheInfo)

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
	if len(sheets) < 3 {
		errorMessage := fmt.Sprintf("与数据模板不匹配：%s文件中没有足够的工作表", fileName)
		fmt.Println(errorMessage)
		return QueryResult{
			Ok:      false,
			Data:    nil,
			Message: errorMessage,
		}
	}

	 err = a.validateTable1Headers(f, sheets)
	 if err != nil {
		errorMessage := fmt.Sprintf("与数据模板不匹配：%s", err.Error())
		fmt.Println(errorMessage)
		return QueryResult{
			Ok:      false,
			Data:    nil,
			Message: errorMessage,
		}
	}

	mainData, equipData, usageData, err := a.parseTable1Excel(f, sheets)
	if err != nil {
		errorMessage := fmt.Sprintf("解析%s文件失败：%s", fileName, err.Error())
		fmt.Println(errorMessage)
		return QueryResult{
			Ok:      false,
			Data:    nil,
			Message: errorMessage,
		}
	}

	errors := a.validateTable1MainData(mainData)
	usageErrors := a.validateTable1UsageData(usageData)
	equipErrors := a.validateTable1EquipData(equipData)

	errors = append(errors, usageErrors...)
	errors = append(errors, equipErrors...)

	if len(errors) > 0 {
		errorMessage := fmt.Sprintf("表1数据验证失败：%s", errors[0].Message)
		fmt.Println(errorMessage)
		return QueryResult{
			Ok:      false,
			Data:    false,
			Message: errorMessage,
		}
	}
	
	return QueryResult{
		Ok:      true,
		Data:    true,
		Message: "验证通过",
	}
}

// ValidateSheetHeaders 验证工作表表头
func (a *App) validateSheetHeaders(rows [][]string, sheetName string, row int, startCol int, expectedHeaders []string) error {
	if len(rows) < row + 1 {
		return fmt.Errorf("%s表格行数不足", sheetName)
	}

	lenHeaders := len(expectedHeaders)
	headerRow := rows[row]	
	if len(headerRow) < len(expectedHeaders) {
		return fmt.Errorf("%s第%d行表头列数与模板不符", sheetName, row + 1)
	}

	for i := startCol; i < lenHeaders; i++ {
		if headerRow[i] != expectedHeaders[i] {
			return fmt.Errorf("%s第%d行表头列数与模板不符", sheetName, row + 1)
		}
	}
	return nil
}

// ValidateTable1Headers 验证表1工作表表头
func (a *App) validateTable1Headers(f *excelize.File, sheets []string) error {
	sheetName := sheets[0]
	rows1, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("%s没有数据", sheetName)
	}

	// 附表1煤炭消费主要信息
	expectedHeadersRow1 := []string{
		"序号",	"年份",	"单位名称",	"统一社会信用代码",	"行业门类",	"行业大类",	"行业中类",	"单位所在省",	"单位所在地市",	"单位所在区县",	"联系电话", "综合能源消费情况", "", "", "煤炭消费情况","","","","","","", "状态",
	}
	err = a.validateSheetHeaders(rows1, sheetName, 0, 0, expectedHeadersRow1)
	if err != nil {
		return err
	}
	
	// 第二行 从L2开始
	expectedHeadersRow2 := []string{ 
		"年综合能耗当量值(万吨标准煤，含原料用能)",	"年综合能耗等价值(万吨标准煤，含原料用能)",	"年原料用能消费量(万吨标准煤)",	"耗煤总量(实物量，万吨)",	"耗煤总量(标准量，万吨标准煤)",	"原料用煤(实物量，万吨)",	"原煤消费(实物量，万吨)",	"洗精煤消费(实物量，万吨)",	"其他煤炭消费(实物量，万吨)",	"焦炭消费(实物量，万吨)",
	}
	err = a.validateSheetHeaders(rows1, sheetName, 1, 11, expectedHeadersRow2)
	if err != nil {
		return err
	}


	// 附表1主要用途情况
	sheetName = sheets[1]
	rows2, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("%s没有数据", sheetName)
	}
	expectedHeadersUsage1 := []string{
		"序号",	"年份",	"单位名称",	"统一社会信用代码",	"煤炭消费主要用途情况",	"","","","","","","","",
	}
	
	err = a.validateSheetHeaders(rows2, sheetName, 0, 0, expectedHeadersUsage1)
	if err != nil {
		return err
	}

	expectedHeadersUsage2 := []string{
		"主要用途",	"具体用途",	"投入品种",	"投入计量单位",	"投入量",	"产出品种品类",	"产出计量单位",	"产出量",	"备注",
	}
	err = a.validateSheetHeaders(rows2, sheetName, 1, 4, expectedHeadersUsage2)
	if err != nil {
		return err
	}

	// 附表1重点耗煤装置（设备）情况
	sheetName = sheets[2]
	rows3, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("%s没有数据", sheetName)
	}
	expectedHeadersEquip1 := []string{
		"序号",	"年份",	"单位名称",	"统一社会信用代码",	"重点耗煤装置（设备）情况",	"","","","","","","","",
	}
	
	err = a.validateSheetHeaders(rows3, sheetName, 0, 0, expectedHeadersEquip1)
	if err != nil {
		return err
	}

	expectedHeadersEquip2 := []string{
		"类型",	"编号",	"累计使用时间",	"设计年限",	"能效水平",	"容量单位",	"容量",	"耗煤品种",	"年耗煤量（单位：吨）",
	}
	err = a.validateSheetHeaders(rows3, sheetName, 1, 4, expectedHeadersEquip2)
	if err != nil {
		return err
	}
	return nil
}

// parseTableSheet 通用表格解析函数
func (a *App) parseTableSheet(f *excelize.File, sheetName string, headerNames []string, startRow int) ([]map[string]interface{}, error) {
    rows, err := f.GetRows(sheetName)
    if err != nil {
        return nil, fmt.Errorf("获取工作表%s数据失败: %v", sheetName, err)
    }

    var result []map[string]interface{}
	rowCount := len(rows)
	
    // 从startRow开始遍历数据行
    for i := startRow; i < rowCount; i++ {
        row := rows[i]
        
        // 检查第一列是否为空，如果为空则结束循环
        if len(row) < 2 {
            break
        }

       // 构建数据行
		dataRow := make(map[string]interface{})
		// 记录实际Excel行号
        dataRow["_excel_row"] = i + 1 // 0索引转换为1索引

		headerCount := len(row)
		for j, fileName := range headerNames {
			if fileName == ""  {
				continue
			}
			if j < headerCount {
				dataRow[fileName] = row[j]
			} else {
				dataRow[fileName] = ""
			}
        }
        
        // 将当前行数据添加到结果中
        result = append(result, dataRow)
    }

    return result, nil
}

// parseTable1Excel 解析附表1Excel文件
func (a *App) parseTable1Excel(f *excelize.File, sheets []string) ([]map[string]interface{}, []map[string]interface{}, []map[string]interface{}, error){
	
	// 解析主表数据（企业基本信息）
	mainData, err := a.parseTableSheet(f, sheets[0], mainDataNames, 2)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("与数据模板不匹配")
	}

	// 解析用途数据（煤炭消费主要用途情况）
	usageData, err :=  a.parseTableSheet(f, sheets[1], usageDataNames, 2)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("与数据模板不匹配")
	}

	// 解析设备数据（重点耗煤装置情况）
	equipData, err := a.parseTableSheet(f, sheets[2], equipDataNames, 2)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("与数据模板不匹配")
	}
	return mainData, usageData, equipData, nil
}

// validateEnergyDevice 验证重点耗煤装置（设备）中类型、能效水平、容量单位、耗煤品种未按照下拉菜单选项填写。
func (a *App) validateEnergyDevice(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	// 装置（设备）中类型
	equipType := getStringValue(row["equip_type"]) // 装置（设备）中类型
	if !table1EquipTypeMap[equipType] {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不在选项中", Flag: 0, Cells: GetCellPosition(equipFieldMapping, rowNumber, "equip_type",), RowNumber: rowNumber,})
	}
	// 能效水平
	energyEfficiency := getStringValue(row["energy_efficiency"]) // 能效水平
	if !table1EnergyEfficiencyMap[energyEfficiency] {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不在选项中", Flag: 0, Cells: GetCellPosition(equipFieldMapping, rowNumber, "energy_efficiency",), RowNumber: rowNumber,})
	}
	// 容量单位
	capacity_unit := getStringValue(row["capacity_unit"]) // 容量单位
	if !table1CapacityUnitMap[capacity_unit] {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不在选项中", Flag: 0, Cells: GetCellPosition(equipFieldMapping, rowNumber, "capacity_unit",), RowNumber: rowNumber,})	
		}
	// 耗煤品种
	coalProduct := getStringValue(row["coal_type"]) // 耗煤品种
	if !table1CoalProductMap[coalProduct] {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不在选项中", Flag: 0, Cells: GetCellPosition(equipFieldMapping, rowNumber, "coal_type",), RowNumber: rowNumber,})	
	}
}

// validateEnergyDeviceConflict 验证重点耗煤装置（设备）中编号存在多个设备合并到1行中填写的情况。
func (a *App) validateEnergyDeviceNumber(fieldMapping map[string]ExcelFieldMapping, row map[string]interface{}, rowNumber int, equipName string, errors *[]ValidationError) {
	equip_no := getStringValue(row[equipName]) // 装置（设备）中编号
	// 检查是否存在多个设备合并到1行中填写的情况
	if len(strings.Split(equip_no, "#")) > 2 || strings.Contains(equip_no, "/") || strings.Contains(equip_no, "、") {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "应为1个设备编号", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, equipName), RowNumber: rowNumber,})
	}
}

// validateEnergyDeviceTime 验证重点耗煤装置（设备）中累计使用时间及设计年限未按照要求填写整数数字。
func (a *App) validateEnergyDeviceTime(fieldMapping map[string]ExcelFieldMapping, row map[string]interface{}, rowNumber int, total_runtime string, errors *[]ValidationError) {

	fieldNames := []string{total_runtime, "design_life", }

	for _, fieldName := range fieldNames {
		value := getStringValue(row[fieldName])
		if isEmpty(value) {
			*errors = append(*errors, ValidationError{ Type: "required", Message: "不能为空", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		} else if !isInteger(value) {
			*errors = append(*errors, ValidationError{ Type: "required", Message: "应为正整数", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		}
	}
}

// validateTable1EquipData 验证附表1重点耗煤装置（设备）情况
func (a *App) validateTable1EquipData(equipData []map[string]interface{}) []ValidationError {
	var errors []ValidationError

	for _, row := range equipData {
		rowNumber := getRowNumber(row)
		// 重点耗煤装置（设备）中类型、能效水平、容量单位、耗煤品种未按照下拉菜单选项填写。
		a.validateEnergyDevice(row, rowNumber, &errors)
		// 重点耗煤装置（设备）中编号存在多个设备合并到1行中填写的情况。
		a.validateEnergyDeviceNumber(equipFieldMapping, row, rowNumber, "equip_no", &errors)
		// 重点耗煤装置（设备）中累计使用时间及设计年限未按照要求填写整数数字。
		a.validateEnergyDeviceTime(equipFieldMapping, row, rowNumber, "total_runtime", &errors)	
		// 一个企业填写了多张表格，即既填写了规模以上企业煤炭消费信息表，也填写了其他耗煤单位重点耗煤装置（设备）煤炭消耗信息表。
		// a.validateEnergyDeviceConflict(row, rowNumber, &errors)
	}
	
	return errors
}

// validateTable1UsageData 验证附表1主要用途情况
func (a *App) validateTable1UsageData(usageData []map[string]interface{}) []ValidationError {
	var errors []ValidationError

	for _, row := range usageData {
		rowNumber := getRowNumber(row)
		// 煤炭消费用主要用途填写存在空缺项。
		a.validateEnergyPurpose(row, rowNumber, &errors)
		// 煤炭消费“主要用途”“具体用途”“产出品种品类”不一致。“投入计量单位”是否为“万吨”，“产出品种品类”与“产出计量单位”是否一致。
		a.validateEnergyPurposeDetail(row, rowNumber, &errors)
		// 煤炭消费主要用途混淆。
		// a.validateEnergyPurposeConfusion(row, rowNumber, &errors)	
	}
	
	return errors
}

// validateEnergyPurpose 验证煤炭消费用主要用途填写存在空缺项。
func (a *App) validateEnergyPurpose(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	fieldNames := []string{
		"main_usage",	"specific_usage",	"input_variety",	"input_unit",		"output_energy_types",	"measurement_unit",	"output_quantity",	"input_quantity",
	}

	for _, fieldName := range fieldNames {
		value := getStringValue(row[fieldName])
		if isEmpty(value) {
			*errors = append(*errors, ValidationError{ Type: "required", Message: "不能为空", Flag: 0, Cells: GetCellPosition(mainFieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		} else if (fieldName == "input_quantity" || fieldName == "output_quantity") && !isNumber(value) {
			*errors = append(*errors, ValidationError{ Type: "required", Message: "应为数字", Flag: 0, Cells: GetCellPosition(mainFieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		}
	}
}

// validateEnergyPurposeDetail 验证煤炭消费“主要用途”“具体用途”“产出品种品类”不一致。
// “投入计量单位”是否为“万吨”，“产出品种品类”与“产出计量单位”是否一致。
// 当原料用煤不为零，但在煤炭消费主要用途中未填写“原料”。 
func (a *App) validateEnergyPurposeDetail(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	input_unit := getStringValue(row["input_unit"]) // 投入计量单位
	main_usage := getStringValue(row["main_usage"]) // 主要用途
	specific_usage := getStringValue(row["specific_usage"]) // 具体用途
	output_energy_types := getStringValue(row["output_energy_types"]) // 产出品种品类
	measurement_unit := getStringValue(row["measurement_unit"]) // 产出计量单位

	if input_unit != "万吨" {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "应为“万吨”", Flag: 0, Cells: GetCellPosition(usageFieldMapping, rowNumber, "input_unit"), RowNumber: rowNumber,})
	}

	// 若企业存在原料用煤,在主要用途中列出“原料”
	mapKey := getCreditCodeStatDateKey(row)
	creditCodeStatDate, exists := creditCodeStatDateMapping[mapKey]
	if exists && creditCodeStatDate.Annual_raw_coal > 0 && main_usage != "原料" {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "应为“原料”", Flag: 0, Cells: GetCellPosition(usageFieldMapping, rowNumber, "main_usage" ), RowNumber: rowNumber,})
		return;	
	}

	usage, exists := mainUsageMapping[main_usage]
	if !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不在选项中", Flag: 0, Cells: GetCellPosition(usageFieldMapping, rowNumber, "main_usage" ), RowNumber: rowNumber,})		
		return;	
	}

	specific, exists := usage[specific_usage]
	if  !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不在选项中", Flag: 0, Cells: GetCellPosition(usageFieldMapping, rowNumber, "specific_usage"), RowNumber: rowNumber,})
		return;
	}

	types, exists := specific[output_energy_types]
	if  !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不在选项中", Flag: 0, Cells: GetCellPosition(usageFieldMapping, rowNumber, "output_energy_types"), RowNumber: rowNumber,})	
		return;
	}

	if _, exists := types[measurement_unit]; !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "与产出品种品类不一致", Flag: 0, Cells: GetCellPosition(usageFieldMapping, rowNumber, "measurement_unit"), RowNumber: rowNumber,})	
	}
}

// validateEnergyPurposeRawMaterial 验证当原料用煤不为零，但在煤炭消费主要用途中未填写“原料”。
func (a *App) validateEnergyPurposeRawMaterial(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	input_quantity := getStringValue(row["input_quantity"])
	main_usage := getStringValue(row["main_usage"])

	if input_quantity != "0" && main_usage != "原料" {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "当原料用煤不为零，但在煤炭消费主要用途中未填写“原料”", Flag: 0, Cells: GetCellPosition(usageFieldMapping, rowNumber, "main_usage"), RowNumber: rowNumber,})	
	}
}

// validateTable1MainData 验证附表1煤炭消费主要信息
func (a *App) validateTable1MainData(mainData []map[string]interface{}) []ValidationError {
	var errors []ValidationError

	for _, row := range mainData {
		rowNumber := getRowNumber(row)
		mapKey := getCreditCodeStatDateKey(row)
		creditCodeStatDateMapping[mapKey] = CreditCacheInfo{
			Annual_raw_coal: parseFloat(getStringValue(row["annual_raw_coal"])),
		}
		// 年份无法正确识别，部分填写为“2023年”“2024年”，部分填写数字为宽体字符。
		a.validateStatDate(mainFieldMapping, row, rowNumber, &errors)
		// 统一社会信用代码填写错误。
		a.validateCreditCode(mainFieldMapping, row, rowNumber, &errors)
		// 行业门类、行业大类、行业中类不一致。
		a.validateTrade(mainFieldMapping, row, rowNumber, &errors)
		// 单位所在省/市/区、单位所在地市、单位所在区县不一致。
		a.validateAddress(mainFieldMapping, row, rowNumber, &errors)
		// 单位联系电话不是数字。
		a.validatePhone(row, rowNumber, &errors)
		// 综合能源消费情况中，年综合能耗等价值和当量值计算错误。
		a.validateEnergyValue(row, rowNumber, &errors)
		// 年原料用能大于年综合能耗。
		a.validateEnergyUse(row, rowNumber, &errors)
		// 煤炭消费情况填写存在空缺项，或填写内容不是数字。
		a.validateEnergyConsumption(row, rowNumber, &errors)
		// 煤炭消费情况单位填写错误。
		a.validateEnergyUnitData(row, rowNumber, &errors)
		// 除煤炭洗选及煤制品加工行业外，耗煤总量（实物量）与原煤消费量、洗精煤消费量、其他煤炭消费量（实物量）加和不相等。
		a.validateEnergyConsumptionTotal(row, rowNumber, &errors)
	}

	return errors
}

// validateEnergyUnitData  煤炭消费情况单位填写错误。
func (a *App) validateEnergyUnitData( row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	// 综合能耗当量值（万吨标准煤，含原料用能）、年综合能耗等价值（万吨标准煤，含原料用能）、年原料用能消费量（万吨标准煤）
	fieldNames := []string{
		"annual_energy_equivalent_value", "annual_energy_equivalent_cost", "annual_raw_material_energy",
	}
	a.validateEnergyUnit(mainFieldMapping, row, rowNumber, 0, fieldNames, errors)
	// 煤炭消费情况, 耗煤总量(实物量，万吨)	耗煤总量(标准量，万吨标准煤)	原料用煤(实物量，万吨)	原煤消费(实物量，万吨)	洗精煤消费(实物量，万吨)	其他煤炭消费(实物量，万吨)	焦炭消费(实物量，万吨)
	fieldNames = []string{
		"annual_total_coal_consumption", "annual_total_coal_products", "annual_raw_coal", "annual_raw_coal_consumption", "annual_clean_coal_consumption", "annual_other_coal_consumption", "annual_coke_consumption",
	}
	a.validateEnergyUnit(usageFieldMapping, row, rowNumber, 1, fieldNames, errors)
}

// validateEnergyUnit 验证数据与单位是否匹配
func (a *App) validateEnergyUnit(fieldMapping map[string]ExcelFieldMapping, row map[string]interface{}, rowNumber int, flag int, fieldNames []string, errors *[]ValidationError) {
	for _, fieldName := range fieldNames {
		value := getStringValue(row[fieldName])
		if isEmpty(value) || !isNumber(value) {
			continue;
		}
		
		// 判断规则，是数值除以10000大于1
		if parseFloat(value) / 10000 > 1 {
			*errors = append(*errors, ValidationError{ Type: "required", Message: "数据与单位不匹配", Flag: flag, Cells: GetCellPosition(fieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		}
	}
}


// validateEnergyConsumptionTotal 验证除煤炭洗选及煤制品加工行业外，耗煤总量（实物量）与原煤消费量、洗精煤消费量、其他煤炭消费量（实物量）加和不相等。
func (a *App) validateEnergyConsumptionTotal(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	annualTotalCoalConsumption := parseFloat(getStringValue(row["annual_total_coal_consumption"]))
	annualRawCoalConsumption := parseFloat(getStringValue(row["annual_raw_coal_consumption"]))
	annualCleanCoalConsumption := parseFloat(getStringValue(row["annual_clean_coal_consumption"]))
	annualOtherCoalConsumption := parseFloat(getStringValue(row["annual_other_coal_consumption"]))

	sumCoalConsumption := addFloat64(annualRawCoalConsumption, annualCleanCoalConsumption, annualOtherCoalConsumption)
	if sumCoalConsumption != annualTotalCoalConsumption {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "耗煤总量（实物量）与原煤消费量、洗精煤消费量、其他煤炭消费量（实物量）加和不相等", Flag: 1, Cells: GetCellPosition(mainFieldMapping, rowNumber, "annual_total_coal_consumption", "annual_raw_coal_consumption", "annual_clean_coal_consumption", "annual_other_coal_consumption"), RowNumber: rowNumber,})
	}
}

// validateEnergyConsumption 验证煤炭消费情况填写存在空缺项，或填写内容不是数字。
func (a *App) validateEnergyConsumption(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	fieldNames := []string{
		"annual_total_coal_consumption",
		"annual_total_coal_products",
		"annual_raw_coal",
		"annual_raw_coal_consumption",
		"annual_clean_coal_consumption",
		"annual_other_coal_consumption",
		"annual_coke_consumption",
	}
	for _, fieldName := range fieldNames {
		value := getStringValue(row[fieldName])
		if isEmpty(value) {
			*errors = append(*errors, ValidationError{ Type: "required", Message: "不能为空,如不存在，请填写“0.00”", Flag: 0, Cells: GetCellPosition(mainFieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		} else if !isNumber(value) {
			*errors = append(*errors, ValidationError{ Type: "required", Message: "应为数字", Flag: 0, Cells: GetCellPosition(mainFieldMapping, rowNumber, fieldName,), RowNumber: rowNumber,})
		}
	}
}

// validateEnergyUse 年原料用能大于年综合能耗。
func (a *App) validateEnergyUse(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	// 年综合能耗当量值、年综合能耗等价值、年原料用能消费量校验
	annualEnergyEquivalentValue := parseFloat(getStringValue(row["annual_energy_equivalent_value"]))
	annualEnergyEquivalentCost := parseFloat(getStringValue(row["annual_energy_equivalent_cost"]))
	annualRawMaterialEnergy := parseFloat(getStringValue(row["annual_raw_material_energy"]))

	// 年原料用能消费量≦年综合能耗当量值
	if isIntegerGreaterThan(annualRawMaterialEnergy, annualEnergyEquivalentValue) {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "年原料用能消费量不能大于年综合能耗当量值", Flag: 1, Cells: GetCellPosition(mainFieldMapping, rowNumber, "annual_raw_material_energy", "annual_energy_equivalent_value"), RowNumber: rowNumber,})
	}

	// 年原料用能消费量≦年综合能耗等价值
	if isIntegerGreaterThan(annualRawMaterialEnergy, annualEnergyEquivalentCost) {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "年原料用能消费量不能大于年综合能耗等价值", Flag: 1, Cells: GetCellPosition(mainFieldMapping, rowNumber, "annual_raw_material_energy", "annual_energy_equivalent_cost"), RowNumber: rowNumber,})
	}
}

// validateEnergyValue 综合能源消费情况中，年综合能耗等价值和当量值计算错误。
// 企业综合能源消费量=各种能源消费（包括能源加工转换的投入量，折标准煤）的合计-能源加工转换产出量（折标准煤）的合计-回收能利用量（折标准煤）的合计
func (a *App) validateEnergyValue(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	
}

// validatePhone 验证单位联系电话是否为数字
func (a *App) validatePhone(row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	tel := getStringValue(row["tel"])
	if !IsPhone(tel) {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "格式错误:应为手机号或固话", Flag: 0, Cells: GetCellPosition(mainFieldMapping, rowNumber, "tel"), RowNumber: rowNumber,})
	}
}

// validateAddress 验证单位所在省/市/区、单位所在地市、单位所在区县是否一致
func (a *App) validateAddress(fieldMapping map[string]ExcelFieldMapping, row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	province_name := getStringValue(row["province_name"])
	city_name := getStringValue(row["city_name"])
	country_name := getStringValue(row["country_name"])

	areaMapping, _ := a.GetChinaAreaMap()
		if _, exists := areaMapping[province_name]; !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不在中国省份选项中", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, "province_name"), RowNumber: rowNumber,})	
		return;	
	}
	if _, exists := areaMapping[province_name][city_name]; !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "单位所在省/市/区不一致", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, "city_name"), RowNumber: rowNumber,})	
		return;	
	}
	if _, exists := areaMapping[province_name][city_name][country_name]; !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "单位所在地市不一致", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, "country_name"), RowNumber: rowNumber,})	
		return;	
	}
}

// validateStatDate 验证年份是否为空或四位数字
func (a *App)validateStatDate(fieldMapping map[string]ExcelFieldMapping, row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	statDate := getStringValue(row["stat_date"])

	if statDate == "" {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不能为空", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, "stat_date"), RowNumber: rowNumber,})
	} else {
		// 请严格填写四位数字并使用常规字体，不得使用宽体字符，如“2023”“2024”，不得填写为“2023年”“2024年”。
		isNumber := isInteger(statDate)
		if !isNumber || len(statDate) != 4 {
			*errors = append(*errors, ValidationError{ Type: "required", Message: "应为四位数并使用常规字体", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, "stat_date"), RowNumber: rowNumber,})
		}
	}
}

// validateCreditCode 验证统一社会信用代码
func (a *App)validateCreditCode(fieldMapping map[string]ExcelFieldMapping, row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	credit_code := getStringValue(row["credit_code"])
	// 统一社会信用代码请填写18位字符，并确保该单元格为非科学计数法格式，如果发现类似错误，请下载最新版本表格填写。
	if len(credit_code) != 18 {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "应为18位字符", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, "credit_code"), RowNumber: rowNumber,})
	}
}

// validateTrade 验证行业分类
func (a *App) validateTrade(fieldMapping map[string]ExcelFieldMapping, row map[string]interface{}, rowNumber int, errors *[]ValidationError) {
	trade_a := getStringValue(row["trade_a"])
	trade_b := getStringValue(row["trade_b"])
	trade_c := getStringValue(row["trade_c"])
	
	tradeMapping, _ := a.GetTradeMapping()

	if _, exists := tradeMapping[trade_a]; !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "不在行业门类选项中", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, "trade_a"), RowNumber: rowNumber,})	
		return;	
	}
	if _, exists := tradeMapping[trade_a][trade_b]; !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "行业门类和行业大类不一致", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, "trade_b"), RowNumber: rowNumber,})
		return;
	}
	if _, exists := tradeMapping[trade_a][trade_b][trade_c]; !exists {
		*errors = append(*errors, ValidationError{ Type: "required", Message: "行业大类和行业中类不一致", Flag: 0, Cells: GetCellPosition(fieldMapping, rowNumber, "trade_c"), RowNumber: rowNumber,})
	}
}


// Table1MainData 表1主要数据
type Table1MainData struct {
	Cell_row_no int  `json:"cell_row_no"`	// 年份
	Stat_date string `json:"stat_date"`	// 年份
	Unit_name string `json:"unit_name"`	// 单位名称
	Credit_code string `json:"credit_code"`	// 统一社会信用代码
	Trade_a string `json:"trade_a"`	// 行业门类
	Trade_b string `json:"trade_b"`	// 行业大类
	Trade_c string `json:"trade_c"`	// 行业中类
	Province_name string `json:"province_name"`	// 单位所在省
	City_name string `json:"city_name"`	// 单位所在地市
	Country_name string `json:"country_name"`	// 单位所在区县
	Tel string `json:"tel"`	// 联系电话
	Annual_energy_equivalent_value string `json:"annual_energy_equivalent_value"`	// 年综合能耗当量值(万吨标准煤，含原料用能)
	Annual_energy_equivalent_cost string `json:"annual_energy_equivalent_cost"`	// 年综合能耗等价值(万吨标准煤，含原料用能)
	Annual_raw_material_energy string `json:"annual_raw_material_energy"`	// 年原料用能消费量(万吨标准煤)
	Annual_total_coal_consumption string `json:"annual_total_coal_consumption"`	// 年耗煤总量(实物量，万吨)
	Annual_total_coal_products string `json:"annual_total_coal_products"`	// 年耗煤总量(标准量，万吨标准煤)
	Annual_raw_coal string `json:"annual_raw_coal"`	// 年原煤消费(实物量，万吨)	
	Annual_raw_coal_consumption string `json:"annual_raw_coal_consumption"`// 原煤消费(实物量，万吨)
	Annual_clean_coal_consumption string `json:"annual_clean_coal_consumption"` // 洗精煤消费(实物量，万吨)
	Annual_other_coal_consumption string `json:"annual_other_coal_consumption"` // 其他煤炭消费(实物量，万吨)
	Annual_coke_consumption string `json:"annual_coke_consumption"` // 焦炭消费(实物量，万吨)
	State string `json:"state"` // 状态
}

// Table1UsageData 表1主要用途数据
type Table1UsageData struct {
	Cell_row_no int  `json:"cell_row_no"`	// 年份
	Stat_date string `json:"stat_date"`	
	Unit_name string `json:"unit_name"`
	Credit_code string `json:"credit_code"`
	Main_usage string `json:"main_usage"`
	Specific_usage string `json:"specific_usage"`
	Input_variety string `json:"input_variety"`
	Input_unit string `json:"input_unit"`
	Input_quantity string `json:"input_quantity"`
	Output_energy_types string `json:"output_energy_types"`
	Output_quantity string `json:"output_quantity"`
	Measurement_unit string `json:"measurement_unit"`
	Remarks string `json:"remarks"`
}

// Table1EquipData 表1重点耗煤装置（设备）数据
type Table1EquipData struct {
	Cell_row_no int  `json:"cell_row_no"`	// 年份
	Stat_date string `json:"stat_date"`	
	Unit_name string `json:"unit_name"`
	Credit_code string `json:"credit_code"`
	Equip_type string `json:"equip_type"`
	Equip_no string `json:"equip_no"`
	Total_runtime string `json:"total_runtime"`
	Design_life string `json:"design_life"`
	Energy_efficiency string `json:"energy_efficiency"`
	Capacity string `json:"capacity"`
	Capacity_unit string `json:"capacity_unit"`
	Coal_type string `json:"coal_type"`
	Annual_coal_consumption string `json:"annual_coal_consumption"`
}

// CreditCacheInfo 统一社会信用代码 + 年份 缓存信息
type CreditCacheInfo struct {
	Annual_raw_coal float64 `json:"annual_raw_coal"`
}