package shadowinput

import "arion.local/desktop/infrastructure/macos"

func needsAccess() bool { return (macos.Permissions{}).Accessibility() != macos.AuthGranted }
func requestAccess() {
	p := macos.Permissions{}
	p.RequestAccessibility()
	p.Open(macos.PaneAccessibility)
}
