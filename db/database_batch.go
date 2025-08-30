package db

import (
	"fmt"
	"strings"
	"time"
)

// BatchInsertResult 批量插入结果
type BatchInsertResult struct {
	TotalRows    int64    `json:"totalRows"`    // 总行数
	SuccessRows  int64    `json:"successRows"`  // 成功行数
	FailedRows   int64    `json:"failedRows"`   // 失败行数
	LastInsertID int64    `json:"lastInsertId"` // 最后插入ID
	Errors       []string `json:"errors"`       // 错误信息
	Duration     int64    `json:"duration"`     // 执行时间(毫秒)
}

// BatchUpdateResult 批量更新结果
type BatchUpdateResult struct {
	TotalRows    int64    `json:"totalRows"`    // 总行数
	SuccessRows  int64    `json:"successRows"`  // 成功行数
	FailedRows   int64    `json:"failedRows"`   // 失败行数
	RowsAffected int64    `json:"rowsAffected"` // 影响行数
	Errors       []string `json:"errors"`       // 错误信息
	Duration     int64    `json:"duration"`     // 执行时间(毫秒)
}

// BatchData 批量数据项
type BatchData struct {
	Data map[string]interface{} `json:"data"` // 数据
	ID   interface{}            `json:"id"`   // 用于更新的ID
}

// BatchInsert 高性能批量插入
func (d *Database) BatchInsert(tableName string, dataList []map[string]interface{}, batchSize int) (BatchInsertResult, error) {
	startTime := time.Now()
	result := BatchInsertResult{
		TotalRows: int64(len(dataList)),
		Errors:    make([]string, 0),
	}

	if len(dataList) == 0 {
		return result, nil
	}

	// 开始事务
	tx, err := d.Begin()
	if err != nil {
		return result, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取列名
	var columns []string
	for col := range dataList[0] {
		columns = append(columns, col)
	}

	// 构建插入语句
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	baseQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// 准备批量插入语句
	stmt, err := tx.Prepare(baseQuery)
	if err != nil {
		return result, fmt.Errorf("准备语句失败: %v", err)
	}
	defer stmt.Close()

	// 分批处理
	for i := 0; i < len(dataList); i += batchSize {
		end := i + batchSize
		if end > len(dataList) {
			end = len(dataList)
		}

		batch := dataList[i:end]
		for _, data := range batch {
			var values []interface{}
			for _, col := range columns {
				values = append(values, data[col])
			}

			res, err := stmt.Exec(values...)
			if err != nil {
				result.FailedRows++
				result.Errors = append(result.Errors, fmt.Sprintf("行 %d: %v", i+1, err))
			} else {
				result.SuccessRows++
				if lastID, _ := res.LastInsertId(); lastID > result.LastInsertID {
					result.LastInsertID = lastID
				}
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("提交事务失败: %v", err)
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// BatchInsertWithMultiValue 使用多值插入语法的高性能批量插入
func (d *Database) BatchInsertWithMultiValue(tableName string, dataList []map[string]interface{}, batchSize int) (BatchInsertResult, error) {
	startTime := time.Now()
	result := BatchInsertResult{
		TotalRows: int64(len(dataList)),
		Errors:    make([]string, 0),
	}

	if len(dataList) == 0 {
		return result, nil
	}

	// 开始事务
	tx, err := d.Begin()
	if err != nil {
		return result, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取列名
	var columns []string
	for col := range dataList[0] {
		columns = append(columns, col)
	}

	// 分批处理
	for i := 0; i < len(dataList); i += batchSize {
		end := i + batchSize
		if end > len(dataList) {
			end = len(dataList)
		}

		batch := dataList[i:end]

		// 构建多值插入语句
		var valueStrings []string
		var allValues []interface{}

		for _, data := range batch {
			placeholders := make([]string, len(columns))
			for j := range placeholders {
				placeholders[j] = "?"
			}
			valueStrings = append(valueStrings, "("+strings.Join(placeholders, ", ")+")")

			for _, col := range columns {
				allValues = append(allValues, data[col])
			}
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(valueStrings, ", "))

		res, err := tx.Exec(query, allValues...)
		if err != nil {
			result.FailedRows += int64(len(batch))
			result.Errors = append(result.Errors, fmt.Sprintf("批次 %d-%d: %v", i+1, end, err))
		} else {
			result.SuccessRows += int64(len(batch))
			if lastID, _ := res.LastInsertId(); lastID > result.LastInsertID {
				result.LastInsertID = lastID
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("提交事务失败: %v", err)
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// BatchUpdate 高性能批量更新
func (d *Database) BatchUpdate(tableName string, dataList []BatchData, idColumn string, batchSize int) (BatchUpdateResult, error) {
	startTime := time.Now()
	result := BatchUpdateResult{
		TotalRows: int64(len(dataList)),
		Errors:    make([]string, 0),
	}

	if len(dataList) == 0 {
		return result, nil
	}

	// 开始事务
	tx, err := d.Begin()
	if err != nil {
		return result, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取更新列名（排除ID列）
	var updateColumns []string
	for col := range dataList[0].Data {
		if col != idColumn {
			updateColumns = append(updateColumns, col)
		}
	}

	// 构建更新语句
	var setClauses []string
	for _, col := range updateColumns {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
	}

	baseQuery := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?",
		tableName,
		strings.Join(setClauses, ", "),
		idColumn)

	// 准备批量更新语句
	stmt, err := tx.Prepare(baseQuery)
	if err != nil {
		return result, fmt.Errorf("准备语句失败: %v", err)
	}
	defer stmt.Close()

	// 分批处理
	for i := 0; i < len(dataList); i += batchSize {
		end := i + batchSize
		if end > len(dataList) {
			end = len(dataList)
		}

		batch := dataList[i:end]
		for _, item := range batch {
			var values []interface{}
			for _, col := range updateColumns {
				values = append(values, item.Data[col])
			}
			values = append(values, item.ID)

			res, err := stmt.Exec(values...)
			if err != nil {
				result.FailedRows++
				result.Errors = append(result.Errors, fmt.Sprintf("行 %d: %v", i+1, err))
			} else {
				result.SuccessRows++
				if rowsAffected, _ := res.RowsAffected(); rowsAffected > 0 {
					result.RowsAffected += rowsAffected
				}
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("提交事务失败: %v", err)
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// BatchDelete 高性能批量删除
func (d *Database) BatchDelete(tableName string, ids []interface{}, idColumn string, batchSize int) (BatchUpdateResult, error) {
	startTime := time.Now()
	result := BatchUpdateResult{
		TotalRows: int64(len(ids)),
		Errors:    make([]string, 0),
	}

	if len(ids) == 0 {
		return result, nil
	}

	// 开始事务
	tx, err := d.Begin()
	if err != nil {
		return result, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 分批处理
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		batch := ids[i:end]

		// 构建IN查询
		placeholders := make([]string, len(batch))
		for j := range placeholders {
			placeholders[j] = "?"
		}

		query := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)",
			tableName,
			idColumn,
			strings.Join(placeholders, ", "))

		res, err := tx.Exec(query, batch...)
		if err != nil {
			result.FailedRows += int64(len(batch))
			result.Errors = append(result.Errors, fmt.Sprintf("批次 %d-%d: %v", i+1, end, err))
		} else {
			result.SuccessRows += int64(len(batch))
			if rowsAffected, _ := res.RowsAffected(); rowsAffected > 0 {
				result.RowsAffected += rowsAffected
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("提交事务失败: %v", err)
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// BatchUpsert 批量插入或更新（使用REPLACE INTO）
func (d *Database) BatchUpsert(tableName string, dataList []map[string]interface{}, batchSize int) (BatchInsertResult, error) {
	startTime := time.Now()
	result := BatchInsertResult{
		TotalRows: int64(len(dataList)),
		Errors:    make([]string, 0),
	}

	if len(dataList) == 0 {
		return result, nil
	}

	// 开始事务
	tx, err := d.Begin()
	if err != nil {
		return result, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取列名
	var columns []string
	for col := range dataList[0] {
		columns = append(columns, col)
	}

	// 分批处理
	for i := 0; i < len(dataList); i += batchSize {
		end := i + batchSize
		if end > len(dataList) {
			end = len(dataList)
		}

		batch := dataList[i:end]

		// 构建REPLACE INTO语句
		var valueStrings []string
		var allValues []interface{}

		for _, data := range batch {
			placeholders := make([]string, len(columns))
			for j := range placeholders {
				placeholders[j] = "?"
			}
			valueStrings = append(valueStrings, "("+strings.Join(placeholders, ", ")+")")

			for _, col := range columns {
				allValues = append(allValues, data[col])
			}
		}

		query := fmt.Sprintf("REPLACE INTO %s (%s) VALUES %s",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(valueStrings, ", "))

		res, err := tx.Exec(query, allValues...)
		if err != nil {
			result.FailedRows += int64(len(batch))
			result.Errors = append(result.Errors, fmt.Sprintf("批次 %d-%d: %v", i+1, end, err))
		} else {
			result.SuccessRows += int64(len(batch))
			if lastID, _ := res.LastInsertId(); lastID > result.LastInsertID {
				result.LastInsertID = lastID
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("提交事务失败: %v", err)
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// BatchInsertWithIgnore 批量插入（忽略重复）
func (d *Database) BatchInsertWithIgnore(tableName string, dataList []map[string]interface{}, batchSize int) (BatchInsertResult, error) {
	startTime := time.Now()
	result := BatchInsertResult{
		TotalRows: int64(len(dataList)),
		Errors:    make([]string, 0),
	}

	if len(dataList) == 0 {
		return result, nil
	}

	// 开始事务
	tx, err := d.Begin()
	if err != nil {
		return result, fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 获取列名
	var columns []string
	for col := range dataList[0] {
		columns = append(columns, col)
	}

	// 分批处理
	for i := 0; i < len(dataList); i += batchSize {
		end := i + batchSize
		if end > len(dataList) {
			end = len(dataList)
		}

		batch := dataList[i:end]

		// 构建INSERT OR IGNORE语句
		var valueStrings []string
		var allValues []interface{}

		for _, data := range batch {
			placeholders := make([]string, len(columns))
			for j := range placeholders {
				placeholders[j] = "?"
			}
			valueStrings = append(valueStrings, "("+strings.Join(placeholders, ", ")+")")

			for _, col := range columns {
				allValues = append(allValues, data[col])
			}
		}

		query := fmt.Sprintf("INSERT OR IGNORE INTO %s (%s) VALUES %s",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(valueStrings, ", "))

		res, err := tx.Exec(query, allValues...)
		if err != nil {
			result.FailedRows += int64(len(batch))
			result.Errors = append(result.Errors, fmt.Sprintf("批次 %d-%d: %v", i+1, end, err))
		} else {
			result.SuccessRows += int64(len(batch))
			if lastID, _ := res.LastInsertId(); lastID > result.LastInsertID {
				result.LastInsertID = lastID
			}
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("提交事务失败: %v", err)
	}

	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}
