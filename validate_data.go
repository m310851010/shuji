package main

import (
	"fmt"
	"strings"
	"path/filepath"
	"os"
	"archive/zip"
	"io"
	"github.com/xuri/excelize/v2"
)

// ValidateData 验证数据
func (a *App) ValidateData(province string, filePaths []string) QueryResult {
	if len(filePaths) != 4 {
		return QueryResult{Ok: false, Message: "请上传4个文件"}
	}

	// 分类并校验文件
	validationTasks, err := a.classifyAndValidate(filePaths)
	if err != nil {
		return QueryResult{Ok: false, Message: err.Error()}
	}

	// 检查是否有错误
	hasErrors := false
	for _, task := range validationTasks {
		if hasSheetErrors(task.Results) {
			hasErrors = true
			break
		}
	}

	// 没有错误，清理并返回
	if !hasErrors {
		cleanupTasks(validationTasks)
		return QueryResult{Ok: true, Data: nil, Message: ""}
	}

	// 生成错误报告
	zipPath, err := a.generateErrorReport(validationTasks)
	if err != nil {
		cleanupTasks(validationTasks)
		return QueryResult{Ok: false, Message: err.Error()}
	}

	// 清理缓存文件
	cleanupTasks(validationTasks)

	return QueryResult{
		Ok:      true,
		Data:    zipPath,
		Message: "校验完成，发现错误，请下载报告查看",
	}
}

// ValidationTask 校验任务
type ValidationTask struct {
	Prefix    string                 // 文件前缀（如"表1"）
	Validate  func(string) ([]ValidateSheetResult, error) // 校验函数
	CachePath string                 // 缓存文件路径
	Results   []ValidateSheetResult  // 校验结果
}

// classifyAndValidate 分类文件并执行校验
func (a *App) classifyAndValidate(filePaths []string) ([]*ValidationTask, error) {
	// 定义校验任务
	tasks := []*ValidationTask{
		{Prefix: "表1", Validate: a.validateTable1},
		{Prefix: "表2", Validate: a.validateTable2},
		{Prefix: "表3", Validate: a.validateTable3},
		{Prefix: "附件2", Validate: a.validateAttachment2},
	}

	// 为每个文件匹配对应的任务
	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)
		matched := false
		
		for _, task := range tasks {
			if strings.HasPrefix(fileName, task.Prefix) {
				// 复制到缓存
				cachePath, err := CopyCacheFile(filePath, "")
				if err != nil {
					return tasks, err
				}
				task.CachePath = cachePath

				// 执行校验
				results, err := task.Validate(cachePath)
				if err != nil {
					return tasks, err
				}
				task.Results = results
				matched = true
				break
			}
		}
		
		if !matched {
			return tasks, fmt.Errorf("文件名称必须以表1、表2、表3、附件2开头")
		}
	}

	// 检查是否所有任务都已完成
	for _, task := range tasks {
		if task.CachePath == "" {
			return tasks, fmt.Errorf("缺少%s文件", task.Prefix)
		}
	}

	return tasks, nil
}

// generateErrorReport 生成错误报告
func (a *App) generateErrorReport(tasks []*ValidationTask) (string, error) {
	var errorFiles []string

	for _, task := range tasks {
		if hasSheetErrors(task.Results) {
			if err := a.markErrorsInExcel(task.CachePath, task.Results); err != nil {
				return "", fmt.Errorf("标记错误失败: %v", err)
			}
			errorFiles = append(errorFiles, task.CachePath)
		}
	}

	zipPath := GetPath(filepath.Join(CACHE_FILE_DIR_NAME, "校验报告.zip"))
	if err := createZipFile(zipPath, errorFiles); err != nil {
		return "", fmt.Errorf("生成压缩包失败: %v", err)
	}

	return zipPath, nil
}

// hasSheetErrors 检查是否有错误
func hasSheetErrors(results []ValidateSheetResult) bool {
	for _, result := range results {
		if len(result.Errors) > 0 {
			return true
		}
	}
	return false
}

// cleanupTasks 清理任务缓存文件
func cleanupTasks(tasks []*ValidationTask) {
	for _, task := range tasks {
		if task.CachePath != "" {
			os.Remove(task.CachePath)
		}
	}
}

var border = []excelize.Border{
	{Type: "left", Color: "#000000", Style: 1},
	{Type: "top", Color: "#000000", Style: 1},
	{Type: "bottom", Color: "#000000", Style: 1},
	{Type: "right", Color: "#000000", Style: 1},
}

// markErrorsInExcel 在Excel中标记错误
func (a *App) markErrorsInExcel(filePath string, results []ValidateSheetResult) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 创建样式
	blueStyle, yellowStyle, err := createCellStyles(f)
	if err != nil {
		return err
	}

	// 处理每个sheet
	for _, sheetResult := range results {
		if len(sheetResult.Errors) == 0 {
			continue
		}

		fmt.Printf("处理Sheet: %s, 错误数量: %d\n", sheetResult.SheetName, len(sheetResult.Errors))

		// 标记错误单元格
		if err := markErrorCells(f, sheetResult, blueStyle, yellowStyle); err != nil {
			return err
		}

		// 添加统计信息
		if err := addStatistics(f, sheetResult); err != nil {
			return err
		}
	}

	return f.Save()
}

// createCellStyles 创建单元格样式
func createCellStyles(f *excelize.File) (int, int, error) {
	blueStyle, err := f.NewStyle(&excelize.Style{
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"#00b0f0"}, Pattern: 1},
		Border: border,
	})
	if err != nil {
		return 0, 0, err
	}

	yellowStyle, err := f.NewStyle(&excelize.Style{
		Fill:   excelize.Fill{Type: "pattern", Color: []string{"#ffff00"}, Pattern: 1},
		Border: border,
	})
	if err != nil {
		return 0, 0, err
	}

	return blueStyle, yellowStyle, nil
}

// markErrorCells 标记错误单元格
func markErrorCells(f *excelize.File, sheetResult ValidateSheetResult, blueStyle, yellowStyle int) error {
	// 合并同一单元格的错误
	cellErrorsMap := make(map[string][]ValidationError)
	for _, errItem := range sheetResult.Errors {
		for _, cell := range errItem.Cells {
			cellErrorsMap[cell] = append(cellErrorsMap[cell], errItem)
		}
	}

	// 标记每个单元格
	for cell, errors := range cellErrorsMap {
		messages, flag := mergeErrorMessages(errors)
		
		// 选择样式
		style := yellowStyle
		if flag == 0 {
			style = blueStyle
		}

		// 设置单元格样式
		if err := f.SetCellStyle(sheetResult.SheetName, cell, cell, style); err != nil {
			return err
		}

		// 添加批注
		if err := f.AddComment(sheetResult.SheetName, excelize.Comment{
			Cell:   cell,
			Author: "系统校验",
			Paragraph: []excelize.RichTextRun{{Text: messages}},
		}); err != nil {
			return err
		}
	}

	return nil
}

// mergeErrorMessages 合并错误消息
func mergeErrorMessages(errors []ValidationError) (string, int) {
	var messages []string
	flag := 1 // 默认黄色
	for _, err := range errors {
		messages = append(messages, err.Message)
		if err.Flag == 0 {
			flag = 0 // 有蓝色标记则使用蓝色
		}
	}
	return strings.Join(messages, "\n"), flag
}

// addStatistics 添加统计信息
func addStatistics(f *excelize.File, sheetResult ValidateSheetResult) error {
	// 统计每类问题数量
	typeCount := make(map[string]int)
	for _, errItem := range sheetResult.Errors {
		typeCount[errItem.Type]++
	}

	// 确保统计行可见
	lastRow := sheetResult.RowNumber + 2
	f.SetRowVisible(sheetResult.SheetName, sheetResult.RowNumber+1, true)
	f.SetRowVisible(sheetResult.SheetName, lastRow, true)

	// 添加统计内容
	f.SetCellValue(sheetResult.SheetName, fmt.Sprintf("A%d", lastRow), "错误统计:")

	var statParts []string
	for errorType, count := range typeCount {
		statParts = append(statParts, fmt.Sprintf("%s: %d个", errorType, count))
	}
	f.SetCellValue(sheetResult.SheetName, fmt.Sprintf("B%d", lastRow), strings.Join(statParts, ", "))

	return nil
}

// createZipFile 创建压缩包
func createZipFile(zipPath string, files []string) error {
	// 确保目录存在
	dir := filepath.Dir(zipPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	// 创建zip文件
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 添加文件到压缩包
	for _, file := range files {
		err := addFileToZip(zipWriter, file)
		if err != nil {
			return err
		}
	}

	return nil
}

// addFileToZip 添加文件到压缩包
func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// 获取文件信息
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// 使用文件名作为压缩包内的名称
	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}
