// Package shadowbrand owns shadowinput's yellow SI mark. Generate platform
// assets from mark.svg with go run ./apps/desktop/cmd/shadowinput-icon.
package shadowbrand

import (
	_ "embed"
	"github.com/egoist/mygo/ui"
)

//go:embed mark.svg
var svg []byte

var Mark = ui.MustParseSVG(svg)

//go:embed icon-32.png
var TrayPNG []byte

//go:embed icon-512.png
var AppPNG []byte
