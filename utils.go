package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
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

// getStringValue 安全获取字符串值
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func parseFloat(value string) float64 {
	if value == "" {
		return 0
	}

	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return result
}

func addFloat64(a, b float64) float64 {
	intA := int64(a * 1000)
	intB := int64(b * 1000)
	result := intA + intB
	return float64(result) / 1000
}
