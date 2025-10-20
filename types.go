package main

// QueryResult 查询结果
type QueryResult struct {
	Ok      bool        `json:"ok"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// EnvResult 环境变量结果
type EnvResult struct {
	AppName     string `json:"appName"`
	AppFileName string `json:"appFileName"`
	BasePath    string `json:"basePath"`
	OS          string `json:"os"`
	ARCH        string `json:"arch"`
	X64Level    int    `json:"x64Level"`
	ExePath     string `json:"exePath"`
	AssetsDir   string `json:"assetsDir"`
}

type FlagResult struct {
	Ok   bool   `json:"ok"`
	Data string `json:"data"`
}

// FileInfo 自定义结构体，用于返回前端
type FileInfo struct {
	Name         string `json:"name"`         // 文件名（不含路径）
	FullPath     string `json:"fullPath"`     // 完整路径
	Size         int64  `json:"size"`         // 大小（字节）
	IsDirectory  bool   `json:"isDirectory"`  // 是否是文件夹
	IsFile       bool   `json:"isFile"`       // 是否是文件
	LastModified int64  `json:"lastModified"` // 修改时间
	Ext          string `json:"ext"`          // 扩展名（文件）| 空（文件夹）
	ParentDir    string `json:"parentDir"`    // 父目录路径
}

// MessageBoxOptions 消息框选项
type MessageBoxOptions struct {
	Title     string   `json:"title,omitempty"`     // 标题
	Message   string   `json:"message"`             // 主消息内容
	Detail    string   `json:"detail,omitempty"`    // 详细信息
	Type      string   `json:"type,omitempty"`      // 类型: "none", "info", "error", "question", "warning"
	Buttons   []string `json:"buttons,omitempty"`   // 按钮文本数组
	DefaultId int      `json:"defaultId,omitempty"` // 默认按钮索引
	CancelId  int      `json:"cancelId,omitempty"`  // 取消按钮索引
}

type MessageBoxResult struct {
	Response        int  `json:"response"`        // 用户点击的按钮索引
	CheckboxChecked bool `json:"checkboxChecked"` // 复选框状态（如有）
}

// FileDialogOptions 文件选择对话框选项
type FileDialogOptions struct {
	Title   string       `json:"title,omitempty"`
	Filters []FileFilter `json:"filters,omitempty"`
	// OpenDirectory 是否可以选择目录
	OpenDirectory bool `json:"openDirectory,omitempty"`
	// CreateDirectory 是否可以创建目录
	CreateDirectory bool `json:"createDirectory,omitempty"`
	// DefaultPath 默认路径
	DefaultPath string `json:"defaultPath,omitempty"`
	// 默认文件名
	DefaultFilename string `json:"defaultFilename,omitempty"`
	// MultiSelections 是否可以多选
	MultiSelections bool `json:"multiSelections,omitempty"`
}

// FileFilter 文件选择对话框筛选器
type FileFilter struct {
	// Name 显示名称
	Name string `json:"name"`
	// Pattern 筛选模式
	Pattern string `json:"pattern"`
}

// FileDialogResult 文件选择对话框结果
type FileDialogResult struct {
	// Canceled 是否取消选择
	Canceled bool `json:"canceled"`
	// Path 选中的文件路径
	FilePaths []string `json:"filePaths"`
}

// AppConfig 应用配置
type AppConfig struct {
	// 耗煤总量阈值, “耗煤总量(实物量，万吨)”数值与“原煤消费(实物量，万吨)”“洗精煤消费(实物量，万吨)”“其他煤炭消费(实物量，万吨)”加和的数值相等
	ThresholdTotalCoalConsumption float64 `json:"thresholdTotalCoalConsumption"`
	// 耗煤用途阈值 “耗煤总量(实物量，万吨)”与附表1主要用途情况，企业的“投入量”之和相等
    ThresholdMainUsage float64 `json:"thresholdMainUsage"`
	// 耗煤设备阈值 “耗煤总量(实物量，万吨)”与附表1重点耗煤装置（设备）情况这重点耗煤装置的年耗煤量之和相等
    ThresholdCoalEquipment float64 `json:"thresholdCoalEquipment"`
	// 对于某一个区域（省、市、县）年度煤合计数值，预留差异空间配置
    ThresholdTotalCoal_province float64 `json:"thresholdTotalCoal_province"`
	// 对于某一个区域（省、市、县）年度煤合计数值，预留差异空间配置
    ThresholdTotalCoal_city float64 `json:"thresholdTotalCoal_city"`
	// 对于某一个区域（省、市、县）年度煤合计数值，预留差异空间配置
    ThresholdTotalCoal_country float64 `json:"thresholdTotalCoal_country"`
	// 对于某一个区域（省、市、县）的煤炭消费品种数据，预留差异空间配置
    ThresholdEnergyTypes float64 `json:"thresholdEnergyTypes"`
	// 对于某一个区域（省、市、县）的煤炭消费数据，预留差异空间配置
    ThresholdCoalConsumption float64 `json:"thresholdCoalConsumption"`
	// 对于某一个区域（省、市）的煤炭消费数据，本年度“煤合计”数值与下辖所有区域（市、县）“煤合计”数值，预留差异空间配置
    ThresholdTotalCoal_area_province float64 `json:"thresholdTotalCoal_area_province"`
	// 对于某一个区域（省、市）的煤炭消费数据，本年度“煤合计”数值与下辖所有区域（市、县）“煤合计”数值，预留差异空间配置
    ThresholdTotalCoal_area_city float64 `json:"thresholdTotalCoal_area_city"`
	// 对于某一个区域（省、市）的煤炭消费数据，本年度“煤合计”数值与下辖所有区域（市、县）“煤合计”数值，预留差异空间配置
    ThresholdTotalCoal_area_country float64 `json:"thresholdTotalCoal_area_country"`
}

// ValidationError 验证错误结构
type ValidationError struct {
	RowNumber int      `json:"row_number"` // 错误行号
	Message   string   `json:"message"`    // 错误信息
	Type      string   `json:"type"`       // 错误类型 "文字格式填写错误","未按选项填写错误","数据与单位不匹配","数据校验逻辑错误","非数字格式错误","缺数据错误"
	Flag      int   `json:"flag"`      // 错误标识, 用于高亮单元格 0: 蓝色, 1: 黄色
	Cells      []string   `json:"cells"`      // 单元格位置，如["A1", "B1", "C1"]
	SheetName string `json:"sheetName"` // 表格名称
}

// ExcelFieldMapping Excel字段映射
type ExcelFieldMapping struct {
	FieldName string `json:"fieldName"` // 字段名
	Index int `json:"index"` // 字段索引
	Column    string `json:"column"` // Excel列名，如 "A", "B", "C"	
}

// ValidateSheetResult 验证表格结果
type ValidateSheetResult struct {
	SheetName string `json:"sheetName"` // 表格名称
	RowNumber int `json:"lastRowNumber"` // 总行数
	ColumnNumber int `json:"lastColumnNumber"` // 总列数
	Errors []ValidationError `json:"errors"` // 错误列表
}

// ParseSheetResult 解析表格结果
type ParseSheetResult struct {
	SheetName string `json:"sheetName"` // 表格名称
	RowNumber int `json:"lastRowNumber"` // 总行数
	ColumnNumber int `json:"lastColumnNumber"` // 总列数
	Data []map[string]interface{} `json:"data"` // 数据
}