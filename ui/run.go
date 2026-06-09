package ui

import (
	"log"
	"runtime"

	"github.com/Zyko0/go-sdl3/sdl"
)

const (
	width  = 256
	height = 240
	title  = "NES"
)

func Run(paths []string, scale int) {
	// initialize SDL
	if err := sdl.LoadLibrary(libraryPath()); err != nil {
		log.Fatalln(err)
	}
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_JOYSTICK); err != nil {
		log.Fatalln(err)
	}
	defer sdl.Quit()
	driver := sdl.GetCurrentVideoDriver()
	bounds := must(sdl.GetPrimaryDisplay().Bounds())
	log.Printf("SDL %s %s - display:%dx%d scale:%d", sdl.GetVersion(), driver, bounds.W, bounds.H, scale)

	// initialize audio
	audio := NewAudio()
	if err := audio.Start(); err != nil {
		log.Fatalln(err)
	}
	defer audio.Stop()

	// create window
	w, h := width*scale, height*scale
	opts := sdl.WINDOW_HIGH_PIXEL_DENSITY
	if driver == "kmsdrm" {
		w, h = int(bounds.W), int(bounds.H)
		opts = sdl.WINDOW_FULLSCREEN
		sdl.HideCursor()
	}
	window, renderer, err := sdl.CreateWindowAndRenderer(title, w, h, opts)
	if err != nil {
		log.Fatalln(err)
	}
	defer window.Destroy()
	if err = renderer.SetVSync(1); err != nil {
		log.Fatalln(err)
	}
	window.Raise()

	// run director
	director := NewDirector(window, renderer, audio)
	director.Start(paths)
}

func libraryPath() string {
	switch runtime.GOOS {
	case "windows":
		return "SDL3.dll"
	case "linux", "freebsd":
		return "libSDL3.so.0"
	case "darwin":
		return "/usr/local/lib/libSDL3.dylib"
	default:
		return ""
	}
}
