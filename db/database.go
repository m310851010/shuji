package db

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Database SQLite数据库管理器
type Database struct {
	db *sql.DB
}

// retryOnBusy 在数据库锁定错误时重试操作
func (d *Database) retryOnBusy(operation func() error) error {
	maxRetries := 20
	retryDelay := time.Millisecond * 100

	for i := 0; i < maxRetries; i++ {
		err := operation()
		if err == nil {
			return nil
		}

		// 检查是否是锁定错误
		if strings.Contains(err.Error(), "database is locked") || strings.Contains(err.Error(), "SQLITE_BUSY") {
			if i < maxRetries-1 {
				log.Printf("数据库锁定，重试第%d次: %v", i+1, err)
				time.Sleep(retryDelay)
				retryDelay *= 2 // 指数退避
				continue
			}
		}

		// 如果不是锁定错误，或者已经重试完所有次数，直接返回错误
		return err
	}

	return fmt.Errorf("重试次数已用完")
}

// QueryResult 查询结果
type QueryResult struct {
	Ok      bool        `json:"ok"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// NewDatabase 创建新的数据库连接
func NewDatabase(dbPath string, password string) (*Database, error) {
	// 对于加密的数据库，使用 modernc.org/sqlite 的连接方式

	password = hex.EncodeToString([]byte(password))

	// 尝试不同的连接配置
	type driverConfig struct {
		driver string
		format string
	}

	configs := []driverConfig{
		// 使用 modernc.org/sqlite 驱动（注册为 sqlite）
		{"sqlite", dbPath + "?_pragma=key('" + password + "')&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?_pragma=key=\"" + password + "\"&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?_pragma=key=" + password + "&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?_pragma=hexkey=" + password + "&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		// 尝试不同的加密算法
		{"sqlite", dbPath + "?_pragma=key('" + password + "')&_pragma=cipher=aes256cbc&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?_pragma=key('" + password + "')&_pragma=cipher=chacha20&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?_pragma=key('" + password + "')&_pragma=cipher=aes256ctr&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?_pragma_key=" + password + "&_pragma_cipher_page_size=4096&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?_pragma_key=" + password + "&_pragma_cipher_page_size=4096&_pragma_cipher=aes256ctr&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?_pragma_key=" + password + "&_pragma_cipher_page_size=4096&_pragma_cipher=aes256cbc&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?_pragma_key=" + password + "&_pragma_cipher_page_size=4096&_pragma_cipher=chacha20&_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
		{"sqlite", dbPath + "?mode=rw&_journal_mode=WAL&_busy_timeout=30000&key=" + password + "&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY"},
	}

	var lastErr error
	for i, config := range configs {
		fmt.Printf("尝试连接配置 %d: %s - %s\n", i+1, config.driver, config.format)
		db, err := sql.Open(config.driver, config.format)
		if err != nil {
			fmt.Printf("配置 %d 打开失败: %v\n", i+1, err)
			lastErr = err
			continue
		}

		// 测试连接
		if err := db.Ping(); err != nil {
			fmt.Printf("配置 %d 连接测试失败: %v\n", i+1, err)
			db.Close()
			lastErr = err
			continue
		}

		fmt.Printf("配置 %d 连接成功\n", i+1)
		// 连接成功，设置连接池参数
		db.SetMaxOpenConns(150)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Hour)

		return &Database{db: db}, nil
	}

	// 如果所有格式都失败，尝试不加密的方式
	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY")
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(150)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := db.Ping(); err != nil {
		db.Close()
		// 返回最后一个错误信息
		return nil, fmt.Errorf("数据库连接测试失败: %v", lastErr)
	}

	return &Database{db: db}, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Exec 执行SQL语句
func (d *Database) Exec(query string, args ...interface{}) (QueryResult, error) {
	if d.db == nil {
		return QueryResult{Ok: false, Message: "数据库未初始化"}, fmt.Errorf("数据库未初始化")
	}

	var result sql.Result
	var err error

	err = d.retryOnBusy(func() error {
		result, err = d.db.Exec(query, args...)
		return err
	})

	if err != nil {
		log.Printf("SQL执行失败: %v", err)
		return QueryResult{Ok: false, Message: err.Error()}, err
	}

	lastID, _ := result.LastInsertId()
	rowsAffected, _ := result.RowsAffected()

	return QueryResult{
		Ok: true,
		Data: map[string]interface{}{
			"lastInsertId": lastID,
			"rowsAffected": rowsAffected,
		},
		Message: "执行成功",
	}, nil
}

// Query 执行查询语句
func (d *Database) Query(query string, args ...interface{}) (QueryResult, error) {
	if d.db == nil {
		return QueryResult{Ok: false, Message: "数据库未初始化"}, fmt.Errorf("数据库未初始化")
	}

	var rows *sql.Rows
	var err error

	err = d.retryOnBusy(func() error {
		rows, err = d.db.Query(query, args...)
		return err
	})

	if err != nil {
		return QueryResult{Ok: false, Message: err.Error()}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return QueryResult{Ok: false, Message: err.Error()}, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return QueryResult{Ok: false, Message: err.Error()}, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				switch v := val.(type) {
				case []byte:
					row[col] = string(v)
				default:
					row[col] = v
				}
			} else {
				row[col] = nil
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return QueryResult{Ok: false, Message: err.Error()}, err
	}

	return QueryResult{
		Ok:      true,
		Data:    results,
		Message: fmt.Sprintf("查询成功，返回 %d 条记录", len(results)),
	}, nil
}

// Insert 插入数据
func (d *Database) Insert(tableName string, data map[string]interface{}) (QueryResult, error) {
	if len(data) == 0 {
		return QueryResult{Ok: false, Message: "数据不能为空"}, nil
	}

	var columns []string
	var placeholders []string
	var values []interface{}

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return d.Exec(query, values...)
}

// Update 更新数据
func (d *Database) Update(tableName string, data map[string]interface{}, where string, args ...interface{}) (QueryResult, error) {
	if len(data) == 0 {
		return QueryResult{Ok: false, Message: "更新数据不能为空"}, nil
	}

	var setClauses []string
	var values []interface{}

	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
		values = append(values, val)
	}

	values = append(values, args...)

	query := fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(setClauses, ", "))
	if where != "" {
		query += " WHERE " + where
	}

	return d.Exec(query, values...)
}

// Delete 删除数据
func (d *Database) Delete(tableName string, where string, args ...interface{}) (QueryResult, error) {
	query := fmt.Sprintf("DELETE FROM %s", tableName)
	if where != "" {
		query += " WHERE " + where
	}

	return d.Exec(query, args...)
}

// Select 查询数据
func (d *Database) Select(tableName string, columns []string, where string, args ...interface{}) (QueryResult, error) {
	cols := "*"
	if len(columns) > 0 {
		cols = strings.Join(columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, tableName)
	if where != "" {
		query += " WHERE " + where
	}

	return d.Query(query, args...)
}

// GetTables 获取所有表名
func (d *Database) GetTables() (QueryResult, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
	return d.Query(query)
}

// Begin 开始事务
func (d *Database) Begin() (*sql.Tx, error) {
	if d.db == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var tx *sql.Tx
	var err error

	err = d.retryOnBusy(func() error {
		tx, err = d.db.Begin()
		return err
	})

	return tx, err
}

// QueryRow 查询单行数据
func (d *Database) QueryRow(query string, args ...interface{}) (QueryResult, error) {
	// 先执行查询获取列信息
	var rows *sql.Rows
	var err error

	err = d.retryOnBusy(func() error {
		rows, err = d.db.Query(query, args...)
		return err
	})

	if err != nil {
		return QueryResult{Ok: false, Message: err.Error()}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return QueryResult{Ok: false, Message: err.Error()}, err
	}

	// 只读取第一行
	if rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return QueryResult{Ok: false, Message: err.Error()}, err
		}

		result := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				switch v := val.(type) {
				case []byte:
					result[col] = string(v)
				default:
					result[col] = v
				}
			} else {
				result[col] = nil
			}
		}

		return QueryResult{Ok: true, Data: result, Message: "查询成功"}, nil
	}

	// 没有找到数据
	return QueryResult{Ok: true, Data: nil, Message: "未找到数据"}, nil
}

// Count 统计记录数
func (d *Database) Count(tableName string, where string, args ...interface{}) (QueryResult, error) {
	query := fmt.Sprintf("SELECT COUNT(1) as count FROM %s", tableName)
	if where != "" {
		query += " WHERE " + where
	}
	return d.QueryRow(query, args...)
}

// SelectOne 查询单条数据
func (d *Database) SelectOne(tableName string, columns []string, where string, args ...interface{}) (QueryResult, error) {
	cols := "*"
	if len(columns) > 0 {
		cols = strings.Join(columns, ", ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s", cols, tableName)
	if where != "" {
		query += " WHERE " + where
	}
	query += " LIMIT 1"

	return d.QueryRow(query, args...)
}
