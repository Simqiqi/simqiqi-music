// Command icon draws the app icons: the cat that fills them, its layer in
// AppIcon.icon, the Icon Composer document macOS 26 and later render in
// Liquid Glass, and resources/icon.png, the same icon flattened, for older
// macOS and Windows; and from AppIcon-dev.icon, cream where the app's is
// red, resources/icon-dev.png, the icon of mygo dev builds.
//
// Run it from the repository root:
//
//	go run ./icon
//
// The cat is a picture rather than drawn paths, so this needs no Xcode:
// icon/cat.png is scaled to the grid and laid on the background, and
// icon.png is that same drawing flattened, so the document and the PNG
// cannot disagree. The background of each document, and its dark-mode
// colors, are still edited by hand in icon.json or in Icon Composer.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// size is the grid the icons are drawn on: Icon Composer's full-bleed
// canvas, which macOS masks to its own shape, and the size of the PNGs.
const size = 1024

// portrait is the cat, keyed off the color it was drawn on, black on
// nothing. It fills the grid the way the picture fills its own frame.
const portrait = "icon/cat.png"

// icons are the documents drawn, with the PNG rendered from each and the
// two ends of its background, from the top of the icon down.
var icons = []struct {
	document, output string
	top, bottom      string
}{
	// Red, with the cat sitting on it.
	{"icon/AppIcon.icon", "resources/icon.png", "#FF3357", "#E50133"},
	// Cream, so a dev build's icon is not taken for the app's.
	{"icon/AppIcon-dev.icon", "resources/icon-dev.png", "#F9EEE3", "#E4C7AC"},
}

func main() {
	src, err := os.Open(portrait)
	if err != nil {
		log.Fatal(err)
	}
	picture, err := png.Decode(src)
	src.Close()
	if err != nil {
		log.Fatal(err)
	}
	cat := fit(picture, size)
	for _, ic := range icons {
		// The layer the document draws: the cat alone, on nothing.
		if err := write(filepath.Join(ic.document, "Assets", "cat.png"), cat); err != nil {
			log.Fatal(err)
		}
		icon := ground(ic.top, ic.bottom)
		over(icon, cat, image.Pt(0, 0))
		if err := write(ic.output, icon); err != nil {
			log.Fatal(err)
		}
	}
}

// fit scales src to a size x size frame, sampling bilinearly: the picture
// is drawn a touch larger than it was taken, so its edges, feathered when
// it was keyed, stay soft.
func fit(src image.Image, n int) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewNRGBA(image.Rect(0, 0, n, n))
	for y := 0; y < n; y++ {
		sy := (float64(y)+0.5)*float64(h)/float64(n) - 0.5
		y0 := int(math.Floor(sy))
		fy := sy - float64(y0)
		for x := 0; x < n; x++ {
			sx := (float64(x)+0.5)*float64(w)/float64(n) - 0.5
			x0 := int(math.Floor(sx))
			fx := sx - float64(x0)
			// The four corners about the sample, weighted. Colors are
			// summed premultiplied by their alpha, so that a transparent
			// neighbor adds no color of its own.
			var r, g, bl, a float64
			for _, c := range [...]struct {
				dx, dy int
				w      float64
			}{
				{0, 0, (1 - fx) * (1 - fy)}, {1, 0, fx * (1 - fy)},
				{0, 1, (1 - fx) * fy}, {1, 1, fx * fy},
			} {
				px := b.Min.X + min(max(x0+c.dx, 0), w-1)
				py := b.Min.Y + min(max(y0+c.dy, 0), h-1)
				p := color.NRGBAModel.Convert(src.At(px, py)).(color.NRGBA)
				k := float64(p.A) / 255 * c.w
				r, g, bl, a = r+float64(p.R)*k, g+float64(p.G)*k, bl+float64(p.B)*k, a+k
			}
			o := dst.PixOffset(x, y)
			if a > 0 {
				dst.Pix[o] = clamp8(r / a)
				dst.Pix[o+1] = clamp8(g / a)
				dst.Pix[o+2] = clamp8(bl / a)
			}
			dst.Pix[o+3] = clamp8(a * 255)
		}
	}
	return dst
}

// ground is the background of an icon: top at the top, bottom at the foot,
// a straight line between them, and opaque to the very corner, so that
// nothing of the desktop shows through the icon.
func ground(top, bottom string) *image.NRGBA {
	a, b := rgb(top), rgb(bottom)
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		t := float64(y) / float64(size-1)
		c := color.NRGBA{
			clamp8(255 * (a[0] + (b[0]-a[0])*t)),
			clamp8(255 * (a[1] + (b[1]-a[1])*t)),
			clamp8(255 * (a[2] + (b[2]-a[2])*t)),
			255,
		}
		for x := 0; x < size; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

// over lays src on dst at at, source over: a half-transparent pixel of the
// cat lets the background through rather than punching a hole in it.
func over(dst, src *image.NRGBA, at image.Point) {
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			s := src.NRGBAAt(x, y)
			if s.A == 0 {
				continue
			}
			px, py := at.X+x, at.Y+y
			if !(image.Pt(px, py).In(dst.Bounds())) {
				continue
			}
			k := float64(s.A) / 255
			d := dst.NRGBAAt(px, py)
			dst.SetNRGBA(px, py, color.NRGBA{
				clamp8(float64(s.R)*k + float64(d.R)*(1-k)),
				clamp8(float64(s.G)*k + float64(d.G)*(1-k)),
				clamp8(float64(s.B)*k + float64(d.B)*(1-k)),
				255,
			})
		}
	}
}

// rgb reads a color written #RRGGBB, as the backgrounds are.
func rgb(s string) [3]float64 {
	v, err := strconv.ParseUint(strings.TrimPrefix(s, "#"), 16, 32)
	if err != nil {
		log.Fatalf("color %q: %v", s, err)
	}
	return [3]float64{float64(v>>16&0xFF) / 255, float64(v>>8&0xFF) / 255, float64(v&0xFF) / 255}
}

func clamp8(v float64) uint8 {
	return uint8(min(max(math.Round(v), 0), 255))
}

func write(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return fmt.Errorf("%s: %w", path, err)
	}
	return f.Close()
}
