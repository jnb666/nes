### Summary

This is a fork of the excellent NES emulator written by Michael Fogleman in Go.

Changes are:
  - Ported from using glfw/openGL/portaudio to SDL3.
  - Title screens imported from libretro project which has a very complete collection.

### Screenshots

![Screenshots](http://i.imgur.com/vD3FXVh.png)

### Title Screens

From [libretro-thumbnails](https://github.com/jnb666/nes/tree/master/thumbnails) project.
Metadata to map from ROM file names to thumbnails is from the [libretro-database](https://github.com/libretro/libretro-database).

### Dependencies

    github.com/Zyko0/go-sdl3

Is a wrapper for the SDL3 shared libraries which can be downloaded from [here](https://github.com/libsdl-org/SDL/releases/latest).

### Installation

The `go get` command will automatically fetch the dependencies listed above,
compile the binary and place it in your `$GOPATH/bin` directory.

    go get github.com/jnb666/nes

### Usage

    nes [rom_file|rom_directory]

1. If no arguments are specified, the program will look for rom files in
the current working directory.

2. If a directory is specified, the program will look for rom files in that
directory.

3. If a file is specified, the program will run that rom.

For 1 & 2, the program will display a menu screen to select which rom to play.
The thumbnails are downloaded from an online database keyed by the md5 sum of
the rom file.

![Menu Screenshot](http://i.imgur.com/pwetBLv.png)

### Controls

Joysticks are supported, although the button mapping is currently hard-coded.
Keyboard controls are indicated below.

| Nintendo              | Emulator    |
| --------------------- | ----------- |
| Up, Down, Left, Right | Arrow Keys  |
| Start                 | Enter       |
| Select                | Right Shift |
| A                     | Z           |
| B                     | X           |
| A (Turbo)             | A           |
| B (Turbo)             | S           |
| Reset                 | R           |

### Mappers

The following mappers have been implemented:

* NROM (0)
* MMC1 (1)
* UNROM (2)
* CNROM (3)
* MMC3 (4)
* AOROM (7)

These mappers cover about 85% of all NES games. I hope to implement more
mappers soon. To see what games should work, consult this list:

[NES Mapper List](http://tuxnes.sourceforge.net/nesmapper.txt)

### Known Issues

* there are some minor issues with PPU timing, but most games work OK anyway
* the APU emulation isn't quite perfect, but not far off

### Documentation

Interested in writing your own emulator? Curious about the NES internals? Here
are some good resources:

* [NES Documentation (PDF)](http://nesdev.com/NESDoc.pdf)
* [NES Reference Guide (Wiki)](https://www.nesdev.org/wiki/NES_reference_guide)
* [6502 CPU Reference](https://www.nesdev.org/obelisk-6502-guide/reference.html)
