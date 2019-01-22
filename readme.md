# 目标

一个贪吃蛇游戏。

# 安装依赖包和依赖工具：

	go get -u github.com/asticode/go-astilectron-bootstrap
	go get -u github.com/asticode/go-astilectron-bundler/...

# 打包：

**需把本项目根目录添加到 GOPATH 中**

	astilectron-bundler -v

# 直接运行：

	go run src/gosnake/main.go src/gosnake/bind_windows_amd64.go

# 参考资料

[go-astilectron](https://github.com/asticode/go-astilectron)
