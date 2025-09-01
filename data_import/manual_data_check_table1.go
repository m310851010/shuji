package data_import

import (
	"fmt"
	"shuji/db"
	"strings"
)

// 人工数据检查附表1
func (s *DataImportService) QueryDataTable1() db.QueryResult {
	// 查询主表数据，只返回未确认的数据
	query := `
		SELECT 
			obj_id, unit_name, stat_date, credit_code, create_time, create_user, is_confirm, is_check
		FROM enterprise_coal_consumption_main 
		ORDER BY create_time DESC
	`
	result, err := s.app.GetDB().Query(query)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附表1数据失败: %v", err),
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

	// 处理数据字段
	var data []map[string]interface{}
	for _, record := range rawData {
		// 格式化时间

		processedRecord := map[string]interface{}{
			"obj_id":      record["obj_id"],
			"unit_name":   record["unit_name"],
			"stat_date":   record["stat_date"],
			"credit_code": record["credit_code"],
			"create_time": record["create_time"],
			"create_user": record["create_user"],
			"is_confirm":  s.getDecryptedStatus(record["is_confirm"]),
			"is_check":    s.getDecryptedStatus(record["is_check"]),
		}
		data = append(data, processedRecord)
	}

	return db.QueryResult{
		Ok:   true,
		Data: data,
	}
}

// 查询附表1数据，指定数据库文件路径
func (s *DataImportService) QueryDataDetailTable1ByDBFile(obj_ids []string, dbFilePath string) db.QueryResult {

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
			Message: fmt.Sprintf("查询附表1数据失败: %v", err),
		}
	}

	defer database.Close()

	var allData []map[string]interface{}
	for _, obj_id := range obj_ids {
		result := s.queryDataDetailTable1Forinner(obj_id, database)
		if result.Ok && result.Data != nil {
			if data, ok := result.Data.(map[string]interface{}); ok {
				allData = append(allData, data)
			}
		}
	}

	defer database.Close()

	return db.QueryResult{
		Ok:   true,
		Data: allData,
	}
}

func (s *DataImportService) queryDataDetailTable1Forinner(obj_id string, database *db.Database) db.QueryResult {
	// 查询主表数据
	mainQuery := `
		SELECT 
			obj_id, unit_name, stat_date, tel, credit_code, trade_a, trade_b, trade_c,
			province_name, city_name, country_name,
			annual_energy_equivalent_value, annual_energy_equivalent_cost, annual_raw_material_energy,
			annual_total_coal_consumption, annual_total_coal_products, annual_raw_coal,
			annual_raw_coal_consumption, annual_clean_coal_consumption, annual_other_coal_consumption,
			annual_coke_consumption, create_time, create_user, is_confirm, is_check
		FROM enterprise_coal_consumption_main 
		WHERE obj_id = ?
	`

	mainResult, err := database.QueryRow(mainQuery, obj_id)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附表1主表数据失败: %v", err),
		}
	}

	if !mainResult.Ok {
		return mainResult
	}

	mainData, ok := mainResult.Data.(map[string]interface{})
	if !ok || mainData == nil {
		return db.QueryResult{
			Ok:      false,
			Message: "未找到指定的主表数据",
		}
	}

	mainObjId := obj_id

	// 解密主表数值字段
	decryptedMainData := map[string]interface{}{
		"obj_id":                         mainData["obj_id"],
		"unit_name":                      mainData["unit_name"],
		"stat_date":                      mainData["stat_date"],
		"tel":                            mainData["tel"],
		"credit_code":                    mainData["credit_code"],
		"trade_a":                        mainData["trade_a"],
		"trade_b":                        mainData["trade_b"],
		"trade_c":                        mainData["trade_c"],
		"province_name":                  mainData["province_name"],
		"city_name":                      mainData["city_name"],
		"country_name":                   mainData["country_name"],
		"annual_energy_equivalent_value": s.decryptValue(mainData["annual_energy_equivalent_value"]),
		"annual_energy_equivalent_cost":  s.decryptValue(mainData["annual_energy_equivalent_cost"]),
		"annual_raw_material_energy":     s.decryptValue(mainData["annual_raw_material_energy"]),
		"annual_total_coal_consumption":  s.decryptValue(mainData["annual_total_coal_consumption"]),
		"annual_total_coal_products":     s.decryptValue(mainData["annual_total_coal_products"]),
		"annual_raw_coal":                s.decryptValue(mainData["annual_raw_coal"]),
		"annual_raw_coal_consumption":    s.decryptValue(mainData["annual_raw_coal_consumption"]),
		"annual_clean_coal_consumption":  s.decryptValue(mainData["annual_clean_coal_consumption"]),
		"annual_other_coal_consumption":  s.decryptValue(mainData["annual_other_coal_consumption"]),
		"annual_coke_consumption":        s.decryptValue(mainData["annual_coke_consumption"]),
		"create_time":                    mainData["create_time"],
		"create_user":                    mainData["create_user"],
		"is_confirm":                     s.getDecryptedStatus(mainData["is_confirm"]),
		"is_check":                       s.getDecryptedStatus(mainData["is_check"]),
	}

	// 查询用途表数据
	usageQuery := `
		SELECT 
			obj_id, fk_id, stat_date, main_usage, specific_usage, input_variety, input_unit,
			input_quantity, output_energy_types, output_quantity, measurement_unit, remarks, row_no,
			create_time, is_confirm, is_check
		FROM enterprise_coal_consumption_usage 
		WHERE fk_id = ?
		ORDER BY create_time DESC
	`

	usageResult, err := database.Query(usageQuery, mainObjId)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附表1用途表数据失败: %v", err),
		}
	}

	var usageData []map[string]interface{}
	if usageResult.Ok {
		if usageDataList, ok := usageResult.Data.([]map[string]interface{}); ok {
			for _, usage := range usageDataList {
				usageRecord := map[string]interface{}{
					"obj_id":              usage["obj_id"],
					"fk_id":               usage["fk_id"],
					"stat_date":           usage["stat_date"],
					"main_usage":          usage["main_usage"],
					"specific_usage":      usage["specific_usage"],
					"input_variety":       usage["input_variety"],
					"input_unit":          usage["input_unit"],
					"input_quantity":      s.decryptValue(usage["input_quantity"]),
					"output_energy_types": usage["output_energy_types"],
					"output_quantity":     s.decryptValue(usage["output_quantity"]),
					"measurement_unit":    usage["measurement_unit"],
					"remarks":             usage["remarks"],
					"row_no":              usage["row_no"],
					"create_time":         usage["create_time"],
					"is_confirm":          s.getDecryptedStatus(usage["is_confirm"]),
					"is_check":            s.getDecryptedStatus(usage["is_check"]),
				}
				usageData = append(usageData, usageRecord)
			}
		}
	}

	// 查询设备表数据
	equipQuery := `
		SELECT 
			obj_id, fk_id, stat_date, equip_type, equip_no, total_runtime, design_life,
			energy_efficiency, capacity_unit, capacity, coal_type, annual_coal_consumption, row_no,
			create_time, is_confirm, is_check
		FROM enterprise_coal_consumption_equip 
		WHERE fk_id = ?
		ORDER BY create_time DESC
	`

	equipResult, err := database.Query(equipQuery, mainObjId)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("查询附表1设备表数据失败: %v", err),
		}
	}

	var equipData []map[string]interface{}
	if equipResult.Ok {
		if equipDataList, ok := equipResult.Data.([]map[string]interface{}); ok {
			for _, equip := range equipDataList {
				equipRecord := map[string]interface{}{
					"obj_id":                  equip["obj_id"],
					"fk_id":                   equip["fk_id"],
					"stat_date":               equip["stat_date"],
					"equip_type":              equip["equip_type"],
					"equip_no":                equip["equip_no"],
					"total_runtime":           s.decryptValue(equip["total_runtime"]),
					"design_life":             s.decryptValue(equip["design_life"]),
					"energy_efficiency":       s.decryptValue(equip["energy_efficiency"]),
					"capacity_unit":           equip["capacity_unit"],
					"capacity":                s.decryptValue(equip["capacity"]),
					"coal_type":               equip["coal_type"],
					"annual_coal_consumption": s.decryptValue(equip["annual_coal_consumption"]),
					"row_no":                  equip["row_no"],
					"create_time":             equip["create_time"],
					"is_confirm":              s.getDecryptedStatus(equip["is_confirm"]),
					"is_check":                s.getDecryptedStatus(equip["is_check"]),
				}
				equipData = append(equipData, equipRecord)
			}
		}
	}

	result := map[string]interface{}{
		"main":  decryptedMainData,
		"usage": usageData,
		"equip": equipData,
	}

	return db.QueryResult{
		Ok:   true,
		Data: result,
	}
}

func (s *DataImportService) QueryDataDetailTable1(obj_id string) db.QueryResult {
	return s.queryDataDetailTable1Forinner(obj_id, s.app.GetDB())
}

func (s *DataImportService) ConfirmDataTable1(obj_id []string) db.QueryResult {
	if len(obj_id) == 0 {
		return db.QueryResult{
			Ok:      false,
			Message: "请选择要确认的数据",
		}
	}

	// 构建IN查询的占位符
	placeholders := strings.Repeat("?,", len(obj_id))
	placeholders = placeholders[:len(placeholders)-1] // 移除最后一个逗号

	// 更新主表确认状态
	mainQuery := fmt.Sprintf(`
		UPDATE enterprise_coal_consumption_main 
		SET is_confirm = ? 
		WHERE obj_id IN (%s)
	`, placeholders)

	args := []interface{}{EncryptedOne}
	args = append(args, s.convertToInterfaceSlice(obj_id)...)

	_, err := s.app.GetDB().Exec(mainQuery, args...)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("确认附表1主表数据失败: %v", err),
		}
	}

	// 更新用途表确认状态
	usageQuery := fmt.Sprintf(`
		UPDATE enterprise_coal_consumption_usage 
		SET is_confirm = ? 
		WHERE fk_id IN (%s)
	`, placeholders)

	usageArgs := []interface{}{EncryptedOne}
	usageArgs = append(usageArgs, s.convertToInterfaceSlice(obj_id)...)

	_, err = s.app.GetDB().Exec(usageQuery, usageArgs...)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("确认附表1用途表数据失败: %v", err),
		}
	}

	// 更新设备表确认状态
	equipQuery := fmt.Sprintf(`
		UPDATE enterprise_coal_consumption_equip 
		SET is_confirm = ? 
		WHERE fk_id IN (%s)
	`, placeholders)

	equipArgs := []interface{}{EncryptedOne}
	equipArgs = append(equipArgs, s.convertToInterfaceSlice(obj_id)...)

	_, err = s.app.GetDB().Exec(equipQuery, equipArgs...)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: fmt.Sprintf("确认附表1设备表数据失败: %v", err),
		}
	}

	return db.QueryResult{
		Ok:      true,
		Message: fmt.Sprintf("成功确认 %d 条附表1数据", len(obj_id)),
	}
}
