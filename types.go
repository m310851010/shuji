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
	ThresholdTotalCoalConsumption int `json:"thresholdTotalCoalConsumption"`
	// 耗煤用途阈值 “耗煤总量(实物量，万吨)”与附表1主要用途情况，企业的“投入量”之和相等
    ThresholdMainUsage int `json:"thresholdMainUsage"`
	// 耗煤设备阈值 “耗煤总量(实物量，万吨)”与附表1重点耗煤装置（设备）情况这重点耗煤装置的年耗煤量之和相等
    ThresholdCoalEquipment int `json:"thresholdCoalEquipment"`
	// 煤合计阈值 本年度的“煤合计”与“原煤”“洗精煤”“其他”应该相等
    ThresholdTotalCoal3 int `json:"thresholdTotalCoal3"`
	// 煤合计阈值 本年度的“煤合计”数值与“能源加工转换”和“终端消费”之和应该相等
    ThresholdTotalCoalSum int `json:"thresholdTotalCoalSum"`
}