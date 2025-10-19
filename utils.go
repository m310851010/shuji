package main

import (
	"time"
	"fmt"
	"regexp"
	"io"
	"os"
	"os/user"
	"strings"
	"path/filepath"
	"strconv"
	"github.com/xuri/excelize/v2"
)


func GetPath(path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(Env.BasePath, path)
	}
	return filepath.Clean(path)
}

// GetCurrentOSUser 获取当前操作系统登录账号
func GetCurrentOSUser() string {
	currentUser, err := user.Current()
	if err != nil {
		return "系统"
	}
	return currentUser.Username
}

// copyCacheFile 带异常处理的复制缓存文件函数
func CopyCacheFile(filePath string, tableType string) (string, error) {

	fileName := filepath.Base(filePath)
	cachePath := GetPath(filepath.Join(CACHE_FILE_DIR_NAME, tableType, fileName))

	// 检查缓存目录是否存在, 不存在则创建
	cacheDir := filepath.Join(Env.BasePath, CACHE_FILE_DIR_NAME, tableType)
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		os.MkdirAll(cacheDir, os.ModePerm)
	}

	if err := copyFile(filePath, cachePath); err != nil {
		return "", err
	}

	return cachePath, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// getStringValue 获取字符串值的通用函数
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return value.(string)
}

// parseDateValue 解析日期值
func  parseDateValue(excelDate string) (time.Time, error) {
	baseDate:= time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	days, err := strconv.Atoi(excelDate)
	if err != nil {
		return time.Time{}, err
	}
	return baseDate.Add(time.Second * time.Duration(days * 24*60*60)), nil
}

// parseDateValueToString 解析日期值为字符串
func  parseDateValueToString(excelDate string, format string) string {
	date, err := parseDateValue(excelDate)
	if err != nil {
		return excelDate
	}

	if format == "" ||  format == "2006-01-02" {
		return date.Format("2006-01-02")
	}
	return date.Format(format)
}


// parseFloat 解析浮点数
func  parseFloat(value string) float64 {
	if value == "" {
		return 0
	}

	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return result
}

// tryParseFloat 尝试解析浮点数
func  tryParseFloat(value string) (float64, error) {
	if value == "" {
        return 0, nil
    }
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}


// isIntegerEqual 使用整数计算判断两个float64是否相等（乘以1000转换为整数计算）
func  isIntegerEqual(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA == intB
}

// isIntegerLessThan 使用整数计算判断a是否小于b（乘以1000转换为整数计算）
func  isIntegerLessThan(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA < intB
}

// isIntegerGreaterThan 使用整数计算判断a是否大于b（乘以1000转换为整数计算）
func  isIntegerGreaterThan(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA > intB
}

// isIntegerLessThanOrEqual 使用整数计算判断a是否小于等于b（乘以1000转换为整数计算）
func  isIntegerLessThanOrEqual(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA <= intB
}

// isIntegerGreaterThanOrEqual 使用整数计算判断a是否大于等于b（乘以1000转换为整数计算）
func  isIntegerGreaterThanOrEqual(a, b float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	return intA >= intB
}

// isIntegerInteger 使用整数计算判断float64是否为整数（乘以1000转换为整数计算）
func  isIntegerInteger(value float64) bool {
	// 将浮点数乘以1000转换为整数进行计算
	intValue := int64(value * 1000)
	// 判断是否能被1000整除
	return intValue%1000 == 0
}

// 精度安全的算术运算函数（基于整数计算）
// addFloat64 精度安全的浮点数加法
func  addFloat64(a ...float64) float64 {
	var total int64
	for _, value := range a {
		total += int64(value * 1000)
	}
	return float64(total) / 1000
}

// subtractFloat64 精度安全的浮点数减法
func  subtractFloat64(a, b float64) float64 {
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	result := intA - intB
	return float64(result) / 1000
}

// multiplyFloat64 精度安全的浮点数乘法
func  multiplyFloat64(a, b float64) float64 {
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	result := intA * intB
	return float64(result) / 1000000 // 除以1000000是因为两个数都乘以了1000
}

// divideFloat64 精度安全的浮点数除法
func  divideFloat64(a, b float64) float64 {
	if b == 0 {
		return 0 // 避免除零错误
	}
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	result := intA / intB
	return float64(result) // 不需要再除以1000，因为分子分母都乘以了1000
}

// sumFloat64 精度安全的浮点数求和（可变参数）
func  sumFloat64(values ...float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var total int64
	for _, value := range values {
		total += int64(value * 1000)
	}
	return float64(total) / 1000
}

// getRowNumber 获取记录中的Excel行号
func getRowNumber(data map[string]interface{}) int {
	// 尝试获取记录的行号
	if rowNum, ok := data["_excel_row"].(int); ok {
		return rowNum
	}
	// 如果没有找到行号，返回默认值
	return 1
}

// isNumber 判断字符串是否为数字（包含小数点）
func isNumber(value string) bool {
    if value == "" {
        return false
    }
    
    for i := 0; i < len(value); i++ {
        if (value[i] < '0' || value[i] > '9') && value[i] != '.' {
            return false
        }
    }
    return true
}

// isInteger 判断字符串是否为数字
func isInteger(value string) bool {
    if value == "" {
        return false
    }
    
    for i := 0; i < len(value); i++ {
        if (value[i] < '0' || value[i] > '9') {
            return false
        }
    }
    return true
}

// isEmpty 判断字符串是否为空（包含空格）
func isEmpty(value string) bool {
    if value == "" {
        return true
    }
    
	value = strings.TrimSpace(value)
    return value == ""
}

// isNotEmpty 判断字符串是否不为空（包含空格）
func isNotEmpty(value string) bool {
	return !isEmpty(value)
}

// containsList 判断字符串列表是否包含指定字符串
func containsList(list []string, value string ) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

// notContainsList 判断字符串列表是否不包含指定字符串
func notContainsList(list []string, value string ) bool {
	return !containsList(list, value)
}

// 查找元素第一次出现的索引，找不到返回 -1
func IndexOf(slice []string, target string) int {
    for i, v := range slice {
        if v == target {
            return i
        }
    }
    return -1
}

// getCellPosition 获取记录中的Excel单元格位置
func GetCellPosition(fieldMapping map[string]ExcelFieldMapping, rowNumber int, fieldNames ...string) []string {
	// 查找字段在列名列表中的索引
	var cellPositions []string
    for _, fieldName := range fieldNames {
        field, ok := fieldMapping[fieldName]
        if !ok {
            return []string{}
        }
        cellPositions = append(cellPositions, fmt.Sprintf("%s%d", field.Column, rowNumber))
    }
	return cellPositions
}

// CreateExcelFieldMapping 创建Excel字段映射
func CreateExcelFieldMapping(columnNames []string) map[string]ExcelFieldMapping {
	mapping := make(map[string]ExcelFieldMapping)
	for index, fieldName := range columnNames {
		columnName, _ := excelize.ColumnNumberToName(index)
		mapping[fieldName] = ExcelFieldMapping{
			FieldName: fieldName,
			Index: index,
			Column: columnName,
		}
	}
	return mapping
}

// 将 TreeNode 转换为 3级map 结构
func toTreeMap(treeNode []TreeNode) map[string]map[string]map[string]bool {
	// 构建3级联动map结构
	// 结构: map[省份名称]map[城市名称]map[区县名称]bool
	result := make(map[string]map[string]map[string]bool)

	for _, treeNode := range treeNode {
		// 创建省份级别的map
		citiesMap := make(map[string]map[string]bool)
		result[treeNode.Name] = citiesMap

		// 判断省份是否有城市数据
		if treeNode.Children != nil && len(treeNode.Children) > 0 {
			for _, secondLevel := range treeNode.Children {
				// 创建城市级别的map
				districtsMap := make(map[string]bool)
				citiesMap[secondLevel.Name] = districtsMap

				// 判断城市是否有区县数据
				if secondLevel.Children != nil && len(secondLevel.Children) > 0 {
					for _, district := range secondLevel.Children {	
						// 存储区县信息
						districtsMap[district.Name] = true
					}
				}
			}
		}
	}

	return result
}

// IsMobile 验证是否是手机号
func IsMobile(phone string) bool {
     matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
    return matched
}

// IsLandline 验证是否是固定电话
func IsLandline(phone string) bool {
   matched, _ := regexp.MatchString(`^0\d{2,3}-?\d{7,8}$`, phone)
    return matched
}

// IsPhone 验证是否是手机号或固话
func IsPhone(phone string) bool {
    return IsMobile(phone) || IsLandline(phone)
}
