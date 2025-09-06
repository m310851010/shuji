package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	sysruntime "runtime"
	"syscall"

	"github.com/google/uuid"
	"github.com/tjfoc/gmsm/sm4"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Windows常量定义
const (
	CREATE_NO_WINDOW  = 0x08000000
	DETACHED_PROCESS  = 0x00000008
)

func GetPath(path string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(Env.BasePath, path)
	}
	return filepath.Clean(path)
}

// SM4Encrypt SM4加密函数
// 加密模式：ECB
// 填充方式：PKCS#7
// 输出格式：十六进制字符串
// 不足16字节时用0填充，超过16字节时截断
// 加密时使用PKCS#7填充

func SM4Encrypt(plaintext string) (string, error) {
	key := DEFAULT_ENCRYPTION_KEY
	// 将密钥转换为字节数组
	keyBytes := []byte(key)
	if len(keyBytes) != SM4_KEY_LENGTH {
		// 如果密钥长度不是16字节，进行填充或截断
		if len(keyBytes) < SM4_KEY_LENGTH {
			// 填充到16字节
			for len(keyBytes) < SM4_KEY_LENGTH {
				keyBytes = append(keyBytes, 0)
			}
		} else {
			// 截断到16字节
			keyBytes = keyBytes[:SM4_KEY_LENGTH]
		}
	}

	// 创建SM4加密器
	cipher, err := sm4.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("创建SM4加密器失败: %v", err)
	}

	// 将明文转换为字节数组
	plaintextBytes := []byte(plaintext)

	// 计算需要填充的字节数
	padding := SM4_BLOCK_SIZE - (len(plaintextBytes) % SM4_BLOCK_SIZE)
	if padding == SM4_BLOCK_SIZE {
		padding = 0
	}

	// 填充明文
	for i := 0; i < padding; i++ {
		plaintextBytes = append(plaintextBytes, byte(padding))
	}

	// 加密
	ciphertext := make([]byte, len(plaintextBytes))
	for i := 0; i < len(plaintextBytes); i += SM4_BLOCK_SIZE {
		cipher.Encrypt(ciphertext[i:i+SM4_BLOCK_SIZE], plaintextBytes[i:i+SM4_BLOCK_SIZE])
	}

	// 返回十六进制字符串
	return hex.EncodeToString(ciphertext), nil
}

// SM4Decrypt SM4解密函数
func SM4Decrypt(ciphertextHex string) (string, error) {
	key := DEFAULT_ENCRYPTION_KEY
	// 将密钥转换为字节数组
	keyBytes := []byte(key)
	if len(keyBytes) != SM4_KEY_LENGTH {
		// 如果密钥长度不是16字节，进行填充或截断
		if len(keyBytes) < SM4_KEY_LENGTH {
			// 填充到16字节
			for len(keyBytes) < SM4_KEY_LENGTH {
				keyBytes = append(keyBytes, 0)
			}
		} else {
			// 截断到16字节
			keyBytes = keyBytes[:SM4_KEY_LENGTH]
		}
	}

	// 创建SM4解密器
	cipher, err := sm4.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("创建SM4解密器失败: %v", err)
	}

	// 将十六进制字符串转换为字节数组
	ciphertextBytes, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("解析十六进制字符串失败: %v", err)
	}

	// 解密
	plaintext := make([]byte, len(ciphertextBytes))
	for i := 0; i < len(ciphertextBytes); i += SM4_BLOCK_SIZE {
		cipher.Decrypt(plaintext[i:i+SM4_BLOCK_SIZE], ciphertextBytes[i:i+SM4_BLOCK_SIZE])
	}

	// 去除填充
	if len(plaintext) > 0 {
		padding := int(plaintext[len(plaintext)-1])
		if padding > 0 && padding <= SM4_BLOCK_SIZE {
			plaintext = plaintext[:len(plaintext)-padding]
		}
	}

	// 返回明文
	return string(plaintext), nil
}

// GetCurrentOSUser 获取当前操作系统登录账号
func GetCurrentOSUser() string {
	currentUser, err := user.Current()
	if err != nil {
		return "系统"
	}
	return currentUser.Username
}

func GenerateUUID() string {
	return uuid.New().String()
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

// SendImportResultNotification 发送导入结果通知
func SendImportResultNotification(ctx context.Context, result map[string]interface{}, messageID string) {
	result["messageId"] = messageID
	runtime.EventsEmit(ctx, "import_result", result)
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

// CreateSilentCommand 创建一个静默执行的命令（Windows专用）
// 使用最彻底的隐藏方法，适用于Windows 11等新版本系统
func CreateSilentCommand(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	// 在Windows系统上设置静默执行属性
	if sysruntime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
			CreationFlags: CREATE_NO_WINDOW,
		}
		// 重定向标准输出和错误输出到空设备
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
	}
	return cmd
}
