package app

import (
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"netease-music-mygo/internal/netease"
)

// nowPlayingScene opens Now Playing on a song of a cover that is not
// loaded, so that the backdrop is painted of its tint alone, and returns
// the app and the tester showing it.
func nowPlayingScene(t *testing.T, style string, w, h int) (*app, *ui.Tester) {
	t.Helper()
	a := newTestApp()
	s := defaultSettings()
	s.setBackdrop(style)
	a.settings = s
	url := "https://p1.music.126.net/star.jpg"
	a.art.art[netease.Sized(url, 600)+"#backdrop"] = &artwork{tint: ui.RGB(52, 84, 176)}
	a.pl.Cue([]netease.Track{{ID: 7, Name: "Starman", Duration: 240000,
		Album: netease.Album{ID: 1, Name: "Space", PicURL: url}, Artists: []netease.Artist{{Name: "Bowie"}}}})
	a.pl.SetLyrics(nil)
	a.nowPlaying, a.npPane = true, paneLyrics
	tt := ui.NewTester(a.view, w, h)
	tt.SetDark(true)
	tt.Frame()
	tt.Advance(1500 * time.Millisecond)
	tt.Frame()
	return a, tt
}

// TestBackdropPickerPicksOne: the pill at the top right of Now Playing
// names the backdrop it wears, opens onto the three, and the one picked
// is the one the settings keep.
func TestBackdropPickerPicksOne(t *testing.T) {
	a, tt := nowPlayingScene(t, "", 1400, 900)
	if !tt.HasText("渐变") {
		t.Fatal("the picker does not name the backdrop it wears")
	}
	if dir := os.Getenv("NM_SHOTS"); dir != "" {
		saveFrame(t, dir, "picker-closed.png", tt.Image())
	}
	if err := tt.Click("播放页背景"); err != nil {
		t.Fatalf("the picker does not open: %v", err)
	}
	tt.Frame()
	if dir := os.Getenv("NM_SHOTS"); dir != "" {
		saveFrame(t, dir, "picker-open.png", tt.Image())
	}
	for _, b := range backdrops {
		if !tt.HasText(b.label) {
			t.Errorf("the picker does not offer %s", b.label)
		}
	}
	if err := tt.Click("星野"); err != nil {
		t.Fatalf("the picker does not take a pick: %v", err)
	}
	tt.Frame()
	if got := a.settings.backdropStyle(); got != "stars" {
		t.Errorf("backdrop %q after picking the stars, want stars", got)
	}
	if a.settings.Fog {
		t.Error("the fog is on after picking the stars")
	}
	if !tt.HasText("星野") {
		t.Error("the picker still names the backdrop it wore before")
	}
	if a.backdropCache.style != "stars" {
		t.Errorf("the backdrop is %q, want stars", a.backdropCache.style)
	}
}

// TestBackdropSwatches paints each of the three small, the way the picker
// shows them.
func TestBackdropSwatches(t *testing.T) {
	view := func(c *ui.Context) {
		ui.Row(c).Gap(6).Children(func() {
			for _, b := range backdrops {
				style := b.id
				ui.Box(c).Size(24, 24).Draw(func(p *ui.Painter, r ui.Rect) {
					backdropSwatch(p, r, style, ui.RGB(52, 84, 176))
				})
			}
		})
	}
	tt := ui.NewTester(view, len(backdrops)*30+12, 40)
	tt.Frame()
	img := tt.Image()
	// Each swatch is drawn, and each differently from the others.
	seen := map[uint32]bool{}
	for i := range backdrops {
		x := i*30 + 12
		r, g, bl, _ := img.At(x, 20).RGBA()
		seen[r<<16|g<<8|bl] = true
		if r == 0 && g == 0 && bl == 0 {
			t.Errorf("the %s swatch is blank", backdrops[i].label)
		}
	}
	if len(seen) != len(backdrops) {
		t.Errorf("%d swatches look alike, want %d apart", len(backdrops)-len(seen)+1, len(backdrops))
	}
}
