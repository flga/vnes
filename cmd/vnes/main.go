package main

//go:generate go run ../embed -root ../../ -o assets.go -exclude ../../assets/**/*.{ttf} ../../assets/**

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/flga/nes/cmd/internal/gui"
	"github.com/flga/nes/nes"
	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	runtime.LockOSThread()
}

func initSDL() (func(), error) {
	if err := sdl.Init(sdl.INIT_GAMECONTROLLER | sdl.INIT_JOYSTICK | sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return func() {}, fmt.Errorf("initSDL: unable to init sdl: %s", err)
	}

	return sdl.Quit, nil
}

func initTTF() (gui.FontMap, error) {
	fontPath := filepath.Join("assets", "runescape_uf.fnt")
	f, err := assets.Open(fontPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	openFunc := func(path string) (io.ReadCloser, error) {
		// return os.Open(filepath.Join("assets", path))
		return assets.Open(filepath.Join("assets", path))
	}

	fontMap := make(gui.FontMap)
	if err := fontMap.LoadXML(f, openFunc); err != nil {
		return nil, fmt.Errorf("initTTF: unable to load font %s: %s", fontPath, err)
	}

	return fontMap, nil
}

func loadRom(path string) (*nes.Cartridge, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open rom: %s", err)
	}
	defer f.Close()

	return nes.LoadINES(f)
}

func run(romPath string, trace bool, cpuprof, memprof string) error {
	var out io.Writer
	if trace {
		out = os.Stderr
	}

	console := nes.NewConsole(0, out)

	if romPath != "" {
		cartridge, err := loadRom(romPath)
		if err != nil {
			return err
		}
		console.Load(cartridge)
	}

	quitSDL, err := initSDL()
	if err != nil {
		return err
	}
	defer quitSDL()

	fontCache, err := initTTF()
	if err != nil {
		return err
	}

	audioEngine := &audioEngine{
		AudioChan: console.APU.Channel(),
	}

	if err := audioEngine.init(true); err != nil {
		return err
	}
	defer audioEngine.quit()

	zoom := 4
	engine, err := newEngine("vnes", zoom, audioEngine, fontCache)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func() {
		<-sigchan
		cancel()
	}()

	if cpuprof != "" {
		cpuf, err := os.Create(cpuprof)
		if err != nil {
			return fmt.Errorf("could not create CPU profile: %s", err)
		}
		defer cpuf.Close()
		if err := pprof.StartCPUProfile(cpuf); err != nil {
			return fmt.Errorf("could not start CPU profile: %s", err)
		}
		defer pprof.StopCPUProfile()
	}
	if memprof != "" {
		memf, err := os.Create(memprof)
		if err != nil {
			return fmt.Errorf("could not create memory profile: %s", err)
		}
		defer memf.Close()
		defer func() {
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(memf); err != nil {
				panic("could not write memory profile: " + err.Error())
			}
		}()
	}

	return engine.run(ctx, console)
}

func main() {

	trace := flag.Bool("trace", false, "Print a trace of the CPU execution into stdout. WARNING: this is not fully implemented and will bug out graphics")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile := flag.String("memprofile", "", "write memory profile to file")

	flag.Parse()

	// if cartridge.Mapper != 0 {
	// 	panic(fmt.Sprintf("Unexpected mapper %d\n", cartridge.Mapper))
	// }

	if err := run(flag.Arg(0), *trace, *cpuprofile, *memprofile); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
