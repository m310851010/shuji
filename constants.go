package main

// 应用相关常量
const (
	// 应用名称
	APP_NAME = "数据校验工具"

	// 应用版本
	APP_VERSION = "v1.0.0"

	// 数据目录名称
	DATA_DIR_NAME = "data"

	// 缓存文件目录名称
	CACHE_FILE_DIR_NAME = DATA_DIR_NAME + "/files"

	// 配置文件名
	CONFIG_FILE_NAME = "config.json"

	// 前端文件目录名称
	FRONTEND_FILE_DIR_NAME = "frontend/dist/"

	// 中国区域信息文件路径
	CHINA_AREA_FILE_PATH = FRONTEND_FILE_DIR_NAME + "China.json"
	// 行业门类映射文件路径
	TRADE_MAPPING_FILE_PATH = FRONTEND_FILE_DIR_NAME + "TradeMapping.json"
)

// 验证错误类型
const (
	ERROR_TYPE_TEXT_FORMAT_ERROR = "文字格式填写错误"
	ERROR_TYPE_OPTION_ERROR = "未按选项填写错误"
	ERROR_TYPE_DATA_UNIT_MISMATCH = "数据与单位不匹配"
	ERROR_TYPE_DATA_LOGIC_ERROR = "数据校验逻辑错误"
	ERROR_TYPE_NON_NUMBER_ERROR = "非数字格式错误"
	ERROR_TYPE_MISSING_DATA_ERROR = "缺数据错误"
)
