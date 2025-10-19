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

	// 根据文件名识别并分类文件
	fileMap := make(map[string]string)
	for _, filePath := range filePaths {
		fileName := filepath.Base(filePath)
		if strings.HasPrefix(fileName, "表1") {
			fileMap["table1"] = filePath
		} else if strings.HasPrefix(fileName, "表2") {
			fileMap["table2"] = filePath
		} else if strings.HasPrefix(fileName, "表3") {
			fileMap["table3"] = filePath
		} else if strings.HasPrefix(fileName, "附件2") {
			fileMap["attachment2"] = filePath
		}
	}

	// 检查是否所有文件都已识别
	if len(fileMap) != 4 {
		return QueryResult{Ok: false, Message: "文件名称必须以表1、表2、表3、附件2开头"}
	}

	// 复制文件到缓存并校验
	cacheFiles := make(map[string]string)
	allResults := make(map[string][]ValidateSheetResult)
	hasErrors := false

	// 表1校验
	if filePath, ok := fileMap["table1"]; ok {
		cachePath, err := CopyCacheFile(filePath, "")
		if err != nil {
			return QueryResult{Ok: false, Message: err.Error()}
		}
		cacheFiles["table1"] = cachePath

		results, err := a.validateTable1(cachePath)
		if err != nil {
			return QueryResult{Ok: false, Message: err.Error()}
		}
		allResults["table1"] = results
		for _, result := range results {
			if len(result.Errors) > 0 {
				hasErrors = true
				break
			}
		}
	}

	// 表2校验
	if filePath, ok := fileMap["table2"]; ok {
		cachePath, err := CopyCacheFile(filePath, "")
		if err != nil {
			return QueryResult{Ok: false, Message: err.Error()}
		}
		cacheFiles["table2"] = cachePath

		results, err := a.validateTable2(cachePath)
		if err != nil {
			return QueryResult{Ok: false, Message: err.Error()}
		}
		allResults["table2"] = results
		for _, result := range results {
			if len(result.Errors) > 0 {
				hasErrors = true
				break
			}
		}
	}

	// 表3校验
	if filePath, ok := fileMap["table3"]; ok {
		cachePath, err := CopyCacheFile(filePath, "")
		if err != nil {
			return QueryResult{Ok: false, Message: err.Error()}
		}
		cacheFiles["table3"] = cachePath

		results, err := a.validateTable3(cachePath)
		if err != nil {
			return QueryResult{Ok: false, Message: err.Error()}
		}
		allResults["table3"] = results
		for _, result := range results {
			if len(result.Errors) > 0 {
				hasErrors = true
				break
			}
		}
	}

	// 附件2校验
	if filePath, ok := fileMap["attachment2"]; ok {
		cachePath, err := CopyCacheFile(filePath, "")
		if err != nil {
			return QueryResult{Ok: false, Message: err.Error()}
		}
		cacheFiles["attachment2"] = cachePath

		results, err := a.validateAttachment2(cachePath)
		if err != nil {
			return QueryResult{Ok: false, Message: err.Error()}
		}
		allResults["attachment2"] = results
		for _, result := range results {
			if len(result.Errors) > 0 {
				hasErrors = true
				break
			}
		}
	}

	// 如果没有错误，直接返回成功
	if !hasErrors {
		// 清理缓存文件
		for _, cachePath := range cacheFiles {
			os.Remove(cachePath)
		}
		return QueryResult{
			Ok:      true,
			Data:    nil,
			Message: "",
		}
	}

	// 标记错误并生成压缩包
	errorFiles := []string{}
	for tableType, results := range allResults {
		cachePath := cacheFiles[tableType]
		hasError := false
		for _, result := range results {
			if len(result.Errors) > 0 {
				hasError = true
				break
			}
		}

		if hasError {
			// 标记错误
			err := a.markErrorsInExcel(cachePath, results)
			if err != nil {
				// 清理所有缓存文件
				for _, path := range cacheFiles {
					os.Remove(path)
				}
				return QueryResult{Ok: false, Message: fmt.Sprintf("标记错误失败: %v", err)}
			}
			errorFiles = append(errorFiles, cachePath)
		}
	}

	// 生成压缩包
	zipPath := GetPath(filepath.Join(CACHE_FILE_DIR_NAME, "校验报告.zip"))
	err := createZipFile(zipPath, errorFiles)
	if err != nil {
		// 清理所有缓存文件
		for _, path := range cacheFiles {
			os.Remove(path)
		}
		return QueryResult{Ok: false, Message: fmt.Sprintf("生成压缩包失败: %v", err)}
	}

	// 清理缓存的excel文件
	for _, cachePath := range cacheFiles {
		os.Remove(cachePath)
	}

	return QueryResult{
		Ok:      true,
		Data:    zipPath,
		Message: "校验完成，发现错误，请下载报告查看",
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

	// 定义颜色样式
	blueStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#00b0f0"}, // 蓝色
			Pattern: 1,
		},
		Border: border,
	})
	if err != nil {
		return err
	}

	yellowStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffff00"}, // 黄色
			Pattern: 1,
		},
		Border: border,
	})
	if err != nil {
		return err
	}

	// 遍历每个sheet的错误
	for _, sheetResult := range results {
		fmt.Println(fmt.Sprintf("处理Sheet: %s, 错误数量: %d, RowNumber: %d", sheetResult.SheetName, len(sheetResult.Errors), sheetResult.RowNumber))
		
		if len(sheetResult.Errors) == 0 {
			continue
		}

		// 统计每类问题的数量
		typeCount := make(map[string]int)
		for _, errItem := range sheetResult.Errors {
			typeCount[errItem.Type]++
		}

		// 合并同一单元格的错误
		cellErrorsMap := make(map[string][]ValidationError)
		for _, errItem := range sheetResult.Errors {
			for _, cell := range errItem.Cells {
				cellErrorsMap[cell] = append(cellErrorsMap[cell], errItem)
			}
		}

		// 标记每个单元格
		for cell, errors := range cellErrorsMap {
			// 合并错误消息
			var messages []string
			flag := 1 // 默认黄色
			for _, err := range errors {
				messages = append(messages, err.Message)
				if err.Flag == 0 {
					flag = 0 // 如果有蓝色标记，则使用蓝色
				}
			}
			comment := strings.Join(messages, "\n")

			// 设置单元格颜色
			style := yellowStyle
			if flag == 0 {
				style = blueStyle
			}
			err := f.SetCellStyle(sheetResult.SheetName, cell, cell, style)
			if err != nil {
				return err
			}

			// 添加批注
			err = f.AddComment(sheetResult.SheetName, excelize.Comment{
				Cell:   cell,
				Author: "系统校验",
				Paragraph: []excelize.RichTextRun{
					{Text: comment},
				},
			})
			if err != nil {
				return err
			}
		}

		// 取消隐藏
		f.SetRowVisible(sheetResult.SheetName, sheetResult.RowNumber + 1, true)
		// 在最后一行添加统计信息
		lastRow := sheetResult.RowNumber + 2 // 空一行后添加统计
		f.SetRowVisible(sheetResult.SheetName, lastRow, true)
		
		// 添加统计标题
		f.SetCellValue(sheetResult.SheetName, fmt.Sprintf("A%d", lastRow), "错误统计:")
		
		// 添加每类问题的统计
		var statParts []string
		for errorType, count := range typeCount {
			statParts = append(statParts, fmt.Sprintf("%s: %d个", errorType, count))
		}
		statText := strings.Join(statParts, ", ")
		f.SetCellValue(sheetResult.SheetName, fmt.Sprintf("B%d", lastRow), statText)
	}

	// 保存文件
	return f.Save()
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
