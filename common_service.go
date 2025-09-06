package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"shuji/db"
	"strings"

	"github.com/google/uuid"
)

// SetUserPassword 设置用户密码
func (a *App) SetUserPassword(password string) db.QueryResult {
	// 使用包装函数来处理异常
	return a.setUserPasswordWithRecover(password)
}

// setUserPasswordWithRecover 带异常处理的设置用户密码函数
func (a *App) setUserPasswordWithRecover(password string) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("SetUserPassword 发生异常: %v", r)
		}
	}()

	result := db.QueryResult{
		Ok:      false,
		Message: "",
	}

	// 验证密码不能为空
	if strings.TrimSpace(password) == "" {
		result.Message = "密码不能为空"
		return result
	}

	encryptedPassword, err := SM4Encrypt(password)
	if err != nil {
		result.Message = "密码加密失败: " + err.Error()
		return result
	}

	// 开始事务
	tx, err := a.db.Begin()
	if err != nil {
		result.Message = "开始事务失败: " + err.Error()
		return result
	}
	defer func() {
		if result.Ok {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// 检查pws_info表是否存在记录
	var existingObjID string
	err = tx.QueryRow("SELECT obj_id FROM pws_info LIMIT 1").Scan(&existingObjID)

	if err == sql.ErrNoRows {
		// 没有记录，返回前端数据异常
		return db.QueryResult{Ok: false, Message: "数据异常，请联系管理员！"}
	}

	if err == nil {
		// 有记录，更新用户密码
		_, err = tx.Exec(`
			UPDATE pws_info 
			SET user_pws = ?
			WHERE obj_id = ?
		`, encryptedPassword, existingObjID)
	} else {
		// 其他错误
		result.Message = "查询密码记录失败: " + err.Error()
		return result
	}

	if err != nil {
		result.Message = "设置密码失败: " + err.Error()
		return result
	}

	result.Ok = true
	result.Message = "用户密码设置成功"
	return result
}

// GetPasswordInfo 获取密码信息
func (a *App) GetPasswordInfo() db.QueryResult {
	// 使用包装函数来处理异常
	return a.getPasswordInfoWithRecover()
}

// getPasswordInfoWithRecover 带异常处理的获取密码信息函数
func (a *App) getPasswordInfoWithRecover() db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetPasswordInfo 发生异常: %v", r)
		}
	}()

	result, err := a.db.Query("SELECT obj_id, user_pws FROM pws_info LIMIT 1")
	if err != nil {
		return db.QueryResult{Ok: false, Message: "查询密码信息失败: " + err.Error()}
	}

	return result
}

// 对密码加密后从数据库中查询密码是否正确
func (a *App) Login(password string) db.QueryResult {
	// 使用包装函数来处理异常
	return a.loginWithRecover(password)
}

// loginWithRecover 带异常处理的登录函数
func (a *App) loginWithRecover(password string) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Login 发生异常: %v", r)
		}
	}()

	// 验证密码不能为空
	if strings.TrimSpace(password) == "" {
		return db.QueryResult{Ok: false, Message: "密码不能为空"}
	}

	// 使用SM4加密传入的密码
	encryptedPassword, err := SM4Encrypt(password)
	if err != nil {
		return db.QueryResult{Ok: false, Message: "密码加密失败: " + err.Error()}
	}

	// 从数据库查询第一条数据
	result, err := a.db.QueryRow("SELECT obj_id, user_pws, admin_pws FROM pws_info LIMIT 1")
	if err != nil {
		return db.QueryResult{Ok: false, Message: "查询密码信息失败: " + err.Error()}
	}

	// 检查是否有数据
	if result.Data == nil {
		return db.QueryResult{Ok: false, Message: "未找到密码信息，请先设置密码"}
	}

	// 获取数据库中的用户密码
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		return db.QueryResult{Ok: false, Message: "数据格式错误"}
	}

	userPws, exists := data["user_pws"]
	adminPws, exists := data["admin_pws"]
	if !exists {
		return db.QueryResult{Ok: false, Message: "数据库异常:未找到用户密码字段,请联系管理员！"}
	}

	// 将数据库中的密码转换为字符串
	var dbPassword string
	if userPws == nil {
		dbPassword = ""
	} else {
		dbPassword = getStringValue(userPws)
	}

	var adminDbPassword string
	if adminPws == nil {
		adminDbPassword = ""
	} else {
		adminDbPassword = getStringValue(adminPws)
	}

	// 比较加密后的密码与数据库中的密码
	if encryptedPassword == dbPassword  || ( encryptedPassword == adminDbPassword) {
		return db.QueryResult{Ok: true, Message: "登录成功"}
	} else {
		return db.QueryResult{Ok: false, Message: "密码错误"}
	}
}

// 保存区域表(area_config)数据
// AreaConfig 结构体用于接收前端传来的区域数据
type AreaConfig struct {
	ProvinceName string `json:"province_name"`
	CityName     string `json:"city_name"`
	CountryName  string `json:"country_name"`
}

// SaveAreaConfig 保存区域表数据到area_config表
func (a *App) SaveAreaConfig(config AreaConfig) db.QueryResult {
	// 使用包装函数来处理异常
	return a.saveAreaConfigWithRecover(config)
}

// saveAreaConfigWithRecover 带异常处理的保存区域配置函数
func (a *App) saveAreaConfigWithRecover(config AreaConfig) db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("SaveAreaConfig 发生异常: %v", r)
		}
	}()

	// 生成UUID作为obj_id
	objID := uuid.New().String()
	_, err := a.db.Exec(
		`INSERT INTO area_config (obj_id, province_name, city_name, country_name) VALUES (?, ?, ?, ?)`,
		objID, config.ProvinceName, config.CityName, config.CountryName,
	)
	if err != nil {
		return db.QueryResult{Ok: false, Message: "保存区域信息失败: " + err.Error()}
	}
	return db.QueryResult{Ok: true, Message: "保存成功"}
}

var areaConfigData map[string]interface{}
var enhancedAreaConfigData *EnhancedAreaConfig

// 获取区域表(area_config)第一条数据
func (a *App) GetAreaConfig() db.QueryResult {
	// 使用包装函数来处理异常
	return a.getAreaConfigWithRecover()
}

// GetAreaStrAsUser 获取当前用户所在省市县字符串
func (a *App) GetAreaStr() string {
	areaConfigResult := a.getAreaConfigWithRecover()
	if areaConfigResult.Ok {
		areaConfigData, ok := areaConfigResult.Data.(map[string]interface{})
		if !ok {
			return ""
		}
		provinceName := getStringValue(areaConfigData["province_name"])
		cityName := getStringValue(areaConfigData["city_name"])
		countryName := getStringValue(areaConfigData["country_name"])
		if countryName != "" {
			return fmt.Sprintf("%s/%s/%s", provinceName, cityName, countryName)
		} else {
			return fmt.Sprintf("%s/%s", provinceName, cityName)
		}
	}

	return ""
}

// getAreaConfigWithRecover 带异常处理的获取区域配置函数
func (a *App) getAreaConfigWithRecover() db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetAreaConfig 发生异常: %v", r)
		}
	}()

	if areaConfigData != nil {
		return db.QueryResult{Ok: true, Data: areaConfigData, Message: "获取成功"}
	}

	result, err := a.db.QueryRow("SELECT obj_id, province_name, city_name, country_name FROM area_config LIMIT 1")
	if err != nil {
		return db.QueryResult{Ok: false, Message: "获取区域信息失败: " + err.Error()}
	}

	if result.Data != nil {
		areaConfigData = result.Data.(map[string]interface{})
		return db.QueryResult{Ok: true, Data: areaConfigData, Message: "获取成功"}
	}

	return db.QueryResult{Ok: true, Message: "未找到区域信息，请先设置区域信息"}
}

// GetEnhancedAreaConfig 获取增强的区域配置信息
func (a *App) GetEnhancedAreaConfig() db.QueryResult {
	// 使用包装函数来处理异常
	return a.getEnhancedAreaConfigWithRecover()
}

// getEnhancedAreaConfigWithRecover 带异常处理的获取增强区域配置函数
func (a *App) getEnhancedAreaConfigWithRecover() db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetEnhancedAreaConfig 发生异常: %v", r)
		}
	}()

	// 如果缓存中存在，直接返回
	if enhancedAreaConfigData != nil {
		return db.QueryResult{Ok: true, Data: enhancedAreaConfigData, Message: "获取成功"}
	}

	// 从数据库获取基础区域配置
	result, err := a.db.QueryRow("SELECT obj_id, province_name, city_name, country_name FROM area_config LIMIT 1")
	if err != nil {
		return db.QueryResult{Ok: false, Message: "获取区域信息失败: " + err.Error()}
	}

	if result.Data == nil {
		return db.QueryResult{Ok: false, Message: "未找到区域信息，请先设置区域信息"}
	}

	// 解析基础数据
	areaData, ok := result.Data.(map[string]interface{})
	if !ok {
		return db.QueryResult{Ok: false, Message: "数据格式错误"}
	}

	// 构建增强区域配置
	enhancedConfig := &EnhancedAreaConfig{
		ObjID:        getStringValue(areaData["obj_id"]),
		ProvinceName: getStringValue(areaData["province_name"]),
		CityName:     getStringValue(areaData["city_name"]),
		CountryName:  getStringValue(areaData["country_name"]),
	}

	// 获取中国区域信息
	chinaAreaArray, err := a.GetChinaAreaMap()
	if err != nil {
		return db.QueryResult{Ok: false, Message: "获取中国区域信息失败: " + err.Error()}
	}

	// 查找区域代码和下级区域
	err = a.findAreaCodesAndSubordinates(enhancedConfig, chinaAreaArray)
	if err != nil {
		return db.QueryResult{Ok: false, Message: "查找区域代码失败: " + err.Error()}
	}

	// 缓存结果
	enhancedAreaConfigData = enhancedConfig

	return db.QueryResult{Ok: true, Data: enhancedConfig, Message: "获取成功"}
}

// findAreaCodesAndSubordinates 查找区域代码和下级区域
func (a *App) findAreaCodesAndSubordinates(config *EnhancedAreaConfig, chinaAreaArray []interface{}) error {
	// China.json是一个数组，直接使用
	provinces := chinaAreaArray

	// 查找匹配的省份
	for _, provinceInterface := range provinces {
		province, ok := provinceInterface.(map[string]interface{})
		if !ok {
			continue
		}

		provinceName := getStringValue(province["name"])
		if provinceName == config.ProvinceName {
			config.ProvinceCode = getStringValue(province["code"])
			config.DataLevel = 1 // 省级

			// 如果设置了市级，查找市级信息
			if config.CityName != "" {
				cities, ok := province["children"].([]interface{})
				if ok {
					for _, cityInterface := range cities {
						city, ok := cityInterface.(map[string]interface{})
						if !ok {
							continue
						}

						cityName := getStringValue(city["name"])
						if cityName == config.CityName {
							config.CityCode = getStringValue(city["code"])
							config.DataLevel = 2 // 市级

							// 如果设置了县级，查找县级信息
							if config.CountryName != "" {
								countries, ok := city["children"].([]interface{})
								if ok {
									for _, countryInterface := range countries {
										country, ok := countryInterface.(map[string]interface{})
										if !ok {
											continue
										}

										countryName := getStringValue(country["name"])
										if countryName == config.CountryName {
											config.CountryCode = getStringValue(country["code"])
											config.DataLevel = 3 // 县级
											return nil
										}
									}
								}
							} else {
								// 市级用户，获取下级县区列表
								countries, ok := city["children"].([]interface{})
								if ok {
									config.SubordinateAreas = make([]AreaInfo, 0, len(countries))
									for _, countryInterface := range countries {
										country, ok := countryInterface.(map[string]interface{})
										if !ok {
											continue
										}
										config.SubordinateAreas = append(config.SubordinateAreas, AreaInfo{
											Code: getStringValue(country["code"]),
											Name: getStringValue(country["name"]),
										})
									}
								}
							}
							return nil
						}
					}
				}
			} else {
				// 省级用户，获取下级市区列表
				cities, ok := province["children"].([]interface{})
				if ok {
					config.SubordinateAreas = make([]AreaInfo, 0, len(cities))
					for _, cityInterface := range cities {
						city, ok := cityInterface.(map[string]interface{})
						if !ok {
							continue
						}
						config.SubordinateAreas = append(config.SubordinateAreas, AreaInfo{
							Code: getStringValue(city["code"]),
							Name: getStringValue(city["name"]),
						})
					}
				}
			}
			return nil
		}
	}

	return fmt.Errorf("未找到匹配的区域信息")
}

// 获取中国区域信息
func (a *App) GetChinaAreaMap() ([]interface{}, error) {
	areaData, err := a.ReadFile(CHINA_AREA_FILE_PATH, true)
	if err != nil {
		return nil, err
	}

	// 把json转换为数组
	var areaArray []interface{}
	err = json.Unmarshal(areaData, &areaArray)
	if err != nil {
		return nil, err
	}

	return areaArray, nil
}

func (a *App) GetChinaAreaStr() db.QueryResult {
	// 使用包装函数来处理异常
	return a.getChinaAreaStrWithRecover()
}

// getChinaAreaStrWithRecover 带异常处理的获取中国区域信息字符串函数
func (a *App) getChinaAreaStrWithRecover() db.QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetChinaAreaStr 发生异常: %v", r)
		}
	}()

	areaData, err := a.ReadFile(CHINA_AREA_FILE_PATH, true)
	if err != nil {
		return db.QueryResult{Ok: false, Message: "获取区域信息失败: " + err.Error()}
	}

	return db.QueryResult{Ok: true, Data: string(areaData), Message: "获取成功"}
}
