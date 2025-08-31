# 针对 AMD64 (Intel/AMD CPU) 架构的麒麟/UOS
wails build -platform linux/amd64 -o

# 针对 ARM64 (鲲鹏、飞腾等 CPU) 架构的麒麟/UOS
# wails build -platform linux/arm64 -package



wails build -platform linux/amd64 -s -m -nosyncgomod -skipembedcreate
