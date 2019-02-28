package nes

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"testing"
)

func TestConsole_nestest(t *testing.T) {
	testRom, err := os.Open("../roms/cpu/nestest/nestest.nes")
	if err != nil {
		t.Fatal("unable to open rom")
	}
	cartridge, err := LoadINES(testRom)
	if err != nil {
		t.Fatal("unable to load rom")
	}

	buf := bytes.NewBuffer(nil)
	out := io.MultiWriter(buf, os.Stderr)

	console := NewConsole(cartridge, 0xC000, out)

	log, err := os.Open("../roms/cpu/nestest/nestest.log.txt")
	if err != nil {
		t.Fatalf("unable to open log: %v", err)
	}

	scanner := bufio.NewScanner(log)

	for scanner.Scan() {
		want := scanner.Bytes()
		want = append(want, '\n')

		console.Step()

		t1, t2 := console.Read(0x02), console.Read(0x03)
		if t1 != 0 || t2 != 0 {
			t.Fatalf("%02x%02x", t1, t2)
		}

		if got := buf.Bytes(); !bytes.Equal(got, want) {
			t.Fatalf("nestest: want %q, got %q", want, got)
		}

		buf.Reset()
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("unable to read log: %v", err)
	}
}
