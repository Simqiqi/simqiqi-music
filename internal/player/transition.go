package player

import (
	"time"
)

// transitions reports whether the smart transitions are on, as the
// settings say.
func (p *Player) transitions() bool {
	return p.Transition != nil && p.Transition()
}

// armTail schedules the fade out at the end of the track playing, over
// its last seconds: longer after a long song, shorter after a short one,
// and none at all when transitions are off or the listener paused. A new
// arm of a track left behind is overtaken, and does nothing.
func (p *Player) armTail(gen int) {
	p.tailGen++
	if !p.transitions() || p.out == nil || !p.playing {
		return
	}
	dur := p.Duration()
	if dur <= 0 {
		return
	}
	over := tailFor(dur)
	wait := dur - p.Position() - over
	if wait < 0 {
		wait = 0
	}
	mine := p.tailGen
	go func() {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		<-timer.C
		p.post(func() {
			if mine != p.tailGen || gen != p.gen || p.out == nil || !p.playing || p.seeking {
				return
			}
			left := p.Duration() - p.Position()
			if left <= tailLeast {
				// Too little left to fade: it goes as it is.
				return
			}
			p.fade.fade(0, min(over, left), nil)
		})
	}()
}

// rearm brings the music back up when a seek took it out of the song's
// tail, then arms the tail again where the song now stands.
func (p *Player) rearm() {
	if p.out == nil || !p.playing {
		return
	}
	if d, at := p.Duration(), p.Position(); d > 0 && d-at > tailFor(d) {
		if p.fade.silent() {
			p.fade.fade(1, fadeIn, nil)
		}
	}
	p.armTail(p.gen)
}
