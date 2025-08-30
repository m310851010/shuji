package data_import

import (
	"fmt"
	"math/big"
	"strings"
)

// ValidationContext 校验上下文，用于管理校验状态和错误收集
type ValidationContext struct {
	errors      []ValidationError
	rowNum      int
	tableType   string
	fieldErrors map[string]bool    // 记录字段是否已有错误，避免重复校验
	service     *DataImportService // 引用DataImportService以使用其方法
}

// NewValidationContext 创建新的校验上下文
func NewValidationContext(rowNum int, tableType string, service *DataImportService) *ValidationContext {
	return &ValidationContext{
		errors:      []ValidationError{},
		rowNum:      rowNum,
		tableType:   tableType,
		fieldErrors: make(map[string]bool),
		service:     service,
	}
}

// AddError 添加错误到校验上下文
func (vc *ValidationContext) AddError(message string, cells []string) {
	vc.errors = append(vc.errors, ValidationError{
		RowNumber: vc.rowNum,
		Message:   message,
		Cells:     cells,
	})
}

// AddFieldError 添加字段错误并标记字段为错误状态
func (vc *ValidationContext) AddFieldError(fieldName, message string) {
	cells := []string{vc.getCellPosition(fieldName)}
	vc.AddError(message, cells)
	vc.fieldErrors[fieldName] = true
}

// HasFieldError 检查字段是否已有错误
func (vc *ValidationContext) HasFieldError(fieldName string) bool {
	return vc.fieldErrors[fieldName]
}

// GetErrors 获取所有错误
func (vc *ValidationContext) GetErrors() []ValidationError {
	return vc.errors
}

// getCellPosition 获取单元格位置
func (vc *ValidationContext) getCellPosition(fieldName string) string {
	return vc.service.getCellPosition(vc.tableType, fieldName, vc.rowNum)
}

// parseBigFloat 解析字符串为 *big.Float，简化定点数处理
func parseBigFloat(value string) *big.Float {
	// 移除逗号和空格
	value = strings.ReplaceAll(value, ",", "")
	value = strings.TrimSpace(value)

	if value == "" {
		return big.NewFloat(0)
	}

	result := new(big.Float)
	result.SetString(value)
	return result
}

// ValidationRule 校验规则接口
type ValidationRule interface {
	Validate(data map[string]interface{}, vc *ValidationContext) bool
}

// RequiredFieldRule 必填字段校验规则
type RequiredFieldRule struct {
	FieldName string
	FieldDesc string
}

// Validate 校验必填字段
func (r *RequiredFieldRule) Validate(data map[string]interface{}, vc *ValidationContext) bool {
	value := vc.getStringValue(data[r.FieldName])
	if strings.TrimSpace(value) == "" {
		vc.AddFieldError(r.FieldName, fmt.Sprintf("%s不能为空", r.FieldDesc))
		return false
	}
	return true
}

// NumericRangeRule 数值范围校验规则
type NumericRangeRule struct {
	FieldName string
	FieldDesc string
	Min       *big.Float
	Max       *big.Float
}

// Validate 校验数值范围
func (r *NumericRangeRule) Validate(data map[string]interface{}, vc *ValidationContext) bool {
	// 如果字段已有错误，跳过校验
	if vc.HasFieldError(r.FieldName) {
		return false
	}

	value := vc.getStringValue(data[r.FieldName])
	if strings.TrimSpace(value) == "" {
		return true // 空值跳过范围校验
	}

	decimalValue := parseBigFloat(value)

	if r.Min != nil && decimalValue.Cmp(r.Min) < 0 {
		vc.AddFieldError(r.FieldName, fmt.Sprintf("%s不能小于%s", r.FieldDesc, r.Min.Text('f', -1)))
		return false
	}

	if r.Max != nil && decimalValue.Cmp(r.Max) > 0 {
		vc.AddFieldError(r.FieldName, fmt.Sprintf("%s不能大于%s", r.FieldDesc, r.Max.Text('f', -1)))
		return false
	}

	return true
}

// NumericComparisonRule 数值比较校验规则
type NumericComparisonRule struct {
	FieldName1 string
	FieldDesc1 string
	FieldName2 string
	FieldDesc2 string
	Operator   string // ">=", "<=", "==", "!="
	Message    string
}

// Validate 校验数值比较
func (r *NumericComparisonRule) Validate(data map[string]interface{}, vc *ValidationContext) bool {
	// 如果任一字段已有错误，跳过校验
	if vc.HasFieldError(r.FieldName1) || vc.HasFieldError(r.FieldName2) {
		return false
	}

	value1 := vc.getStringValue(data[r.FieldName1])
	value2 := vc.getStringValue(data[r.FieldName2])

	// 如果任一字段为空，跳过校验
	if strings.TrimSpace(value1) == "" || strings.TrimSpace(value2) == "" {
		return true
	}

	decimal1 := parseBigFloat(value1)
	decimal2 := parseBigFloat(value2)

	var isValid bool
	switch r.Operator {
	case ">=":
		isValid = decimal1.Cmp(decimal2) >= 0
	case "<=":
		isValid = decimal1.Cmp(decimal2) <= 0
	case "==":
		isValid = decimal1.Cmp(decimal2) == 0
	case "!=":
		isValid = decimal1.Cmp(decimal2) != 0
	case ">":
		isValid = decimal1.Cmp(decimal2) > 0
	case "<":
		isValid = decimal1.Cmp(decimal2) < 0
	default:
		isValid = true
	}

	if !isValid {
		cells := []string{
			vc.getCellPosition(r.FieldName1),
			vc.getCellPosition(r.FieldName2),
		}
		vc.AddError(r.Message, cells)
		return false
	}

	return true
}

// SumValidationRule 求和校验规则
type SumValidationRule struct {
	SumFieldName    string
	SumFieldDesc    string
	ComponentFields []string
	ComponentDescs  []string
	Message         string
}

// Validate 校验求和
func (r *SumValidationRule) Validate(data map[string]interface{}, vc *ValidationContext) bool {
	// 检查所有字段是否都有错误
	hasError := false
	for _, field := range append([]string{r.SumFieldName}, r.ComponentFields...) {
		if vc.HasFieldError(field) {
			hasError = true
			break
		}
	}
	if hasError {
		return false
	}

	// 获取总和字段值
	sumValue := vc.getStringValue(data[r.SumFieldName])
	if strings.TrimSpace(sumValue) == "" {
		return true // 总和为空时跳过校验
	}

	sumDecimal := parseBigFloat(sumValue)

	// 计算组件字段的和
	expectedSum := big.NewFloat(0)
	for _, field := range r.ComponentFields {
		value := vc.getStringValue(data[field])
		if strings.TrimSpace(value) != "" {
			fieldValue := parseBigFloat(value)
			expectedSum.Add(expectedSum, fieldValue)
		}
	}

	// 比较总和
	if sumDecimal.Cmp(expectedSum) != 0 {
		cells := []string{vc.getCellPosition(r.SumFieldName)}
		for _, field := range r.ComponentFields {
			cells = append(cells, vc.getCellPosition(field))
		}
		vc.AddError(r.Message, cells)
		return false
	}

	return true
}

// ValidationEngine 校验引擎
type ValidationEngine struct {
	rules []ValidationRule
}

// NewValidationEngine 创建新的校验引擎
func NewValidationEngine() *ValidationEngine {
	return &ValidationEngine{
		rules: []ValidationRule{},
	}
}

// AddRule 添加校验规则
func (ve *ValidationEngine) AddRule(rule ValidationRule) {
	ve.rules = append(ve.rules, rule)
}

// Validate 执行所有校验规则
func (ve *ValidationEngine) Validate(data map[string]interface{}, rowNum int, tableType string, service *DataImportService) []ValidationError {
	vc := NewValidationContext(rowNum, tableType, service)

	for _, rule := range ve.rules {
		rule.Validate(data, vc)
	}

	return vc.GetErrors()
}

// getStringValue 获取字符串值的辅助方法
func (vc *ValidationContext) getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// 预定义的常用校验规则
var (
	// 常用数值范围
	NonNegativeRule = func(fieldName, fieldDesc string) *NumericRangeRule {
		return &NumericRangeRule{
			FieldName: fieldName,
			FieldDesc: fieldDesc,
			Min:       big.NewFloat(0),
			Max:       nil,
		}
	}

	Max100000Rule = func(fieldName, fieldDesc string) *NumericRangeRule {
		return &NumericRangeRule{
			FieldName: fieldName,
			FieldDesc: fieldDesc,
			Min:       big.NewFloat(0),
			Max:       big.NewFloat(100000),
		}
	}

	// 常用比较规则
	GreaterEqualRule = func(field1, desc1, field2, desc2, message string) *NumericComparisonRule {
		return &NumericComparisonRule{
			FieldName1: field1,
			FieldDesc1: desc1,
			FieldName2: field2,
			FieldDesc2: desc2,
			Operator:   ">=",
			Message:    message,
		}
	}

	EqualRule = func(field1, desc1, field2, desc2, message string) *NumericComparisonRule {
		return &NumericComparisonRule{
			FieldName1: field1,
			FieldDesc1: desc1,
			FieldName2: field2,
			FieldDesc2: desc2,
			Operator:   "==",
			Message:    message,
		}
	}
)
