package app

import (
	"fmt"
	"sync"
	"time"

	"github.com/egoist/mygo/ui"
)

// The sleep timer leaves the music after a while, faded rather than cut,
// so that the listener may fall asleep to it and the room goes quiet of
// itself. The timer lives here rather than in the player, which knows
// nothing of a night's plans: it only knows how to leave.

// sleepFade is how long the music takes to leave when the timer fires.
const sleepFade = 4 * time.Second

// sleepChoice is one way the timer may end.
type sleepChoice struct {
	label string
	dur   time.Duration // zero: at the end of the song playing
}

var sleepChoices = []sleepChoice{
	{"15 分钟后", 15 * time.Minute},
	{"30 分钟后", 30 * time.Minute},
	{"45 分钟后", 45 * time.Minute},
	{"60 分钟后", 60 * time.Minute},
	{"播完当前这首", 0},
}

// sleepTimer is the timer as the view and the goroutine that fires it see
// it: under a lock, since one holds it while the other reads, and an arm
// overtakes the one before it.
type sleepTimer struct {
	mu    sync.Mutex
	ends  time.Time // when it fires; zero while it is off
	mode  string    // the choice armed, for the menu to check
	gen   int       // counts arms, to drop the goroutine of one overtaken
	fires bool      // its fade is under way: the music is leaving
}

// on reports whether a timer runs, and what it was set to.
func (s *sleepTimer) on() (string, bool) {
	if s == nil {
		return "", false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mode, !s.ends.IsZero() || s.fires
}

// left is how long the timer has still to run.
func (s *sleepTimer) left() time.Duration {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ends.IsZero() {
		return 0
	}
	return max(0, time.Until(s.ends))
}

// arm sets the timer to fire at when, with the mode it was chosen by,
// unless another arm overtakes it first.
func (s *sleepTimer) arm(when time.Time, mode string, fire func()) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.gen++
	gen := s.gen
	s.ends, s.mode, s.fires = when, mode, false
	s.mu.Unlock()
	go func() {
		if wait := time.Until(when); wait > 0 {
			t := time.NewTimer(wait)
			defer t.Stop()
			<-t.C
		}
		s.mu.Lock()
		if s.gen != gen {
			// Overtaken: another arm, or a cancel, has taken its place.
			s.mu.Unlock()
			return
		}
		s.ends, s.fires = time.Time{}, true
		s.mu.Unlock()
		fire()
		// It stays on, in its fade, until the music has left.
		s.mu.Lock()
		if s.gen == gen {
			s.mode, s.fires = "", false
		}
		s.mu.Unlock()
	}()
}

// cancel stops the timer, whether it waits or its fade is under way.
func (s *sleepTimer) cancel() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.gen++
	s.ends, s.mode, s.fires = time.Time{}, "", false
	s.mu.Unlock()
}

// sleepIn arms the timer: after d, or at the end of the song playing when
// d is zero. Its fade starts early enough that the music is gone as the
// time is up, rather than playing on four seconds past it.
func (a *app) sleepIn(d time.Duration, label string) {
	at := d
	if d <= 0 {
		at = max(0, a.pl.Duration()-a.pl.Position())
	}
	fade := min(sleepFade, at)
	a.sleep.arm(time.Now().Add(at-fade), label, func() {
		a.pl.FadeOutAndPause(fade)
		a.win.Invalidate()
	})
}

// sleepCancel stops the timer, and brings the music back if its fade had
// begun.
func (a *app) sleepCancel() {
	a.sleep.cancel()
	if !a.pl.Playing() && a.pl.Ready() {
		// The fade had paused it: the listener wants it back.
		a.pl.TogglePlay()
	}
	a.win.Invalidate()
}

// sleepButton is the moon in the player's bar: it opens the timer's
// lengths, and wears the accent, with how long is left, while one runs.
func (a *app) sleepButton(c *ui.Context, color, on ui.Color) {
	mode, running := a.sleep.on()
	col, label := color, "睡眠定时"
	if running {
		col = on
		label = fmt.Sprintf("睡眠定时：%s，还剩 %s", mode, rough(a.sleep.left()))
	}
	b := iconButton(c.Key("sleep"), iconMoon, 16, col, label)
	if left := a.sleep.left(); running && left > 0 && left < time.Minute {
		// Under a minute, the count shows itself on the bar.
		ui.Text(c, fmt.Sprintf("%ds", int(left.Seconds()+0.5))).FontSize(10).FontFeatures("tnum").TextColor(col)
	}
	b.Menu(func(m *ui.Menu) {
		for _, ch := range sleepChoices {
			item := m.Item(ch.label)
			item.Checked(running && mode == ch.label)
			if item.Chosen() {
				a.sleepIn(ch.dur, ch.label)
			}
		}
		m.Separator()
		off := m.Item("取消定时").Disabled(!running)
		if off.Chosen() {
			a.sleepCancel()
		}
	})
}

// rough says how long is left in words, as the timer's label shows it:
// a minute at least, the hours when they matter.
func rough(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "不到 1 分钟"
	case d < time.Hour:
		return fmt.Sprintf("%d 分钟", int(d.Minutes()+0.5))
	default:
		h := int(d.Hours())
		if m := int(d.Minutes()) - h*60; m > 0 {
			return fmt.Sprintf("%d 小时 %d 分钟", h, m)
		}
		return fmt.Sprintf("%d 小时", h)
	}
}
