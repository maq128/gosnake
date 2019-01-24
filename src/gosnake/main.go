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
	w         *astilectron.Window
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
					bootstrap.SendMessage(w, "about", nil)
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
			w = ws[0]
			go engine()
			return nil
		},
	}
	if err := bootstrap.Run(options); err != nil {
		astilog.Fatal(errors.Wrap(err, "running bootstrap failed"))
	}
}

func handleMessages(_ *astilectron.Window, m bootstrap.MessageIn) (payload interface{}, err error) {
	astilog.Debug("handleMessages:", m.Name, m.Payload)
	switch m.Name {
	case "keydown":
		// 收到用户操作，转给 engine
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
	Food   int     `json:"food"`
	Snakes []Snake `json:"snakes"`
}

type FrameParams struct {
	Num     int `json:"num"`
	KeyCode int `json:"keycode"`
}

func engine() {
	// time.Sleep(time.Second)
	bootstrap.SendMessage(w, "kick-off", KickOffParams{
		Food: 43,
		Snakes: []Snake{
			Snake{
				ID:   1,
				Body: []int{41, 40},
			},
		},
	})
	ticker := time.NewTicker(time.Millisecond * 250)
	keyCode := 0
	for {
		select {
		case <-ticker.C:
			bootstrap.SendMessage(w, "frame", FrameParams{
				Num:     0,
				KeyCode: keyCode,
			})
		case kc := <-chKeyCode:
			keyCode = kc
		}
	}
}
