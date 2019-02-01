package main

import (
	"gosnake-server/comm"
	"log"
	"math/rand"
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

		// 接收并解析来自 client 的数据包
		n, err := conn.Read(buffer[:])
		if err != nil {
			log.Println("conn.Read:", err)
			return
		}

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
			if room == nil || cmd.Op.Keycode < 37 || cmd.Op.Keycode > 40 {
				break
			}
			room.chOp <- ClientKeyCode{
				cid:     cid,
				keycode: cmd.Op.Keycode,
			}
		}
	}
}

type ClientKeyCode struct {
	cid     int32
	keycode int32
}

type Room struct {
	id            int32
	mode          int32 // 本房间的模式 1/2/3
	num           int32 // 本房间当前人数
	width, height int32 // 盘面的宽度、高度
	foods         []int32
	conns         []net.Conn // 客户端的通信连接
	snakes        []*comm.Down_Snake
	keycodes      []int32
	ticker        *time.Ticker

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
				Body: room.newSnakeBody(),
			}
		}
	}

	// 初始化食物
	foodNum := room.num + 2
	for foodNum > 0 {
		foodNum--
		room.newFood()
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
		conn.Write(out)
	}

	for {
		select {
		case op := <-room.chOp:
			// 把用户操作记录到帧数据中
			room.keycodes[op.cid] = op.keycode

		case <-room.ticker.C:
			// 服务端演算
			newFoods, finish := room.playFrame()

			// 推送帧数据
			down := &comm.Down{
				M: &comm.Down_Frame{
					Frame: &comm.Down_DownFrame{
						Foods:    newFoods,
						Keycodes: room.keycodes,
					},
				},
			}
			out, _ := proto.Marshal(down)
			for _, conn := range room.conns {
				conn.Write(out)
			}

			// 清空 keycodes
			for cid := range room.keycodes {
				room.keycodes[cid] = 0
			}

			// 若已结束，通知所有 client
			if finish {
				down := &comm.Down{
					M: &comm.Down_Finish{
						Finish: &comm.Down_DownFinish{
							Winer: 0,
						},
					},
				}
				out, _ := proto.Marshal(down)
				for _, conn := range room.conns {
					conn.Write(out)
				}

				// 释放 room 资源
				room.release()
				return
			}
		}
	}
}

func (room *Room) release() {
	room.ticker.Stop()
	for _, conn := range room.conns {
		conn.Close()
	}
	for idx, r := range rooms {
		if r == room {
			rooms = append(rooms[0:idx], rooms[idx+1:]...)
			break
		}
	}
	log.Println("room.release:", room.id)
}

func (room *Room) playFrame() (newFoods []int32, finish bool) {
	foodNum := 0 // 本帧中被吃掉的食物数量
loop:
	for cid, snake := range room.snakes {
		if snake == nil { // 该蛇已死
			continue
		}
		// 计算出蛇头的位置
		origDir := snake.Body[0] - snake.Body[1]
		newDir := origDir
		switch room.keycodes[cid] {
		case 37:
			newDir = -1
		case 38:
			newDir = -room.width
		case 39:
			newDir = 1
		case 40:
			newDir = room.width
		}
		if newDir+origDir == 0 {
			newDir = origDir
		}
		head := snake.Body[0] + newDir

		// 判断是否撞到墙壁
		if head < 0 || head >= room.width*room.height || (newDir == 1 && head%room.width == 0) || (newDir == -1 && head%room.width == room.width-1) {
			room.snakes[cid] = nil
			continue
		}

		// 判断是否撞到自己或其它蛇身
		for _, other := range room.snakes {
			if other == nil { // 该蛇已死
				continue
			}
			for _, body := range other.Body {
				if head == body {
					room.snakes[cid] = nil
					continue loop
				}
			}
		}

		// 蛇头并入身体
		snake.Body = append([]int32{head}, snake.Body...)

		// 判断是否吃到食物
		hit := -1
		for idx, food := range room.foods {
			if head == food {
				hit = idx
				break
			}
		}

		if hit >= 0 {
			// 吃到，清除食物
			room.foods = append(room.foods[:hit], room.foods[hit+1:]...)
			foodNum++
		} else {
			// 没有吃到，清除蛇尾
			snake.Body = snake.Body[0 : len(snake.Body)-1]
		}
	}

	// 产生新食物
	for foodNum > 0 {
		foodNum--
		newFoods = append(newFoods, room.newFood())
	}

	// 判断本局是否已经结束
	finish = true
	for _, snake := range room.snakes {
		if snake != nil {
			finish = false
			return
		}
	}
	return
}

func (room *Room) newFood() int32 {
loop:
	for {
		food := rand.Int31() % (room.width * room.height)
		// 不能在蛇身上
		for _, snake := range room.snakes {
			if snake == nil { // 该蛇已死
				continue
			}
			for _, body := range snake.Body {
				if food == body {
					continue loop
				}
			}
		}
		// 不能在已有的食物上
		for _, n := range room.foods {
			if food == n {
				continue loop
			}
		}
		room.foods = append(room.foods, food)
		return food
	}
}

func (room *Room) newSnakeBody() []int32 {
	h := room.height / room.mode
	y := (rand.Int31() % (h - 4)) + 2
	x := (rand.Int31() % (room.width - 10)) + 5
	head := (y+h*(room.num-1))*room.width + x
	tail := head - 1 // TODO: 再随机一点
	return []int32{head, tail}
}

var roomID int32
var rooms = make([]*Room, 0)

func joinRoom(conn net.Conn, mode int32) (room *Room, cid int32) {
	// 找到一个空闲的 room
	// FIXME: rooms 访问冲突
	for _, r := range rooms {
		if r.mode == mode && r.num < mode {
			room = r
			cid = room.num
			break
		}
	}

	if room == nil {
		// 没有空闲的 room，新开一个
		roomID++
		room = &Room{
			id:       roomID,
			mode:     mode,
			num:      0,
			width:    20,
			height:   20,
			foods:    make([]int32, 0),
			conns:    make([]net.Conn, mode),
			snakes:   make([]*comm.Down_Snake, mode),
			keycodes: make([]int32, mode),
			ticker:   time.NewTicker(time.Millisecond * 250),
			chJoin:   make(chan net.Conn),
			chOp:     make(chan ClientKeyCode),
		}
		switch mode {
		case 2:
			room.width = 30
			room.height = 30
		case 3:
			room.width = 40
			room.height = 40
		}

		cid = 0
		rooms = append(rooms, room)

		log.Println("joinRoom: new room:", room.id)
		go room.run()
	}

	room.chJoin <- conn
	return
}
