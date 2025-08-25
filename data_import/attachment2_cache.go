package data_import

import (
	"fmt"
)

// Attachment2DatabaseCache 附件2数据库缓存结构
type Attachment2DatabaseCache struct {
	CacheKey     string // 缓存键：province_city_country_statDate
	TotalCoal    float64
	RawCoal      float64
	WashedCoal   float64
	OtherCoal    float64
	PowerGen     float64
	Heating      float64
	CoalWashing  float64
	Coking       float64
	OilRefining  float64
	GasProd      float64
	Industry     float64
	RawMaterials float64
	OtherUses    float64
	Coke         float64
}

// Attachment2CacheManager 附件2缓存管理器
type Attachment2CacheManager struct {
	service *DataImportService
}

// 全局缓存变量
var (
	attachment2DatabaseCache    map[string]*Attachment2DatabaseCache
	attachment2DataExistsCache  map[string]bool
	attachment2CacheInitialized bool
)

// NewAttachment2CacheManager 创建附件2缓存管理器
func NewAttachment2CacheManager(service *DataImportService) *Attachment2CacheManager {
	return &Attachment2CacheManager{
		service: service,
	}
}

// InitDatabaseCache 初始化附件2数据库缓存
func (m *Attachment2CacheManager) InitDatabaseCache() {
	if attachment2DatabaseCache == nil {
		attachment2DatabaseCache = make(map[string]*Attachment2DatabaseCache)
	}
}

// InitDataExistsCache 初始化附件2数据存在性缓存
func (m *Attachment2CacheManager) InitDataExistsCache() {
	if attachment2DataExistsCache == nil {
		attachment2DataExistsCache = make(map[string]bool)
	}
}

// GetDatabaseCacheKey 生成数据库缓存键
func (m *Attachment2CacheManager) GetDatabaseCacheKey(provinceName, cityName, countryName, statDate string) string {
	return fmt.Sprintf("%s_%s_%s_%s", provinceName, cityName, countryName, statDate)
}

// GetDataExistsCacheKey 生成数据存在性缓存键
func (m *Attachment2CacheManager) GetDataExistsCacheKey(statDate, provinceName, cityName, countryName string) string {
	return fmt.Sprintf("exists_%s_%s_%s_%s", statDate, provinceName, cityName, countryName)
}

// PreloadDatabaseCache 预加载附件2数据库缓存（一次性加载所有相关数据）
func (m *Attachment2CacheManager) PreloadDatabaseCache() error {
	if attachment2CacheInitialized {
		return nil // 缓存已初始化，无需重复加载
	}

	m.InitDatabaseCache()

	// 从数据库加载所有附件2数据
	query := `
		SELECT 
			stat_date, province_name, city_name, country_name,
			total_coal, raw_coal, washed_coal, other_coal,
			power_generation, heating, coal_washing, coking, oil_refining, gas_production,
			industry, raw_materials, other_uses, coke
		FROM coal_consumption_report 
		WHERE is_confirm = 1 AND is_check = 1
	`

	result, err := m.service.app.GetDB().Query(query)
	if err != nil || result.Data == nil {
		// 如果没有数据，标记为已初始化并返回
		attachment2CacheInitialized = true
		return nil
	}

	data, ok := result.Data.([]map[string]interface{})
	if !ok || len(data) == 0 {
		// 如果没有数据，标记为已初始化并返回
		attachment2CacheInitialized = true
		return nil
	}

	// 按地区和时间分组数据
	cacheMap := make(map[string]*Attachment2DatabaseCache)

	for _, record := range data {
		statDate := m.service.getStringValue(record["stat_date"])
		provinceName := m.service.getStringValue(record["province_name"])
		cityName := m.service.getStringValue(record["city_name"])
		countryName := m.service.getStringValue(record["country_name"])

		// 生成缓存键
		cacheKey := m.GetDatabaseCacheKey(provinceName, cityName, countryName, statDate)

		// 解密数值字段
		totalCoal, _ := m.service.parseFloat(m.service.decryptValue(record["total_coal"]))
		rawCoal, _ := m.service.parseFloat(m.service.decryptValue(record["raw_coal"]))
		washedCoal, _ := m.service.parseFloat(m.service.decryptValue(record["washed_coal"]))
		otherCoal, _ := m.service.parseFloat(m.service.decryptValue(record["other_coal"]))
		powerGeneration, _ := m.service.parseFloat(m.service.decryptValue(record["power_generation"]))
		heating, _ := m.service.parseFloat(m.service.decryptValue(record["heating"]))
		coalWashing, _ := m.service.parseFloat(m.service.decryptValue(record["coal_washing"]))
		coking, _ := m.service.parseFloat(m.service.decryptValue(record["coking"]))
		oilRefining, _ := m.service.parseFloat(m.service.decryptValue(record["oil_refining"]))
		gasProduction, _ := m.service.parseFloat(m.service.decryptValue(record["gas_production"]))
		industry, _ := m.service.parseFloat(m.service.decryptValue(record["industry"]))
		rawMaterials, _ := m.service.parseFloat(m.service.decryptValue(record["raw_materials"]))
		otherUses, _ := m.service.parseFloat(m.service.decryptValue(record["other_uses"]))
		coke, _ := m.service.parseFloat(m.service.decryptValue(record["coke"]))

		// 累加到缓存中
		if cache, exists := cacheMap[cacheKey]; exists {
			cache.TotalCoal += totalCoal
			cache.RawCoal += rawCoal
			cache.WashedCoal += washedCoal
			cache.OtherCoal += otherCoal
			cache.PowerGen += powerGeneration
			cache.Heating += heating
			cache.CoalWashing += coalWashing
			cache.Coking += coking
			cache.OilRefining += oilRefining
			cache.GasProd += gasProduction
			cache.Industry += industry
			cache.RawMaterials += rawMaterials
			cache.OtherUses += otherUses
			cache.Coke += coke
		} else {
			cacheMap[cacheKey] = &Attachment2DatabaseCache{
				CacheKey:     cacheKey,
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

		// 同时更新数据存在性缓存
		existsCacheKey := m.GetDataExistsCacheKey(statDate, provinceName, cityName, countryName)
		attachment2DataExistsCache[existsCacheKey] = true
	}

	// 将缓存数据存储到全局缓存中
	for cacheKey, cache := range cacheMap {
		attachment2DatabaseCache[cacheKey] = cache
	}

	// 标记缓存已初始化
	attachment2CacheInitialized = true
	return nil
}

// UpdateDatabaseCache 更新附件2数据库缓存
func (m *Attachment2CacheManager) UpdateDatabaseCache(statDate, provinceName, cityName, countryName string, record map[string]interface{}) {
	m.InitDatabaseCache()

	// 获取当前记录的所有上级区域缓存键
	cacheKeys := m.getParentCacheKeys(provinceName, cityName, countryName, statDate)

	// 解析当前记录的数值
	totalCoal, _ := m.service.parseFloat(m.service.getStringValue(record["total_coal"]))
	rawCoal, _ := m.service.parseFloat(m.service.getStringValue(record["raw_coal"]))
	washedCoal, _ := m.service.parseFloat(m.service.getStringValue(record["washed_coal"]))
	otherCoal, _ := m.service.parseFloat(m.service.getStringValue(record["other_coal"]))
	powerGeneration, _ := m.service.parseFloat(m.service.getStringValue(record["power_generation"]))
	heating, _ := m.service.parseFloat(m.service.getStringValue(record["heating"]))
	coalWashing, _ := m.service.parseFloat(m.service.getStringValue(record["coal_washing"]))
	coking, _ := m.service.parseFloat(m.service.getStringValue(record["coking"]))
	oilRefining, _ := m.service.parseFloat(m.service.getStringValue(record["oil_refining"]))
	gasProduction, _ := m.service.parseFloat(m.service.getStringValue(record["gas_production"]))
	industry, _ := m.service.parseFloat(m.service.getStringValue(record["industry"]))
	rawMaterials, _ := m.service.parseFloat(m.service.getStringValue(record["raw_materials"]))
	otherUses, _ := m.service.parseFloat(m.service.getStringValue(record["other_uses"]))
	coke, _ := m.service.parseFloat(m.service.getStringValue(record["coke"]))

	// 更新所有相关缓存
	for _, cacheKey := range cacheKeys {
		if cache, exists := attachment2DatabaseCache[cacheKey]; exists {
			// 累加到现有缓存
			cache.TotalCoal += totalCoal
			cache.RawCoal += rawCoal
			cache.WashedCoal += washedCoal
			cache.OtherCoal += otherCoal
			cache.PowerGen += powerGeneration
			cache.Heating += heating
			cache.CoalWashing += coalWashing
			cache.Coking += coking
			cache.OilRefining += oilRefining
			cache.GasProd += gasProduction
			cache.Industry += industry
			cache.RawMaterials += rawMaterials
			cache.OtherUses += otherUses
			cache.Coke += coke
		} else {
			// 创建新缓存
			attachment2DatabaseCache[cacheKey] = &Attachment2DatabaseCache{
				CacheKey:     cacheKey,
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
}

// UpdateDatabaseCacheForUpdate 更新附件2数据库缓存（用于UPDATE操作）
func (m *Attachment2CacheManager) UpdateDatabaseCacheForUpdate(statDate, provinceName, cityName, countryName string, oldRecord, newRecord map[string]interface{}) {
	m.InitDatabaseCache()

	// 获取当前记录的所有上级区域缓存键
	cacheKeys := m.getParentCacheKeys(provinceName, cityName, countryName, statDate)

	// 解析旧数据和新数据的数值
	oldTotalCoal, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["total_coal"]))
	oldRawCoal, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["raw_coal"]))
	oldWashedCoal, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["washed_coal"]))
	oldOtherCoal, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["other_coal"]))
	oldPowerGeneration, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["power_generation"]))
	oldHeating, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["heating"]))
	oldCoalWashing, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["coal_washing"]))
	oldCoking, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["coking"]))
	oldOilRefining, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["oil_refining"]))
	oldGasProduction, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["gas_production"]))
	oldIndustry, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["industry"]))
	oldRawMaterials, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["raw_materials"]))
	oldOtherUses, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["other_uses"]))
	oldCoke, _ := m.service.parseFloat(m.service.getStringValue(oldRecord["coke"]))

	newTotalCoal, _ := m.service.parseFloat(m.service.getStringValue(newRecord["total_coal"]))
	newRawCoal, _ := m.service.parseFloat(m.service.getStringValue(newRecord["raw_coal"]))
	newWashedCoal, _ := m.service.parseFloat(m.service.getStringValue(newRecord["washed_coal"]))
	newOtherCoal, _ := m.service.parseFloat(m.service.getStringValue(newRecord["other_coal"]))
	newPowerGeneration, _ := m.service.parseFloat(m.service.getStringValue(newRecord["power_generation"]))
	newHeating, _ := m.service.parseFloat(m.service.getStringValue(newRecord["heating"]))
	newCoalWashing, _ := m.service.parseFloat(m.service.getStringValue(newRecord["coal_washing"]))
	newCoking, _ := m.service.parseFloat(m.service.getStringValue(newRecord["coking"]))
	newOilRefining, _ := m.service.parseFloat(m.service.getStringValue(newRecord["oil_refining"]))
	newGasProduction, _ := m.service.parseFloat(m.service.getStringValue(newRecord["gas_production"]))
	newIndustry, _ := m.service.parseFloat(m.service.getStringValue(newRecord["industry"]))
	newRawMaterials, _ := m.service.parseFloat(m.service.getStringValue(newRecord["raw_materials"]))
	newOtherUses, _ := m.service.parseFloat(m.service.getStringValue(newRecord["other_uses"]))
	newCoke, _ := m.service.parseFloat(m.service.getStringValue(newRecord["coke"]))

	// 计算差值
	deltaTotalCoal := newTotalCoal - oldTotalCoal
	deltaRawCoal := newRawCoal - oldRawCoal
	deltaWashedCoal := newWashedCoal - oldWashedCoal
	deltaOtherCoal := newOtherCoal - oldOtherCoal
	deltaPowerGeneration := newPowerGeneration - oldPowerGeneration
	deltaHeating := newHeating - oldHeating
	deltaCoalWashing := newCoalWashing - oldCoalWashing
	deltaCoking := newCoking - oldCoking
	deltaOilRefining := newOilRefining - oldOilRefining
	deltaGasProduction := newGasProduction - oldGasProduction
	deltaIndustry := newIndustry - oldIndustry
	deltaRawMaterials := newRawMaterials - oldRawMaterials
	deltaOtherUses := newOtherUses - oldOtherUses
	deltaCoke := newCoke - oldCoke

	// 更新所有相关缓存
	for _, cacheKey := range cacheKeys {
		if cache, exists := attachment2DatabaseCache[cacheKey]; exists {
			// 累加差值到现有缓存
			cache.TotalCoal += deltaTotalCoal
			cache.RawCoal += deltaRawCoal
			cache.WashedCoal += deltaWashedCoal
			cache.OtherCoal += deltaOtherCoal
			cache.PowerGen += deltaPowerGeneration
			cache.Heating += deltaHeating
			cache.CoalWashing += deltaCoalWashing
			cache.Coking += deltaCoking
			cache.OilRefining += deltaOilRefining
			cache.GasProd += deltaGasProduction
			cache.Industry += deltaIndustry
			cache.RawMaterials += deltaRawMaterials
			cache.OtherUses += deltaOtherUses
			cache.Coke += deltaCoke
		}
	}
}

// getParentCacheKeys 获取所有上级区域的缓存键
func (m *Attachment2CacheManager) getParentCacheKeys(provinceName, cityName, countryName, statDate string) []string {
	var cacheKeys []string

	// 如果有县，需要更新省、市、县的缓存
	if countryName != "" {
		// 省级缓存
		cacheKeys = append(cacheKeys, m.GetDatabaseCacheKey(provinceName, "", "", statDate))
		// 市级缓存
		cacheKeys = append(cacheKeys, m.GetDatabaseCacheKey(provinceName, cityName, "", statDate))
		// 县级缓存
		cacheKeys = append(cacheKeys, m.GetDatabaseCacheKey(provinceName, cityName, countryName, statDate))
	} else if cityName != "" {
		// 如果有市无县，需要更新省、市的缓存
		// 省级缓存
		cacheKeys = append(cacheKeys, m.GetDatabaseCacheKey(provinceName, "", "", statDate))
		// 市级缓存
		cacheKeys = append(cacheKeys, m.GetDatabaseCacheKey(provinceName, cityName, "", statDate))
	} else if provinceName != "" {
		// 如果只有省，只需要更新省级缓存
		cacheKeys = append(cacheKeys, m.GetDatabaseCacheKey(provinceName, "", "", statDate))
	}

	return cacheKeys
}

// IsDataExistsInCache 检查附件2数据是否存在于缓存中
func (m *Attachment2CacheManager) IsDataExistsInCache(statDate, provinceName, cityName, countryName string) bool {
	m.InitDataExistsCache()
	cacheKey := m.GetDataExistsCacheKey(statDate, provinceName, cityName, countryName)
	exists, found := attachment2DataExistsCache[cacheKey]
	return found && exists
}

// CacheDataExists 缓存附件2数据存在性
func (m *Attachment2CacheManager) CacheDataExists(statDate, provinceName, cityName, countryName string, exists bool) {
	m.InitDataExistsCache()
	cacheKey := m.GetDataExistsCacheKey(statDate, provinceName, cityName, countryName)
	attachment2DataExistsCache[cacheKey] = exists
}

// GetDatabaseCache 获取数据库缓存
func (m *Attachment2CacheManager) GetDatabaseCache(cacheKey string) (*Attachment2DatabaseCache, bool) {
	cache, exists := attachment2DatabaseCache[cacheKey]
	return cache, exists
}

// ClearCache 清除附件2数据库缓存
func (m *Attachment2CacheManager) ClearCache() {
	attachment2DatabaseCache = nil
	attachment2DataExistsCache = nil
	attachment2CacheInitialized = false
}
