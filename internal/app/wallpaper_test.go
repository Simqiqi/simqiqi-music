package app

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"
)

// TestNewWallpaperIsARandomOne: each launch's wallpaper holds two hues a
// step apart, at a diagonal, with its light somewhere along the way.
func TestNewWallpaperIsARandomOne(t *testing.T) {
	same := 0
	for range 40 {
		w := newWallpaper()
		if w.hue < 0 || w.hue >= 360 || w.next < 0 || w.next >= 360 {
			t.Fatalf("hues %0.1f and %0.1f off the wheel", w.hue, w.next)
		}
		step := math.Abs(float64(shortestArc(w.hue, w.next)))
		if step < 26 || step > 74 {
			t.Errorf("the hues are %0.0f apart, want 26 to 74", step)
		}
		if (w.angle < 105 || w.angle > 165) && (w.angle < 195 || w.angle > 255) {
			t.Errorf("the angle %0.0f is not a diagonal", w.angle)
		}
		if w.lift < 0.38 || w.lift > 0.62 {
			t.Errorf("the light gathers at %0.2f, want between 0.38 and 0.62", w.lift)
		}
		if w == (wallpaper{hue: 0, next: 0, angle: 105, lift: 0.38}) {
			same++
		}
	}
	if same > 36 {
		t.Errorf("%d of 40 wallpapers are the same one", same)
	}
}

// TestWallpaperStops: the wash runs from one hue to the other through the
// light, darker at night and paler in daylight, and never two stops alike.
func TestWallpaperStops(t *testing.T) {
	w := wallpaper{hue: 214, next: 262, angle: 132, lift: 0.5}
	for _, dark := range []bool{true, false} {
		stops := w.stops(dark)
		if len(stops) != 3 {
			t.Fatalf("the wash holds %d stops, want three", len(stops))
		}
		if stops[0].At != 0 || stops[1].At != 0.5 || stops[2].At != 1 {
			t.Errorf("the stops sit at %0.2f, %0.2f, %0.2f", stops[0].At, stops[1].At, stops[2].At)
		}
		seen := map[uint32]bool{}
		var alpha uint8
		for i, s := range stops {
			c := s.Color
			if i == 0 {
				alpha = c.A
			} else if c.A != alpha {
				t.Errorf("the %d. stop is %d opaque, the first %d: the wash must be even", i+1, c.A, alpha)
			}
			seen[uint32(c.R)<<16|uint32(c.G)<<8|uint32(c.B)] = true
		}
		if len(seen) != 3 {
			t.Errorf("the wash's stops are %d colors, want three", len(seen))
		}
	}
}

// TestWallpaperIsDrawn: the wash lands over the ground, running from one
// corner toward the other, and what it leaves is not the ground itself.
func TestWallpaperIsDrawn(t *testing.T) {
	w := wallpaper{hue: 214, next: 262, angle: 132, lift: 0.5}
	view := func(c *ui.Context) {
		ui.Box(c).Fill().Background(ui.Hex("#141416")).Draw(func(p *ui.Painter, r ui.Rect) {
			w.paint(p, r, true, 0)
		})
	}
	tt := ui.NewTester(view, 400, 300)
	tt.Frame()
	img := tt.Image()
	at := func(x, y int) (r, g, b int) {
		return int(img.RGBAAt(x, y).R), int(img.RGBAAt(x, y).G), int(img.RGBAAt(x, y).B)
	}
	trR, trG, trB := at(380, 20)
	blR, blG, blB := at(20, 280)
	if trR == blR && trG == blG && trB == blB {
		t.Errorf("the corners are one color (%d %d %d): the wash does not run", trR, trG, trB)
	}
	ground := func(x, y int) bool {
		r, g, b := at(x, y)
		return r == 20 && g == 20 && b == 22
	}
	if ground(200, 150) {
		t.Error("the middle is the bare ground: the wash is not painted")
	}
	// Blue of the hue, the wash is, over the gray ground.
	if trB <= trR {
		t.Errorf("the lit corner is %d red to %d blue: the hue is lost", trR, trB)
	}
}

// TestWallpaperBreathes: minutes on, the wash has gone round the wheel,
// and its angle has swayed either way of the one it opened at, by no more
// than the sway allows.
func TestWallpaperBreathes(t *testing.T) {
	w := wallpaper{hue: 214, next: 262, angle: 132, lift: 0.5}
	if d := shortestArc(w.hue, w.posed(0).hue); math.Abs(float64(d)) > 0.001 {
		t.Errorf("the wash moved %0.3f degrees at once", d)
	}
	if d := shortestArc(w.hue, w.posed(600).hue); math.Abs(float64(d)) < 20 {
		t.Errorf("ten minutes on the hue has gone %0.1f degrees round", d)
	}
	lo, hi := float32(999), float32(-999)
	for s := range 240 {
		sw := w.posed(float32(s)).angle - w.angle
		if math.Abs(float64(sw)) > wallpaperSway+0.01 {
			t.Fatalf("%d seconds on the angle swung %0.2f degrees", s, sw)
		}
		lo, hi = min(lo, sw), max(hi, sw)
	}
	if hi-lo < 14 {
		t.Errorf("the angle swings %0.1f degrees only, want about %d", hi-lo, 2*wallpaperSway)
	}
}
