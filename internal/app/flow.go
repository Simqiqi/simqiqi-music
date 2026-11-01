package app

import (
	"sync"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/shader"
)

// The flow is the wallpaper as a field of color rather than a line of it:
// three colors of the cover — the two hues the app opened onto and the one
// between them — lying on the page in soft pools, pushed about by a slow
// current, so that the pools wander and the borders between them bend and
// give. It is the gradient the app opened onto, alive: where the renderer
// cannot paint shaders, the gradient stands in for it, as it did.
//
// The colors it holds are the wallpaper's own, drifting round the wheel
// and swaying as they do (wallpaper.go), so the flow changes with the
// evening as the gradient does.

// flowSource paints it: noise of four octaves, warped by noise of its own,
// two of its fields deciding where each color pools.
const flowSource = `package main

var Time float
var Wash float
var C1 vec4
var C2 vec4
var C3 vec4

func hash(p vec2) float {
	return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453)
}

func noise(p vec2) float {
	i := floor(p)
	f := fract(p)
	u := f * f * (3.0 - 2.0*f)
	a := hash(i)
	b := hash(i + vec2(1.0, 0.0))
	c := hash(i + vec2(0.0, 1.0))
	d := hash(i + vec2(1.0, 1.0))
	return mix(mix(a, b, u.x), mix(c, d, u.x), u.y)
}

func fbm(p vec2) float {
	v := 0.0
	amp := 0.5
	for i := 0; i < 4; i++ {
		v = v + amp*noise(p)
		p = p*2.07 + vec2(1.7, 9.2)
		amp = amp * 0.5
	}
	return v
}

func Fragment(dst vec4, pos vec2) vec4 {
	size := rectSize()
	p := pos / size.y * 1.5
	t := Time * 0.05

	// The current: noise pushed by noise of its own, so that the pools
	// wander rather than slide.
	q := vec2(fbm(p+vec2(t, 0.0)), fbm(p+vec2(5.2, 1.3)-vec2(0.0, t*0.8)))
	f := fbm(p + 1.35*q)
	g := fbm(p*0.7 + 0.9*q + vec2(3.7, 2.2) - vec2(t*0.6, 0.0))

	c := mix(C1.rgb, C2.rgb, smoothstep(0.28, 0.62, f))
	c = mix(c, C3.rgb, smoothstep(0.42, 0.80, g))
	// A little air in it, so that the pools read as light and shade.
	c = c * (0.92 + 0.16*smoothstep(0.3, 0.75, f))
	return vec4(c*Wash, Wash)
}
`

var flowShader = sync.OnceValue(func() *shader.Shader { return shader.MustCompile(flowSource) })

// drawFlow paints the wallpaper's colors as a field of them in r, at t
// seconds after the app opened.
func drawFlow(p *ui.Painter, r ui.Rect, w wallpaper, dark bool, t float32) {
	d := w.posed(t)
	var c1, c2, c3 ui.Color
	if dark {
		c1 = ui.Oklch(0.34, 0.11, d.hue)
		c2 = ui.Oklch(0.52, 0.12, wrapHue(d.hue+shortestArc(d.hue, d.next)/2))
		c3 = ui.Oklch(0.26, 0.09, d.next)
	} else {
		c1 = ui.Oklch(0.92, 0.065, d.hue)
		c2 = ui.Oklch(0.86, 0.090, wrapHue(d.hue+shortestArc(d.hue, d.next)/2))
		c3 = ui.Oklch(0.90, 0.055, d.next)
	}
	// Held at one opacity, so that the page's ground shows through, as
	// the gradient's stops do: less by night, more by day.
	wash := float32(0.75)
	if !dark {
		wash = 0.88
	}
	flowShader().Draw(p, r, shader.Options{Uniforms: map[string]any{
		"Time": t, "Wash": wash, "C1": c1, "C2": c2, "C3": c3,
	}})
}
