package app

import (
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// The mini player is a window of its own: a card in the app's colors with
// the cover, what plays and the controls, floating over the desktop where
// the window stood, for keeping the music at hand while the window itself
// is out of the way.
const (
	miniWidth  = 420
	miniHeight = 104
)

// miniPlayer is the mini player's window, made the first time it shows.
type miniPlayer struct {
	win *mygo.Window
}

// miniOpen reports whether the mini player is the one showing.
func (a *app) miniOpen() bool {
	return a.mini.win != nil && a.mini.win.IsVisible()
}

// toggleMini hands the window over to the mini player, or takes it back.
func (a *app) toggleMini() {
	if a.miniOpen() {
		a.expandMini()
	} else {
		a.showMini()
	}
}

// showMini puts the window away and shows the mini player where its foot
// on the left was, clear of the Dock.
func (a *app) showMini() {
	if a.win == nil || a.pl.Current() == nil {
		return
	}
	if a.mini.win == nil {
		a.mini.win = mygo.NewWindow(mygo.WindowOptions{
			Title:         mygo.App.Name() + " 迷你播放器",
			Panel:         true,
			Frameless:     true,
			Transparent:   true,
			AlwaysOnTop:   true,
			DisableResize: true,
			Width:         miniWidth,
			Height:        miniHeight,
			Content:       ui.View(a.miniView),
		})
		// Closing it hands back to the window, as a panel does not.
		a.mini.win.OnClose(func(e *mygo.CloseEvent) {
			if a.quitting {
				return
			}
			e.PreventDefault()
			a.expandMini()
		})
	}
	b := a.win.Bounds()
	a.mini.win.SetBounds(mygo.Rectangle{
		X:      b.X,
		Y:      b.Y + b.Height - miniHeight - 48,
		Width:  miniWidth,
		Height: miniHeight,
	})
	// Shown before the window goes, so the app is never without one.
	a.mini.win.Show()
	a.mini.win.Focus()
	a.win.Hide()
}

// expandMini puts the window back, in the mini player's place.
func (a *app) expandMini() {
	if a.mini.win != nil {
		a.mini.win.Hide()
	}
	a.showMain()
}

// miniView draws the mini player: the cover and what plays over a line of
// its progress, and the controls, on a card whose rounded corners let the
// desktop show through.
func (a *app) miniView(c *ui.Context) {
	t := *c.Theme()
	a.pal = paletteFor(&t, a.settings.Accent)
	p := a.pal
	a.songAccent = a.coverAccent(&p)
	c.Root().Background(ui.Transparent)

	pl := a.pl
	tr := pl.Current()
	if tr != nil && pl.Playing() {
		// Redraw with the clock, as the bar does.
		pos := pl.Position()
		c.After(time.Second - pos%time.Second + 10*time.Millisecond)
	}
	c.OnShortcut(0, ui.KeyEscape, a.expandMini)

	ui.Box(c).Fill().Padding(10).Children(func() {
		card := ui.Row(c).Fill().Radius(20).Padding(10).Gap(12).AlignItems(ui.Center).
			Background(p.bg).Border(1, p.hairline).
			Shadow(0, 16, 36, -10, ui.RGBA(0, 0, 0, 0.4)).DragWindow()
		card.Draw(func(pt *ui.Painter, r ui.Rect) {
			// How far the song is, along the foot of the card.
			if tr == nil {
				return
			}
			d := pl.Duration()
			if d <= 0 {
				return
			}
			v := float32(min(1, max(0, float64(pl.Position())/float64(d))))
			y := r.Y + r.H - 3
			pt.Fill(ui.Rect{X: r.X + 18, Y: y, W: r.W - 36, H: 3}, p.press, 1.5)
			pt.Fill(ui.Rect{X: r.X + 18, Y: y, W: max(3, (r.W-36)*v), H: 3}, a.songAccent, 1.5)
		})
		card.Children(func() {
			if tr == nil {
				ui.Text(c, "未在播放").FontSize(13).TextColor(p.muted).Grow(1)
				return
			}
			info := ui.ButtonBase(c.Key("mini-info")).Gap(12).AlignItems(ui.Center).
				FocusRing(false).Label("展开播放器").Shrink(1).MinWidth(0)
			info.Children(func() {
				a.cover(c, tr.Album.PicURL, 200, 56, 10).Shadow(0, 2, 8, 0, p.cardShade)
				ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
					ui.Text(c, tr.Name).FontSize(13.5).FontWeight(600).SingleLine()
					a.artistLinks(c, tr, p.muted, p.text).FontSize(11.5).SingleLine()
				})
			})
			if info.Clicked() {
				a.expandMini()
			}
			ui.Row(c).Gap(2).AlignItems(ui.Center).Shrink(0).Children(func() {
				if iconButton(c.Key("mini-prev"), iconPrev, 13, p.text.Alpha(0.8), "上一首").Clicked() {
					pl.Prev()
				}
				if playButton(c.Key("mini-play"), pl.Playing(), 20, p.text).Clicked() {
					pl.TogglePlay()
				}
				if iconButton(c.Key("mini-next"), iconNext, 13, p.text.Alpha(0.8), "下一首").Clicked() {
					pl.Next()
				}
				if iconButton(c.Key("mini-expand"), iconExpand, 13, p.muted, "展开").Clicked() {
					a.expandMini()
				}
			})
		})
	})
}
