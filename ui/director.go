package ui

import (
	"log"
	"slices"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/jnb666/nes/nes"
)

// joystick mapping for https://thepihut.com/products/nes-style-raspberry-pi-compatible-usb-gamepad-controller
var JoystickButtons = map[uint8]uint16{
	8: nes.ButtonSelect,
	9: nes.ButtonStart,
	1: nes.ButtonA,
	0: nes.ButtonB,
}

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
	buttons   [2][8]bool
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
			d.view.OnText(event.TextInputEvent())

		case sdl.EVENT_JOYSTICK_AXIS_MOTION:
			ev := event.JoyAxisEvent()
			if controller, ok := d.joystick(ev.Which); ok {
				if ev.Axis == 1 {
					d.buttons[controller][nes.ButtonUp] = ev.Value < -16384
					d.buttons[controller][nes.ButtonDown] = ev.Value > 16384
				} else {
					d.buttons[controller][nes.ButtonLeft] = ev.Value < -16384
					d.buttons[controller][nes.ButtonRight] = ev.Value > 16384
				}
			}

		case sdl.EVENT_JOYSTICK_BUTTON_DOWN, sdl.EVENT_JOYSTICK_BUTTON_UP:
			ev := event.JoyButtonEvent()
			if controller, ok := d.joystick(ev.Which); ok {
				if id, ok := JoystickButtons[ev.Button]; ok {
					d.buttons[controller][id] = ev.Down
				}
			}

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

func (d *Director) joystick(id sdl.JoystickID) (int, bool) {
	ix := slices.IndexFunc(d.joysticks, func(j Joystick) bool { return j.id == id })
	if ix >= 0 && ix < 2 {
		return ix, true
	}
	log.Printf("Joystick event with invalid ID %d", id)
	return 0, false
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
