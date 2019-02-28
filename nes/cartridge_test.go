package nes

import (
	"bytes"
	"fmt"
	"testing"
)

type check func(*Cartridge) error
type romfn func([]byte) ([]byte, check)

func TestLoadINES(t *testing.T) {
	empty := func([]byte) ([]byte, check) {
		return []byte{}, isNil
	}
	tooShort := func([]byte) ([]byte, check) {
		return []byte{'N', 'E', 'S', 0x1A, 0, 0, 0, 0, 0, 0}, isNil
	}
	invalidMagic1 := func([]byte) ([]byte, check) {
		return []byte{'N', 'O', 'S', 0x1A, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, isNil
	}
	invalidMagic2 := func([]byte) ([]byte, check) {
		return []byte{'N', 'E', 'S', ' ', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, isNil
	}

	tests := []struct {
		name    string
		rom     []romfn
		wantErr bool
	}{
		{
			name: "empty",
			rom: []romfn{
				empty,
			},
			wantErr: true,
		},
		{
			name: "too short",
			rom: []romfn{
				tooShort,
			},
			wantErr: true,
		},
		{
			name: "invalidMagic 1",
			rom: []romfn{
				invalidMagic1,
			},
			wantErr: true,
		},
		{
			name: "invalidMagic 2",
			rom: []romfn{
				invalidMagic2,
			},
			wantErr: true,
		},
		{
			name: "horizontal mirroring",
			rom: []romfn{
				withHorizontal,
			},
			wantErr: false,
		},
		{
			name: "vertical mirroring",
			rom: []romfn{
				withVertical,
			},
			wantErr: false,
		},
		{
			name: "has ram",
			rom: []romfn{
				withRAM,
			},
			wantErr: false,
		},
		{
			name: "no ram",
			rom: []romfn{
				withoutRAM,
			},
			wantErr: false,
		},
		{
			name: "has trainer",
			rom: []romfn{
				withTrainer,
			},
			wantErr: false,
		},
		{
			name: "no trainer",
			rom: []romfn{
				withoutTrainer,
			},
			wantErr: false,
		},
		{
			name: "has four screen",
			rom: []romfn{
				withFourScreen,
			},
			wantErr: false,
		},
		{
			name: "no four screen",
			rom: []romfn{
				withoutFourScreen,
			},
			wantErr: false,
		},
		{
			name: "with mapper 42",
			rom: []romfn{
				withMapper(42),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rom := []byte{'N', 'E', 'S', 0x1a, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
			var checks []check

			for _, fn := range tt.rom {
				var c check
				rom, c = fn(rom)
				checks = append(checks, c)
			}

			got, err := LoadINES(bytes.NewBuffer(rom))
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadINES() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, fn := range checks {
				if err := fn(got); err != nil {
					t.Errorf("LoadINES(): %s", err)
				}
			}
		})
	}
}

func TestLoadINES_MapperRange(t *testing.T) {
	for i := byte(0); i < 255; i++ {
		rom := []byte{'N', 'E', 'S', 0x1a, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		rom, _ = withMapper(i)(rom)

		got, err := LoadINES(bytes.NewBuffer(rom))
		if err != nil {
			t.Errorf("TestLoadINES_MapperRange() error = %v, wantErr %v", err, nil)
			return
		}

		if got.Mapper != i {
			t.Errorf("TestLoadINES_MapperRange(): wanted mapper %v, got %v", i, got.Mapper)
		}
	}
}

func withHorizontal(rom []byte) ([]byte, check) {
	rom[6] = unset(rom[6], rc1MirrorModeVertical)
	return rom, hasMode(Horizontal)
}

func withVertical(rom []byte) ([]byte, check) {
	rom[6] = set(rom[6], rc1MirrorModeVertical)
	return rom, hasMode(Vertical)
}

func withRAM(rom []byte) ([]byte, check) {
	rom[6] = set(rom[6], rc1SaveRAM)
	return rom, hasRAM(true)
}

func withoutRAM(rom []byte) ([]byte, check) {
	rom[6] = unset(rom[6], rc1SaveRAM)
	return rom, hasRAM(false)
}

func withTrainer(rom []byte) ([]byte, check) {
	rom[6] = set(rom[6], rc1Trainer)
	rom = append(rom, make([]byte, trainerLen)...)
	return rom, hasTrainer(true)
}

func withoutTrainer(rom []byte) ([]byte, check) {
	rom[6] = unset(rom[6], rc1Trainer)
	return rom, hasTrainer(false)
}

func withFourScreen(rom []byte) ([]byte, check) {
	rom[6] = set(rom[6], rc1FourScreen)
	return rom, hasFourScreen(true)
}

func withoutFourScreen(rom []byte) ([]byte, check) {
	rom[6] = unset(rom[6], rc1FourScreen)
	return rom, hasFourScreen(false)
}

func withMapper(m byte) romfn {
	lo := m & 0x0F
	hi := m & 0xF0

	return func(rom []byte) ([]byte, check) {
		rom[6] = (rom[6] & 0x0F) | (lo << 4)
		rom[7] = (rom[7] & 0x0F) | hi
		return rom, hasMapper(m)
	}
}

func isNil(c *Cartridge) error {
	if c != nil {
		return fmt.Errorf("%s() expected %s to be %v, got %v", "isNil", "cartridge", nil, c)
	}
	return nil
}

func hasMode(v MirrorMode) check {
	return func(c *Cartridge) error {
		if c.MirrorMode != v {
			return fmt.Errorf("%s() expected %s to be %v, got %v", "hasMode", "MirrorMode", v, c.MirrorMode)
		}
		return nil
	}
}

func hasRAM(v bool) check {
	return func(c *Cartridge) error {
		if c.SaveRAM != v {
			return fmt.Errorf("%s() expected %s to be %v, got %v", "hasRAM", "SaveRAM", v, c.SaveRAM)
		}
		return nil
	}
}

func hasTrainer(v bool) check {
	var want int
	if v {
		want = trainerLen
	}
	return func(c *Cartridge) error {
		if len(c.Trainer) != want {
			return fmt.Errorf("%s() expected %s to be %v, got %v", "hasTrainer", "len(trainer)", want, len(c.Trainer))
		}
		return nil
	}
}

func hasFourScreen(v bool) check {
	return func(c *Cartridge) error {
		if c.FourScreen != v {
			return fmt.Errorf("%s() expected %s to be %v, got %v", "hasFourScreen", "FourScreen", v, c.FourScreen)
		}
		return nil
	}
}

func hasMapper(v byte) check {
	return func(c *Cartridge) error {
		if c.Mapper != v {
			return fmt.Errorf("%s() expected %s to be %v, got %v", "hasMapper", "Mapper", v, c.Mapper)
		}
		return nil
	}
}

func set(v byte, mask byte) byte {
	return v | mask
}

func unset(v byte, mask byte) byte {
	return v &^ mask
}
