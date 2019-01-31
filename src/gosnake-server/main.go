package main

import (
	"gosnake-server/comm"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	kcp "github.com/xtaci/kcp-go"
)

func main() {
	listener, err := kcp.Listen("0.0.0.0:6688")
	if err != nil {
		log.Fatal("kcp.Listen:", err)
	}
	defer listener.Close()

	log.Println("Serve UDP at:", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("listen.Accept:", err)
		}
		go handleClientConn(conn)
	}
}

func handleClientConn(conn net.Conn) {
	defer conn.Close()
	var room *Room
	var cid int32
	for {
		var buffer [512]byte

		log.Println("conn.Read:")
		n, err := conn.Read(buffer[:])
		if err != nil {
			log.Println("conn.Read:", err)
			return
		}

		log.Println("conn.Read:", n)

		up := &comm.Up{}
		err = proto.Unmarshal(buffer[:n], up)
		if err != nil {
			log.Println("proto.Unmarshal:", err)
			return
		}

		log.Println("proto.Unmarshal:", up)
		switch cmd := up.M.(type) {
		case *comm.Up_Join:
			log.Println("join:", cmd.Join.Mode)
			room, cid = joinRoom(conn, int(cmd.Join.Mode))

		case *comm.Up_Op:
			if room == nil {
				break
			}
			room.chOp <- ClientKeyCode{
				cid:     cid,
				keycode: int32(cmd.Op.Keycode),
			}
		}
	}
}

type ClientKeyCode struct {
	cid     int32
	keycode int32
}

type Room struct {
	id       int
	mode     int        // 本房间的模式 1/2/3
	num      int        // 本房间当前人数
	conns    []net.Conn // 客户端的通信连接
	ticker   *time.Ticker
	foods    []int32
	snakes   []*comm.Down_Snake
	keycodes []int32

	chJoin chan net.Conn
	chOp   chan ClientKeyCode
}

func (room *Room) run() {
	// 若人数不足，则等待新人加入
	for room.num < room.mode {
		select {
		case conn := <-room.chJoin:
			cid := room.num
			room.num++
			room.conns[cid] = conn
			room.snakes[cid] = &comm.Down_Snake{
				Cid:  int32(cid),
				Body: []int32{41, 40}, // FIXME: 随机初始化
			}
			log.Println("new join:", room.id, cid)
		}
	}

	// 通知所有人 kick-off
	for cid, conn := range room.conns {
		down := &comm.Down{
			M: &comm.Down_Kickoff{
				Kickoff: &comm.Down_DownKickOff{
					Cid:    int32(cid),
					Foods:  room.foods,
					Snakes: room.snakes,
				},
			},
		}
		out, _ := proto.Marshal(down)
		n, err := conn.Write(out)
		log.Println("conn.Write: KickOff:", n, err)
	}

	for {
		select {
		case <-room.ticker.C:

		case op := <-room.chOp:
			log.Println("op:", op)
		}
	}
}

var roomID = 0
var rooms = make([]*Room, 0)

func joinRoom(conn net.Conn, mode int) (room *Room, cid int32) {
	// FIXME: rooms 访问冲突
	// 找到一个空闲的 room
	for _, r := range rooms {
		if r.mode == mode && r.num < mode {
			room = r
			cid = int32(room.num)
			break
		}
	}

	log.Println("joinRoom:", room)
	if room == nil {
		// 没有空闲的 room，新开一个
		roomID++
		room = &Room{
			id:       roomID,
			mode:     mode,
			num:      0,
			conns:    make([]net.Conn, mode),
			ticker:   time.NewTicker(time.Millisecond * 250),
			foods:    []int32{78, 208},
			snakes:   make([]*comm.Down_Snake, mode),
			keycodes: make([]int32, mode),
			chJoin:   make(chan net.Conn),
			chOp:     make(chan ClientKeyCode),
		}
		cid = 0
		rooms = append(rooms, room)

		log.Println("joinRoom: new room:", room.id)
		go room.run()
	}

	room.chJoin <- conn
	return
}
