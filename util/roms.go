package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jnb666/nes/nes"
)

func testRom(path string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	console, err := nes.NewConsole(path)
	if err != nil {
		return err
	}
	console.StepSeconds(3)
	return nil
}

func main() {
	failedDir := flag.String("failed", "", "optionally move unsupported roms to this directory")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalln("Usage: go run util/roms.go [-failed dir] roms_directory")
	}
	dir := flag.Arg(0)
	infos, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, info := range infos {
		name := info.Name()
		if !strings.HasSuffix(name, ".nes") {
			continue
		}
		file := filepath.Join(dir, name)
		err := testRom(file)
		if err == nil {
			fmt.Println("OK  ", name)
		} else {
			fmt.Printf("FAIL %-60s %s\n", name, err)
			if failedDir != nil {
				err := os.Rename(file, filepath.Join(*failedDir, name))
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
