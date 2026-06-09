package ui

import (
	"image"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/jnb666/nes/nes"
)

type GameView struct {
	director *Director
	console  *nes.Console
	title    string
	hash     string
	texture  *sdl.Texture
	viewport *sdl.FRect
	record   bool
	frames   []image.Image
}

func NewGameView(director *Director, console *nes.Console, title, hash string) View {
	width, height, _ := director.window.SizeInPixels()
	w, h := float32(width), float32(height)
	texw, texh := console.Buffer().Rect.Dx(), console.Buffer().Rect.Dy()
	aspect := float32(texw) / float32(texh)
	var viewport sdl.FRect
	if w/h > aspect {
		viewport.W = h * aspect
		viewport.H = h
		viewport.X = (w - viewport.W) / 2
	} else {
		viewport.W = w
		viewport.H = w / aspect
		viewport.Y = (h - viewport.H) / 2
	}
	texture := createTexture(director.renderer, texw, texh, sdl.TEXTUREACCESS_STREAMING)
	return &GameView{director, console, title, hash, texture, &viewport, false, nil}
}

func (view *GameView) load(snapshot int) {
	// load state
	if err := view.console.LoadState(savePath(view.hash, snapshot)); err == nil {
		return
	} else {
		view.console.Reset()
	}
	// load sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		if sram, err := readSRAM(sramPath(view.hash, snapshot)); err == nil {
			cartridge.SRAM = sram
		}
	}
}

func (view *GameView) save(snapshot int) {
	// save sram
	cartridge := view.console.Cartridge
	if cartridge.Battery != 0 {
		writeSRAM(sramPath(view.hash, snapshot), cartridge.SRAM)
	}
	// save state
	view.console.SaveState(savePath(view.hash, snapshot))
}

func (view *GameView) Enter() {
	view.director.renderer.SetDrawColor(0, 0, 0, 255)
	view.director.SetTitle(view.title)
	view.console.SetAudioChannel(view.director.audio.channel)
	view.console.SetAudioSampleRate(AudioSampleRate)
	view.load(-1)
}

func (view *GameView) Exit() {
	view.console.SetAudioChannel(nil)
	view.console.SetAudioSampleRate(0)
	view.save(-1)
}

func (view *GameView) Update(t, dt float64) {
	if dt > 1 {
		dt = 0
	}
	console := view.console
	joysticks := view.director.joysticks
	if readKey(sdl.SCANCODE_ESCAPE) || joystickReset(view.director) {
		view.director.ShowMenu()
	}
	turbo := console.PPU.Frame%6 < 3
	j1 := readKeys(turbo)
	if len(joysticks) >= 1 {
		j1 = combineButtons(view.director.buttons[0], j1)
	}
	console.SetButtons1(j1)
	if len(joysticks) >= 2 {
		console.SetButtons2(view.director.buttons[1])
	}
	console.StepSeconds(dt)
	setTexture(view.texture, view.console.Buffer())
	view.director.renderer.RenderTexture(view.texture, nil, view.viewport)

	if view.record {
		view.frames = append(view.frames, copyImage(console.Buffer()))
	}
}

func (view *GameView) OnKey(ev *sdl.KeyboardEvent) {
	switch ev.Key {
	case sdl.K_SPACE:
		screenshot(view.console.Buffer())
	case sdl.K_R:
		view.console.Reset()
	case sdl.K_TAB:
		if view.record {
			view.record = false
			animation(view.frames)
			view.frames = nil
		} else {
			view.record = true
		}
	default:
		if ev.Key >= sdl.K_0 && ev.Key <= sdl.K_9 {
			snapshot := int(ev.Key - sdl.K_9)
			if ev.Mod&(sdl.KMOD_LSHIFT|sdl.KMOD_RSHIFT) != 0 {
				view.load(snapshot)
			} else {
				view.save(snapshot)
			}
		}
	}
}

func (view *GameView) OnText(ev *sdl.TextInputEvent) {}
