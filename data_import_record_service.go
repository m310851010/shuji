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
	ObjID       string `json:"obj_id" db:"obj_id"`             // 主键
	FileName    string `json:"file_name" db:"file_name"`       // 导入文件名
	FileType    string `json:"file_type" db:"file_type"`       // 文件类型
	ImportTime  int64  `json:"import_time" db:"import_time"`   // 导入时间
	ImportState string `json:"import_state" db:"import_state"` // 导入状态，导入成功，导入失败
	Describe    string `json:"describe" db:"describe"`         // 说明
	CreateUser  string `json:"create_user" db:"create_user"`   // 导入用户
}

// DataImportRecordService 导入记录服务
type DataImportRecordService struct {
	db       *db.Database
	logQueue chan *DataImportRecord
	app      *App
}

var (
	instance *DataImportRecordService
	once     sync.Once
)

// NewDataImportRecordService 创建导入记录服务实例（单例模式）
func NewDataImportRecordService(db *db.Database, app *App) *DataImportRecordService {
	once.Do(func() {
		instance = &DataImportRecordService{
			db:       db,
			logQueue: make(chan *DataImportRecord, 10000), // 队列大小10000
			app:      app,
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
		ImportTime:  time.Now().UnixMilli(),
		ImportState: importState,
		Describe:    describe,
		CreateUser:  s.app.GetAreaStr(),
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
	// 使用包装函数来处理异常
	return s.getImportRecordsByFileTypeWithRecover(fileType)
}

// getImportRecordsByFileTypeWithRecover 带异常处理的获取导入记录函数
func (s *DataImportRecordService) getImportRecordsByFileTypeWithRecover(fileType string) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetImportRecordsByFileType 发生异常: %v", r)
		}
	}()

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
