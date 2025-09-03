@echo off
set GOARCH=386
set GOOS=windows
wails dev -s -m -nosyncgomod -skipembedcreate
