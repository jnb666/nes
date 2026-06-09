package ui

import (
	"log"
	"slices"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/jnb666/nes/nes"
)

type View interface {
	Enter()
	Exit()
	Update(t, dt float64)
	OnKey(ev *sdl.KeyboardEvent)
	OnText(ev *sdl.TextInputEvent)
}

type Director struct {
	window    *sdl.Window
	renderer  *sdl.Renderer
	audio     *Audio
	joysticks []Joystick
	view      View
	menuView  View
	timestamp float64
}

type Joystick struct {
	id  sdl.JoystickID
	dev *sdl.Joystick
}

func NewDirector(window *sdl.Window, renderer *sdl.Renderer, audio *Audio) *Director {
	director := Director{}
	director.window = window
	director.renderer = renderer
	director.audio = audio
	return &director
}

func (d *Director) SetTitle(title string) {
	d.window.SetTitle(title)
}

func (d *Director) SetView(view View) {
	if d.view != nil {
		d.view.Exit()
	}
	d.view = view
	if d.view != nil {
		d.view.Enter()
	}
	d.timestamp = getTime()
}

func (d *Director) Step() {
	d.renderer.Clear()
	timestamp := getTime()
	dt := timestamp - d.timestamp
	d.timestamp = timestamp
	if d.view != nil {
		d.view.Update(timestamp, dt)
	}
}

func (d *Director) Start(paths []string) {
	d.menuView = NewMenuView(d, paths)
	if len(paths) == 1 {
		d.PlayGame(paths[0])
	} else {
		d.ShowMenu()
	}
	d.Run()
}

func (d *Director) Run() {
	for d.PollEvents() {
		d.Step()
		d.renderer.Present()
	}
	d.SetView(nil)
}

func (d *Director) PlayGame(path string) {
	hash, err := hashFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	console, err := nes.NewConsole(path)
	if err != nil {
		log.Fatalln(err)
	}
	d.SetView(NewGameView(d, console, path, hash))
}

func (d *Director) ShowMenu() {
	d.SetView(d.menuView)
}

func (d *Director) PollEvents() bool {
	var event sdl.Event
	for sdl.PollEvent(&event) {
		switch event.Type {
		case sdl.EVENT_QUIT:
			return false

		case sdl.EVENT_KEY_DOWN:
			d.view.OnKey(event.KeyboardEvent())

		case sdl.EVENT_TEXT_INPUT:
			d.menuView.OnText(event.TextInputEvent())

		case sdl.EVENT_JOYSTICK_ADDED:
			ev := event.JoyDeviceEvent()
			js := must(ev.Which.OpenJoystick())
			log.Printf("added joystick %d: %s", ev.Which, must(js.Name()))
			d.joysticks = append(d.joysticks, Joystick{id: ev.Which, dev: js})

		case sdl.EVENT_JOYSTICK_REMOVED:
			ev := event.JoyDeviceEvent()
			ix := slices.IndexFunc(d.joysticks, func(j Joystick) bool { return j.id == ev.Which })
			if ix > 0 {
				d.joysticks[ix].dev.Close()
				log.Printf("removed joystick %d", ev.Which)
				d.joysticks = slices.Delete(d.joysticks, ix, ix+1)
			} else {
				log.Printf("Error: got joystick removed event for %d but it was not previously added", ev.Which)
			}
		}
	}
	return true
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
