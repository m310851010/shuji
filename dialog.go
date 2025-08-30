package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// OpenSaveDialog 选择需要处理的文件
func (a *App) OpenSaveDialog(option FileDialogOptions) FileDialogResult {
	// 使用包装函数来处理异常
	return a.openSaveDialogWithRecover(option)
}

// openSaveDialogWithRecover 带异常处理的打开保存对话框函数
func (a *App) openSaveDialogWithRecover(option FileDialogOptions) FileDialogResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("OpenSaveDialog 发生异常: %v", r)
		}
	}()

	// DefaultPath 默认值
	if option.DefaultPath == "" {
		option.DefaultPath = "."
	}

	// option.Filters 默认值
	if len(option.Filters) == 0 {
		option.Filters = []FileFilter{
			{
				Name:    "所有文件",
				Pattern: "*.*",
			},
		}
	}

	var filters = a.transformFileFilters(option.Filters)

	result, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                option.Title,
		Filters:              filters,
		CanCreateDirectories: option.CreateDirectory,
		DefaultFilename:      option.DefaultFilename,
		DefaultDirectory:     option.DefaultPath,
	})
	if err != nil || result == "" {
		return FileDialogResult{
			Canceled:  true,
			FilePaths: []string{fmt.Sprintf("err %s!", err)},
		}
	}
	return FileDialogResult{
		Canceled:  false,
		FilePaths: []string{result},
	}
}

func padRight(s string) string {
	// 计算需要填充的空格数
	padding := 90 - len(s)
	if padding <= 0 {
		return s[:90] // 如果字符串长度超过或等于90，截取前90个字符
	}
	// 使用strings.Repeat创建空格字符串并拼接
	return s + strings.Repeat(" ", padding)
}

// ShowMessageBox 弹出消息框，支持自定义按钮，模仿 electron 的 showMessageBox
// type: info、error、question、warning四种类型
// buttons: 按钮文本数组
// defaultId: 默认按钮索引
// cancelId: 取消按钮索引
func (a *App) ShowMessageBox(options MessageBoxOptions) MessageBoxResult {
	// 使用包装函数来处理异常
	return a.showMessageBoxWithRecover(options)
}

// showMessageBoxWithRecover 带异常处理的显示消息框函数
func (a *App) showMessageBoxWithRecover(options MessageBoxOptions) MessageBoxResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("ShowMessageBox 发生异常: %v", r)
		}
	}()

	// 设置默认按钮
	buttons := options.Buttons
	if len(buttons) == 0 {
		buttons = []string{"确定"}
	}
	defaultId := options.DefaultId
	if defaultId < 0 || defaultId >= len(buttons) {
		defaultId = 0
	}
	cancelId := options.CancelId
	if cancelId < 0 || cancelId >= len(buttons) {
		cancelId = -1
	}

	// 根据options.Type设置对话框类型
	dialogType := runtime.InfoDialog
	switch options.Type {
	case "info":
		dialogType = runtime.InfoDialog
	case "error":
		dialogType = runtime.ErrorDialog
	case "warning":
		dialogType = runtime.WarningDialog
	case "question":
		dialogType = runtime.QuestionDialog
	}

	title := options.Title
	if title == "" {
		title = "提示"
	}

	message := padRight(options.Message)

	var response int

	log.Println(buttons)
	log.Println(len(buttons))

	// 使用runtime.MessageDialog的Buttons参数支持自定义按钮
	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    dialogType,
		Title:   title,
		Message: message,
		Buttons: buttons,
	})
	if err != nil {
		// 处理错误
		response = -1
	} else {
		// 根据用户点击的按钮返回对应的索引
		for i, button := range buttons {
			if result == button {
				response = i
				break
			}
		}
		// 如果没有找到匹配的按钮（用户取消或关闭对话框），返回取消按钮索引
		if response == -1 {
			response = cancelId
		}
	}

	return MessageBoxResult{
		Response:        response,
		CheckboxChecked: false, // 暂不支持复选框
	}
}

// option.Filters 转换为runtime.FileFilter
func (a *App) transformFileFilters(filters []FileFilter) []runtime.FileFilter {
	var _filters []runtime.FileFilter
	for _, filter := range filters {
		_filters = append(_filters, runtime.FileFilter{
			DisplayName: filter.Name,
			Pattern:     filter.Pattern,
		})
	}
	return _filters
}

// OpenFileDialog 选择需要处理的文件
func (a *App) OpenFileDialog(option FileDialogOptions) FileDialogResult {
	// 使用包装函数来处理异常
	return a.openFileDialogWithRecover(option)
}

// openFileDialogWithRecover 带异常处理的打开文件对话框函数
func (a *App) openFileDialogWithRecover(option FileDialogOptions) FileDialogResult {
	// 添加异常处理，防止函数崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("OpenFileDialog 发生异常: %v", r)
		}
	}()

	// option.Filters 默认值
	if len(option.Filters) == 0 {
		option.Filters = []FileFilter{
			{
				Name:    "所有文件",
				Pattern: "*.*",
			},
		}
	}

	var filters = a.transformFileFilters(option.Filters)

	// DefaultPath 默认值
	if option.DefaultPath == "" {
		option.DefaultPath = "."
	}

	print(option.OpenDirectory)
	// OpenDirectory 选择目录
	if option.OpenDirectory {
		selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
			Title:                option.Title,
			Filters:              filters,
			CanCreateDirectories: option.CreateDirectory,
			DefaultFilename:      option.DefaultFilename,
			DefaultDirectory:     option.DefaultPath,
		})

		if err != nil {
			return FileDialogResult{
				Canceled:  true,
				FilePaths: []string{fmt.Sprintf("err %s!", err)},
			}
		}

		if len(selection) == 0 {
			return FileDialogResult{
				Canceled:  true,
				FilePaths: []string{},
			}
		}

		return FileDialogResult{
			Canceled:  false,
			FilePaths: []string{selection},
		}
	}

	// 选择文件 多选
	if option.MultiSelections {
		selection, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
			Title:                option.Title,
			Filters:              filters,
			CanCreateDirectories: option.CreateDirectory,
			DefaultFilename:      option.DefaultFilename,
			DefaultDirectory:     option.DefaultPath,
		})

		if err != nil {
			return FileDialogResult{
				Canceled:  true,
				FilePaths: []string{fmt.Sprintf("err %s!", err)},
			}
		}

		if len(selection) == 0 {
			return FileDialogResult{
				Canceled:  true,
				FilePaths: []string{},
			}
		}
		return FileDialogResult{
			Canceled:  false,
			FilePaths: selection,
		}
	}

	// 选择文件, 单选
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                option.Title,
		Filters:              filters,
		CanCreateDirectories: option.CreateDirectory,
		DefaultFilename:      option.DefaultFilename,
		DefaultDirectory:     option.DefaultPath,
	})
	if err != nil {
		return FileDialogResult{
			Canceled:  true,
			FilePaths: []string{fmt.Sprintf("err %s!", err)},
		}
	}

	if selection == "" {
		return FileDialogResult{
			Canceled:  true,
			FilePaths: []string{},
		}
	}

	// 单选
	return FileDialogResult{
		Canceled:  false,
		FilePaths: []string{selection},
	}
}
