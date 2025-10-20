package main

import (
	"encoding/json"
	"path/filepath"
	"log"
	"os"
)

// 定义区域数据结构
type TreeNode struct {
	Code     string     `json:"code"`
	Name     string     `json:"name"`
	Children []TreeNode `json:"children"`
}

var CHINA_AREA_MAPPING_CACHE map[string]map[string]map[string]bool // 中国区域映射缓存
var TRADE_MAPPING_CACHE map[string]map[string]map[string]bool // 行业门类映射缓存
var APP_CONFIG_CACHE *AppConfig // 应用配置缓存

// 获取中国区域信息 - 返回3级联动map结构（省->市->区县）
func (a *App) GetChinaAreaMap() (map[string]map[string]map[string]bool, error) {
	if CHINA_AREA_MAPPING_CACHE != nil {
		return CHINA_AREA_MAPPING_CACHE, nil
	}

	areaData, err := a.ReadFile(CHINA_AREA_FILE_PATH, true)
	if err != nil {
		return nil, err
	}

	// 解析为结构化数据
	var provinces []TreeNode
	err = json.Unmarshal(areaData, &provinces)
	if err != nil {
		return nil, err
	}

	CHINA_AREA_MAPPING_CACHE = toTreeMap(provinces)
	return CHINA_AREA_MAPPING_CACHE, nil
}

// 行业门类映射信息 - 返回3级联动map结构（行业门类->行业大类->行业中类）
func (a *App) GetTradeMapping() (map[string]map[string]map[string]bool, error) {
	if TRADE_MAPPING_CACHE != nil {
		return TRADE_MAPPING_CACHE, nil
	}

	tradeData, err := a.ReadFile(TRADE_MAPPING_FILE_PATH, true)
	if err != nil {
		return nil, err
	}

	// 解析为结构化数据
	var trades []TreeNode
	err = json.Unmarshal(tradeData, &trades)
	if err != nil {
		return nil, err
	}

	TRADE_MAPPING_CACHE = toTreeMapHangye(trades)
	return TRADE_MAPPING_CACHE, nil
}

// 获取应用配置信息
func (a *App) GetAppConfig() (AppConfig, error) {
	if APP_CONFIG_CACHE != nil {
		return *APP_CONFIG_CACHE, nil
	}
	
	configPath := filepath.Join(Env.BasePath, DATA_DIR_NAME, CONFIG_FILE_NAME)
	configData, err := a.ReadFile(configPath, false)
	if err != nil {
		return AppConfig{}, err
	}
	
	// 解析为结构化数据
	err = json.Unmarshal(configData, &APP_CONFIG_CACHE)
	if err != nil {
		return AppConfig{}, err
	}
	
	return *APP_CONFIG_CACHE, nil
}

// 写入应用配置信息
func (a *App) WriteAppConfig(appConfig AppConfig) QueryResult {
	configPath := filepath.Join(Env.BasePath, DATA_DIR_NAME, CONFIG_FILE_NAME)
	data, err := json.MarshalIndent(appConfig, "", "  ")
	if err != nil {
		return QueryResult{Ok: false, Message: "序列化配置文件失败: " + err.Error()}
	}
		
	if err := os.WriteFile(configPath, data, os.ModePerm); err != nil {
		return QueryResult{Ok: false, Message: "写入配置文件失败: " + err.Error()}
	}

	APP_CONFIG_CACHE = &appConfig
	return QueryResult{Ok: true, Message: "写入配置文件成功"}
}

func (a *App) GetChinaAreaStr() QueryResult {
	// 使用包装函数来处理异常
	return a.getChinaAreaStrWithRecover()
}

// getChinaAreaStrWithRecover 带异常处理的获取中国区域信息字符串函数
func (a *App) getChinaAreaStrWithRecover() QueryResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GetChinaAreaStr 发生异常: %v", r)
		}
	}()

	areaData, err := a.ReadFile(CHINA_AREA_FILE_PATH, true)
	if err != nil {
		return QueryResult{Ok: false, Message: "获取区域信息失败: " + err.Error()}
	}

	return QueryResult{Ok: true, Data: string(areaData), Message: "获取成功"}
}
