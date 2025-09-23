## 打包

- 替换文件 `pkg/DEBIAN/control` 

> 检查下`Architecture`是否是`arm64`

## 备用方案,让用户执行install脚本安装依赖

1. 先给`install`文件设置权限

```shell
chmod 755 ./install
```

2. 执行install脚本

```shell
./install
```
