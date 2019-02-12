package main

import (
	"gosnake-server/comm"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
	kcp "github.com/xtaci/kcp-go"
)

func main() {
	listener, err := kcp.Listen("0.0.0.0:6688")
	if err != nil {
		log.Fatal("kcp.Listen:", err)
	}
	defer listener.Close()

	log.Println("GoSnake Serve UDP at:", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("listen.Accept:", err)
		}
		go handleClientConn(conn)
	}
}

// 本函数在一个单独的协程中运行，用于接收一个 client 的上行通信
func handleClientConn(conn net.Conn) {
	defer conn.Close()
	var room *Room // 此 client 所在的 room
	var cid int32  // 此 client 在 room 内的 id
	for {
		var buffer [512]byte

		// 接收数据包
		n, err := conn.Read(buffer[:])
		if err != nil {
			log.Println("conn.Read:", err)
			return
		}

		// 解析数据包
		up := &comm.Up{}
		err = proto.Unmarshal(buffer[:n], up)
		if err != nil {
			log.Println("proto.Unmarshal:", err)
			return
		}

		switch cmd := up.M.(type) {
		case *comm.Up_Join:
			// 新开局，加入一个 room
			if cmd.Join.Mode < 1 || cmd.Join.Mode > 3 {
				continue
			}
			room, cid = joinRoom(conn, cmd.Join.Mode)

		case *comm.Up_Op:
			// 把用户操作转发给 room
			if room == nil {
				break
			}
			room.postOp(cid, cmd.Op.Keycode)
		}
	}
}
