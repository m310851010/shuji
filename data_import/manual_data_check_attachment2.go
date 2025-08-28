package data_import

import (
	"fmt"
	"shuji/db"
	"strings"
)

// 人工数据检查附件2
func (s *DataImportService) QueryDataAttachment2() db.QueryResult {
	// 查询附件2数据，只返回未确认的数据
	query := `
		SELECT 
			obj_id, stat_date, sg_code, unit_id, unit_name, unit_level, province_name, city_name, country_name,
			total_coal, raw_coal, washed_coal, other_coal, power_generation, heating, coal_washing,
			coking, oil_refining, gas_production, industry, raw_materials, other_uses, coke,
			create_user, create_time, is_confirm, is_check
		FROM coal_consumption_report 
		ORDER BY create_time DESC
	`

	result, err := s.app.GetDB().Query(query)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附件2数据失败: %v", err),
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
			"obj_id":           record["obj_id"],
			"stat_date":        record["stat_date"],
			"sg_code":          record["sg_code"],
			"unit_id":          record["unit_id"],
			"unit_name":        record["unit_name"],
			"unit_level":       record["unit_level"],
			"province_name":    record["province_name"],
			"city_name":        record["city_name"],
			"country_name":     record["country_name"],
			"total_coal":       s.decryptValue(record["total_coal"]),
			"raw_coal":         s.decryptValue(record["raw_coal"]),
			"washed_coal":      s.decryptValue(record["washed_coal"]),
			"other_coal":       s.decryptValue(record["other_coal"]),
			"power_generation": s.decryptValue(record["power_generation"]),
			"heating":          s.decryptValue(record["heating"]),
			"coal_washing":     s.decryptValue(record["coal_washing"]),
			"coking":           s.decryptValue(record["coking"]),
			"oil_refining":     s.decryptValue(record["oil_refining"]),
			"gas_production":   s.decryptValue(record["gas_production"]),
			"industry":         s.decryptValue(record["industry"]),
			"raw_materials":    s.decryptValue(record["raw_materials"]),
			"other_uses":       s.decryptValue(record["other_uses"]),
			"coke":             s.decryptValue(record["coke"]),
			"create_user":      record["create_user"],
			"create_time":      record["create_time"],
			"is_confirm":       s.getDecryptedStatus(record["is_confirm"]),
			"is_check":         s.getDecryptedStatus(record["is_check"]),
		}
		data = append(data, decryptedRecord)
	}

	return db.QueryResult{
		Ok:   true,
		Data: data,
	}
}

// 查询附件2数据，指定数据库文件路径
func (s *DataImportService) QueryDataDetailAttachment2ByDBFile(obj_id string, dbFilePath string) db.QueryResult {
	database, err := db.NewDatabase(dbFilePath, s.app.GetDBPassword())
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附件2数据失败: %v", err),
		}
	}

	return s.queryDataDetailAttachment2Forinner(obj_id, database)
}

func (s *DataImportService) queryDataDetailAttachment2Forinner(obj_id string, database *db.Database) db.QueryResult {
	// 查询附件2数据
	query := `
		SELECT 
			obj_id, stat_date, sg_code, unit_id, unit_name, unit_level, province_name, city_name, country_name,
			total_coal, raw_coal, washed_coal, other_coal, power_generation, heating, coal_washing,
			coking, oil_refining, gas_production, industry, raw_materials, other_uses, coke,
			create_user, create_time, is_confirm, is_check
		FROM coal_consumption_report 
		WHERE obj_id = ?
	`

	result, err := database.QueryRow(query, obj_id)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附件2数据失败: %v", err),
		}
	}

	if !result.Ok {
		return result
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok || data == nil {
		return db.QueryResult{
			Ok:      false,
			Message: "未找到指定的附件2数据",
		}
	}

	// 解密数值字段
	decryptedData := map[string]interface{}{
		"obj_id":           data["obj_id"],
		"stat_date":        data["stat_date"],
		"sg_code":          data["sg_code"],
		"unit_id":          data["unit_id"],
		"unit_name":        data["unit_name"],
		"unit_level":       data["unit_level"],
		"province_name":    data["province_name"],
		"city_name":        data["city_name"],
		"country_name":     data["country_name"],
		"total_coal":       s.decryptValue(data["total_coal"]),
		"raw_coal":         s.decryptValue(data["raw_coal"]),
		"washed_coal":      s.decryptValue(data["washed_coal"]),
		"other_coal":       s.decryptValue(data["other_coal"]),
		"power_generation": s.decryptValue(data["power_generation"]),
		"heating":          s.decryptValue(data["heating"]),
		"coal_washing":     s.decryptValue(data["coal_washing"]),
		"coking":           s.decryptValue(data["coking"]),
		"oil_refining":     s.decryptValue(data["oil_refining"]),
		"gas_production":   s.decryptValue(data["gas_production"]),
		"industry":         s.decryptValue(data["industry"]),
		"raw_materials":    s.decryptValue(data["raw_materials"]),
		"other_uses":       s.decryptValue(data["other_uses"]),
		"coke":             s.decryptValue(data["coke"]),
		"create_user":      data["create_user"],
		"create_time":      data["create_time"],
		"is_confirm":       s.getDecryptedStatus(data["is_confirm"]),
		"is_check":         s.getDecryptedStatus(data["is_check"]),
	}

	return db.QueryResult{
		Ok:   true,
		Data: decryptedData,
	}
}

func (s *DataImportService) QueryDataDetailAttachment2(obj_id string) db.QueryResult {
	return s.queryDataDetailAttachment2Forinner(obj_id, s.app.GetDB())
}

func (s *DataImportService) ConfirmDataAttachment2(obj_id []string) db.QueryResult {
	if len(obj_id) == 0 {
		return db.QueryResult{
			Ok:      false,
			Message: "请选择要确认的数据",
		}
	}

	// 构建IN查询的占位符
	placeholders := strings.Repeat("?,", len(obj_id))
	placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号

	// 更新附件2确认状态
	query := fmt.Sprintf(`
		UPDATE coal_consumption_report 
		SET is_confirm = ? 
		WHERE obj_id IN (%s)
	`, placeholders)

	args := []interface{}{EncryptedOne}
	args = append(args, s.convertToInterfaceSlice(obj_id)...)

	_, err := s.app.GetDB().Exec(query, args...)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("确认附件2数据失败: %v", err),
		}
	}

	return db.QueryResult{
		Ok:      true,
		Message: fmt.Sprintf("成功确认 %d 条附件2数据", len(obj_id)),
	}
}
