package main

import (
	"log"
	"shuji/db"
	"sync"
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
	db       *db.Database
	logQueue chan *DataImportRecord
}

var (
	instance *DataImportRecordService
	once     sync.Once
)

// NewDataImportRecordService 创建导入记录服务实例（单例模式）
func NewDataImportRecordService(db *db.Database) *DataImportRecordService {
	once.Do(func() {
		instance = &DataImportRecordService{
			db:       db,
			logQueue: make(chan *DataImportRecord, 10000), // 队列大小10000
		}

		// 启动异步日志处理协程
		go instance.asyncLogWorker()
	})

	return instance
}

// asyncLogWorker 异步日志处理工作协程
func (s *DataImportRecordService) asyncLogWorker() {
	for record := range s.logQueue {
		s.insertRecordToDB(record)
	}
}

// insertRecordToDB 实际插入记录到数据库
func (s *DataImportRecordService) insertRecordToDB(record *DataImportRecord) {
	query := `
		INSERT INTO data_import_record (
			obj_id, file_name, file_type, import_time, 
			import_state, describe, create_user
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

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
		log.Printf("异步插入导入记录失败: %v", err)
	}
}

// InsertImportRecord 异步插入导入记录
func (s *DataImportRecordService) InsertImportRecord(fileName, fileType, importState, describe string) {
	record := &DataImportRecord{
		ObjID:       uuid.New().String(),
		FileName:    fileName,
		FileType:    fileType,
		ImportTime:  time.Now(),
		ImportState: importState,
		Describe:    describe,
		CreateUser:  GetCurrentOSUser(),
	}

	// 异步发送到日志队列
	select {
	case s.logQueue <- record:
		// 日志已成功加入队列
	default:
		// 队列已满，记录警告但不阻塞主流程
		log.Printf("日志队列已满，丢弃日志记录: %s - %s", fileName, describe)
	}
}

// GetAllImportRecords 获取所有导入记录
func (s *DataImportRecordService) GetImportRecordsByFileType(fileType string) db.QueryResult {
	query := "SELECT * FROM data_import_record WHERE file_type = ? ORDER BY import_time DESC"
	result, err := s.db.Query(query, fileType)
	if err != nil {
		return db.QueryResult{
			Ok:      false,
			Message: err.Error(),
			Data:    nil,
		}
	}

	return result
}
