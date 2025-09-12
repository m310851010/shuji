package slices

// Contains 泛型函数：检查切片中是否包含指定元素
// T 是类型参数，约束为 comparable（支持 == 和 != 比较）
func Contains[T comparable](slice []T, value T) bool {
    for _, v := range slice {
        if v == value {
            return true
        }
    }
    return false
}