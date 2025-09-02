package data_import

import (
	"fmt"
)

// YearlyAggregatedData 年份累计数据缓存结构
type YearlyAggregatedData struct {
	StatDate     string  // 年份
	TotalCoal    float64 // 煤合计累计
	RawCoal      float64 // 原煤累计
	WashedCoal   float64 // 洗精煤累计
	OtherCoal    float64 // 其他累计
	PowerGen     float64 // 火力发电累计
	Heating      float64 // 供热累计
	CoalWashing  float64 // 煤炭洗选累计
	Coking       float64 // 炼焦累计
	OilRefining  float64 // 炼油及煤制油累计
	GasProd      float64 // 制气累计
	Industry     float64 // 工业累计
	RawMaterials float64 // 工业（#用作原料、材料）累计
	OtherUses    float64 // 其他用途累计
	Coke         float64 // 焦炭累计
}

// CityData 当前市的数据缓存结构
type CityData struct {
	ProvinceName string  // 省份名称
	CityName     string  // 市名称
	StatDate     string  // 年份
	TotalCoal    float64 // 煤合计
	RawCoal      float64 // 原煤
	WashedCoal   float64 // 洗精煤
	OtherCoal    float64 // 其他
	PowerGen     float64 // 火力发电
	Heating      float64 // 供热
	CoalWashing  float64 // 煤炭洗选
	Coking       float64 // 炼焦
	OilRefining  float64 // 炼油及煤制油
	GasProd      float64 // 制气
	Industry     float64 // 工业
	RawMaterials float64 // 工业（#用作原料、材料）
	OtherUses    float64 // 其他用途
	Coke         float64 // 焦炭
}

// Attachment2CacheManager 附件2缓存管理器
type Attachment2CacheManager struct {
	service *DataImportService
}

// 全局缓存变量
var (
	// 优化缓存
	yearlyAggregatedCache     map[string]*YearlyAggregatedData // 年份累计数据缓存 key: statDate (下辖县区数据累加)
	cityDataCache             map[string]*CityData             // 当前市数据缓存 key: province_city_statDate (本市数据)
	importedDataCache         map[string]bool                  // 已导入数据缓存 key: province_city_country_statDate
	optimizedCacheInitialized bool                             // 优化缓存是否已初始化
)

// NewAttachment2CacheManager 创建附件2缓存管理器
func NewAttachment2CacheManager(service *DataImportService) *Attachment2CacheManager {
	return &Attachment2CacheManager{
		service: service,
	}
}

// InitOptimizedCache 初始化优化缓存（只初始化一次）
func (m *Attachment2CacheManager) InitOptimizedCache() {
	if !optimizedCacheInitialized {
		yearlyAggregatedCache = make(map[string]*YearlyAggregatedData)
		cityDataCache = make(map[string]*CityData)
		importedDataCache = make(map[string]bool)
		optimizedCacheInitialized = true
	}
}

// GetYearlyAggregatedData 获取年份累计数据
func (m *Attachment2CacheManager) GetYearlyAggregatedData(statDate string) (*YearlyAggregatedData, bool) {
	data, exists := yearlyAggregatedCache[statDate]
	return data, exists
}

// SetYearlyAggregatedData 设置年份累计数据
func (m *Attachment2CacheManager) SetYearlyAggregatedData(statDate string, data *YearlyAggregatedData) {
	m.InitOptimizedCache()
	yearlyAggregatedCache[statDate] = data
}

// GetCityData 获取当前市数据
func (m *Attachment2CacheManager) GetCityData(provinceName, cityName, statDate string) (*CityData, bool) {
	key := fmt.Sprintf("%s|%s|%s", provinceName, cityName, statDate)
	data, exists := cityDataCache[key]
	return data, exists
}

// SetCityData 设置当前市数据
func (m *Attachment2CacheManager) SetCityData(provinceName, cityName, statDate string, data *CityData) {
	key := fmt.Sprintf("%s|%s|%s", provinceName, cityName, statDate)
	cityDataCache[key] = data
}

// IsDataImported 检查数据是否已导入stat_date (1)
func (m *Attachment2CacheManager) IsDataImported(statDate, provinceName, cityName, countryName string) bool {
	key := fmt.Sprintf("%s|%s|%s|%s", provinceName, cityName, countryName, statDate)
	fmt.Println()
	return importedDataCache[key]
}

// MarkDataAsImported 标记数据为已导入
func (m *Attachment2CacheManager) MarkDataAsImported(statDate, provinceName, cityName, countryName string) {
	key := fmt.Sprintf("%s|%s|%s|%s", provinceName, cityName, countryName, statDate)
	importedDataCache[key] = true
}

// RemoveDataFromImported 从已导入缓存中移除数据
func (m *Attachment2CacheManager) RemoveDataFromImported(statDate, provinceName, cityName, countryName string) {
	m.InitOptimizedCache()
	key := fmt.Sprintf("%s|%s|%s|%s", provinceName, cityName, countryName, statDate)
	delete(importedDataCache, key)
}

// UpdateYearlyAggregatedData 更新年份累计数据（累加）
func (m *Attachment2CacheManager) UpdateYearlyAggregatedData(statDate string, data *YearlyAggregatedData) {
	existing, exists := yearlyAggregatedCache[statDate]
	if !exists {
		// 如果不存在，直接设置
		yearlyAggregatedCache[statDate] = data
		return
	}

	// 累加数据
	existing.TotalCoal = m.service.addFloat64(existing.TotalCoal, data.TotalCoal)
	existing.RawCoal = m.service.addFloat64(existing.RawCoal, data.RawCoal)
	existing.WashedCoal = m.service.addFloat64(existing.WashedCoal, data.WashedCoal)
	existing.OtherCoal = m.service.addFloat64(existing.OtherCoal, data.OtherCoal)
	existing.PowerGen = m.service.addFloat64(existing.PowerGen, data.PowerGen)
	existing.Heating = m.service.addFloat64(existing.Heating, data.Heating)
	existing.CoalWashing = m.service.addFloat64(existing.CoalWashing, data.CoalWashing)
	existing.Coking = m.service.addFloat64(existing.Coking, data.Coking)
	existing.OilRefining = m.service.addFloat64(existing.OilRefining, data.OilRefining)
	existing.GasProd = m.service.addFloat64(existing.GasProd, data.GasProd)
	existing.Industry = m.service.addFloat64(existing.Industry, data.Industry)
	existing.RawMaterials = m.service.addFloat64(existing.RawMaterials, data.RawMaterials)
	existing.OtherUses = m.service.addFloat64(existing.OtherUses, data.OtherUses)
	existing.Coke = m.service.addFloat64(existing.Coke, data.Coke)
}

// IsDataExistsInOptimizedCache 检查数据是否存在于优化缓存中
func (m *Attachment2CacheManager) IsDataExistsInOptimizedCache(statDate, provinceName, cityName, countryName string) bool {
	// 使用已导入数据缓存检查
	return m.IsDataImported(statDate, provinceName, cityName, countryName)
}

// ClearOptimizedCache 清除优化缓存
func (m *Attachment2CacheManager) ClearOptimizedCache() {
	yearlyAggregatedCache = nil
	cityDataCache = nil
	importedDataCache = nil
	optimizedCacheInitialized = false
}

// PreloadOptimizedCache 预加载优化缓存（从数据库加载所有相关数据到新结构）
func (m *Attachment2CacheManager) PreloadOptimizedCache() error {
	// 初始化缓存
	m.InitOptimizedCache()

	// 从数据库加载所有附件2数据
	query := `
		SELECT 
			stat_date, province_name, city_name, country_name,
			total_coal, raw_coal, washed_coal, other_coal,
			power_generation, heating, coal_washing, coking, oil_refining, gas_production,
			industry, raw_materials, other_uses, coke
		FROM coal_consumption_report group by stat_date, province_name, city_name, country_name
	`

	result, err := m.service.app.GetDB().Query(query)
	if err != nil || result.Data == nil {
		return nil // 如果没有数据，直接返回
	}

	data, ok := result.Data.([]map[string]interface{})
	if !ok || len(data) == 0 {
		return nil // 如果没有数据，直接返回
	}

	// 按年份和区域分组数据
	yearlyDataMap := make(map[string]*YearlyAggregatedData)
	cityDataMap := make(map[string]*CityData)

	for _, record := range data {
		statDate := m.service.getStringValue(record["stat_date"])
		provinceName := m.service.getStringValue(record["province_name"])
		cityName := m.service.getStringValue(record["city_name"])
		countryName := m.service.getStringValue(record["country_name"])

		// 解密数值字段
		totalCoal := m.service.parseFloat(m.service.decryptValue(record["total_coal"]))
		rawCoal := m.service.parseFloat(m.service.decryptValue(record["raw_coal"]))
		washedCoal := m.service.parseFloat(m.service.decryptValue(record["washed_coal"]))
		otherCoal := m.service.parseFloat(m.service.decryptValue(record["other_coal"]))
		powerGeneration := m.service.parseFloat(m.service.decryptValue(record["power_generation"]))
		heating := m.service.parseFloat(m.service.decryptValue(record["heating"]))
		coalWashing := m.service.parseFloat(m.service.decryptValue(record["coal_washing"]))
		coking := m.service.parseFloat(m.service.decryptValue(record["coking"]))
		oilRefining := m.service.parseFloat(m.service.decryptValue(record["oil_refining"]))
		gasProduction := m.service.parseFloat(m.service.decryptValue(record["gas_production"]))
		industry := m.service.parseFloat(m.service.decryptValue(record["industry"]))
		rawMaterials := m.service.parseFloat(m.service.decryptValue(record["raw_materials"]))
		otherUses := m.service.parseFloat(m.service.decryptValue(record["other_uses"]))
		coke := m.service.parseFloat(m.service.decryptValue(record["coke"]))

		// 1. 累加到年份累计数据（下辖县区数据累加，即countryName不为空的数据）
		if countryName != "" {
			if yearlyData, exists := yearlyDataMap[statDate]; exists {
				yearlyData.TotalCoal = m.service.addFloat64(yearlyData.TotalCoal, totalCoal)
				yearlyData.RawCoal = m.service.addFloat64(yearlyData.RawCoal, rawCoal)
				yearlyData.WashedCoal = m.service.addFloat64(yearlyData.WashedCoal, washedCoal)
				yearlyData.OtherCoal = m.service.addFloat64(yearlyData.OtherCoal, otherCoal)
				yearlyData.PowerGen = m.service.addFloat64(yearlyData.PowerGen, powerGeneration)
				yearlyData.Heating = m.service.addFloat64(yearlyData.Heating, heating)
				yearlyData.CoalWashing = m.service.addFloat64(yearlyData.CoalWashing, coalWashing)
				yearlyData.Coking = m.service.addFloat64(yearlyData.Coking, coking)
				yearlyData.OilRefining = m.service.addFloat64(yearlyData.OilRefining, oilRefining)
				yearlyData.GasProd = m.service.addFloat64(yearlyData.GasProd, gasProduction)
				yearlyData.Industry = m.service.addFloat64(yearlyData.Industry, industry)
				yearlyData.RawMaterials = m.service.addFloat64(yearlyData.RawMaterials, rawMaterials)
				yearlyData.OtherUses = m.service.addFloat64(yearlyData.OtherUses, otherUses)
				yearlyData.Coke = m.service.addFloat64(yearlyData.Coke, coke)
			} else {
				yearlyDataMap[statDate] = &YearlyAggregatedData{
					StatDate:     statDate,
					TotalCoal:    totalCoal,
					RawCoal:      rawCoal,
					WashedCoal:   washedCoal,
					OtherCoal:    otherCoal,
					PowerGen:     powerGeneration,
					Heating:      heating,
					CoalWashing:  coalWashing,
					Coking:       coking,
					OilRefining:  oilRefining,
					GasProd:      gasProduction,
					Industry:     industry,
					RawMaterials: rawMaterials,
					OtherUses:    otherUses,
					Coke:         coke,
				}
			}
		}

		// 2. 累加到市数据（本市数据，即countryName为空且cityName不为空的数据）
		if countryName == "" && cityName != "" {
			cityKey := fmt.Sprintf("%s|%s|%s", provinceName, cityName, statDate)
			if cityData, exists := cityDataMap[cityKey]; exists {
				cityData.TotalCoal = m.service.addFloat64(cityData.TotalCoal, totalCoal)
				cityData.RawCoal = m.service.addFloat64(cityData.RawCoal, rawCoal)
				cityData.WashedCoal = m.service.addFloat64(cityData.WashedCoal, washedCoal)
				cityData.OtherCoal = m.service.addFloat64(cityData.OtherCoal, otherCoal)
				cityData.PowerGen = m.service.addFloat64(cityData.PowerGen, powerGeneration)
				cityData.Heating = m.service.addFloat64(cityData.Heating, heating)
				cityData.CoalWashing = m.service.addFloat64(cityData.CoalWashing, coalWashing)
				cityData.Coking = m.service.addFloat64(cityData.Coking, coking)
				cityData.OilRefining = m.service.addFloat64(cityData.OilRefining, oilRefining)
				cityData.GasProd = m.service.addFloat64(cityData.GasProd, gasProduction)
				cityData.Industry = m.service.addFloat64(cityData.Industry, industry)
				cityData.RawMaterials = m.service.addFloat64(cityData.RawMaterials, rawMaterials)
				cityData.OtherUses = m.service.addFloat64(cityData.OtherUses, otherUses)
				cityData.Coke = m.service.addFloat64(cityData.Coke, coke)
			} else {
				cityDataMap[cityKey] = &CityData{
					ProvinceName: provinceName,
					CityName:     cityName,
					StatDate:     statDate,
					TotalCoal:    totalCoal,
					RawCoal:      rawCoal,
					WashedCoal:   washedCoal,
					OtherCoal:    otherCoal,
					PowerGen:     powerGeneration,
					Heating:      heating,
					CoalWashing:  coalWashing,
					Coking:       coking,
					OilRefining:  oilRefining,
					GasProd:      gasProduction,
					Industry:     industry,
					RawMaterials: rawMaterials,
					OtherUses:    otherUses,
					Coke:         coke,
				}
			}
		}

		// 3. 标记数据为已导入
		m.MarkDataAsImported(statDate, provinceName, cityName, countryName)
		fmt.Println(" 标记数据为已导入==, statDate==", statDate, "provinceName==", provinceName, "cityName==", cityName, "countryName==", countryName)
	}

	// 将数据设置到全局缓存中
	for statDate, yearlyData := range yearlyDataMap {
		yearlyAggregatedCache[statDate] = yearlyData
		fmt.Println("加载完成==, yearlyData==", yearlyData, "statDate==", statDate)
	}

	for cityKey, cityData := range cityDataMap {
		cityDataCache[cityKey] = cityData
		fmt.Println("加载完成==, cityData==", cityData, "statDate==", cityKey)
	}

	return nil
}
