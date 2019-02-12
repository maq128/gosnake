# 目标

一个网络版的贪吃蛇游戏。

# Server

## 通信协议

- Up: join

	申请开始一局游戏。可以指定单人、双人、三人模式。

- Up: Op

	上报按键操作。

- Down: kick-off

	开始一局。

- Down: frame

	帧驱动数据。

- Down: finish

	本局结束。

## 生成 protobuffer 代码

	protoc --go_out=src gosnake.proto

## 打包

**需把本项目根目录添加到 GOPATH 中**

	go build gosnake-server

## 直接运行

**需把本项目根目录添加到 GOPATH 中**

	go run src/gosnake-server/main.go src/gosnake-server/room.go

# Client

## 安装依赖包和依赖工具

	go get -u github.com/asticode/go-astilectron-bootstrap
	go get -u github.com/asticode/go-astilectron-bundler/...
	go get -u github.com/xtaci/kcp-go

## 打包

**需把本项目根目录添加到 GOPATH 中**

	astilectron-bundler -v

## 仅重新生成 bindata 文件（用于开发过程中修改了 H5 内容之后）

	astilectron-bundler bd -v

## 直接运行

**需把本项目根目录添加到 GOPATH 中**

	go run src/gosnake/main.go -d -v

客户端运行时会通过域名 `gosnake.game` 找到服务器，所以需要在 hosts 文件中设置域名解析。

## 运行时的文件展开位置：

	%UserProfile%\AppData\Roaming\GoSnake
	%UserProfile%\AppData\Roaming\Electron

# 参考资料

[go-astilectron](https://github.com/asticode/go-astilectron)

[20行代码的贪吃蛇](https://kongchenglc.github.io/blog/%E8%B4%AA%E5%90%83%E8%9B%8720170613/)

[An absurdly small jQuery alternative for modern browsers](https://github.com/kenwheeler/cash)

[Go socket编程实践: UDP服务器和客户端实现](https://colobu.com/2014/12/02/go-socket-programming-UDP/)

[A Production-Grade Reliable-UDP Library for golang](https://github.com/xtaci/kcp-go)

[Go support for Google's protocol buffers](https://github.com/golang/protobuf)

[Language Guide (proto3)](https://developers.google.com/protocol-buffers/docs/proto3)
