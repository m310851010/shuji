package main

import (
	"encoding/json"
	"log"
)

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
