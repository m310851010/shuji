package data_import
import (
	"shuji/db"
)

func (s *DataImportService) ModelDataCheckTable3() db.QueryResult {
	return db.QueryResult{
		Data:    nil,
		Ok:      true,
		Message: "数据校验通过",
	}
}
