package nes

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestPPURegisters(t *testing.T) {
	type result struct {
		t, v uint16
		x, w byte
	}

	type prev result
	type want result

	parse := func(s string) uint64 {
		s = strings.Replace(s, " ", "", -1)
		s = strings.Replace(s, ".", "0", -1)
		n, err := strconv.ParseUint(s, 2, 64)
		if err != nil {
			panic(err)
		}
		return n
	}
	p16 := func(s string) uint16 { return uint16(parse(s)) }
	p8 := func(s string) uint8 { return uint8(parse(s)) }

	ppu := &PPU{}

	tests := []struct {
		name  string
		op    func()
		prev  prev
		want  want
		tmask uint16
	}{
		{
			// tests are from https://wiki.nesdev.com/w/index.php?title=PPU_scrolling&redirect=no#Summary
			name:  "0x2000 write",
			op:    func() { ppu.WritePort(0x2000, 0x00, nil) },
			prev:  prev{t: p16("........ ........"), v: p16("........ ........"), x: p8("........"), w: p8("........")},
			want:  want{t: p16("....00.. ........"), v: p16("........ ........"), x: p8("........"), w: p8("........")},
			tmask: 0x0C00,
		},
		{
			// tests are from https://wiki.nesdev.com/w/index.php?title=PPU_scrolling&redirect=no#Summary
			name:  "0x2002 read",
			op:    func() { ppu.ReadPort(0x2002) },
			prev:  prev{t: p16("....00.. ........"), v: p16("........ ........"), x: p8("........"), w: p8("........")},
			want:  want{t: p16("....00.. ........"), v: p16("........ ........"), x: p8("........"), w: p8(".......0")},
			tmask: 0x0C00,
		},
		{
			// tests are from https://wiki.nesdev.com/w/index.php?title=PPU_scrolling&redirect=no#Summary
			name:  "0x2005 write 1",
			op:    func() { ppu.WritePort(0x2005, 0x7D, nil) },
			prev:  prev{t: p16("....00.. ........"), v: p16("........ ........"), x: p8("........"), w: p8(".......0")},
			want:  want{t: p16("....00.. ...01111"), v: p16("........ ........"), x: p8(".....101"), w: p8(".......1")},
			tmask: 0x0C1F,
		},
		{
			// tests are from https://wiki.nesdev.com/w/index.php?title=PPU_scrolling&redirect=no#Summary
			name:  "0x2005 write 2",
			op:    func() { ppu.WritePort(0x2005, 0x5E, nil) },
			prev:  prev{t: p16("....00.. ...01111"), v: p16("........ ........"), x: p8(".....101"), w: p8(".......1")},
			want:  want{t: p16(".1100001 01101111"), v: p16("........ ........"), x: p8(".....101"), w: p8(".......0")},
			tmask: 0x7FFF,
		},
		{
			// tests are from https://wiki.nesdev.com/w/index.php?title=PPU_scrolling&redirect=no#Summary
			name:  "0x2006 write 1",
			op:    func() { ppu.WritePort(0x2006, 0x3D, nil) },
			prev:  prev{t: p16(".1100001 01101111"), v: p16("........ ........"), x: p8(".....101"), w: p8(".......0")},
			want:  want{t: p16(".0111101 01101111"), v: p16("........ ........"), x: p8(".....101"), w: p8(".......1")},
			tmask: 0x7FFF,
		},
		{
			// tests are from https://wiki.nesdev.com/w/index.php?title=PPU_scrolling&redirect=no#Summary
			name:  "0x2006 write 2",
			op:    func() { ppu.WritePort(0x2006, 0xF0, nil) },
			prev:  prev{t: p16(".0111101 01101111"), v: p16("........ ........"), x: p8(".....101"), w: p8(".......1")},
			want:  want{t: p16(".0111101 11110000"), v: p16(".0111101 11110000"), x: p8(".....101"), w: p8(".......0")},
			tmask: 0x7FFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ppu.t&tt.tmask != tt.prev.t {
				t.Errorf("got prev t = %016b, want prev = %016b", ppu.t&tt.tmask, tt.prev.t)
			}
			if ppu.v != tt.prev.v {
				t.Errorf("got prev v = %016b, want prev = %016b", ppu.v, tt.prev.v)
			}
			if ppu.x != tt.prev.x {
				t.Errorf("got prev x = %016b, want prev = %016b", ppu.x, tt.prev.x)
			}
			if ppu.w != tt.prev.w {
				t.Errorf("got prev w = %016b, want prev = %016b", ppu.w, tt.prev.w)
			}

			tt.op()

			if ppu.t&tt.tmask != tt.want.t {
				t.Errorf("got t = %016b, want = %016b", ppu.t&tt.tmask, tt.want.t)
			}
			if ppu.v != tt.want.v {
				t.Errorf("got v = %016b, want = %016b", ppu.v, tt.want.v)
			}
			if ppu.x != tt.want.x {
				t.Errorf("got x = %016b, want = %016b", ppu.x, tt.want.x)
			}
			if ppu.w != tt.want.w {
				t.Errorf("got w = %016b, want = %016b", ppu.w, tt.want.w)
			}
		})
	}
}

func TestPPUNametableMirroring(t *testing.T) {
	writeData := func(p *PPU, addr uint16, val byte) {
		for i := uint16(0); i < 960; i++ {
			p.Write(addr+i, val)
		}
	}

	t.Run("horizontal", func(t *testing.T) {
		ppu := &PPU{Cartridge: &Cartridge{MirrorMode: Horizontal}}

		// Horizontal
		// 2000 A
		// 2400 A
		// 2800 B
		// 2C00 B
		writeData(ppu, 0x2000, 1)
		writeData(ppu, 0x2800, 2)

		// writes
		if !bytes.Equal(ppu.nametable0[:960], bytes.Repeat([]byte{1}, 960)) {
			t.Fatalf("expected nametable 0 to have been set, got %v", ppu.nametable0[:960])
		}
		if !bytes.Equal(ppu.nametable1[:960], ppu.nametable0[:960]) {
			t.Fatalf("expected nametable 1 to mirror nametable 0, got %v", ppu.nametable1[:960])
		}
		if !bytes.Equal(ppu.nametable2[:960], bytes.Repeat([]byte{2}, 960)) {
			t.Fatalf("expected nametable 2 to have been set, got %v", ppu.nametable2[:960])
		}
		if !bytes.Equal(ppu.nametable3[:960], ppu.nametable2[:960]) {
			t.Fatalf("expected nametable 3 to mirror nametable 2, got %v", ppu.nametable3[:960])
		}

		// reads
		if got := ppu.readNametable(0x2000); got != 1 {
			t.Fatalf("read from 0x%X, want %v, got %v", 0x2000, 1, got)
		}
		if got := ppu.readNametable(0x2400); got != 1 {
			t.Fatalf("read from 0x%X, want %v, got %v", 0x2400, 1, got)
		}
		if got := ppu.readNametable(0x2800); got != 2 {
			t.Fatalf("read from 0x%X, want %v, got %v", 0x2800, 2, got)
		}
		if got := ppu.readNametable(0x2C00); got != 2 {
			t.Fatalf("read from 0x%X, want %v, got %v", 0x2C00, 2, got)
		}
	})

	t.Run("vertical", func(t *testing.T) {
		ppu := &PPU{Cartridge: &Cartridge{MirrorMode: Vertical}}

		// Vertical
		// 2000 A
		// 2400 B
		// 2800 A
		// 2C00 B
		writeData(ppu, 0x2000, 1)
		writeData(ppu, 0x2400, 2)

		// writes
		if !bytes.Equal(ppu.nametable0[:960], bytes.Repeat([]byte{1}, 960)) {
			t.Fatalf("expected nametable 0 to have been set, got %v", ppu.nametable0[:960])
		}
		if !bytes.Equal(ppu.nametable2[:960], ppu.nametable0[:960]) {
			t.Fatalf("expected nametable 2 to mirror nametable 0, got %v", ppu.nametable2[:960])
		}
		if !bytes.Equal(ppu.nametable1[:960], bytes.Repeat([]byte{2}, 960)) {
			t.Fatalf("expected nametable 1 to have been set, got %v", ppu.nametable1[:960])
		}
		if !bytes.Equal(ppu.nametable3[:960], ppu.nametable1[:960]) {
			t.Fatalf("expected nametable 3 to mirror nametable 1, got %v", ppu.nametable3[:960])
		}

		// reads
		if got := ppu.readNametable(0x2000); got != 1 {
			t.Fatalf("read from 0x%X, want %v, got %v", 0x2000, 1, got)
		}
		if got := ppu.readNametable(0x2400); got != 2 {
			t.Fatalf("read from 0x%X, want %v, got %v", 0x2400, 2, got)
		}
		if got := ppu.readNametable(0x2800); got != 1 {
			t.Fatalf("read from 0x%X, want %v, got %v", 0x2800, 1, got)
		}
		if got := ppu.readNametable(0x2C00); got != 2 {
			t.Fatalf("read from 0x%X, want %v, got %v", 0x2C00, 2, got)
		}
	})

}
