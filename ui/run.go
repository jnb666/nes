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

	// initialize audio
	audio := NewAudio()
	if err := audio.Start(); err != nil {
		log.Fatalln(err)
	}
	defer audio.Stop()

	// create window
	window, renderer, err := sdl.CreateWindowAndRenderer(title, width*scale, height*scale, sdl.WINDOW_HIGH_PIXEL_DENSITY)
	if err != nil {
		log.Fatalln(err)
	}
	defer window.Destroy()
	if err = renderer.SetVSync(1); err != nil {
		log.Fatalln(err)
	}

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
