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
	flag.Parse()
	paths := getPaths()
	if len(paths) == 0 {
		log.Fatalln("no rom files specified or found")
	}
	ui.Run(paths, *scale)
}

// assume HiDPI - on MacOS it is already taken into account
func defaultScale() int {
	if runtime.GOOS == "darwin" {
		return 3
	} else {
		return 6
	}
}

func getPaths() []string {
	var arg string
	args := os.Args[1:]
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
