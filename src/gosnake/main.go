package main

import (
	"flag"
	"strconv"
	"time"

	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"
	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

// Vars
var (
	debug     = flag.Bool("d", false, "enables the debug mode")
	AppName   = "GoSnake"
	chKeyCode = make(chan int, 10) // 传递用户操作按键给 engine
	chMode    = make(chan int, 1)  // 传递 mode 信号给 engine
	chExit    = make(chan int)     // 传递退出信号给 engine
	mainWin   *astilectron.Window
)

func main() {
	// Init
	flag.Parse()
	astilog.FlagInit()

	// Run bootstrap
	options := bootstrap.Options{
		Debug:         *debug,
		Asset:         Asset,
		AssetDir:      AssetDir,
		RestoreAssets: RestoreAssets,
		AstilectronOptions: astilectron.Options{
			AppName:            AppName,
			AppIconDarwinPath:  "resources/icon.icns",
			AppIconDefaultPath: "resources/icon.png",
		},
		MenuOptions: []*astilectron.MenuItemOptions{{
			Label: astilectron.PtrStr("File"),
			SubMenu: []*astilectron.MenuItemOptions{{
				Label: astilectron.PtrStr("About"),
				OnClick: func(e astilectron.Event) (deleteListener bool) {
					bootstrap.SendMessage(mainWin, "about", nil)
					return
				},
			}, {
				Role: astilectron.MenuItemRoleClose,
			}},
		}},
		Windows: []*bootstrap.Window{{
			Homepage:       "index.html",
			MessageHandler: handleMessages,
			Options: &astilectron.WindowOptions{
				BackgroundColor: astilectron.PtrStr("#fff"),
				Center:          astilectron.PtrBool(true),
				Height:          astilectron.PtrInt(500),
				Width:           astilectron.PtrInt(500),
			},
		}},
		OnWait: func(_ *astilectron.Astilectron, ws []*astilectron.Window, _ *astilectron.Menu, _ *astilectron.Tray, _ *astilectron.Menu) error {
			mainWin = ws[0]
			// Astilectron 已经 ready，启动 engine
			go engine()
			return nil
		},
		Adapter: func(a *astilectron.Astilectron) {
			// Astilectron 已经结束，通知 engine 退出
			a.On(astilectron.EventNameAppCrash, func(e astilectron.Event) (deleteListener bool) {
				close(chExit)
				return
			})
		},
	}
	if err := bootstrap.Run(options); err != nil {
		astilog.Fatal(errors.Wrap(err, "running bootstrap failed"))
	}
	time.Sleep(time.Second)
}

func handleMessages(_ *astilectron.Window, m bootstrap.MessageIn) (payload interface{}, err error) {
	// astilog.Debug("handleMessages:", m.Name, m.Payload)
	switch m.Name {
	case "start":
		// 从 Astilectron 收到 start 通知，转给 engine
		mode, err := strconv.Atoi(string(m.Payload))
		if err != nil || mode < 1 || mode > 3 {
			break
		}
		chMode <- mode

	case "keydown":
		// 从 Astilectron 收到用户按键操作，转给 engine
		kc, err := strconv.Atoi(string(m.Payload))
		if err != nil {
			break
		}
		chKeyCode <- kc
	}
	return
}

type Snake struct {
	ID   int   `json:"id"`
	Body []int `json:"body"`
}

type KickOffParams struct {
	Foods  []int   `json:"foods"`
	Snakes []Snake `json:"snakes"`
	MyID   int     `json:"myid"`
}

type FrameParams struct {
	Num      int   `json:"num"`
	KeyCodes []int `json:"keycodes"`
	Foods    []int `json:"foods"`
}

func engine() {
	mode := 0
	ticker := time.NewTicker(time.Millisecond * 250)
	var keyCodes []int
	var myID int
loop:
	for {
		select {
		case m := <-chMode:
			if m == 1 {
				// 启动 1P 模式
				mode = 1
				myID = 0
				keyCodes = make([]int, 1)
				bootstrap.SendMessage(mainWin, "kick-off", KickOffParams{
					Foods: []int{78, 208},
					Snakes: []Snake{
						Snake{
							ID:   myID,
							Body: []int{41, 40},
						},
					},
					MyID: myID,
				})
			}

		case kc := <-chKeyCode:
			keyCodes[myID] = kc

		case <-ticker.C:
			if mode <= 0 {
				break
			}
			bootstrap.SendMessage(mainWin, "frame", FrameParams{
				Num:      0,
				KeyCodes: keyCodes,
				Foods:    []int{},
			})
			for id := 0; id < len(keyCodes); id++ {
				keyCodes[id] = 0
			}

		case <-chExit:
			break loop
		}
	}
}
