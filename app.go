package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	sysruntime "runtime"
	"shuji/db"
	"strings"
	"time"

	"shuji/data_import"

	"github.com/klauspost/cpuid/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var dbDst string
var dbDstPath string

// App struct
type App struct {
	ctx     context.Context
	fs      embed.FS
	db      *db.Database
	dbError error
}

var Config = &AppConfig{}

var Env = &EnvResult{
	AppName:     "",
	AppFileName: "",
	BasePath:    "",
	OS:          sysruntime.GOOS,
	ARCH:        sysruntime.GOARCH,
	X64Level:    cpuid.CPU.X64Level(),
	AssetsDir:   "",
	ExePath:     "",
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func CreateApp(fs embed.FS) *App {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	Env.ExePath = exePath
	Env.BasePath = filepath.Dir(exePath)
	Env.AppName = APP_NAME
	Env.AppFileName = filepath.Base(exePath)
	Env.AssetsDir = "frontend/dist"

	app := NewApp()
	app.fs = fs

	// Use absolute path for database
	dbDst = filepath.Join(Env.BasePath, DATA_DIR_NAME)
	dbDstPath = filepath.Join(dbDst, DB_FILE_NAME)

	// 保证数据库目录存在，防止抽取数据库文件失败导致后续找不到db文件
	if _, err := os.Stat(dbDst); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDst, os.ModePerm); err != nil {
			log.Fatalf("创建数据库目录失败: %v", err)
		}
	}
	if _, err := os.Stat(dbDstPath); os.IsNotExist(err) {
		extractEmbeddedFiles(fs)
		time.Sleep(1 * time.Second)
	}

	newDb, err := db.NewDatabase(dbDstPath, DB_PASSWORD)
	if err != nil {
		log.Printf("创建数据库失败: %v", err)
		app.dbError = err
	} else {
		app.db = newDb
	}

	log.Printf("数据库路径: %s", dbDstPath)
	log.Printf("exePath 路径: %s", exePath)
	log.Printf("基础路径: %s", Env.BasePath)

	createCacheDirs(Env.BasePath)
	return app
}

// 检查并创建缓存目录及子目录
func createCacheDirs(basePath string) {
	cacheDir := filepath.Join(basePath, CACHE_FILE_DIR_NAME)
	subDirs := []string{
		TableType1,
		TableType2,
		TableType3,
		TableTypeAttachment2,
	}

	// 检查并创建缓存主目录
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		os.MkdirAll(cacheDir, os.ModePerm)
	}

	// 循环检查并创建各子目录
	for _, dirName := range subDirs {
		subDir := filepath.Join(cacheDir, dirName)
		if _, err := os.Stat(subDir); os.IsNotExist(err) {
			os.MkdirAll(subDir, os.ModePerm)
		}
	}
}

func extractEmbeddedFiles(fs embed.FS) {
	dbSrcPath := "frontend/dist/" + DB_FILE_NAME
	if _, err := os.Stat(dbDstPath); os.IsNotExist(err) {
		extractFile(fs, dbSrcPath, dbDstPath)
	}
}

func extractFiles(fs embed.FS, srcDir, dstDir string) {
	files, _ := fs.ReadDir(srcDir)
	for _, file := range files {
		fileName := file.Name()
		dstPath := GetPath(dstDir + "/" + fileName)
		extractFile(fs, srcDir+"/"+fileName, dstPath)
	}
}

func extractFile(fs embed.FS, srcPath, dstPath string) {
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		log.Printf("抽取文件 [%s]", dstPath)
		data, err := fs.ReadFile(srcPath)
		if err != nil {
			log.Printf("抽取文件失败: %s: %v", srcPath, err)
			return
		}
		if err := os.WriteFile(dstPath, data, os.ModePerm); err != nil {
			log.Printf("抽取文件失败: %s: %v", dstPath, err)
		} else {
			log.Printf("抽取文件成功: %s", dstPath)
		}
	} else {
		log.Printf("文件已存在: %s", dstPath)
	}
}

// 启动程序
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if a.dbError != nil {
		runtime.WindowHide(ctx)
		errorMsg := fmt.Sprintf("数据库初始化失败：%v\n\n", a.dbError)

		// 根据错误类型提供更具体的建议
		if strings.Contains(a.dbError.Error(), "out of memory") {
			errorMsg += "可能的原因：\n• 数据库密码错误\n• 数据库文件损坏\n\n"
		} else if strings.Contains(a.dbError.Error(), "file is not a database") {
			errorMsg += "可能的原因：\n• 数据库文件损坏或不是有效的SQLite文件\n• 文件被其他程序占用\n\n"
		} else {
			errorMsg += "可能的原因：\n• 数据库文件不存在\n• 文件权限问题\n• 磁盘空间不足\n\n"
		}

		errorMsg += "请检查数据库文件或联系技术支持。\n\n程序将退出。"

		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "数据库错误",
			Message: errorMsg,
		})
		runtime.Quit(ctx)
		os.Exit(0)
	}
}

// 退出程序
func (a *App) ExitApp() {
	runtime.Quit(a.ctx)
	os.Exit(0)
}

func (a *App) GetCtx() context.Context {
	return a.ctx
}

// 获取运行环境变量
func (a *App) GetEnv() EnvResult {
	return EnvResult{
		AppName:     Env.AppName,
		AppFileName: Env.AppFileName,
		BasePath:    Env.BasePath,
		OS:          Env.OS,
		ARCH:        Env.ARCH,
		X64Level:    Env.X64Level,
		ExePath:     Env.ExePath,
		AssetsDir:   Env.AssetsDir,
	}
}

// ReadFile 读取文件内容
// path: 文件路径
// isEmbed: 是否为嵌入式文件
func (a *App) ReadFile(path string, isEmbed bool) ([]byte, error) {
	if isEmbed {
		// 读取嵌入式文件
		return a.fs.ReadFile(path)
	} else {
		// 读取普通文件
		return os.ReadFile(GetPath(path))
	}
}

// GetFileInfo 根据路径获取文件信息
func (a *App) GetFileInfo(path string) (*FileInfo, error) {
	fullPath := GetPath(path)
	// 检查文件是否存在
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("文件不存在: %s", path)
		}
		return nil, err
	}

	// 获取父目录
	parentDir := filepath.Dir(path)

	// 获取扩展名（如果是文件）
	ext := ""
	if !info.IsDir() {
		ext = strings.ToLower(filepath.Ext(path))
	}

	// 构造返回对象
	fileInfo := &FileInfo{
		Name:         info.Name(),               // 文件名（如：document.txt）
		FullPath:     fullPath,                  // 完整路径
		Size:         info.Size(),               // 字节大小
		IsDirectory:  info.IsDir(),              // 是否是目录
		LastModified: info.ModTime().UnixNano(), // 修改时间
		Ext:          ext,                       // 扩展名（如：.txt）
		IsFile:       !info.IsDir(),             // 是否是文件
		ParentDir:    parentDir,                 // 父目录路径
	}

	return fileInfo, nil
}

func (a *App) FileExists(path string) FlagResult {
	log.Printf("FileExists: %s", path)
	path = GetPath(path)
	_, err := os.Stat(path)
	if err == nil {
		return FlagResult{true, "true"}
	}
	if os.IsNotExist(err) {
		return FlagResult{true, "false"}
	}
	return FlagResult{false, err.Error()}
}

func (a *App) Movefile(source string, target string) FlagResult {
	log.Printf("Movefile: %s -> %s", source, target)

	fullSource := GetPath(source)
	fullTarget := GetPath(target)

	if err := os.MkdirAll(filepath.Dir(fullTarget), os.ModePerm); err != nil {
		return FlagResult{false, err.Error()}
	}

	if err := os.Rename(fullSource, fullTarget); err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func (a *App) Removefile(path string) FlagResult {
	log.Printf("RemoveFile: %s", path)

	fullPath := GetPath(path)

	if err := os.RemoveAll(fullPath); err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

// Copyfile 复制文件
func (a *App) Copyfile(src string, dst string) FlagResult {
	log.Printf("Copyfile: %s -> %s", src, dst)

	srcPath := GetPath(src)
	dstPath := GetPath(dst)

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return FlagResult{false, err.Error()}
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

// Makedir 创建目录
func (a *App) Makedir(path string) FlagResult {
	log.Printf("Makedir: %s", path)

	fullPath := GetPath(path)

	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

// Readdir 读取目录内容
func (a *App) Readdir(path string) FlagResult {
	log.Printf("Readdir: %s", path)

	fullPath := GetPath(path)

	files, err := os.ReadDir(fullPath)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	var result []string

	for _, file := range files {
		if info, err := file.Info(); err == nil {
			result = append(result, fmt.Sprintf("%v,%v,%v", info.Name(), info.Size(), info.IsDir()))
		}
	}

	return FlagResult{true, strings.Join(result, "|")}
}

func (a *App) AbsolutePath(path string) FlagResult {
	log.Printf("绝对路径: %s", path)
	absPath := GetPath(path)
	return FlagResult{true, absPath}
}

// OpenExternal 执行外部命令
func (a *App) OpenExternal(path string) error {
	log.Printf("OpenExternal: %s", path)

	exePath := GetPath(path)

	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		exePath = path
	}

	var cmd *exec.Cmd
	switch sysruntime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", exePath)
	case "darwin":
		cmd = exec.Command("open", exePath)
	case "linux":
		cmd = exec.Command("xdg-open", exePath)
	default:
		return fmt.Errorf("不支持的操作系统")
	}

	return cmd.Start()
}

// GetCurrentOSUser 获取当前操作系统用户
func (a *App) GetCurrentOSUser() string {
	return GetCurrentOSUser()
}

// GetCachePath 获取缓存路径
func (a *App) GetCachePath(tableType string) string {
	return GetPath(filepath.Join(CACHE_FILE_DIR_NAME, tableType))
}

// CacheFileExists 检查缓存文件是否存在
func (a *App) CacheFileExists(tableType string, fileName string) db.QueryResult {
	cachePath := GetPath(filepath.Join(CACHE_FILE_DIR_NAME, tableType, fileName))
	_, err := os.Stat(cachePath)
	if err == nil {
		return db.QueryResult{Ok: true, Message: "缓存文件存在", Data: cachePath}
	}
	if os.IsNotExist(err) {
		return db.QueryResult{Ok: false, Message: "缓存文件不存在"}
	}
	return db.QueryResult{Ok: false, Message: err.Error(), Data: err.Error()}
}

// CopyFileToCache 复制文件到缓存目录
func (a *App) CopyFileToCache(tableType string, filePath string) db.QueryResult {
	cachePath, err := CopyCacheFile(filePath, tableType)
	if err != nil {
		return db.QueryResult{Ok: false, Message: err.Error()}
	}
	return db.QueryResult{Ok: true, Message: "文件复制成功", Data: cachePath}
}

// GetDB 获取数据库实例
func (a *App) GetDB() *db.Database {
	return a.db
}

// ==================== 校验文件 API ====================

// ValidateTable1File 校验附表1文件
func (a *App) ValidateTable1File(filePath string, isCover bool) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ValidateTable1File(filePath, isCover)
}

// ValidateTable2File 校验附表2文件
func (a *App) ValidateTable2File(filePath string, isCover bool) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ValidateTable2File(filePath, isCover)
}

// ValidateTable3File 校验附表3文件
func (a *App) ValidateTable3File(filePath string, isCover bool) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ValidateTable3File(filePath, isCover)
}

// ValidateAttachment2File 校验附件2文件
func (a *App) ValidateAttachment2File(filePath string, isCover bool) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ValidateAttachment2File(filePath, isCover)
}

// ==================== 模型校验 API ====================

// ModelDataCheckTable1 附表1模型校验
func (a *App) ModelDataCheckTable1() db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ModelDataCheckTable1()
}

// ModelDataCheckTable2 附表2模型校验
func (a *App) ModelDataCheckTable2() db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ModelDataCheckTable2()
}

// ModelDataCheckTable3 附表3模型校验
func (a *App) ModelDataCheckTable3() db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ModelDataCheckTable3()
}

// ModelDataCheckAttachment2 附件2模型校验
func (a *App) ModelDataCheckAttachment2() db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ModelDataCheckAttachment2()
}

// ModelDataCheckReportDownload 下载报告
func (a *App) ModelDataCheckReportDownload(tableType string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ModelDataCheckReportDownload(tableType)
}

// ==================== 数据覆盖 API ====================

// ModelDataCoverTable1 覆盖附表1数据
func (a *App) ModelDataCoverTable1(fileNames []string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ModelDataCoverTable1(fileNames)
}

// ModelDataCoverTable2 覆盖附表2数据
func (a *App) ModelDataCoverTable2(fileNames []string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ModelDataCoverTable2(fileNames)
}

// ModelDataCoverTable3 覆盖附表3数据
func (a *App) ModelDataCoverTable3(fileNames []string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ModelDataCoverTable3(fileNames)
}

// ModelDataCoverAttachment2 覆盖附件2数据
func (a *App) ModelDataCoverAttachment2(fileNames []string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ModelDataCoverAttachment2(fileNames)
}

// ==================== 导入记录服务 API ====================

// InsertImportRecord 插入导入记录
func (a *App) InsertImportRecord(fileName, fileType, importState, describe string) {
	if a.dbError != nil {
		log.Printf("数据库连接失败，无法插入日志")
		return
	}

	service := NewDataImportRecordService(a.db)
	service.InsertImportRecord(fileName, fileType, importState, describe)
}

// GetImportRecordsByFileType 根据文件类型查询导入记录
func (a *App) GetImportRecordsByFileType(fileType string) db.QueryResult {
	if a.dbError != nil {
		return db.QueryResult{Ok: false, Message: "数据库连接失败"}
	}

	service := NewDataImportRecordService(a.db)
	return service.GetImportRecordsByFileType(fileType)
}

// ==================== 人工校验 API ====================

// QueryDataTable1 查询附表1数据
func (a *App) QueryDataTable1() db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.QueryDataTable1()
}

// QueryDataDetailTable1 查询附表1详细数据
func (a *App) QueryDataDetailTable1(obj_id string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.QueryDataDetailTable1(obj_id)
}

// ConfirmDataTable1 确认附表1数据
func (a *App) ConfirmDataTable1(obj_id []string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ConfirmDataTable1(obj_id)
}

// QueryDataTable2 查询附表2数据
func (a *App) QueryDataTable2() db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.QueryDataTable2()
}

// ConfirmDataTable2 确认附表2数据
func (a *App) ConfirmDataTable2(obj_id []string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ConfirmDataTable2(obj_id)
}

// QueryDataTable3 查询附表3数据
func (a *App) QueryDataTable3() db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.QueryDataTable3()
}

// ConfirmDataTable3 确认附表3数据
func (a *App) ConfirmDataTable3(obj_id []string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ConfirmDataTable3(obj_id)
}

// QueryDataAttachment2 查询附件2数据
func (a *App) QueryDataAttachment2() db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.QueryDataAttachment2()
}

// ConfirmDataAttachment2 确认附件2数据
func (a *App) ConfirmDataAttachment2(obj_id []string) db.QueryResult {
	dataImportService := data_import.NewDataImportService(a)
	return dataImportService.ConfirmDataAttachment2(obj_id)
}

// ========================SM4加密========================

// SM4Encrypt 加密
func (a *App) SM4Encrypt(plaintext string) (string, error) {
	string, error := SM4Encrypt(plaintext)
	return string, error
}

// SM4Decrypt 解密
func (a *App) SM4Decrypt(ciphertext string) (string, error) {
	string, error := SM4Decrypt(ciphertext)
	return string, error
}
