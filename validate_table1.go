package main
import (
	"fmt"
)

// ValidateTable1 验证表1数据
func (a *App) validateTable1(f *excelize.File)  {
	
}

func (a *App) parseTable1MainSheet(f *excelize.File, sheetName string) ([]map[string]interface{}, error) {
	var mainData []map[string]interface{}
	// 读取企业基本信息表格
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("表格行数不足")
	}

	// 企业基本信息表格表头
	expectedHeadersRow1 := []string{
		"序号",	"年份",	"单位名称",	"统一社会信用代码",	"行业门类",	"行业大类",	"行业中类",	"单位所在省",	"单位所在地市",	"单位所在区县",	"联系电话", "综合能源消费情况", "", "", "煤炭消费情况","","","","","","", "状态",
	}

	// 第二行 从L2开始
	expectedHeadersRow2 := []string{ 
		"年综合能耗当量值(万吨标准煤，含原料用能)",	"年综合能耗等价值(万吨标准煤，含原料用能)",	"年原料用能消费量(万吨标准煤)",	"耗煤总量(实物量，万吨)",	"耗煤总量(标准量，万吨标准煤)",	"原料用煤(实物量，万吨)",	"原煤消费(实物量，万吨)",	"洗精煤消费(实物量，万吨)",	"其他煤炭消费(实物量，万吨)",	"焦炭消费(实物量，万吨)",
	}

	// 检查第一行是否为表头
	firstRow := rows[0]
	if len(firstRow) < len(expectedHeadersRow1) {
		return nil, fmt.Errorf("与数据模板不匹配：第一行表头列数与模板不符")
	}

	for i, header := range expectedHeadersRow1 {
		if firstRow[i] != header {
			return nil, fmt.Errorf("与数据模板不匹配：第一行表头列数与模板不符")
		}
	}

	// 检查第二行是否为表头
	secondRow := rows[1]
	if len(secondRow) < len(expectedHeadersRow2) {
		return nil, fmt.Errorf("与数据模板不匹配：第二行表头列数与模板不符")	
	}

	for i, header := range expectedHeadersRow2 {
		if secondRow[i] != header {
			return nil, fmt.Errorf("与数据模板不匹配：第二行表头第 %d 列预期为 %s，实际为 %s", i+1, header, secondRow[i])
		}
	}

	columns := []string{
		"stat_date","unit_name","credit_code",
	}
}