package main

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/xuri/excelize/v2"
)


func (a *App) validateExcel(filePaths []string) (bool, string) {
	// 打开 Excel 文件
	f, err := excelize.OpenFile(filePaths[0])
	if err != nil {
		return false, "打开文件失败"
	}
	defer func() {
		// 关闭文件
		if err := f.Close(); err != nil {
			log.Fatalf("关闭文件失败: %v", err)
		}
	}()

	// 获取所有工作表名称
	sheetNames := f.GetSheetList()
	if len(sheetNames) == 0 {
		return false, "文件中没有工作表"
	}

	// 校验每个工作表
	for _, sheetName := range sheetNames {
		// 获取工作表
		rows := f.GetRows(sheetName)
		if len(rows) == 0 {
			return false, "工作表 " + sheetName + " 为空"
		}

		// 校验表头
		header := rows[0]
		if len(header) != 10 {
			return false, "工作表 " + sheetName + " 的表头列数不是10列"
		}
