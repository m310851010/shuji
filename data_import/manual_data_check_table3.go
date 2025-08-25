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
