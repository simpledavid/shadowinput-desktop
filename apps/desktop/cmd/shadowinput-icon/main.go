// Command shadowinput-icon exports platform icon containers from the canonical
// SVG already used by the website. It does not redraw or alter the brand mark.
package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/egoist/mygo/ui"
)

func main() {
	check := flag.Bool("check", false, "verify generated assets without writing")
	flag.Parse()
	base := "apps/desktop/ui/shared/shadowbrand"
	svg, err := os.ReadFile(filepath.Join(base, "mark.svg"))
	if err != nil {
		log.Fatal("run from the repository root: ", err)
	}
	mark := ui.MustParseSVG(svg)
	images := map[int]*image.RGBA{}
	pngs := map[int][]byte{}
	for _, size := range []int{16, 24, 32, 48, 64, 128, 180, 256, 512, 1024} {
		img := ui.NewTester(func(c *ui.Context) { ui.Image(c, mark).Size(float32(size), float32(size)) }, size, size).Image()
		var b bytes.Buffer
		if err := png.Encode(&b, img); err != nil {
			log.Fatal(err)
		}
		images[size], pngs[size] = img, b.Bytes()
	}
	assets := map[string][]byte{}
	for _, size := range []int{16, 32, 256, 512, 1024} {
		assets[filepath.Join(base, fmt.Sprintf("icon-%d.png", size))] = pngs[size]
	}
	assets[filepath.Join(base, "shadowinput.ico")] = ico(images, pngs)
	assets[filepath.Join(base, "shadowinput.icns")] = icns(pngs)
	assets["apps/shadowinput-site/public/favicon.svg"] = svg
	assets["apps/shadowinput-site/public/favicon.ico"] = assets[filepath.Join(base, "shadowinput.ico")]
	assets["apps/shadowinput-site/public/apple-touch-icon.png"] = pngs[180]
	for name, data := range assets {
		if *check {
			current, err := os.ReadFile(name)
			if err != nil || !bytes.Equal(current, data) {
				log.Fatalf("icon asset is stale: %s; run go run ./apps/desktop/cmd/shadowinput-icon", name)
			}
		} else if err := os.WriteFile(name, data, 0644); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("shadowinput: %d icon assets %s\n", len(assets), map[bool]string{true: "verified", false: "exported"}[*check])
}

func put(w *bytes.Buffer, value any) {
	if err := binary.Write(w, binary.LittleEndian, value); err != nil {
		panic(err)
	}
}

// Small ICO entries use DIBs for Windows shell/installer compatibility. The
// 256px entry uses PNG. Transparency is carried by alpha and the AND mask.
func ico(images map[int]*image.RGBA, pngs map[int][]byte) []byte {
	sizes := []int{16, 24, 32, 48, 64, 128, 256}
	entries := make([][]byte, len(sizes))
	for i, n := range sizes {
		if n == 256 {
			entries[i] = pngs[n]
			continue
		}
		var b bytes.Buffer
		maskStride := ((n + 31) / 32) * 4
		put(&b, uint32(40))
		put(&b, int32(n))
		put(&b, int32(n*2))
		put(&b, uint16(1))
		put(&b, uint16(32))
		put(&b, uint32(0))
		put(&b, uint32(n*n*4+maskStride*n))
		put(&b, [4]uint32{})
		mask := make([]byte, maskStride*n)
		for row := 0; row < n; row++ {
			for x := 0; x < n; x++ {
				c := color.NRGBAModel.Convert(images[n].At(x, n-1-row)).(color.NRGBA)
				b.Write([]byte{c.B, c.G, c.R, c.A})
				if c.A == 0 {
					mask[row*maskStride+x/8] |= 1 << uint(7-x%8)
				}
			}
		}
		b.Write(mask)
		entries[i] = b.Bytes()
	}
	var out bytes.Buffer
	put(&out, [3]uint16{0, 1, uint16(len(sizes))})
	offset := 6 + 16*len(sizes)
	for i, n := range sizes {
		out.Write([]byte{byte(n % 256), byte(n % 256), 0, 0})
		put(&out, [2]uint16{1, 32})
		put(&out, [2]uint32{uint32(len(entries[i])), uint32(offset)})
		offset += len(entries[i])
	}
	for _, entry := range entries {
		out.Write(entry)
	}
	return out.Bytes()
}

func icns(pngs map[int][]byte) []byte {
	var body bytes.Buffer
	for _, e := range []struct {
		kind string
		size int
	}{{"icp4", 16}, {"icp5", 32}, {"ic11", 32}, {"ic12", 64}, {"ic07", 128}, {"ic13", 256}, {"ic08", 256}, {"ic14", 512}, {"ic09", 512}, {"ic10", 1024}} {
		body.WriteString(e.kind)
		_ = binary.Write(&body, binary.BigEndian, uint32(8+len(pngs[e.size])))
		body.Write(pngs[e.size])
	}
	var out bytes.Buffer
	out.WriteString("icns")
	_ = binary.Write(&out, binary.BigEndian, uint32(body.Len()+8))
	out.Write(body.Bytes())
	return out.Bytes()
}
