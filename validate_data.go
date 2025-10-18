package main

// ValidateData 验证数据
func (a *App) ValidateData(filePaths []string) QueryResult {
	return QueryResult{
		Ok:      true,
		Data:    nil,
		Message: "验证通过",
	}
}

// DownloadValidateReport 下载验证报告
func (a *App) DownloadValidateReport() QueryResult {
	return QueryResult{
		Ok:      true,
		Data:    nil,
		Message: "下载验证报告成功",
	}
}
