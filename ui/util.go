package ui

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/jnb666/nes/nes"
)

var homeDir = getHomeDir()

func getHomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}
	return dir
}

func thumbnailPath(hash, romPath string) string {
	return filepath.Join(filepath.Dir(romPath), "thumbnails", hash+".png")
}

func sramPath(hash string, snapshot int) string {
	if snapshot >= 0 {
		return filepath.Join(homeDir, ".nes", "sram", fmt.Sprintf("%s-%d.dat", hash, snapshot))
	}
	return filepath.Join(homeDir, ".nes", "sram", hash+".dat")
}

func savePath(hash string, snapshot int) string {
	if snapshot >= 0 {
		return filepath.Join(homeDir, ".nes", "save", fmt.Sprintf("%s-%d.dat", hash, snapshot))
	}
	return filepath.Join(homeDir, ".nes", "save", hash+".dat")
}

func getTime() float64 {
	return time.Duration(sdl.TicksNS()).Seconds()
}

func readKey(key sdl.Scancode) bool {
	return sdl.GetKeyboardState()[key]
}

func readKeys(turbo bool) (result [8]bool) {
	keys := sdl.GetKeyboardState()
	result[nes.ButtonA] = keys[sdl.SCANCODE_Z] || (turbo && keys[sdl.SCANCODE_A])
	result[nes.ButtonB] = keys[sdl.SCANCODE_X] || (turbo && keys[sdl.SCANCODE_S])
	result[nes.ButtonSelect] = keys[sdl.SCANCODE_RSHIFT]
	result[nes.ButtonStart] = keys[sdl.SCANCODE_RETURN]
	result[nes.ButtonUp] = keys[sdl.SCANCODE_UP]
	result[nes.ButtonDown] = keys[sdl.SCANCODE_DOWN]
	result[nes.ButtonLeft] = keys[sdl.SCANCODE_LEFT]
	result[nes.ButtonRight] = keys[sdl.SCANCODE_RIGHT]
	return result
}

func joystickReset(d *Director) bool {
	return d.buttons[0][nes.ButtonSelect] && d.buttons[0][nes.ButtonB] ||
		d.buttons[1][nes.ButtonSelect] && d.buttons[1][nes.ButtonB]
}

func combineButtons(a, b [8]bool) [8]bool {
	var result [8]bool
	for i := 0; i < 8; i++ {
		result[i] = a[i] || b[i]
	}
	return result
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(data)), nil
}

func createTexture(r *sdl.Renderer, width, height int, access sdl.TextureAccess) *sdl.Texture {
	texture, err := r.CreateTexture(sdl.PIXELFORMAT_ABGR8888, access, width, height)
	if err != nil {
		log.Fatal(err)
	}
	texture.SetScaleMode(sdl.SCALEMODE_NEAREST)
	return texture
}

func setTexture(tex *sdl.Texture, im *image.RGBA) {
	buf, _, err := tex.Lock(nil)
	if err != nil {
		panic(err)
	}
	copy(buf, im.Pix)
	tex.Unlock()
}

func copyImage(src image.Image) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Rect, src, image.Point{}, draw.Src)
	return dst
}

func loadPNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

func savePNG(path string, im image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, im)
}

func saveGIF(path string, frames []image.Image) error {
	var palette []color.Color
	for _, c := range nes.Palette {
		palette = append(palette, c)
	}
	g := gif.GIF{}
	for i, src := range frames {
		if i%3 != 0 {
			continue
		}
		dst := image.NewPaletted(src.Bounds(), palette)
		draw.Draw(dst, dst.Rect, src, image.Point{}, draw.Src)
		g.Image = append(g.Image, dst)
		g.Delay = append(g.Delay, 5)
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return gif.EncodeAll(file, &g)
}

func screenshot(im image.Image) {
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("%03d.png", i)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			savePNG(path, im)
			return
		}
	}
}

func animation(frames []image.Image) {
	for i := 0; i < 1000; i++ {
		path := fmt.Sprintf("%03d.gif", i)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			saveGIF(path, frames)
			return
		}
	}
}

func writeSRAM(filename string, sram []byte) error {
	dir, _ := path.Split(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return binary.Write(file, binary.LittleEndian, sram)
}

func readSRAM(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	sram := make([]byte, 0x2000)
	if err := binary.Read(file, binary.LittleEndian, sram); err != nil {
		return nil, err
	}
	return sram, nil
}
