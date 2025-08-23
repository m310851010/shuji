package main

import (
	"fmt"
	"shuji/db"
	"time"

	"github.com/google/uuid"
)


// DataImportRecord 导入记录
type DataImportRecord struct {
	ObjID       string    `json:"obj_id" db:"obj_id"`             // 主键
	FileName    string    `json:"file_name" db:"file_name"`       // 上传文件名
	FileType    string    `json:"file_type" db:"file_type"`       // 文件类型
	ImportTime  time.Time `json:"import_time" db:"import_time"`   // 上传时间
	ImportState string    `json:"import_state" db:"import_state"` // 上传状态，上传成功，上传失败
	Describe    string    `json:"describe" db:"describe"`         // 说明
	CreateUser  string    `json:"create_user" db:"create_user"`   // 上传用户
}

// DataImportRecordService 导入记录服务
type DataImportRecordService struct {
	db *db.Database
}

// NewDataImportRecordService 创建导入记录服务实例
func NewDataImportRecordService(db *db.Database) *DataImportRecordService {
	return &DataImportRecordService{
		db: db,
	}
}

// InsertImportRecord 插入导入记录
func (s *DataImportRecordService) InsertImportRecord(record *DataImportRecord) error {
	// 生成UUID
	record.ObjID = uuid.New().String()

	// 如果没有设置时间，使用当前时间
	if record.ImportTime.IsZero() {
		record.ImportTime = time.Now()
	}

	// 构建SQL语句
	query := `
		INSERT INTO data_import_record (
			obj_id, file_name, file_type, import_time, 
			import_state, describe, create_user
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	// 执行插入
	_, err := s.db.Exec(
		query,
		record.ObjID,
		record.FileName,
		record.FileType,
		record.ImportTime,
		record.ImportState,
		record.Describe,
		record.CreateUser,
	)

	if err != nil {
		return fmt.Errorf("插入导入记录失败: %v", err)
	}

	return nil
}

// GetAllImportRecords 获取所有导入记录
func (s *DataImportRecordService) GetImportRecordsByFileType(fileType string) ([]map[string]interface{}, error) {
	query := "SELECT * FROM data_import_record WHERE file_type = ? ORDER BY import_time DESC"
	result, err := s.db.Query(query, fileType)
	if err != nil {
		return nil, err
	}

	return result.Data.([]map[string]interface{}), err
}
