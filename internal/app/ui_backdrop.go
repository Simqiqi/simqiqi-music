package app

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// backdropPicker is Now Playing's own choice of backdrop: a quiet pill
// naming the one it wears, which opens onto all of them, each drawn small
// in the colors of the song playing. It sits at the top right of the view,
// out of the way of the song, where it is there to be found when the
// backdrop is what the listener is looking at.
func (a *app) backdropPicker(c *ui.Context, tint ui.Color, key string) ui.Element {
	style := a.settings.backdropStyle()
	chosen := style
	sel := ui.SelectBase(c.Key(key), &chosen)
	// As tall as the pane tabs, and their weight, so that the two read
	// as buttons of one kind.
	tr := sel.Trigger.Height(npTabsHeight).Padding(0, 9, 0, 12).Gap(6).Radius(999).AlignItems(ui.Center).Label("播放页背景")
	fill := white.Alpha(0.08)
	if tr.Hovered() || sel.Open() {
		fill = white.Alpha(0.18)
	}
	tr.Background(tr.AnimateColor("bg", fill, ui.Animation{Duration: 120 * time.Millisecond}))
	chevron := ui.Pose{}
	if sel.Open() {
		chevron.Rotate = 180
	}
	tr.Children(func() {
		ui.Text(c, backdropLabel(style)).FontSize(13).FontWeight(600).TextColor(white.Alpha(0.75)).SingleLine()
		ui.Icon(c, iconDown).Size(11, 11).TextColor(white.Alpha(0.5)).AnimatePose(chevron, ui.Snappy)
	})
	sel.Popup(func(panel ui.Element) {
		panel.MinWidth(156).Padding(5).Radius(14).Background(ui.RGBA(24, 24, 28, 0.96)).
			Border(1, white.Alpha(0.12)).Shadow(0, 18, 44, -10, ui.RGBA(0, 0, 0, 0.45)).
			Transition(ui.ElementTransition{Animation: &ui.Snappy, Enter: &ui.Motion{Scale: 0.97, Y: -4}, Exit: &ui.Motion{Opacity: 0}})
		for _, b := range backdrops {
			choice := b
			it := sel.Item(choice.id).Padding(5, 8).Radius(9).Gap(9).AlignItems(ui.Center)
			if it.Highlighted() {
				it.Background(white.Alpha(0.08))
			}
			picked := choice.id == style
			it.Children(func() {
				ui.Box(c).Size(24, 24).Radius(7).Clip().
					Draw(func(p *ui.Painter, r ui.Rect) { backdropSwatch(p, r, choice.id, tint) })
				ui.Text(c, choice.label).FontSize(13).Grow(1).TextColor(white.Alpha(0.92)).SingleLine()
				if picked {
					ui.Icon(c, iconCheck).Size(12, 12).TextColor(white.Alpha(0.85))
				}
			})
		}
	})
	if chosen != style {
		// Only the backdrop changed: nothing to apply of the appearance,
		// so it is kept and painted again.
		a.settings.setBackdrop(chosen)
		a.settings.save()
		a.redraw()
	}
	return sel.Trigger
}

// backdropLabel returns what the settings call a backdrop by its id.
func backdropLabel(style string) string {
	for _, b := range backdrops {
		if b.id == style {
			return b.label
		}
	}
	return backdrops[0].label
}

// backdropSwatch paints style small in box, in the colors of a song of
// tint: the gradient of the cover, its fog, its sky of stars, its hills
// or its water.
func backdropSwatch(p *ui.Painter, box ui.Rect, style string, tint ui.Color) {
	switch style {
	case "fog":
		deep, fog := fogColors(tint)
		p.Fill(box, deep, 7)
		// Puffs of it, thicker above the middle.
		p.Fill(ui.Rect{X: box.X + box.W*0.12, Y: box.Y + box.H*0.34, W: box.W * 0.52, H: box.H * 0.4}, fog.Alpha(0.34), box.H)
		p.Fill(ui.Rect{X: box.X + box.W*0.46, Y: box.Y + box.H*0.2, W: box.W * 0.44, H: box.H * 0.44}, fog.Alpha(0.26), box.H)
		p.Fill(ui.Rect{X: box.X + box.W*0.26, Y: box.Y + box.H*0.58, W: box.W * 0.5, H: box.H * 0.3}, fog.Alpha(0.22), box.H)
	case "stars":
		deep, glow := starColors(tint)
		p.Fill(box, deep, 7)
		p.Fill(ui.Rect{X: box.X + box.W*0.4, Y: box.Y + box.H*0.36, W: box.W * 0.66, H: box.H * 0.6}, glow.Alpha(0.3), box.H)
		// A few of them, of two sizes, where they fall.
		for _, s := range [][3]float32{{0.2, 0.28, 2.6}, {0.42, 0.16, 1.8}, {0.66, 0.34, 2.0}, {0.78, 0.62, 1.6}, {0.3, 0.66, 1.8}, {0.54, 0.78, 2.2}, {0.86, 0.3, 1.5}, {0.14, 0.52, 1.5}, {0.62, 0.12, 1.4}, {0.36, 0.44, 1.4}} {
			p.Fill(ui.Rect{X: box.X + box.W*s[0] - s[2]/2, Y: box.Y + box.H*s[1] - s[2]/2, W: s[2], H: s[2]}, white.Alpha(0.95), s[2]/2)
		}
	case "terrain":
		deep, glow := nightColors(tint)
		p.Fill(box, deep, 7)
		// Three skylines of the spectrum, one behind another, the far
		// one faintest.
		for _, l := range [][3]float32{{0.52, 0.55, 0.45}, {0.66, 0.75, 0.7}, {0.82, 0.95, 1}} {
			y := box.Y + box.H*l[0]
			for i := range 8 {
				h := box.H * 0.10 * (0.35 + float32((i*29+int(l[0]*100))%5)/4)
				p.Fill(ui.Rect{
					X: box.X + box.W*float32(i)/8,
					Y: y - h,
					W: box.W/8 - 0.4,
					H: 0.9,
				}, glow.Alpha(l[1]), 0)
			}
		}
	case "ripple":
		deep, glow := nightColors(tint)
		p.FillGradient(box, ui.LinearGradient{Angle: 180, Oklab: true, Stops: []ui.GradientStop{
			{Color: ui.RGBA(tint.R, tint.G, tint.B, 0.55), At: 0}, {Color: deep, At: 1},
		}}, 7)
		// The light of the rings, seen across the water.
		for i, y := range []float32{0.34, 0.52, 0.70} {
			p.Fill(ui.Rect{X: box.X + box.W*0.08, Y: box.Y + box.H*y, W: box.W * 0.84, H: 0.9}, glow.Alpha(0.42-float32(i)*0.10), 0)
		}
	case "vinyl":
		deep, glow := nightColors(tint)
		p.Fill(box, deep.Alpha(1), 7)
		// The disc, standing out to the right, and the light on it.
		r := box.H * 0.34
		cx, cy := box.X+box.W*0.62, box.Y+box.H*0.54
		p.Fill(ui.Rect{X: cx - r, Y: cy - r, W: 2 * r, H: 2 * r}, ui.RGBA(18, 18, 21, 255), r)
		p.Fill(ui.Rect{X: cx - r*0.62, Y: cy - r*0.62, W: 1.24 * r, H: 1.24 * r}, ui.RGBA(30, 31, 35, 255), r*0.62)
		p.Fill(ui.Rect{X: cx - r + 1, Y: cy - r*0.9, W: 2*r - 2, H: 0.7}, glow.Alpha(0.5), 0)
		p.Fill(ui.Rect{X: cx + r*0.3, Y: cy - r*0.72, W: 0.8, H: 1.44 * r}, white.Alpha(0.22), 0.4)
	case "water":
		deep, glow := nightColors(tint)
		p.Fill(box, deep.Alpha(1), 7)
		// Three washes, bleeding over one another.
		p.Fill(ui.Rect{X: box.X + box.W*0.02, Y: box.Y + box.H*0.1, W: box.W * 0.62, H: box.H * 0.5}, deep.Mix(glow, 0.3).Alpha(0.85), box.H*0.26)
		p.Fill(ui.Rect{X: box.X + box.W*0.34, Y: box.Y + box.H*0.34, W: box.W * 0.6, H: box.H * 0.42}, deep.Mix(glow, 0.6).Alpha(0.7), box.H*0.22)
		p.Fill(ui.Rect{X: box.X + box.W*0.14, Y: box.Y + box.H*0.58, W: box.W * 0.6, H: box.H * 0.34}, deep.Mix(glow, 0.85).Alpha(0.5), box.H*0.2)
	default:
		top, foot := backdropColors(tint)
		p.FillGradient(box, ui.LinearGradient{Angle: 180, Oklab: true, Stops: []ui.GradientStop{
			{Color: top, At: 0}, {Color: foot, At: 1},
		}}, 7)
	}
}
