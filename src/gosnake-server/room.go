package main

import (
	"gosnake-server/comm"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
)

type ClientKeyCode struct {
	cid     int32
	keycode int32
}

type Room struct {
	id            int32
	mode          int32              // 本房间的模式 1/2/3
	num           int32              // 本房间当前人数
	width, height int32              // 盘面的宽度、高度
	foods         []int32            // 房间内现有的食物
	conns         []net.Conn         // 每个客户端的通信连接
	snakes        []*comm.Down_Snake // 每条蛇
	keycodes      []int32            // 每个客户端的用户操作
	ticker        *time.Ticker       // 帧驱动定时器
	chJoin        chan net.Conn      // 用于把客户端的通信连接传递给处理协程
	chOp          chan ClientKeyCode // 用于把客户端的用户操作传递给处理协程
}

var roomID int32
var rooms = make([]*Room, 0)

// 把一个 client conn 按照指定的游戏模式加入到一个合适的 room，必要的话创建新的 room
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
			ticker:   time.NewTicker(time.Millisecond * 150),
			chJoin:   make(chan net.Conn),
			chOp:     make(chan ClientKeyCode),
		}
		switch mode {
		case 1:
			room.width = 20
			room.height = 20
		case 2:
			room.width = 30
			room.height = 30
		case 3:
			room.width = 40
			room.height = 40
		}

		cid = 0
		rooms = append(rooms, room)

		log.Println("joinRoom: new room:", room.id, mode)
		go room.run()
	}

	room.chJoin <- conn
	return
}

// 把用户操作转发给处理协程
func (room *Room) postOp(cid, keycode int32) {
	if keycode < 37 || keycode > 40 {
		// 滤掉非法操作
		return
	}
	room.chOp <- ClientKeyCode{
		cid:     cid,
		keycode: keycode,
	}
}

// 本函数在一个单独的协程中运行，用于维护一个 room 的整个生命周期
func (room *Room) run() {
	// 等待新人加入，直到数量达到要求
	for room.num < room.mode {
		log.Printf("room.run: room %d waiting for next %d snake(s)...\n", room.id, room.mode-room.num)

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

	log.Printf("room.run: room %d ready\n", room.id)

	// 初始化食物
	foodNum := room.num + 5
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
					Width:  room.width,
					Height: room.height,
					Foods:  room.foods,
					Snakes: room.snakes,
				},
			},
		}
		out, _ := proto.Marshal(down)
		conn.Write(out)
	}

	// 启动一个协程专门用于发送下行数据包
	chDown := make(chan []byte, 10)
	go room.connsWriter(chDown)

loop:
	for {
		select {
		case op := <-room.chOp:
			// 把用户操作记录到帧数据中
			if room.snakes[op.cid] == nil { // 此蛇已死
				continue
			}
			room.keycodes[op.cid] = op.keycode

		case <-room.ticker.C:
			// 服务端演算
			newFoods, finished := room.playFrame()

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
			chDown <- out

			// 清空 keycodes
			for cid := range room.keycodes {
				room.keycodes[cid] = 0
			}

			if finished {
				break loop
			}
		}
	}

	log.Printf("room.run: room %d over\n", room.id)

	// 通知所有 client 本局结束
	down := &comm.Down{
		M: &comm.Down_Finish{
			Finish: &comm.Down_DownFinish{
				Winer: 0, // FIXME: 获胜者
			},
		},
	}
	out, _ := proto.Marshal(down)
	chDown <- out
	close(chDown)
}

// 本函数在一个单独的协程中运行，专门用于发送本 room 中所有的下行数据包
func (room *Room) connsWriter(chDown chan []byte) {
	for {
		out, ok := <-chDown
		if !ok {
			break
		}
		for cid, conn := range room.conns {
			if conn == nil {
				continue
			}
			conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 100))
			_, err := conn.Write(out)
			if err != nil {
				log.Println("room.connsWriter:", cid, err)
				room.conns[cid] = nil
				conn.Close()
			}
		}
	}

	// 释放 room 资源
	room.release()
}

func (room *Room) release() {
	room.ticker.Stop()
	for _, conn := range room.conns {
		if conn != nil {
			conn.Close()
		}
	}
	for idx, r := range rooms {
		if r == room {
			rooms = append(rooms[0:idx], rooms[idx+1:]...)
			break
		}
	}
	log.Println("room.release:", room.id)
}

// 根据帧数据演算本帧的结果
func (room *Room) playFrame() (newFoods []int32, finished bool) {
	foodNum := 0 // 本帧中被吃掉的食物数量
	finished = true
loop:
	for cid, snake := range room.snakes {
		if snake == nil { // 此蛇已死
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
			if other == nil { // 此蛇已死
				continue
			}
			for _, body := range other.Body {
				if head == body {
					room.snakes[cid] = nil
					continue loop
				}
			}
		}

		// 此蛇未死，本局没有结束
		finished = false

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

	return
}

// 在本 room 中产生一粒食物
func (room *Room) newFood() int32 {
loop:
	for {
		food := rand.Int31() % (room.width * room.height)
		// 不能在蛇身上
		for _, snake := range room.snakes {
			if snake == nil { // 此蛇已死
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

// 在本 room 中产生一条蛇，用于初始化
func (room *Room) newSnakeBody() []int32 {
	h := room.height / room.mode
	y := (rand.Int31() % (h - 4)) + 2
	x := (rand.Int31() % (room.width - 10)) + 5
	head := (y+h*(room.num-1))*room.width + x
	tail := head - 1 // TODO: 再随机一点
	return []int32{head, tail}
}
