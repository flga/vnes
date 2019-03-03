package nes

// import (
// 	"os"
// 	"testing"
// )

// func TestCPU_Basic(t *testing.T) {
// 	prg := make([]byte, prgRomSize)
// 	copy(prg, []byte{0xAD, 0xFF, 0x00, 0x8D, 0x00, 0x00})
// 	cartridge := &Cartridge{
// 		PRG: prg,
// 	}

// 	bus := &SysBus{
// 		Cartridge: cartridge,
// 		RAM:       NewRAM(),
// 	}
// 	bus.Write(0x00FF, 42)

// 	c := NewCPU(os.Stderr)
// 	c.Init(bus)
// 	c.SetPC(0x8000)

// 	c.Execute(bus, nil)
// 	if c.A != 42 {
// 		t.Errorf("expected A to be %v, got %v", 42, c.A)
// 	}
// 	c.Execute(bus, nil)
// 	if v := bus.Read(0x0000); v != 42 {
// 		t.Errorf("expected 0x0000 to be %v, got %v", 42, v)
// 	}
// }

// // func TestCPU_resolveAddress(t *testing.T) {
// // 	type args struct {
// // 		pc   uint16
// // 		x, y byte
// // 		mode AddressingMode
// // 		bus  *SysBus
// // 	}

// // 	newBus := func(ram ...byte) *SysBus {
// // 		busRAM := make([]byte, ramSize)
// // 		copy(busRAM, ram)
// // 		return &SysBus{
// // 			RAM: busRAM,
// // 		}
// // 	}

// // 	tests := []struct {
// // 		name        string
// // 		args        args
// // 		wantAddress uint16
// // 		wantPaged   bool
// // 		wantPC      uint16
// // 	}{
// // 		{
// // 			name: "Immediate",
// // 			args: args{
// // 				mode: Immediate,
// // 				bus:  newBus(0x2A, 0x01),
// // 			},
// // 			wantAddress: 0,
// // 			wantPaged:   false,
// // 			wantPC:      1,
// // 		},
// // 		{
// // 			name: "ZeroPage",
// // 			args: args{
// // 				mode: ZeroPage,
// // 				bus:  newBus(0x2A, 0x01),
// // 			},
// // 			wantAddress: 0x2A,
// // 			wantPaged:   false,
// // 			wantPC:      1,
// // 		},
// // 		{
// // 			name: "Absolute",
// // 			args: args{
// // 				mode: Absolute,
// // 				bus:  newBus(0x2A, 0x01),
// // 			},
// // 			wantAddress: 0x012A,
// // 			wantPaged:   false,
// // 			wantPC:      2,
// // 		},
// // 		{
// // 			name: "Relative",
// // 			args: args{
// // 				mode: Relative,
// // 				bus:  newBus(0x2A, 0x06),
// // 			},
// // 			wantAddress: 0x2A + 1,
// // 			wantPaged:   false,
// // 			wantPC:      1,
// // 		},
// // 		{
// // 			name: "Relative+",
// // 			args: args{
// // 				pc:   128,
// // 				mode: Relative,
// // 				bus: newBus(
// // 					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// // 					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// // 					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// // 					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// // 					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// // 					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// // 					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// // 					0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
// // 					0x7F, 0x06),
// // 			},
// // 			wantAddress: 0xFF + 1,
// // 			wantPaged:   true,
// // 			wantPC:      129,
// // 		},
// // 		// {name: "Implied"},
// // 		{
// // 			name: "IndexedX",
// // 			args: args{
// // 				mode: IndexedX,
// // 				x:    0x03,
// // 				y:    0x04,
// // 				bus:  newBus(0x2A, 0x01),
// // 			},
// // 			wantAddress: 0x012A + 0x03,
// // 			wantPaged:   false,
// // 			wantPC:      2,
// // 		},
// // 		{
// // 			name: "IndexedX+",
// // 			args: args{
// // 				mode: IndexedX,
// // 				x:    0x03,
// // 				y:    0x04,
// // 				bus:  newBus(0xFF, 0x01),
// // 			},
// // 			wantAddress: (0x00FF | 0x0100) + 0x03,
// // 			wantPaged:   true,
// // 			wantPC:      2,
// // 		},
// // 		{
// // 			name: "IndexedY",
// // 			args: args{
// // 				mode: IndexedY,
// // 				x:    0x00,
// // 				y:    0x04,
// // 				bus:  newBus(0x2A, 0x01),
// // 			},
// // 			wantAddress: 0x012A + 0x04,
// // 			wantPaged:   false,
// // 			wantPC:      2,
// // 		},
// // 		{
// // 			name: "IndexedY+",
// // 			args: args{
// // 				mode: IndexedY,
// // 				x:    0x00,
// // 				y:    0x04,
// // 				bus:  newBus(0xFF, 0x01),
// // 			},
// // 			wantAddress: (0x0100 | 0x0FF) + 0x04,
// // 			wantPaged:   true,
// // 			wantPC:      2,
// // 		},
// // 		{
// // 			name: "ZeroPageIndexedX",
// // 			args: args{
// // 				mode: ZeroPageIndexedX,
// // 				x:    0x03,
// // 				y:    0x04,
// // 				bus:  newBus(0x2A, 0x01),
// // 			},
// // 			wantAddress: 0x2A + 0x03,
// // 			wantPaged:   false,
// // 			wantPC:      1,
// // 		},
// // 		{
// // 			name: "ZeroPageIndexedY",
// // 			args: args{
// // 				mode: ZeroPageIndexedY,
// // 				x:    0x03,
// // 				y:    0x04,
// // 				bus:  newBus(0x2A, 0x01),
// // 			},
// // 			wantAddress: 0x2A + 0x04,
// // 			wantPaged:   false,
// // 			wantPC:      1,
// // 		},
// // 		{
// // 			name: "Indirect",
// // 			args: args{
// // 				mode: Indirect,
// // 				x:    0x03,
// // 				y:    0x04,
// // 				bus:  newBus(0x2A, 0x01),
// // 			},
// // 			wantAddress: 0x0100 | 0x002A, //TODO
// // 			wantPaged:   false,
// // 			wantPC:      2,
// // 		},
// // 		{
// // 			name: "PreIndexedIndirect",
// // 			args: args{
// // 				mode: PreIndexedIndirect,
// // 				x:    0x03,
// // 				y:    0x04,
// // 				bus:  newBus(0x02, 0, 0, 0, 0, 0x2A),
// // 			},
// // 			wantAddress: 0x2A,
// // 			wantPaged:   false,
// // 			wantPC:      1,
// // 		},
// // 		{
// // 			name: "PreIndexedIndirect Overflow",
// // 			args: args{
// // 				mode: PreIndexedIndirect,
// // 				x:    0x10,
// // 				y:    0x04,
// // 				bus:  newBus(0xFF, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x2A),
// // 			},
// // 			wantAddress: 0x2A,
// // 			wantPaged:   false,
// // 			wantPC:      1,
// // 		},
// // 		{
// // 			name: "PostIndexedIndirect",
// // 			args: args{
// // 				mode: PostIndexedIndirect,
// // 				x:    0x03,
// // 				y:    0x04,
// // 				bus:  newBus(0x02, 0, 0x2A, 0x01),
// // 			},
// // 			wantAddress: (0x0100 | 0x002A) + 0x04,
// // 			wantPaged:   false,
// // 			wantPC:      1,
// // 		},
// // 		{
// // 			name: "PostIndexedIndirect+",
// // 			args: args{
// // 				mode: PostIndexedIndirect,
// // 				x:    0x03,
// // 				y:    0x04,
// // 				bus:  newBus(0x02, 0, 0xFF, 0x01),
// // 			},
// // 			wantAddress: (0x0100 | 0x00FF) + 0x04,
// // 			wantPaged:   true,
// // 			wantPC:      1,
// // 		},
// // 	}
// // 	for _, tt := range tests {
// // 		t.Run(tt.name, func(t *testing.T) {
// // 			c := NewCPU(os.Stderr)
// // 			c.SetPC(tt.args.pc)
// // 			c.X = tt.args.x
// // 			c.Y = tt.args.y
// // 			gotIntermediateAddress, gotAddress := c.resolveAddress(tt.args.mode, tt.args.bus)
// // 			_ = gotIntermediateAddress
// // 			//TODO intermediate address
// // 			if gotAddress != tt.wantAddress {
// // 				t.Errorf("CPU.resolveAddress() gotAddress = %v, want %v", gotAddress, tt.wantAddress)
// // 			}
// // 			if gotPaged != tt.wantPaged {
// // 				t.Errorf("CPU.resolveAddress() gotPaged = %v, want %v", gotPaged, tt.wantPaged)
// // 			}
// // 			if c.PC != tt.wantPC {
// // 				t.Errorf("CPU.resolveAddress() PC = %v, want %v", c.PC, tt.wantPC)
// // 			}
// // 		})
// // 	}
// // }

// func TestCPU_ADC(t *testing.T) {
// 	newBus := func(ram ...byte) *SysBus {
// 		busRAM := NewRAM()
// 		copy(busRAM.data, ram)
// 		return &SysBus{
// 			RAM: busRAM,
// 		}
// 	}
// 	type args struct {
// 		a    byte
// 		addr uint16
// 		bus  *SysBus
// 	}
// 	type want struct {
// 		carry    bool
// 		overflow bool
// 		a        byte
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want want
// 	}{
// 		// M7 N7 C6		C7 S7 V		Carry / Overflow							Hex				Unsigned	Signed
// 		// 0  0  0		0  0  0		No unsigned carry or signed 	overflow	0x50+0x10=0x60	80+16=96	80+16=96
// 		{
// 			name: "0 No unsigned carry or signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0x50,
// 				bus:  newBus(0x10),
// 			},
// 			want: want{
// 				a:        0x60,
// 				carry:    false,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 S7 V		Carry / Overflow							Hex				Unsigned	Signed
// 		// 0  0  1		0  1  1		No unsigned carry but signed 	overflow	0x50+0x50=0xa0	80+80=160	80+80=-96
// 		{
// 			name: "1 No unsigned carry but signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0x50,
// 				bus:  newBus(0x50),
// 			},
// 			want: want{
// 				a:        0xA0,
// 				carry:    false,
// 				overflow: true,
// 			},
// 		},
// 		// M7 N7 C6		C7 S7 V		Carry / Overflow							Hex				Unsigned	Signed
// 		// 0  1  0		0  1  0		No unsigned carry or signed 	overflow	0x50+0x90=0xe0	80+144=224	80+-112=-32
// 		{
// 			name: "2 No unsigned carry or signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0x50,
// 				bus:  newBus(0x90),
// 			}, want: want{
// 				a:        0xE0,
// 				carry:    false,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 S7 V		Carry / Overflow							Hex				Unsigned	Signed
// 		// 0  1  1		1  0  0		Unsigned carry, but no signed 	overflow	0x50+0xd0=0x120	80+208=288	80+-48=32
// 		{
// 			name: "3 Unsigned carry, but no signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0x50,
// 				bus:  newBus(0xD0),
// 			}, want: want{
// 				a:        0x20,
// 				carry:    true,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 S7 V		Carry / Overflow							Hex				Unsigned	Signed
// 		// 1  0  0		0  1  0		No unsigned carry or signed overflow		0xd0+0x10=0xe0	208+16=224	-48+16=-32
// 		{
// 			name: "4 No unsigned carry or signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0xD0,
// 				bus:  newBus(0x10),
// 			}, want: want{
// 				a:        0xE0,
// 				carry:    false,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 S7 V		Carry / Overflow							Hex				Unsigned	Signed
// 		// 1  0  1		1  0  0		Unsigned carry but no signed overflow		0xd0+0x50=0x120	208+80=288	-48+80=32
// 		{
// 			name: "5 Unsigned carry but no signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0xD0,
// 				bus:  newBus(0x50),
// 			}, want: want{
// 				a:        0x20,
// 				carry:    true,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 S7 V		Carry / Overflow							Hex				Unsigned	Signed
// 		// 1  1  0		1  0  1		Unsigned carry and signed overflow			0xd0+0x90=0x160	208+144=352	-48+-112=96
// 		{
// 			name: "6 Unsigned carry and signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0xD0,
// 				bus:  newBus(0x90),
// 			}, want: want{
// 				a:        0x60,
// 				carry:    true,
// 				overflow: true,
// 			},
// 		},
// 		// M7 N7 C6		C7 S7 V		Carry / Overflow							Hex				Unsigned	Signed
// 		// 1  1  1		1  1  0		Unsigned carry, but no signed overflow		0xd0+0xd0=0x1a0	208+208=416	-48+-48=-96
// 		{
// 			name: "7 Unsigned carry, but no signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0xD0,
// 				bus:  newBus(0xD0),
// 			}, want: want{
// 				a:        0xA0,
// 				carry:    true,
// 				overflow: false,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := NewCPU(os.Stderr)
// 			c.SetPC(0)
// 			c.A = tt.args.a

// 			c.ADC(tt.args.bus, 0, tt.args.addr)
// 			gotCarry := c.P&Carry > 0
// 			gotOverflow := c.P&Overflow > 0
// 			if c.A != tt.want.a {
// 				t.Errorf("CPU.ADC(%x,%x) = got A = %x, want = %x", tt.args.a, tt.args.bus.Read(0), c.A, tt.want.a)
// 			}
// 			if gotCarry != tt.want.carry {
// 				t.Errorf("CPU.ADC(%x,%x) = got carry %v, want %v", tt.args.a, tt.args.bus.Read(0), gotCarry, tt.want.carry)
// 			}
// 			if gotOverflow != tt.want.overflow {
// 				t.Errorf("CPU.ADC(%x,%x) = got overflow %v, want %v", tt.args.a, tt.args.bus.Read(0), gotOverflow, tt.want.overflow)
// 			}
// 		})
// 	}
// }

// func TestCPU_SBC(t *testing.T) {
// 	t.Skip()
// 	newBus := func(ram ...byte) *SysBus {
// 		busRAM := NewRAM()
// 		copy(busRAM.data, ram)
// 		return &SysBus{
// 			RAM: busRAM,
// 		}
// 	}
// 	type args struct {
// 		a    byte
// 		addr uint16
// 		bus  *SysBus
// 	}
// 	type want struct {
// 		carry    bool
// 		overflow bool
// 		a        byte
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want want
// 	}{
// 		// M7 N7 C6		C7 B S7 V		Borrow / Overflow						Hex				Unsigned	Signed
// 		// 0  1  0		0  1 0  0		Unsigned borrow but no signed overflow	0x50-0xF0=0x60	80-240=96	80--16=96
// 		{
// 			name: "0 Unsigned borrow but no signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0x50,
// 				bus:  newBus(0xF0),
// 			},
// 			want: want{
// 				a:        0x60,
// 				carry:    false,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 B S7 V		Borrow / Overflow						Hex				Unsigned	Signed
// 		// 0  1  1		0  1 1  1		Unsigned borrow and signed overflow	0x50-0xB0=0xA0	80-176=160	80--80=-96
// 		{
// 			name: "1 Unsigned borrow and signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0x50,
// 				bus:  newBus(0xB0),
// 			},
// 			want: want{
// 				a:        0xA0,
// 				carry:    false,
// 				overflow: true,
// 			},
// 		},
// 		// M7 N7 C6		C7 B S7 V		Borrow / Overflow						Hex				Unsigned	Signed
// 		// 0  0  0		0  1 1  0		Unsigned borrow but no signed overflow	0x50-0x70=0xE0	80-112=224	80-112=-32
// 		{
// 			name: "2 Unsigned borrow but no signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0x50,
// 				bus:  newBus(0x70),
// 			},
// 			want: want{
// 				a:        0xE0,
// 				carry:    false,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 B S7 V		Borrow / Overflow						Hex				Unsigned	Signed
// 		// 0  0  1		1  0 0  0		No unsigned borrow or signed overflow	0x50-0x30=0x120	80-48=32	80-48=32
// 		{
// 			name: "3 No unsigned borrow or signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0x50,
// 				bus:  newBus(0x30),
// 			},
// 			want: want{
// 				a:        0x20,
// 				carry:    true,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 B S7 V		Borrow / Overflow						Hex				Unsigned	Signed
// 		// 1  1  0		0  1 1  0		Unsigned borrow but no signed overflow	0xD0-0xF0=0xE0	208-240=224	-48--16=-32
// 		{
// 			name: "4 Unsigned borrow but no signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0xD0,
// 				bus:  newBus(0xF0),
// 			},
// 			want: want{
// 				a:        0xE0,
// 				carry:    false,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 B S7 V		Borrow / Overflow						Hex				Unsigned	Signed
// 		// 1  1  1		1  0 0  0		No unsigned borrow or signed overflow	0xD0-0xB0=0x120	208-176=32	-48--80=32
// 		{
// 			name: "5 No unsigned borrow or signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0xD0,
// 				bus:  newBus(0xB0),
// 			},
// 			want: want{
// 				a:        0x20,
// 				carry:    true,
// 				overflow: false,
// 			},
// 		},
// 		// M7 N7 C6		C7 B S7 V		Borrow / Overflow						Hex				Unsigned	Signed
// 		// 1  0  0		1  0 0  1		No unsigned borrow but signed overflow	0xD0-0x70=0x160	208-112=96	-48-112=96
// 		{
// 			name: "6 No unsigned borrow but signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0xD0,
// 				bus:  newBus(0x70),
// 			},
// 			want: want{
// 				a:        0x60,
// 				carry:    true,
// 				overflow: true,
// 			},
// 		},
// 		// M7 N7 C6		C7 B S7 V		Borrow / Overflow						Hex				Unsigned	Signed
// 		// 1  0  1		1  0 1  0		No unsigned borrow or signed overflow	0xD0-0x30=0x1A0	208-48=160	-48-48=-96
// 		{
// 			name: "7 No unsigned borrow or signed overflow",
// 			args: args{
// 				addr: 0,
// 				a:    0xD0,
// 				bus:  newBus(0x30),
// 			},
// 			want: want{
// 				a:        0xA0,
// 				carry:    true,
// 				overflow: false,
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := NewCPU(os.Stderr)
// 			c.SetPC(0)
// 			c.A = tt.args.a

// 			c.SBC(tt.args.bus, 0, tt.args.addr)
// 			gotCarry := c.P&Carry > 0
// 			gotOverflow := c.P&Overflow > 0
// 			if c.A != tt.want.a {
// 				t.Errorf("CPU.SBC(%x,%x) = got A = %x, want = %x", tt.args.a, tt.args.bus.Read(0), c.A, tt.want.a)
// 			}
// 			if gotCarry != tt.want.carry {
// 				t.Errorf("CPU.SBC(%x,%x) = got carry %v, want %v", tt.args.a, tt.args.bus.Read(0), gotCarry, tt.want.carry)
// 			}
// 			if gotOverflow != tt.want.overflow {
// 				t.Errorf("CPU.SBC(%x,%x) = got overflow %v, want %v", tt.args.a, tt.args.bus.Read(0), gotOverflow, tt.want.overflow)
// 			}
// 		})
// 	}
// }
