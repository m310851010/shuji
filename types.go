package main

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

// KeyEquipmentListData 装置清单数据结构
type KeyEquipmentListData struct {
	ObjID            string `json:"obj_id" db:"obj_id"`                         // 主键
	ProvinceName     string `json:"province_name" db:"province_name"`           // 单位省级名称
	CityName         string `json:"city_name" db:"city_name"`                   // 单位市级名称
	CountryName      string `json:"country_name" db:"country_name"`             // 单位县级名称
	UnitName         string `json:"unit_name" db:"unit_name"`                   // 单位详细名称
	CreditCode       string `json:"credit_code" db:"credit_code"`               // 统一社会信用代码
	EquipType        string `json:"equip_type" db:"equip_type"`                 // 设备类型
	EquipModelNumber string `json:"equip_model_number" db:"equip_model_number"` // 设备型号
	EquipNo          string `json:"equip_no" db:"equip_no"`                     // 设备编号
}

// EnterpriseList 企业清单数据结构
type EnterpriseListData struct {
	ObjID        string `json:"obj_id" db:"obj_id"`               // 主键
	ProvinceName string `json:"province_name" db:"province_name"` // 单位省级名称
	CityName     string `json:"city_name" db:"city_name"`         // 单位市级名称
	CountryName  string `json:"country_name" db:"country_name"`   // 单位县级名称
	UnitName     string `json:"unit_name" db:"unit_name"`         // 单位详细名称
	CreditCode   string `json:"credit_code" db:"credit_code"`     // 统一社会信用代码
}

// ExcelParseResult Excel解析结果
type ExcelParseResult struct {
	Ok      bool     `json:"ok"`
	Message string   `json:"message"`
	Total   int      `json:"total"`   // 总记录数
	Success int      `json:"success"` // 成功导入数
	Failed  int      `json:"failed"`  // 失败数
	Errors  []string `json:"errors"`  // 错误信息列表
}

// PasswordInfo 密码信息结构
type PasswordInfo struct {
	ObjID    string `json:"obj_id" db:"obj_id"`       // 主键
	AdminPws string `json:"admin_pws" db:"admin_pws"` // 管理员密码，加密
	UserPws  string `json:"user_pws" db:"user_pws"`   // 用户密码，加密
}

// AppConfig 应用配置结构
type AppConfig struct {
	CurrentUnitName string `json:"current_unit_name"` // 当前单位名称
	CurrentProvince string `json:"current_province"`  // 当前省份
	CurrentCity     string `json:"current_city"`      // 当前城市
	CurrentCountry  string `json:"current_country"`   // 当前县区
}

// AreaInfo 区域信息结构
type AreaInfo struct {
	Code string `json:"code"` // 区域代码
	Name string `json:"name"` // 区域名称
}

// EnhancedAreaConfig 增强的区域配置结构
type EnhancedAreaConfig struct {
	ObjID            string     `json:"obj_id"`            // 主键
	ProvinceName     string     `json:"province_name"`     // 省级名称
	CityName         string     `json:"city_name"`         // 市级名称
	CountryName      string     `json:"country_name"`      // 县级名称
	ProvinceCode     string     `json:"province_code"`     // 省级代码
	CityCode         string     `json:"city_code"`         // 市级代码
	CountryCode      string     `json:"country_code"`      // 县级代码
	DataLevel        int        `json:"data_level"`        // 数据级别：1-省级，2-市级，3-县级
	SubordinateAreas []AreaInfo `json:"subordinate_areas"` // 下级区域列表
}
