# 目标

一个贪吃蛇游戏。

# Server

## 通信协议

- Up: join

- Up: Op

- Down: kick-off

- Down: frame

## 生成 protobuffer 代码

	protoc --go_out=src gosnake.proto

## 直接运行

**需把本项目根目录添加到 GOPATH 中**

	go run src/gosnake-server/main.go

# Client

## 安装依赖包和依赖工具

	go get -u github.com/asticode/go-astilectron-bootstrap
	go get -u github.com/asticode/go-astilectron-bundler/...

## 打包

**需把本项目根目录添加到 GOPATH 中**

	astilectron-bundler -v

## 仅重新生成 bindata 文件（用于开发过程中修改了 H5 内容之后）

	astilectron-bundler bd -v

## 直接运行

**需把本项目根目录添加到 GOPATH 中**

	go run src/gosnake/main.go -d -v

## 运行时的文件展开位置：

	%UserProfile%\AppData\Roaming\GoSnake
	%UserProfile%\AppData\Roaming\Electron

# 参考资料

[go-astilectron](https://github.com/asticode/go-astilectron)

[20行代码的贪吃蛇](https://kongchenglc.github.io/blog/%E8%B4%AA%E5%90%83%E8%9B%8720170613/)

[An absurdly small jQuery alternative for modern browsers](https://github.com/kenwheeler/cash)

[A UDP server and client in Go](https://ops.tips/blog/udp-client-and-server-in-go/)

[Go support for Google's protocol buffers](https://github.com/golang/protobuf)
