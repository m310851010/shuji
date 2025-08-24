package main

// 加密相关常量
const (
	// 默认加密密钥 - 生产环境中修改此密钥
	DEFAULT_ENCRYPTION_KEY = "shuji2024secretkey"

	// 密钥最小长度
	MIN_KEY_LENGTH = 8

	// 密钥推荐长度
	RECOMMENDED_KEY_LENGTH = 16

	// 密钥最大长度
	MAX_KEY_LENGTH = 32
)

// 应用相关常量
const (
	// 应用名称
	APP_NAME = "煤炭摸底数据校验软件"

	// 应用版本
	APP_VERSION = "v1.0.0"

	// 数据目录名称
	DATA_DIR_NAME = "data"

    // 缓存目录名称
	CACHE_DIR_NAME = DATA_DIR_NAME +  "/cache"

	// 缓存文件目录名称
	CACHE_FILE_DIR_NAME = CACHE_DIR_NAME + "/files"

	// 数据库文件名
	DB_FILE_NAME = "coal_consumption_data.db"

	// 数据库密码
	DB_PASSWORD = "shuji"
)

// 加密算法相关常量
const (
	// SM4密钥长度（字节）
	SM4_KEY_LENGTH = 16

	// SM4块大小（字节）
	SM4_BLOCK_SIZE = 16
)


// 文件类型常量
const (
	TableName1      = "规上企业"
	TableName2      = "其他单位"
	TableName3      = "新上项目"
	TableAttachment2 = "区域综合"
)

const (
	TableType1      = "table1"
	TableType2      = "table2"
	TableType3      = "table3"
	TableTypeAttachment2 = "attachment2"
)
