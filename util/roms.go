package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"

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
	log.SetFlags(0)
	testROMs := flag.Bool("test", true, "test roms to verify is supported")
	thumbsData := flag.String("thumbs", "", "attempt to get thumbnail files and store in thumbnails subdirectory")
	failedDir := flag.String("failed", "", "optionally move unsupported roms to this directory")

	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatalln("Usage: go run util/roms.go [opts] roms_directory")
	}
	dir := flag.Arg(0)
	infos, err := os.ReadDir(dir)
	check(err)

	thumbnails := filepath.Join(dir, "thumbnails")
	var roms map[string]string
	if *thumbsData != "" {
		if _, err = os.Stat(thumbnails); os.IsNotExist(err) {
			err = os.Mkdir(thumbnails, 0777)
		}
		check(err)
		roms, err = loadThumbnailsDB(*thumbsData)
		check(err)
	}

	for _, info := range infos {
		name := info.Name()
		if !strings.HasSuffix(name, ".nes") {
			continue
		}
		file := filepath.Join(dir, name)
		fmt.Printf("%-60s", strings.TrimSuffix(name, ".nes"))
		if *testROMs {
			err := testRom(file)
			if err != nil {
				fmt.Printf("  FAIL %s\n", err)
				if failedDir != nil {
					err := os.Rename(file, filepath.Join(*failedDir, name))
					check(err)
				}
				continue
			}
		}
		if *thumbsData != "" {
			err := checkThumbnail(roms, *thumbsData, thumbnails, file)
			check(err)
		}
		fmt.Println()
	}
}

func loadThumbnailsDB(dir string) (map[string]string, error) {
	file := filepath.Join(dir, "index.dat")
	fmt.Println("reading", file)
	db, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	roms := map[string]string{}
	scanner := newScanner(bufio.NewReader(db))
	for scanner.error() == nil {
		r := scanner.nextRecord()
		if r.Type == "game" {
			name, err := r.getName()
			if err != nil {
				return roms, err
			}
			sum, err := r.getmd5()
			if err != nil {
				return roms, err
			}
			roms[sum] = name
		}
	}
	fmt.Printf("read metadata for %d roms from %s\n", len(roms), file)
	if err = scanner.error(); errors.Is(err, io.EOF) {
		return roms, nil
	} else {
		return roms, err
	}
}

var replaceChars = []string{"&", "*", "/", ":", "`", "<", ">", "?", "\\", "|", "\""}

func checkThumbnail(roms map[string]string, baseDir, thumbsDir, romPath string) error {
	hash, err := hashFile(romPath)
	if err != nil {
		return err
	}
	filename := filepath.Join(thumbsDir, hash+".png")
	if _, err = os.Stat(filename); !os.IsNotExist(err) {
		return err
	}
	name, ok := roms[hash]
	if !ok {
		fmt.Print("  no metadata found")
		return nil
	}
	for _, ch := range replaceChars {
		name = strings.ReplaceAll(name, ch, "_")
	}
	src := filepath.Join(baseDir, "libretro_thumbnails", "Named_Titles", name+".png")
	if _, err := os.Stat(src); err != nil {
		fmt.Printf("  thumbnail for %q not found", name)
		return nil
	}
	if err = os.Link(src, filename); err != nil {
		return err
	}
	fmt.Printf("  linked thumbnail to %q", name)
	return nil
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(data)), nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Scanner struct {
	scanner.Scanner
	token  rune
	errors []error
}

type Record struct {
	Type  string
	Value any
}

func (r Record) getName() (string, error) {
	vals, ok := r.Value.(Values)
	if r.Type != "game" || !ok || vals["name"] == nil {
		return "", fmt.Errorf("malformed game record: %#v", r)
	}
	return strconv.Unquote(vals["name"].(string))
}

func (r Record) getmd5() (string, error) {
	vals, ok := r.Value.(Values)
	if r.Type != "game" || !ok {
		return "", fmt.Errorf("malformed game record: %#v", r)
	}
	rvals, ok := vals["rom"].(Values)
	if !ok || rvals["md5"] == nil {
		return "", fmt.Errorf("malformed rom record: %#v", r)
	}
	return rvals["md5"].(string), nil
}

type Values map[string]any

func newScanner(src io.Reader) *Scanner {
	s := new(Scanner)
	s.Init(src)
	s.Mode = scanner.ScanIdents | scanner.ScanStrings
	s.IsIdentRune = func(ch rune, i int) bool {
		return unicode.IsLetter(ch) || unicode.IsDigit(ch)
	}
	s.Error = func(_ *scanner.Scanner, msg string) {
		s.addError(msg)
	}
	s.next()
	return s
}

func (s *Scanner) addError(msg string) {
	s.errors = append(s.errors, fmt.Errorf("parse error at %s: %s", s.Position, msg))
}

func (s *Scanner) error() error {
	return errors.Join(s.errors...)
}

func (s *Scanner) next() {
	s.token = s.Scan()
	if s.token == scanner.EOF {
		s.errors = append(s.errors, io.EOF)
	}
}

// when next record is called s.token should be pointing to it's start
func (s *Scanner) nextRecord() (r Record) {
	if len(s.errors) > 0 {
		return r
	}
	if s.token != scanner.Ident {
		s.addError("expected record type - got " + s.TokenText())
		return r
	}
	r.Type = s.TokenText()

	s.next()
	if len(s.errors) > 0 {
		return r
	}
	if s.token != '(' {
		if s.token == scanner.Ident || s.token == scanner.String {
			r.Value = s.TokenText()
			s.next()
			return r
		}
		s.addError("expected ident or string value for " + r.Type + " - got " + s.TokenText())
		return r
	}
	value := Values{}
	s.next()
	for s.token != ')' && len(s.errors) == 0 {
		child := s.nextRecord()
		value[child.Type] = child.Value
	}
	r.Value = value
	s.next()
	return r
}
