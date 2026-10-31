package app

import (
	"fmt"
	"runtime"
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"

	"netease-music-mygo/internal/netease"
)

// sidebarWidth is how wide the sidebar's column is, the gap that holds
// its pane of glass and the room around it.
const sidebarWidth = 232

// sidebarInset is the room the sidebar's pane of glass keeps from the
// window's edges, floating it over the page, and sidebarRadius its
// corners, as macOS 26's floating sidebars round them.
const (
	sidebarInset  = 8
	sidebarRadius = 22
)

func (a *app) view(c *ui.Context) {
	t := *c.Theme()
	a.pal = paletteFor(&t)
	p := a.pal
	a.songAccent = a.coverAccent(&p)
	t.Accent, t.AccentText = p.accent, p.onAccent
	t.Selection = p.accent.Alpha(0.3)
	c.SetTheme(&t)

	a.shortcuts(c)

	if c.Vibrancy() {
		c.Root().Background(ui.Transparent)
	}
	root := ui.Box(c).Fill()
	if !c.Focused(a.searchBox) {
		// Space plays and pauses wherever the focus is, as in Music. On
		// the container around everything, it comes before a button the
		// pointer focused, which would take Space as a click; on the
		// window, it would come after.
		root.OnShortcut(0, ui.KeySpace, a.pl.TogglePlay)
		// And where nothing has the focus, as once Now Playing covers the
		// button that had it.
		c.OnShortcut(0, ui.KeySpace, a.pl.TogglePlay)
	}
	npOpen, npSettled := a.npState(c)
	root.Children(func() {
		page := ui.Row(c).Fill().AlignItems(ui.Stretch)
		page.Children(func() {
			if npOpen && npSettled {
				// Covered by Now Playing: not built, which would cost each
				// of its frames as much again as the stage itself. Where
				// it is scrolled is the app's (pageScroll, navScroll,
				// trackList), so it shows again as it was.
				return
			}
			// Liquid Glass: the page's ground fills the whole window, and
			// the sidebar floats on it as a pane of glass, a step above
			// the page, so the color of what plays has something to bend
			// and frost.
			bg := a.homeTinted(&p)
			liquid := a.settings.glassOn()
			if liquid {
				page.Background(page.AnimateColor("bg", bg, ui.Animation{Duration: 900 * time.Millisecond}))
			}
			side := ui.Column(c).Width(sidebarWidth)
			switch {
			case liquid:
				side.Padding(sidebarInset, sidebarInset, sidebarInset, sidebarInset)
			case !c.Vibrancy():
				side.Background(p.sidebar)
			}
			side.Children(func() {
				pill := ui.Column(c).Fill()
				if liquid {
					pill.Radius(sidebarRadius).Material(glass.Glass{Clarity: 0.5})
				}
				pill.Children(func() { a.sidebar(c) })
			})
			// Keyed, so that a popover of the sidebar opening or closing
			// keeps what the page and its bars hold, their edges' fades.
			main := ui.Box(c.Key("main")).Grow(1)
			if !liquid {
				main.BorderWidth(0, 0, 0, 1).BorderColor(p.hairline)
				main.Background(main.AnimateColor("bg", bg, ui.Animation{Duration: 900 * time.Millisecond}))
			}
			main.Children(func() {
				a.content(c)
				// The bars float over the page, which scrolls under them:
				// positioned, they paint over it.
				a.pageTop(c)
				a.playerBar(c)
			})
		})
		a.nowPlayingLayer(c, npOpen, npSettled)
	})
	a.loginDialog(c)
	a.playlistNameDialog(c)
	a.deletePlaylistDialog(c)
	if a.toast != "" {
		c.Toast(a.toast)
		a.toast = ""
	}
	a.syncWindows()
}

func (a *app) shortcuts(c *ui.Context) {
	// The side buttons of a mouse, the keys of keyboards that have them
	// and swipes between pages, and the keys of Music.
	c.OnShortcut(0, ui.KeyBack, a.back)
	c.OnShortcut(0, ui.KeyForward, a.goForward)
	c.OnShortcut(ui.Cmd, ui.KeyBracketLeft, a.back)
	c.OnShortcut(ui.Cmd, ui.KeyBracketRight, a.goForward)
	c.OnShortcut(ui.Cmd, ui.KeyF, func() {
		a.nowPlaying = false
		a.searchBox.Focus(c)
	})
	if a.nowPlaying {
		c.OnShortcut(0, ui.KeyEscape, func() { a.nowPlaying = false })
		// ⌘ and a number choose Now Playing's tab of that place.
		if a.pl.Current() != nil {
			for i, t := range a.npTabs() {
				c.OnShortcut(ui.Cmd, ui.Key1+ui.Key(i), func() { a.choosePane(t.pane) })
			}
		}
	}
}

// go opens a page, remembering the one before for the back button, and
// closes Now Playing over it, though it is the page open.
func (a *app) open(r route) {
	a.nowPlaying = false
	if r == a.route {
		return
	}
	a.history, a.forward = append(a.history, a.route), nil
	a.show(r)
}

// show shows the page of r.
func (a *app) show(r route) {
	a.route = r
	if r.page == pageSearch {
		a.query, a.searched = r.query, r
	}
	a.load(r)
}

// load loads what the page of r shows, unless it has.
func (a *app) load(r route) {
	switch r.page {
	case pageHome:
		a.freshenHome()
	case pagePlaylist:
		a.loadPlaylist(r.id, false)
	case pageSearch:
		a.loadSearch(r, false)
	case pageArtist:
		a.loadArtist(r.id)
	case pageAlbum:
		a.loadAlbum(r.id)
	case pageUser:
		a.loadUser(r.id)
	case pagePodcast:
		a.loadPodcast(r.id)
	case pagePodcasts:
		a.loadPodcastTab()
	}
}

// back goes to the page before, as the back button of a mouse does, or
// closes Now Playing, which covers the pages.
func (a *app) back() {
	if a.nowPlaying {
		a.nowPlaying = false
		return
	}
	if len(a.history) == 0 {
		return
	}
	a.forward = append(a.forward, a.route)
	r := a.history[len(a.history)-1]
	a.history = a.history[:len(a.history)-1]
	a.show(r)
}

// goForward goes to the page back left, as the forward button of a mouse
// does.
func (a *app) goForward() {
	if len(a.forward) == 0 {
		return
	}
	a.history = append(a.history, a.route)
	r := a.forward[len(a.forward)-1]
	a.forward = a.forward[:len(a.forward)-1]
	a.nowPlaying = false
	a.show(r)
}

func (a *app) sidebar(c *ui.Context) {
	// Under the window buttons where they sit over the sidebar, as the
	// traffic lights do; elsewhere, as close to the top as to the sides.
	top := float32(14)
	if c.TitleBar().Left > 0 {
		top = topBar(c) + 4
	}
	if a.settings.glassOn() {
		// The pane floats, inset from the top: what it leaves is the room
		// its own inset gives back.
		top -= sidebarInset
	}
	// Where nothing in it takes the press, the sidebar moves the window, as
	// the title bar it reaches up into would.
	ui.Column(c).Fill().Padding(top, 14, 14, 14).Gap(4).DragWindow().Children(func() {
		a.searchField(c)

		ui.Scroll(c.Key("nav")).TrackScroll(&a.navScroll).Grow(1).Margin(0, -6).Padding(0, 6).Gap(4).Children(func() {
			a.navItem(c, iconHome, "为你推荐", route{page: pageHome})
			a.navItem(c, iconDaily, "每日推荐", route{page: pageDaily})
			a.navItem(c, iconRadio, "私人 FM", route{page: pageFM})
			a.navItem(c, iconCharts, "排行榜", route{page: pageCharts})
			a.navItem(c, iconPodcast, "播客", route{page: pagePodcasts})
			if a.searched.query != "" {
				a.navItem(c, iconSearch, "搜索结果", a.searched)
			}
			a.myPlaylists(c, "我的歌单", a.mine, true, &a.mineCollapsed)
			a.myPlaylists(c, "收藏的歌单", a.saved, false, &a.savedCollapsed)
		})

		// What plays, as a card at the foot of the sidebar, over who listens.
		if tr := a.pl.Current(); tr != nil {
			a.sidebarNowPlaying(c, tr)
		}
		a.accountRow(c)
	})
}

func (a *app) navItem(c *ui.Context, icon *ui.SVG, label string, r route) {
	p := a.pal
	selected := a.route.page == r.page && !a.nowPlaying
	b := ui.ButtonBase(c.Key(label)).Height(34).PaddingX(10).Gap(10).Radius(8).AlignItems(ui.Center).Justify(ui.Start).FocusRing(false)
	bg := ui.Transparent
	switch {
	case selected:
		bg = a.songAccent.Alpha(0.10)
	case b.Hovered():
		bg = p.hover
	}
	b.Background(b.AnimateColor("bg", bg, ui.Animation{Duration: 160 * time.Millisecond}))
	if selected {
		// The pill carries a thin ink line at its edge.
		b.Border(1, ui.RGB(28, 28, 30))
	}
	color := p.text
	if selected {
		color = a.songAccent
	}
	b.Children(func() {
		ui.Icon(c, icon).Size(17, 17).TextColor(color)
		textColor := p.text
		if selected {
			// The label stays ink, as a pill's label does.
			textColor = p.text
		}
		ui.Text(c, label).FontSize(13.5).FontWeight(500).TextColor(textColor)
		if selected {
			// A bar of the accent at the item's left edge, sliding in
			// from further left as it arrives.
			bar := ui.Box(c.Key("ind")).Absolute().Left(-10).Top(6).Size(3, 22).Radius(1.5).Background(a.songAccent)
			if !c.Preferences().ReduceMotion {
				bar.Transition(ui.ElementTransition{
					Animation: &ui.Animation{Spring: ui.SpringOptions{Duration: 350 * time.Millisecond, Bounce: 0.25}},
					Enter:     &ui.Motion{X: -10, Opacity: 0},
				})
			}
		}
	})
	if b.Clicked() {
		a.open(r)
	}
}

func (a *app) sidebarNowPlaying(c *ui.Context, tr *netease.Track) {
	p := a.pal
	card := ui.ButtonBase(c.Key("np-card")).Gap(10).Padding(8).Radius(12).AlignItems(ui.Center).FocusRing(false).Label("正在播放")
	bg := a.songAccent.Alpha(0.06)
	if card.Hovered() {
		bg = a.songAccent.Alpha(0.11)
	}
	card.Background(card.AnimateColor("bg", bg, ui.Animation{Duration: 160 * time.Millisecond}))
	card.Children(func() {
		a.cover(c, tr.Album.PicURL, 96, 40, 6).Shadow(0, 2, 6, 0, p.cardShade)
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(2).Children(func() {
			ui.Text(c, tr.Name).FontSize(12.5).FontWeight(600).SingleLine()
			a.artistLinks(c, tr, p.muted, p.text).FontSize(11.5).SingleLine()
		})
		// Three slim bars at the foot of the card, dancing while the
		// music does, still when it does not.
		ui.Box(c.Key("bars")).Size(17, 16).Draw(func(dr *ui.Painter, r ui.Rect) {
			lv, live := a.pl.Levels()
			a.sideBars.step(dr.Now(), lv, live)
			drawBars(dr, r, &a.sideBars, 3, ui.RGB(28, 28, 30), 2)
			if live {
				dr.AnimationFrameRate(20)
			}
		})
	})
	if card.Clicked() {
		a.nowPlaying = true
	}
}

// playerBarHeight is the height of the player along the foot of the pages.
const playerBarHeight = 84

// glassBarHeight is the height of the player as a pane of Liquid Glass
// floating above the foot of the pages: with its margin of 8 below and
// 8 of air above, it takes the same room as the bar.
const glassBarHeight = 68

// pageGap is the room between the bar over a page and its content.
const pageGap = 12

// insets returns the room the bars over a page take at its top and its
// bottom: pages pad their content by it, and scroll under the bars.
func (a *app) insets(c *ui.Context) (top, bottom float32) {
	top = topBar(c)
	if runtime.GOOS == "darwin" {
		// The edge reaches a little below the traffic lights' strip.
		top += topEdgePad
	}
	return top, playerBarHeight
}

// topEdgePad is the room the edge over a page leaves below the strip, on
// macOS.
const topEdgePad = 6

// buttonsHeight is the height of the window buttons off macOS, which fill
// the edge over a page.
const buttonsHeight = topBarHeight + topEdgePad

// topBar returns the height of the strip along the top of the window, as
// tall as the window buttons and the room above and below them, which
// what sits in it centers on.
func topBar(c *ui.Context) float32 {
	if runtime.GOOS != "darwin" {
		// The whole edge, which the buttons fill: what sits in the strip
		// centers on their glyphs.
		return buttonsHeight
	}
	return max(c.TitleBar().Height, topBarHeight)
}

// topBarHeight is the least height of the strip along the top.
const topBarHeight = 40

// startMark marks where a page's content starts, first in what scrolls:
// once it has moved up, content is under the bar at the top.
func (a *app) startMark(c *ui.Context) {
	b := ui.Box(c.Key("start-mark")).Height(0).Bounds()
	if b.W > 0 {
		top, _ := a.insets(c)
		a.underTop = b.Y < top+pageGap-0.5
	}
}

// endMark marks where a page's content ends, last in what scrolls: while it
// is below the top of the player, content is under the player.
func (a *app) endMark(c *ui.Context) {
	b := ui.Box(c.Key("end-mark")).Height(0).Bounds()
	if b.W > 0 {
		_, h := c.Size()
		_, bottom := a.insets(c)
		a.underBottom = b.Y > h-bottom+0.5
	}
}

// edgeShown returns how much of a scroll edge shows over el: fully while
// content is under its bar, and none at rest, as AppKit shows its edges.
func edgeShown(el ui.Element, under bool) float32 {
	target := float32(0)
	if under {
		target = 1
	}
	return el.AnimateTo("shown", target, ui.Animation{Duration: 180 * time.Millisecond})
}

// pageTop is the strip over the top of a page, frosting what scrolls under
// it as macOS 26's hard scroll edge does: it drags the window, and holds
// the back button once there is a page to go back to.
func (a *app) pageTop(c *ui.Context) {
	top, _ := a.insets(c)
	// Liquid Glass: the strip is a pane of it, the page's content bending
	// and frosting under its rim as it scrolls by; else, macOS 26's hard
	// scroll edge, which frosts what passes under it against the page's
	// ground.
	var mat ui.Material = glass.ScrollEdge{Hard: true, Background: a.pal.bg}
	if a.settings.glassOn() {
		mat = glass.Glass{Clarity: 0.45}
	}
	edge := ui.Box(c.Key("top-edge")).Absolute().Top(0).Left(0).Right(0).Height(top).DragWindow().Material(mat)
	edge.Opacity(edgeShown(edge, a.underTop))
	if len(a.history) > 0 || len(a.forward) > 0 {
		ui.Row(c).Absolute().Top(0).Left(20).Height(topBar(c)).AlignItems(ui.Center).Children(func() {
			back := iconButton(c.Key("back"), iconBack, 18, a.pal.muted, "返回").Disabled(len(a.history) == 0)
			if len(a.history) == 0 {
				back.Opacity(0.35)
			}
			if back.Clicked() {
				a.back()
			}
			if len(a.forward) > 0 && iconButton(c.Key("forward"), iconForward, 18, a.pal.muted, "前进").Clicked() {
				a.goForward()
			}
		})
	}
}

func (a *app) content(c *ui.Context) {
	// A new page starts at its top. Not as it is opened: a control in a
	// list, as a page's tabs in its header, opens it as the list builds.
	if a.scrolled != a.route {
		a.trackList, a.pageScroll, a.scrolled = ui.ListState{}, ui.ScrollState{}, a.route
	}
	// Pages mark where their content is, as they scroll.
	a.underTop, a.underBottom = false, false
	// Keyed by where it shows, a page is a new one as the listener goes
	// elsewhere, and comes in faint, from a little below, rather than
	// appearing whole.
	key := fmt.Sprintf("page-%d-%d-%s-%d", a.route.page, a.route.id, a.route.query, a.route.kind)
	page := ui.Box(c.Key(key)).Grow(1)
	if !c.Preferences().ReduceMotion && a.settings.transitions() {
		page.Transition(ui.ElementTransition{
			Duration: 220 * time.Millisecond,
			Enter:    &ui.Motion{Y: 10, Opacity: 0},
		})
	}
	page.Children(func() { a.pageOf(c) })
}

// pageOf shows the page the route names.
func (a *app) pageOf(c *ui.Context) {
	switch a.route.page {
	case pageHome:
		a.homePage(c)
	case pageDaily:
		a.dailyPage(c)
	case pageCharts:
		a.chartsPage(c)
	case pagePlaylist:
		a.playlistPage(c)
	case pageSearch:
		a.searchPage(c)
	case pageFM:
		a.fmPage(c)
	case pageArtist:
		a.artistPage(c)
	case pageAlbum:
		a.albumPage(c)
	case pageUser:
		a.userPage(c)
	case pagePodcast:
		a.podcastPage(c)
	case pagePodcasts:
		a.podcastsPage(c)
	}
}

// centered shows a message, with a retry button if retry is set: what a
// page shows in place of what failed to load.
func (a *app) centered(c *ui.Context, err string, retry func()) {
	top, bottom := a.insets(c)
	ui.Column(c).Grow(1).Padding(top, 0, bottom, 0).Center().Gap(12).Children(func() {
		ui.Text(c, err).TextColor(a.pal.muted).FontSize(13)
		if retry != nil && ui.Button(c, "重试").Clicked() {
			retry()
		}
	})
}

// Loading what the pages show.

func (a *app) loadHome() {
	a.homeLoading, a.homeErr = true, ""
	signedIn := a.profile != nil
	at := time.Now()
	go func() {
		featured, err1 := netease.FeaturedPlaylists(30)
		charts, err2 := netease.Charts()
		daily, _ := netease.DailySongs()
		var rec []netease.Playlist
		if signedIn {
			rec, _ = netease.RecommendedPlaylists()
		}
		a.post(func() {
			a.homeLoading = false
			if err1 != nil && err2 != nil {
				// Loading the new day's again, what shows stays, to try
				// once more.
				if len(a.featured) == 0 && len(a.charts) == 0 {
					a.homeErr = "无法连接网易云音乐，请检查网络"
				}
				return
			}
			a.featured, a.charts, a.daily, a.recPlaylists = featured, charts, daily, rec
			a.homeAt = at
		})
	}()
}

// freshenHome loads the home page again once NetEase has picked the day's
// songs anew since it last did: the app may stay open overnight.
func (a *app) freshenHome() {
	if !a.homeLoading && !a.homeAt.IsZero() && dailyRenewed(a.homeAt, time.Now()) {
		a.loadHome()
	}
}

// beijing is NetEase's time, without daylight saving.
var beijing = time.FixedZone("CST", 8*60*60)

// dailyRenewed reports whether NetEase has picked new daily songs between
// loaded and now. It does at 6:00, Beijing time.
func dailyRenewed(loaded, now time.Time) bool {
	b := now.In(beijing)
	six := time.Date(b.Year(), b.Month(), b.Day(), 6, 0, 0, 0, beijing)
	if b.Before(six) {
		six = six.AddDate(0, 0, -1)
	}
	return loaded.Before(six)
}

func (a *app) loadPlaylist(id int64, thenPlay bool) {
	if pl := a.playlists[id]; pl != nil {
		if thenPlay {
			a.pl.PlayQueue(pl.Tracks, firstPlayable(pl.Tracks))
		}
		return
	}
	if a.loadingPl[id] {
		return
	}
	a.loadingPl[id] = true
	delete(a.plErr, id)
	go func() {
		pl, err := netease.PlaylistDetail(id)
		a.post(func() {
			delete(a.loadingPl, id)
			if err != nil {
				a.plErr[id] = "歌单加载失败"
				return
			}
			a.playlists[id] = pl
			if thenPlay {
				a.pl.PlayQueue(pl.Tracks, firstPlayable(pl.Tracks))
			}
		})
	}()
}

func firstPlayable(tracks []netease.Track) int {
	for i := range tracks {
		if tracks[i].Playable() {
			return i
		}
	}
	return 0
}

// homeTinted is the ground of the home's pages: the theme's background
// holding a little of the playing cover's color, dulled toward gray, so
// the room feels of the music without shouting of it. The last color
// stays while a cover loads, and when nothing plays the ground is plain.
func (a *app) homeTinted(p *palette) ui.Color {
	tr := a.pl.Current()
	if tr == nil || tr.Album.PicURL == "" {
		return p.bg
	}
	if art := a.art.get(netease.Sized(tr.Album.PicURL, 64), true); art != nil && !art.failed {
		a.homeArtURL, a.homeTint = tr.Album.PicURL, art.tint
	}
	if a.homeArtURL != tr.Album.PicURL {
		return p.bg // no tint of this cover yet: plain until it arrives
	}
	// Water the color down toward gray, then into the ground, plainly
	// there to see.
	soft := a.homeTint.Mix(ui.RGB(128, 128, 132), 0.25)
	return p.bg.Mix(soft, 0.18)
}

// coverAccent is the accent of the moment: the playing cover's color,
// whitened for tints and text; the theme's accent when nothing plays or
// the cover has yet to arrive.
func (a *app) coverAccent(p *palette) ui.Color {
	tr := a.pl.Current()
	if tr == nil || tr.Album.PicURL == "" {
		return p.accent
	}
	if art := a.art.get(netease.Sized(tr.Album.PicURL, 64), true); art != nil && !art.failed {
		a.homeArtURL, a.homeTint = tr.Album.PicURL, art.tint
	}
	if a.homeArtURL == tr.Album.PicURL {
		return a.homeTint.Mix(white, 0.45)
	}
	return p.accent
}
