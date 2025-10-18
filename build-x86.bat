@echo off

rem 第一种方式
set GOARCH=386
set GOOS=windows

wails build  -s -m -nosyncgomod -skipembedcreate -webview2 embed

rem 第二种方式
rem go build -tags desktop,production -ldflags "-w -s -H windowsgui" -o 数据校验工具.exe
pause
