package ui

import (
	"path"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/jnb666/nes/nes"
)

const (
	border       = 10
	margin       = 10
	initialDelay = 0.3
	repeatDelay  = 0.1
	typeDelay    = 0.5
)

type MenuView struct {
	director     *Director
	paths        []string
	texture      *Texture
	nx, ny, i, j int
	scroll       int
	t            float64
	buttons      [8]bool
	times        [8]float64
	typeBuffer   string
	typeTime     float64
}

func NewMenuView(director *Director, paths []string) View {
	return &MenuView{director: director, paths: paths, texture: NewTexture(director.renderer)}
}

func (view *MenuView) checkButtons() {
	buttons := readKeys(false)
	joysticks := view.director.joysticks
	if len(joysticks) >= 1 {
		buttons = combineButtons(readJoystick(joysticks[0].dev, false), buttons)
	}
	if len(joysticks) >= 2 {
		buttons = combineButtons(readJoystick(joysticks[1].dev, false), buttons)
	}
	now := getTime()
	for i := range buttons {
		if buttons[i] && !view.buttons[i] {
			view.times[i] = now + initialDelay
			view.onPress(i)
		} else if !buttons[i] && view.buttons[i] {
			view.onRelease(i)
		} else if buttons[i] && now >= view.times[i] {
			view.times[i] = now + repeatDelay
			view.onPress(i)
		}
	}
	view.buttons = buttons
}

func (view *MenuView) onPress(index int) {
	switch index {
	case nes.ButtonUp:
		view.j--
	case nes.ButtonDown:
		view.j++
	case nes.ButtonLeft:
		view.i--
	case nes.ButtonRight:
		view.i++
	default:
		return
	}
	view.t = getTime()
}

func (view *MenuView) onRelease(index int) {
	switch index {
	case nes.ButtonStart:
		view.onSelect()
	}
}

func (view *MenuView) onSelect() {
	index := view.nx*(view.j+view.scroll) + view.i
	if index >= len(view.paths) {
		return
	}
	view.director.PlayGame(view.paths[index])
}

func (view *MenuView) OnKey(ev *sdl.KeyboardEvent) {}

func (view *MenuView) OnText(ev *sdl.TextInputEvent) {
	now := getTime()
	if now > view.typeTime {
		view.typeBuffer = ""
	}
	view.typeTime = now + typeDelay
	view.typeBuffer += strings.ToLower(ev.Text)
	for index, p := range view.paths {
		_, p = path.Split(strings.ToLower(p))
		if p >= view.typeBuffer {
			view.highlight(index)
			return
		}
	}
}

func (view *MenuView) highlight(index int) {
	view.scroll = index/view.nx - (view.ny-1)/2
	view.clampScroll(false)
	view.i = index % view.nx
	view.j = (index-view.i)/view.nx - view.scroll
}

func (view *MenuView) Enter() {
	view.director.SetTitle("Select Game")
	view.director.window.StartTextInput()
}

func (view *MenuView) Exit() {
	view.director.window.StopTextInput()
}

func (view *MenuView) Update(t, dt float64) {
	view.checkButtons()
	width, height, _ := view.director.window.SizeInPixels()
	w, h := int(width), int(height)
	const sx = 256 + margin*2
	const sy = 240 + margin*2
	nx := (w - border*2) / sx
	ny := (h - border*2) / sy
	ox := (w-nx*sx)/2 + margin
	oy := (h-ny*sy)/2 + margin
	if nx < 1 {
		nx = 1
	}
	if ny < 1 {
		ny = 1
	}
	view.nx = nx
	view.ny = ny
	view.clampSelection()

	view.director.renderer.SetDrawColor(85, 85, 85, 255)
	view.director.renderer.Clear()
	for j := 0; j < ny; j++ {
		for i := 0; i < nx; i++ {
			x := float32(ox + i*sx)
			y := float32(oy + j*sy)
			index := nx*(j+view.scroll) + i
			if index >= len(view.paths) || index < 0 {
				continue
			}
			path := view.paths[index]
			tx, ty, tw, th := view.texture.Lookup(path)
			drawThumbnail(view.director.renderer, view.texture.texture, x, y, tx, ty, tw, th)
		}
	}
	if int((t-view.t)*4)%2 == 0 {
		x := float32(ox + view.i*sx)
		y := float32(oy + view.j*sy)
		drawSelection(view.director.renderer, x, y, 8)
	}
}

func (view *MenuView) clampSelection() {
	if view.i < 0 {
		view.i = view.nx - 1
	}
	if view.i >= view.nx {
		view.i = 0
	}
	if view.j < 0 {
		view.j = 0
		view.scroll--
	}
	if view.j >= view.ny {
		view.j = view.ny - 1
		view.scroll++
	}
	view.clampScroll(true)
}

func (view *MenuView) clampScroll(wrap bool) {
	n := len(view.paths)
	rows := n / view.nx
	if n%view.nx > 0 {
		rows++
	}
	maxScroll := rows - view.ny
	if view.scroll < 0 {
		if wrap {
			view.scroll = maxScroll
			view.j = view.ny - 1
		} else {
			view.scroll = 0
			view.j = 0
		}
	}
	if view.scroll > maxScroll {
		if wrap {
			view.scroll = 0
			view.j = 0
		} else {
			view.scroll = maxScroll
			view.j = view.ny - 1
		}
	}
}

func drawThumbnail(r *sdl.Renderer, tex *sdl.Texture, x, y, tx, ty, tw, th float32) {
	r.SetDrawColor(51, 51, 51, 255)
	r.RenderFillRect(&sdl.FRect{X: x + 4, Y: y + 4, W: 256, H: 240})
	r.RenderTexture(tex, &sdl.FRect{X: tx, Y: ty, W: tw, H: th}, &sdl.FRect{X: x, Y: y, W: 256, H: 240})
}

func drawSelection(r *sdl.Renderer, x, y, p float32) {
	r.SetDrawColor(255, 255, 255, 255)
	r.RenderRect(&sdl.FRect{X: x - p, Y: y - p, W: 256 + 2*p, H: 240 + 2*p})
}
