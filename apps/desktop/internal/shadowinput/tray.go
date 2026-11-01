package shadowinput

import (
	"arion.local/desktop/ui/shared/copy"
	"arion.local/desktop/ui/shared/shadowbrand"
	state "arion.local/desktop/ui/shared/writing"
	"github.com/egoist/mygo"
)

func (a *app) installTray() {
	tray, err := mygo.NewTray(mygo.TrayOptions{Icon: shadowbrand.TrayPNG, IconIsTemplate: false, ToolTip: state.Name})
	if err != nil {
		a.s.Failure = copy.WriteTrayError
		return
	}
	tray.OnClick(a.capture)
	menu := mygo.NewMenu([]*mygo.MenuItem{{Label: state.Name, Click: func(*mygo.MenuItem, *mygo.Window) { a.show() }}, mygo.Separator(), {Role: mygo.RoleQuit}})
	tray.OnRightClick(func() { tray.SetMenu(menu); defer tray.SetMenu(nil); tray.PopUpMenu() })
	mygo.App.OnWillQuit(func(*mygo.QuitEvent) { tray.Destroy() })
}
