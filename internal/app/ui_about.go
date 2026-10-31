package app

import (
	"bytes"
	_ "embed"
	"image/jpeg"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// Who keeps this: the name shown at the foot of the settings, and where
// their work is found.
const (
	devName = "柒柒"
	devHome = "https://github.com/Simqiqi"
)

//go:embed assets/developer.jpg
var developerPhoto []byte

var (
	developerOnce sync.Once
	developerBmp  *ui.Bitmap
)

// developerAvatar is the developer's portrait, read from what the app
// carries of it, once.
func developerAvatar() *ui.Bitmap {
	developerOnce.Do(func() {
		if img, err := jpeg.Decode(bytes.NewReader(developerPhoto)); err == nil {
			developerBmp = ui.NewBitmap(img)
		}
	})
	return developerBmp
}

// developerCard is the line at the foot of the settings: a portrait where
// one would be, and the name it stands for. Nothing else — the whole of
// it leads to where the work is.
func (a *app) developerCard(c *ui.Context, p palette) {
	b := ui.ButtonBase(c.Key("dev")).Padding(14, 16).Gap(12).Radius(11).AlignItems(ui.Center).Justify(ui.Start).
		Label("打开 " + devName + " 的 GitHub 主页")
	bg := ui.Transparent
	if b.Hovered() {
		bg = p.hover
	}
	b.Background(b.AnimateColor("bg", bg, ui.Animation{Duration: 120 * time.Millisecond}))
	if b.Clicked() {
		go mygo.Shell.OpenExternal(devHome)
	}
	b.Children(func() {
		badge := ui.Box(c).Size(44, 44).Radius(22).Center().Background(p.hover)
		badge.Children(func() {
			// The portrait fills the round, whatever the shape it came
			// in; failing it, the first letter of the name stands in.
			if bmp := developerAvatar(); bmp != nil {
				ui.Image(c, bmp).Absolute().Top(0).Left(0).Right(0).Bottom(0).Fit(ui.Cover).Radius(22)
			} else {
				ui.Text(c, "柒").FontSize(20).FontWeight(700).TextColor(p.muted)
			}
		})
		ui.Text(c, devName).FontSize(14.5).FontWeight(600).TextColor(p.text)
	})
}
