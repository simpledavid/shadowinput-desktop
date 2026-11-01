package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// macOS templates use alpha, not RGB. Opaque white glyphs become a solid tile.
func TestMenuTemplatePreservesCutouts(t *testing.T) {
	for _, size := range []int{16, 32} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			file, err := os.Open(filepath.Join("../../ui/shared/shadowbrand", fmt.Sprintf("menu-%d.png", size)))
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			img, err := png.Decode(file)
			if err != nil {
				t.Fatal(err)
			}
			for _, point := range [][2]int{{0, 0}, {6, 5}, {11, 5}} {
				_, _, _, alpha := img.At(point[0]*size/16, point[1]*size/16).RGBA()
				if alpha != 0 {
					t.Errorf("template cutout at %v is opaque", point)
				}
			}
			_, _, _, alpha := img.At(2*size/16, 8*size/16).RGBA()
			if alpha != 65535 {
				t.Fatal("template tile is missing")
			}
		})
	}
}
