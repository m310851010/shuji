package data_import

import (
	"fmt"
	"shuji/db"
	"strings"
)

// 人工数据检查附表2
func (s *DataImportService) QueryDataTable2() db.QueryResult {
	return s.queryDataTable2Forinner(s.app.GetDB(), "")
}

// 人工数据检查附表2
func (s *DataImportService) queryDataTable2Forinner(database *db.Database, obj_id string) db.QueryResult {
	// 查询附表2数据，只返回未确认的数据
	query := `
			SELECT 
				obj_id, stat_date, create_time, unit_name, credit_code, trade_a, trade_b, trade_c,
				province_name, city_name, country_name, coal_type, coal_no, usage_time, design_life,
				enecrgy_efficienct_bmk, capacity_unit, capacity, use_info, status, annual_coal_consumption,
				row_no, create_user, is_confirm, is_check
			FROM critical_coal_equipment_consumption 
			%s
			ORDER BY create_time DESC
		`

	var result db.QueryResult
	var err error
	if obj_id != "" {
		query = fmt.Sprintf(query, "WHERE obj_id = ?")
		result, err = database.Query(query, obj_id)
	} else {
		query = fmt.Sprintf(query, "")
		result, err = database.Query(query)
	}

	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附表2数据失败: %v", err),
		}
	}

	if !result.Ok {
		return result
	}

	// 从result.Data中获取数据
	rawData, ok := result.Data.([]map[string]interface{})
	if !ok {
		return db.QueryResult{
			Ok:      false,
			Message: "数据格式错误",
		}
	}

	// 解密数值字段
	var data []map[string]interface{}
	for _, record := range rawData {
		decryptedRecord := map[string]interface{}{
			"obj_id":                  record["obj_id"],
			"stat_date":               record["stat_date"],
			"create_time":             record["create_time"],
			"unit_name":               record["unit_name"],
			"credit_code":             record["credit_code"],
			"trade_a":                 record["trade_a"],
			"trade_b":                 record["trade_b"],
			"trade_c":                 record["trade_c"],
			"province_name":           record["province_name"],
			"city_name":               record["city_name"],
			"country_name":            record["country_name"],
			"coal_type":               record["coal_type"],
			"coal_no":                 record["coal_no"],
			"usage_time":              record["usage_time"],
			"design_life":             s.decryptValue(record["design_life"]),
			"enecrgy_efficienct_bmk":  record["enecrgy_efficienct_bmk"],
			"capacity_unit":           record["capacity_unit"],
			"capacity":                s.decryptValue(record["capacity"]),
			"use_info":                record["use_info"],
			"status":                  record["status"],
			"annual_coal_consumption": s.decryptValue(record["annual_coal_consumption"]),
			"row_no":                  record["row_no"],
			"create_user":             record["create_user"],
			"is_confirm":              s.getDecryptedStatus(record["is_confirm"]),
			"is_check":                s.getDecryptedStatus(record["is_check"]),
		}
		data = append(data, decryptedRecord)
	}

	return db.QueryResult{
		Ok:   true,
		Data: data,
	}
}

func (s *DataImportService) ConfirmDataTable2(obj_id []string) db.QueryResult {
	if len(obj_id) == 0 {
		return db.QueryResult{
			Ok:      false,
			Message: "请选择要确认的数据",
		}
	}

	// 构建IN查询的占位符
	placeholders := strings.Repeat("?,", len(obj_id))
	placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号

	// 更新附表2确认状态
	query := fmt.Sprintf(`
		UPDATE critical_coal_equipment_consumption 
		SET is_confirm = ? 
		WHERE obj_id IN (%s)
	`, placeholders)

	args := []interface{}{EncryptedOne}
	args = append(args, s.convertToInterfaceSlice(obj_id)...)

	_, err := s.app.GetDB().Exec(query, args...)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("确认附表2数据失败: %v", err),
		}
	}

	return db.QueryResult{
		Ok:      true,
		Message: fmt.Sprintf("成功确认 %d 条附表2数据", len(obj_id)),
	}
}

// 查询附表2数据，指定数据库文件路径
func (s *DataImportService) QueryDataDetailTable2ByDBFile(obj_ids []string, dbFilePath string) db.QueryResult {

	if len(obj_ids) == 0 {
		return db.QueryResult{
			Ok:      false,
			Message: "没有提供obj_id",
		}
	}

	database, err := db.NewDatabase(dbFilePath, s.app.GetDBPassword())
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附表2数据失败: %v", err),
		}
	}

	defer database.Close()

	var allData []map[string]interface{}
	for _, obj_id := range obj_ids {
		data := s.queryDataTable2Forinner(database, obj_id)
		if data.Ok {
			dataList, ok := data.Data.([]map[string]interface{})
			if ok && len(dataList) > 0 {
				// 只取最新的一条数据
				allData = append(allData, dataList[len(dataList)-1])
			}
		}
	}

	return db.QueryResult{
		Ok:   true,
		Data: allData,
	}
}

// 查询附表2数据
func (s *DataImportService) QueryDataDetailTable2(obj_id string) db.QueryResult {
	return s.queryDataTable2Forinner(s.app.GetDB(), obj_id)
}
