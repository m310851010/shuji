package data_import

import (
	"fmt"
	"shuji/db"
	"strings"
)

// 人工数据检查附表3
func (s *DataImportService) QueryDataTable3() db.QueryResult {
	// 查询附表3数据，只返回未确认的数据
	query := `
		SELECT 
			obj_id, stat_date, project_name, project_code, construction_unit, main_construction_content,
			province_name, city_name, country_name, trade_a, trade_c, examination_approval_time,
			scheduled_time, actual_time, examination_authority, document_number,
			equivalent_value, equivalent_cost, pq_total_coal_consumption, pq_coal_consumption,
			pq_coke_consumption, pq_blue_coke_consumption, sce_total_coal_consumption, sce_coal_consumption,
			sce_coke_consumption, sce_blue_coke_consumption, is_substitution, substitution_source,
			substitution_quantity, pq_annual_coal_quantity, sce_annual_coal_quantity,
			create_time, create_user, is_confirm, is_check
		FROM fixed_assets_investment_project 
		ORDER BY create_time DESC
	`

	result, err := s.app.GetDB().Query(query)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附表3数据失败: %v", err),
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
			"obj_id":                     record["obj_id"],
			"stat_date":                  record["stat_date"],
			"project_name":               record["project_name"],
			"project_code":               record["project_code"],
			"construction_unit":          record["construction_unit"],
			"main_construction_content":  record["main_construction_content"],
			"province_name":              record["province_name"],
			"city_name":                  record["city_name"],
			"country_name":               record["country_name"],
			"trade_a":                    record["trade_a"],
			"trade_c":                    record["trade_c"],
			"examination_approval_time":  record["examination_approval_time"],
			"scheduled_time":             record["scheduled_time"],
			"actual_time":                record["actual_time"],
			"examination_authority":      record["examination_authority"],
			"document_number":            record["document_number"],
			"equivalent_value":           s.decryptValue(record["equivalent_value"]),
			"equivalent_cost":            s.decryptValue(record["equivalent_cost"]),
			"pq_total_coal_consumption":  s.decryptValue(record["pq_total_coal_consumption"]),
			"pq_coal_consumption":        s.decryptValue(record["pq_coal_consumption"]),
			"pq_coke_consumption":        s.decryptValue(record["pq_coke_consumption"]),
			"pq_blue_coke_consumption":   s.decryptValue(record["pq_blue_coke_consumption"]),
			"sce_total_coal_consumption": s.decryptValue(record["sce_total_coal_consumption"]),
			"sce_coal_consumption":       s.decryptValue(record["sce_coal_consumption"]),
			"sce_coke_consumption":       s.decryptValue(record["sce_coke_consumption"]),
			"sce_blue_coke_consumption":  s.decryptValue(record["sce_blue_coke_consumption"]),
			"is_substitution":            record["is_substitution"],
			"substitution_source":        record["substitution_source"],
			"substitution_quantity":      s.decryptValue(record["substitution_quantity"]),
			"pq_annual_coal_quantity":    s.decryptValue(record["pq_annual_coal_quantity"]),
			"sce_annual_coal_quantity":   s.decryptValue(record["sce_annual_coal_quantity"]),
			"create_time":                record["create_time"],
			"create_user":                record["create_user"],
			"is_confirm":                 s.getDecryptedStatus(record["is_confirm"]),
			"is_check":                   s.getDecryptedStatus(record["is_check"]),
		}
		data = append(data, decryptedRecord)
	}

	return db.QueryResult{
		Ok:   true,
		Data: data,
	}
}

// 查询附表3数据，指定数据库文件路径
func (s *DataImportService) QueryDataDetailTable3ByDBFile(obj_ids []string, dbFilePath string) db.QueryResult {
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
			Message: fmt.Sprintf("查询附表3数据失败: %v", err),
		}
	}

	defer database.Close()

	

	var allData []map[string]interface{}
	for _, obj_id := range obj_ids {
		result := s.queryDataDetailTable3Forinner(obj_id, database)
		if result.Ok && result.Data != nil {
			if data, ok := result.Data.(map[string]interface{}); ok {
				allData = append(allData, data)
			}
		}
	}

	return db.QueryResult{
		Ok:   true,
		Data: allData,
	}
}

func (s *DataImportService) queryDataDetailTable3Forinner(obj_id string, database *db.Database) db.QueryResult {
	// 查询附表3数据
	query := `
		SELECT 
			obj_id, stat_date, sg_code, project_name, project_code, construction_unit, main_construction_content,
			unit_id, province_name, city_name, country_name, trade_a, trade_c, examination_approval_time,
			scheduled_time, actual_time, examination_authority, document_number,
			equivalent_value, equivalent_cost, pq_total_coal_consumption, pq_coal_consumption,
			pq_coke_consumption, pq_blue_coke_consumption, sce_total_coal_consumption, sce_coal_consumption,
			sce_coke_consumption, sce_blue_coke_consumption, is_substitution, substitution_source,
			substitution_quantity, pq_annual_coal_quantity, sce_annual_coal_quantity,
			create_time, create_user, is_confirm, is_check
		FROM fixed_assets_investment_project 
		WHERE obj_id = ?
	`

	result, err := database.QueryRow(query, obj_id)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附表3数据失败: %v", err),
		}
	}

	if !result.Ok {
		return result
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok || data == nil {
		return db.QueryResult{
			Ok:      false,
			Message: "未找到指定的附表3数据",
		}
	}

	// 解密数值字段
	decryptedData := map[string]interface{}{
		"obj_id":                     data["obj_id"],
		"stat_date":                  data["stat_date"],
		"sg_code":                    data["sg_code"],
		"project_name":               data["project_name"],
		"project_code":               data["project_code"],
		"construction_unit":          data["construction_unit"],
		"main_construction_content":  data["main_construction_content"],
		"unit_id":                    data["unit_id"],
		"province_name":              data["province_name"],
		"city_name":                  data["city_name"],
		"country_name":               data["country_name"],
		"trade_a":                    data["trade_a"],
		"trade_c":                    data["trade_c"],
		"examination_approval_time":  data["examination_approval_time"],
		"scheduled_time":             data["scheduled_time"],
		"actual_time":                data["actual_time"],
		"examination_authority":      data["examination_authority"],
		"document_number":            data["document_number"],
		"equivalent_value":           s.decryptValue(data["equivalent_value"]),
		"equivalent_cost":            s.decryptValue(data["equivalent_cost"]),
		"pq_total_coal_consumption":  s.decryptValue(data["pq_total_coal_consumption"]),
		"pq_coal_consumption":        s.decryptValue(data["pq_coal_consumption"]),
		"pq_coke_consumption":        s.decryptValue(data["pq_coke_consumption"]),
		"pq_blue_coke_consumption":   s.decryptValue(data["pq_blue_coke_consumption"]),
		"sce_total_coal_consumption": s.decryptValue(data["sce_total_coal_consumption"]),
		"sce_coal_consumption":       s.decryptValue(data["sce_coal_consumption"]),
		"sce_coke_consumption":       s.decryptValue(data["sce_coke_consumption"]),
		"sce_blue_coke_consumption":  s.decryptValue(data["sce_blue_coke_consumption"]),
		"is_substitution":            data["is_substitution"],
		"substitution_source":        data["substitution_source"],
		"substitution_quantity":      s.decryptValue(data["substitution_quantity"]),
		"pq_annual_coal_quantity":    s.decryptValue(data["pq_annual_coal_quantity"]),
		"sce_annual_coal_quantity":   s.decryptValue(data["sce_annual_coal_quantity"]),
		"create_time":                data["create_time"],
		"create_user":                data["create_user"],
		"is_confirm":                 s.getDecryptedStatus(data["is_confirm"]),
		"is_check":                   s.getDecryptedStatus(data["is_check"]),
	}

	return db.QueryResult{
		Ok:   true,
		Data: decryptedData,
	}
}

func (s *DataImportService) QueryDataDetailTable3(obj_id string) db.QueryResult {
	return s.queryDataDetailTable3Forinner(obj_id, s.app.GetDB())
}

func (s *DataImportService) ConfirmDataTable3(obj_id []string) db.QueryResult {
	if len(obj_id) == 0 {
		return db.QueryResult{
			Ok:      false,
			Message: "请选择要确认的数据",
		}
	}

	// 构建IN查询的占位符
	placeholders := strings.Repeat("?,", len(obj_id))
	placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号

	// 更新附表3确认状态
	query := fmt.Sprintf(`
		UPDATE fixed_assets_investment_project 
		SET is_confirm = ? 
		WHERE obj_id IN (%s)
	`, placeholders)

	args := []interface{}{EncryptedOne}
	args = append(args, s.convertToInterfaceSlice(obj_id)...)

	_, err := s.app.GetDB().Exec(query, args...)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("确认附表3数据失败: %v", err),
		}
	}

	return db.QueryResult{
		Ok:      true,
		Message: fmt.Sprintf("成功确认 %d 条附表3数据", len(obj_id)),
	}
}
