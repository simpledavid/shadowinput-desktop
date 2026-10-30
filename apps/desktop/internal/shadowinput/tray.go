package shadowinput

import (
	"arion.local/desktop/ui/shared/copy"
	state "arion.local/desktop/ui/shared/writing"
	"bytes"
	"github.com/egoist/mygo"
	"image"
	"image/color"
	"image/png"
	"runtime"
)

func (a *app) installTray() {
	tray, err := mygo.NewTray(mygo.TrayOptions{Icon: trayIcon(), IconIsTemplate: true, ToolTip: state.Name})
	if err != nil {
		a.s.Failure = copy.WriteTrayError
		return
	}
	tray.OnClick(a.capture)
	menu := mygo.NewMenu([]*mygo.MenuItem{{Label: state.Name, Click: func(*mygo.MenuItem, *mygo.Window) { a.show() }}, mygo.Separator(), {Role: mygo.RoleQuit}})
	tray.OnRightClick(func() { tray.SetMenu(menu); defer tray.SetMenu(nil); tray.PopUpMenu() })
	mygo.App.OnWillQuit(func(*mygo.QuitEvent) { tray.Destroy() })
}

func trayIcon() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	ink := color.NRGBA{R: 245, G: 192, B: 48, A: 255}
	if runtime.GOOS == "darwin" {
		ink = color.NRGBA{A: 255}
	}
	for y := 7; y < 26; y++ {
		for x := 2; x < 30; x++ {
			border := x == 2 || x == 29 || y == 7 || y == 25
			key := y >= 11 && y <= 17 && (y-11)%4 < 2 && x >= 6 && x < 26 && (x-6)%5 < 3
			space := y >= 21 && y < 23 && x >= 9 && x <= 23
			if border || key || space {
				img.Set(x, y, ink)
			}
		}
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}
