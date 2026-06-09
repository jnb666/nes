package ui

import (
	"image"
	"log"
	"path/filepath"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
)

const textureSize = 4096
const textureDim = textureSize / 256
const textureCount = textureDim * textureDim

type Texture struct {
	texture *sdl.Texture
	lookup  map[string]int
	reverse [textureCount]string
	access  [textureCount]int
	counter int
}

func NewTexture(renderer *sdl.Renderer) *Texture {
	return &Texture{
		texture: createTexture(renderer, textureSize, textureSize, sdl.TEXTUREACCESS_STATIC),
		lookup:  make(map[string]int),
	}
}

func (t *Texture) Lookup(path string) (x, y, dx, dy float32) {
	if index, ok := t.lookup[path]; ok {
		return t.coord(index)
	} else {
		return t.coord(t.load(path))
	}
}

func (t *Texture) mark(index int) {
	t.counter++
	t.access[index] = t.counter
}

func (t *Texture) lru() int {
	minIndex := 0
	minValue := t.counter + 1
	for i, n := range t.access {
		if n < minValue {
			minIndex = i
			minValue = n
		}
	}
	return minIndex
}

func (t *Texture) coord(index int) (x, y, dx, dy float32) {
	x = float32(index%textureDim) * 256
	y = float32(index/textureDim) * 256
	dx = 256
	dy = 240
	return
}

func (t *Texture) load(path string) int {
	index := t.lru()
	delete(t.lookup, t.reverse[index])
	t.mark(index)
	t.lookup[path] = index
	t.reverse[index] = path
	x := int32((index % textureDim) * 256)
	y := int32((index / textureDim) * 256)
	im := copyImage(t.loadThumbnail(path))
	size := im.Rect.Size()
	t.texture.Update(&sdl.Rect{X: x, Y: y, W: int32(size.X), H: int32(size.Y)}, im.Pix, int32(im.Stride))
	return index
}

func (t *Texture) loadThumbnail(romPath string) image.Image {
	hash, err := hashFile(romPath)
	if err != nil {
		return genericThumbnail(romPath)
	}
	thumbnail, err := loadPNG(thumbnailPath(hash, romPath))
	if err != nil {
		return genericThumbnail(romPath)
	}
	return thumbnail
}

func genericThumbnail(romPath string) image.Image {
	name := strings.TrimSuffix(filepath.Base(romPath), ".nes")
	name = strings.Replace(name, "_", " ", -1)
	name = strings.Title(name)
	log.Printf("Error getting thumbnail for %s", filepath.Base(romPath))
	return CreateGenericThumbnail(name)
}
