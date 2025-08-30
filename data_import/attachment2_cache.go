package data_import

import (
	"fmt"
	"math/big"
)

// Attachment2DatabaseCache 附件2数据库缓存结构
type Attachment2DatabaseCache struct {
	CacheKey     string // 缓存键：province_city_country_statDate
	TotalCoal    *big.Float
	RawCoal      *big.Float
	WashedCoal   *big.Float
	OtherCoal    *big.Float
	PowerGen     *big.Float
	Heating      *big.Float
	CoalWashing  *big.Float
	Coking       *big.Float
	OilRefining  *big.Float
	GasProd      *big.Float
	Industry     *big.Float
	RawMaterials *big.Float
	OtherUses    *big.Float
	Coke         *big.Float
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
		totalCoal := m.service.parseBigFloat(m.service.decryptValue(record["total_coal"]))
		rawCoal := m.service.parseBigFloat(m.service.decryptValue(record["raw_coal"]))
		washedCoal := m.service.parseBigFloat(m.service.decryptValue(record["washed_coal"]))
		otherCoal := m.service.parseBigFloat(m.service.decryptValue(record["other_coal"]))
		powerGeneration := m.service.parseBigFloat(m.service.decryptValue(record["power_generation"]))
		heating := m.service.parseBigFloat(m.service.decryptValue(record["heating"]))
		coalWashing := m.service.parseBigFloat(m.service.decryptValue(record["coal_washing"]))
		coking := m.service.parseBigFloat(m.service.decryptValue(record["coking"]))
		oilRefining := m.service.parseBigFloat(m.service.decryptValue(record["oil_refining"]))
		gasProduction := m.service.parseBigFloat(m.service.decryptValue(record["gas_production"]))
		industry := m.service.parseBigFloat(m.service.decryptValue(record["industry"]))
		rawMaterials := m.service.parseBigFloat(m.service.decryptValue(record["raw_materials"]))
		otherUses := m.service.parseBigFloat(m.service.decryptValue(record["other_uses"]))
		coke := m.service.parseBigFloat(m.service.decryptValue(record["coke"]))

		// 累加到缓存中
		if cache, exists := cacheMap[cacheKey]; exists {
			cache.TotalCoal.Add(cache.TotalCoal, totalCoal)
			cache.RawCoal.Add(cache.RawCoal, rawCoal)
			cache.WashedCoal.Add(cache.WashedCoal, washedCoal)
			cache.OtherCoal.Add(cache.OtherCoal, otherCoal)
			cache.PowerGen.Add(cache.PowerGen, powerGeneration)
			cache.Heating.Add(cache.Heating, heating)
			cache.CoalWashing.Add(cache.CoalWashing, coalWashing)
			cache.Coking.Add(cache.Coking, coking)
			cache.OilRefining.Add(cache.OilRefining, oilRefining)
			cache.GasProd.Add(cache.GasProd, gasProduction)
			cache.Industry.Add(cache.Industry, industry)
			cache.RawMaterials.Add(cache.RawMaterials, rawMaterials)
			cache.OtherUses.Add(cache.OtherUses, otherUses)
			cache.Coke.Add(cache.Coke, coke)
		} else {
			cacheMap[cacheKey] = &Attachment2DatabaseCache{
				CacheKey:     cacheKey,
				TotalCoal:    new(big.Float).Copy(totalCoal),
				RawCoal:      new(big.Float).Copy(rawCoal),
				WashedCoal:   new(big.Float).Copy(washedCoal),
				OtherCoal:    new(big.Float).Copy(otherCoal),
				PowerGen:     new(big.Float).Copy(powerGeneration),
				Heating:      new(big.Float).Copy(heating),
				CoalWashing:  new(big.Float).Copy(coalWashing),
				Coking:       new(big.Float).Copy(coking),
				OilRefining:  new(big.Float).Copy(oilRefining),
				GasProd:      new(big.Float).Copy(gasProduction),
				Industry:     new(big.Float).Copy(industry),
				RawMaterials: new(big.Float).Copy(rawMaterials),
				OtherUses:    new(big.Float).Copy(otherUses),
				Coke:         new(big.Float).Copy(coke),
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
	totalCoal := m.service.parseBigFloat(m.service.getStringValue(record["total_coal"]))
	rawCoal := m.service.parseBigFloat(m.service.getStringValue(record["raw_coal"]))
	washedCoal := m.service.parseBigFloat(m.service.getStringValue(record["washed_coal"]))
	otherCoal := m.service.parseBigFloat(m.service.getStringValue(record["other_coal"]))
	powerGeneration := m.service.parseBigFloat(m.service.getStringValue(record["power_generation"]))
	heating := m.service.parseBigFloat(m.service.getStringValue(record["heating"]))
	coalWashing := m.service.parseBigFloat(m.service.getStringValue(record["coal_washing"]))
	coking := m.service.parseBigFloat(m.service.getStringValue(record["coking"]))
	oilRefining := m.service.parseBigFloat(m.service.getStringValue(record["oil_refining"]))
	gasProduction := m.service.parseBigFloat(m.service.getStringValue(record["gas_production"]))
	industry := m.service.parseBigFloat(m.service.getStringValue(record["industry"]))
	rawMaterials := m.service.parseBigFloat(m.service.getStringValue(record["raw_materials"]))
	otherUses := m.service.parseBigFloat(m.service.getStringValue(record["other_uses"]))
	coke := m.service.parseBigFloat(m.service.getStringValue(record["coke"]))

	// 更新所有相关缓存
	for _, cacheKey := range cacheKeys {
		if cache, exists := attachment2DatabaseCache[cacheKey]; exists {
					// 累加到现有缓存
		cache.TotalCoal.Add(cache.TotalCoal, totalCoal)
		cache.RawCoal.Add(cache.RawCoal, rawCoal)
		cache.WashedCoal.Add(cache.WashedCoal, washedCoal)
		cache.OtherCoal.Add(cache.OtherCoal, otherCoal)
		cache.PowerGen.Add(cache.PowerGen, powerGeneration)
		cache.Heating.Add(cache.Heating, heating)
		cache.CoalWashing.Add(cache.CoalWashing, coalWashing)
		cache.Coking.Add(cache.Coking, coking)
		cache.OilRefining.Add(cache.OilRefining, oilRefining)
		cache.GasProd.Add(cache.GasProd, gasProduction)
		cache.Industry.Add(cache.Industry, industry)
		cache.RawMaterials.Add(cache.RawMaterials, rawMaterials)
		cache.OtherUses.Add(cache.OtherUses, otherUses)
		cache.Coke.Add(cache.Coke, coke)
		} else {
			// 创建新缓存
			attachment2DatabaseCache[cacheKey] = &Attachment2DatabaseCache{
				CacheKey:     cacheKey,
				TotalCoal:    new(big.Float).Copy(totalCoal),
				RawCoal:      new(big.Float).Copy(rawCoal),
				WashedCoal:   new(big.Float).Copy(washedCoal),
				OtherCoal:    new(big.Float).Copy(otherCoal),
				PowerGen:     new(big.Float).Copy(powerGeneration),
				Heating:      new(big.Float).Copy(heating),
				CoalWashing:  new(big.Float).Copy(coalWashing),
				Coking:       new(big.Float).Copy(coking),
				OilRefining:  new(big.Float).Copy(oilRefining),
				GasProd:      new(big.Float).Copy(gasProduction),
				Industry:     new(big.Float).Copy(industry),
				RawMaterials: new(big.Float).Copy(rawMaterials),
				OtherUses:    new(big.Float).Copy(otherUses),
				Coke:         new(big.Float).Copy(coke),
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
	oldTotalCoal := m.service.parseBigFloat(m.service.getStringValue(oldRecord["total_coal"]))
	oldRawCoal := m.service.parseBigFloat(m.service.getStringValue(oldRecord["raw_coal"]))
	oldWashedCoal := m.service.parseBigFloat(m.service.getStringValue(oldRecord["washed_coal"]))
	oldOtherCoal := m.service.parseBigFloat(m.service.getStringValue(oldRecord["other_coal"]))
	oldPowerGeneration := m.service.parseBigFloat(m.service.getStringValue(oldRecord["power_generation"]))
	oldHeating := m.service.parseBigFloat(m.service.getStringValue(oldRecord["heating"]))
	oldCoalWashing := m.service.parseBigFloat(m.service.getStringValue(oldRecord["coal_washing"]))
	oldCoking := m.service.parseBigFloat(m.service.getStringValue(oldRecord["coking"]))
	oldOilRefining := m.service.parseBigFloat(m.service.getStringValue(oldRecord["oil_refining"]))
	oldGasProduction := m.service.parseBigFloat(m.service.getStringValue(oldRecord["gas_production"]))
	oldIndustry := m.service.parseBigFloat(m.service.getStringValue(oldRecord["industry"]))
	oldRawMaterials := m.service.parseBigFloat(m.service.getStringValue(oldRecord["raw_materials"]))
	oldOtherUses := m.service.parseBigFloat(m.service.getStringValue(oldRecord["other_uses"]))
	oldCoke := m.service.parseBigFloat(m.service.getStringValue(oldRecord["coke"]))

	newTotalCoal := m.service.parseBigFloat(m.service.getStringValue(newRecord["total_coal"]))
	newRawCoal := m.service.parseBigFloat(m.service.getStringValue(newRecord["raw_coal"]))
	newWashedCoal := m.service.parseBigFloat(m.service.getStringValue(newRecord["washed_coal"]))
	newOtherCoal := m.service.parseBigFloat(m.service.getStringValue(newRecord["other_coal"]))
	newPowerGeneration := m.service.parseBigFloat(m.service.getStringValue(newRecord["power_generation"]))
	newHeating := m.service.parseBigFloat(m.service.getStringValue(newRecord["heating"]))
	newCoalWashing := m.service.parseBigFloat(m.service.getStringValue(newRecord["coal_washing"]))
	newCoking := m.service.parseBigFloat(m.service.getStringValue(newRecord["coking"]))
	newOilRefining := m.service.parseBigFloat(m.service.getStringValue(newRecord["oil_refining"]))
	newGasProduction := m.service.parseBigFloat(m.service.getStringValue(newRecord["gas_production"]))
	newIndustry := m.service.parseBigFloat(m.service.getStringValue(newRecord["industry"]))
	newRawMaterials := m.service.parseBigFloat(m.service.getStringValue(newRecord["raw_materials"]))
	newOtherUses := m.service.parseBigFloat(m.service.getStringValue(newRecord["other_uses"]))
	newCoke := m.service.parseBigFloat(m.service.getStringValue(newRecord["coke"]))

	// 计算差值
	deltaTotalCoal := new(big.Float).Sub(newTotalCoal, oldTotalCoal)
	deltaRawCoal := new(big.Float).Sub(newRawCoal, oldRawCoal)
	deltaWashedCoal := new(big.Float).Sub(newWashedCoal, oldWashedCoal)
	deltaOtherCoal := new(big.Float).Sub(newOtherCoal, oldOtherCoal)
	deltaPowerGeneration := new(big.Float).Sub(newPowerGeneration, oldPowerGeneration)
	deltaHeating := new(big.Float).Sub(newHeating, oldHeating)
	deltaCoalWashing := new(big.Float).Sub(newCoalWashing, oldCoalWashing)
	deltaCoking := new(big.Float).Sub(newCoking, oldCoking)
	deltaOilRefining := new(big.Float).Sub(newOilRefining, oldOilRefining)
	deltaGasProduction := new(big.Float).Sub(newGasProduction, oldGasProduction)
	deltaIndustry := new(big.Float).Sub(newIndustry, oldIndustry)
	deltaRawMaterials := new(big.Float).Sub(newRawMaterials, oldRawMaterials)
	deltaOtherUses := new(big.Float).Sub(newOtherUses, oldOtherUses)
	deltaCoke := new(big.Float).Sub(newCoke, oldCoke)

	// 更新所有相关缓存
	for _, cacheKey := range cacheKeys {
		if cache, exists := attachment2DatabaseCache[cacheKey]; exists {
					// 累加差值到现有缓存
		cache.TotalCoal.Add(cache.TotalCoal, deltaTotalCoal)
		cache.RawCoal.Add(cache.RawCoal, deltaRawCoal)
		cache.WashedCoal.Add(cache.WashedCoal, deltaWashedCoal)
		cache.OtherCoal.Add(cache.OtherCoal, deltaOtherCoal)
		cache.PowerGen.Add(cache.PowerGen, deltaPowerGeneration)
		cache.Heating.Add(cache.Heating, deltaHeating)
		cache.CoalWashing.Add(cache.CoalWashing, deltaCoalWashing)
		cache.Coking.Add(cache.Coking, deltaCoking)
		cache.OilRefining.Add(cache.OilRefining, deltaOilRefining)
		cache.GasProd.Add(cache.GasProd, deltaGasProduction)
		cache.Industry.Add(cache.Industry, deltaIndustry)
		cache.RawMaterials.Add(cache.RawMaterials, deltaRawMaterials)
		cache.OtherUses.Add(cache.OtherUses, deltaOtherUses)
		cache.Coke.Add(cache.Coke, deltaCoke)
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
