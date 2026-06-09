package main

import (
	"flag"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/jnb666/nes/ui"
)

func main() {
	log.SetFlags(0)
	scale := flag.Int("scale", defaultScale(), "set pixel scaling")
	volume := flag.Int("volume", 128, "volume level from 0-255")
	flag.Parse()
	paths := getPaths(flag.Args())
	if len(paths) == 0 {
		log.Fatalln("no rom files specified or found")
	}
	ui.Run(paths, *scale, *volume)
}

// assume HiDPI - on MacOS it is already taken into account
func defaultScale() int {
	if runtime.GOOS == "darwin" {
		return 3
	} else {
		return 6
	}
}

func getPaths(args []string) []string {
	var arg string
	if len(args) == 1 {
		arg = args[0]
	} else {
		arg, _ = os.Getwd()
	}
	info, err := os.Stat(arg)
	if err != nil {
		return nil
	}
	if info.IsDir() {
		infos, err := os.ReadDir(arg)
		if err != nil {
			return nil
		}
		var result []string
		for _, info := range infos {
			name := info.Name()
			if !strings.HasSuffix(name, ".nes") {
				continue
			}
			result = append(result, path.Join(arg, name))
		}
		return result
	} else {
		return []string{arg}
	}
}
