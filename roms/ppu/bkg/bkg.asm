; A graphics test (first)
; This to be assembled using the NES Assembler.
; Copyright (C) 2003 Justin Patrick "Beneficii" Butler

	.inesprg 1            ; Only 16kb PRG-ROM needed.
	.ineschr 1            ; Only 8kb CHR-ROM needed.
	.inesmir 0            ; No (though set to horizontal) mirroring needed.
	.inesmap 0            ; No mappers needed.

	.zp                   ; Declaring zero page variables
address_lo = $00
address_hi = $01
a.button = $02
b.button = $03
select.button = $04
start.button = $05
up.button = $06
down.button = $07
left.button = $08
right.button = $09
y.move = $0a
x.move = $0b
y.pos = $0c
x.pos = $0d
temp = $0e
curtile = $0f
nmiflag = $10
read.lo = $11
read.hi = $12
move.type.1 = $13
move.type.2 = $14
sound.flag = $15
check.sound.flag = $16

	.bss                  ; Declaring page seven sprite variables
sprite.y = $0700
sprite.tile = $0701
sprite.attributes = $0702
sprite.x = $0703

	.code                 ; Stajmp finish_nming code with .org at $8000.  It doesn't really matter which I pick, because
	.org $8000            ; the code is copied into the PRG-ROM twice.  (That is, $8000 or $C000.)

wait_vblank:                  ; Wait for VBlank to be turned on.
	lda $2002
	bpl wait_vblank
	rts

main:                         ; Begin the RESET routine.
	sei                   ; The usual, setting the interrupt disable flag and
	cld                   ; clearing the decimal mode flag.

	ldx #$ff              ; Set x to $FF.
	txs                   ; Transfer $FF to the stack, thereby resetting it.

	jsr wait_vblank       ; Wait for VBlank before turning off the screen.
	lda #$00              ; Turning
	sta $2000             ; the screen
	sta $2001             ; off

	inx                   ; Getting ready to clear the zero page memory.

clearzp:
	sta <$00, x           ; No need to reload the accumulator, because I know it's zero.
	inx                   ; Increment.  Needs to go through all 256 bytes of the zero page memory.
	bne clearzp           ; Go back if not finished (if x is not yet back to zero).

	ldx #$07              ; Getting ready to clear the seven other pages of WRAM,
	ldy #$01              ; to ensure that they are clear (the real NES probably
	sty <address_hi       ; has all sorts of strange, loaded values at the start).
	dey                   ; Am going to use indirect addressing to accomplish so.
	sty <address_lo       ; I will begin at $100 and end at $7FF

clearwram:
	sta [address_lo], y   ; There is no need to set the accumulator to zero, because it is already.
	iny                   ; Increment y.
	bne clearwram         ; If not yet zero again, then go back.
	inc address_hi        ; Increment the address indicating that it is time to go the next page.
	dex                   ; Count down to zero.
	bne clearwram         ; Go back if not done.

	ldy #$20              ; Getting ready to load the palette.
	ldx #$3f              ; Loading the variables.
	stx $2006             ; Reading to $2006, high byte first.
	ldx #$00              ; Putting
	stx $2006             ; low byte in.

loadpal:
	lda paldata, x        ; Loading the palette data, which is located
	sta $2007             ; elsewhere in the PRG-ROM byte by byte.
	inx                   ; Read further into the data.
	dey                   ; Decrement y until it is done.
	bne loadpal           ; There are $20 bytes to load, so y starts at $20.

	lda $2002             ; Reset $2006 (since it hasn't been yet).
	stx $2006             ; x I know is 20 from the last loop, and
	sty $2006             ; y I know is 00 from the last loop, so
                              ; I can set $2006 to VRAM $2000

	ldx #$08              ; Getting ready to clear the first two name tables of the PPU
	tya                   ; (the other two are mirrors, so there is no need).

clearppu:
	sta $2007             ; Storing 0 in there to make sure it is cleared.
	iny                   ; Go to next byte.
	bne clearppu
	dex                   ; Get to next page.
	bne clearppu

	jsr wait_vblank       ; Some emulators, such as NESticle, have problems if I don't
                              ; wait for VBlank after clearing the PPU, so I might as well.

	ldx #$20              ; $2002 has already been read to reset $2006 (wait_vblank),
	stx $2006             ; so there is no need to do so again.
	sty $2006
	ldx #$04
	lda #low(mapdata)     ; Getting ready to load the maze data.
	sta <address_lo       ; Prepping for use of indirect addressing.
	lda #high(mapdata)
	sta <address_hi

loadmap:
	lda [address_lo], y   ; Loading name and attribute tables #1.
	sta $2007             ; No need for name and attribute tables #2, as I will not
	iny                   ; be scrolling in this ROM.
	bne loadmap
	inc <address_hi
	dex
	bne loadmap

	lda #212              ; Putting the sprite together
	sta <y.pos            ; to be placed on the screen.
	sta sprite.y          ; Done 216 - 4, so that the sprite's feet
	lda #016              ; are in the middle of the tiles
	sta <x.pos            ; rather than the edge.
	sta sprite.x
	ldx #001
	stx sprite.tile
	dex
	stx sprite.attributes

	stx $2003              ; Reset $4014
	lda #$07              ; Loading the sprites
	sta $4014             ; "automagically."

	lda #%10001000        ; Turning the screen back on so you can actually see something.
	sta $2000
	lda #%00011110
	sta $2001

eternity:                     ; It's a good idea now to throw the RESET routine into an eternal loop,
	jmp eternity          ; otherwise it'll go on like crazy.  ~_^

store_y:                      ; In case the sprite is moving up or down.
	lda #low(sprite.y)    ; This ensures that the code below knows that
	sta <read.lo          ; and writes to the right sprite variable.
	lda #high(sprite.y)
	sta <read.hi
	lda #$00
	sta <y.move
	sta <x.move
	jmp animate_movement

store_x:                      ; Same as above, except for moving right or left.
	lda #low(sprite.x)
	sta <read.lo
	lda #high(sprite.x)
	sta <read.hi
	lda #$00
	sta <y.move
	sta <x.move
	jmp animate_movement

update_sprites:               ; Check in which direction the sprite is moving, if any.
	lda <y.move
	bne store_y
	lda <x.move
	bne store_x

	sta <nmiflag          ; If the sprite isn't moving, then
	sta <read.lo          ; just go ahead and clear the useless
	sta <read.hi          ; flags before they hurt someone.  ~_^
	sta <move.type.1
	sta <move.type.2

	ldx <sound.flag       ; If the sound flag is activated (see below), then
	bne allow_beep2       ; prevent it from being deactivated.
	jmp do_still_move

allow_beep2:                  ; You see, in the first NMI (this is the second), if the sprite
	inx                   ; goes over the "finish" tile, then flag is activated, which signals
	stx <sound.flag       ; this NMI to give the first the go-ahead to play the tone
	stx <check.sound.flag ; for going over the "finish" tile come the first's return.

do_still_move:                ; In case the sprite changed direction and is facing
	lda #$00              ; an obstruction.
	sta $2003
	lda #$07
	sta $4014

	jmp finish_nmi        ; Finish this NMI.

animate_movement:             ; In case the sprite is moving its position.
	lda #$00              ; Ensuring these are cleared.
	tax
	tay

	lda sprite.tile       ; Load the appropriate tile for the sprite.
	clc                   ; The initial values are set in the first NMI.
	adc #$02              ; Clear the carry flag and add two (left foot out first for the sprite).
	sta sprite.tile
	lda [read.lo], y      ; Add or subtract two (set in the first NMI)
	clc                   ; to the sprite.pos value.
	adc <move.type.1      ; Of course this is done by indirect addressing,
	sta [read.lo], y      ; so I don't have to keep checking to see, if I'm changing
	inc <nmiflag          ; the x or y positions of the sprite.
	sty $2003
	lda #$07              ; Store the updated sprite
	sta $4014             ; into the SPR-RAM.

	jmp finish_nmi        ; Terminate this NMI without further ado.

third_frame:                  ; Same as last, except that the sprite.pos
	lda #$00              ; is incremented or decremented by four.
	tax
	tay

	ldx sprite.tile       ; Load the sprite.tile into x (so I can easily decrement).
	dex                   ; This will cause the sprite to step out with his right foot.
	stx sprite.tile
	lda [read.lo], y
	clc
	adc <move.type.2
	sta [read.lo], y
	inc <nmiflag
	sty $2003
	lda #$07
	sta $4014

	jmp finish_nmi

fourth_frame:                 ; Pretty much the same 
	lda #$00
	tax
	tay

	ldx sprite.tile       ; Add or subtract 2 to the sprite.pos this time, in order to
	dex                   ; make the complete 8 pixels.
	stx sprite.tile       ; Also decrement the sprite's tile to make it stand still
	lda [read.lo], y      ; when its movement is finished.
	clc
	adc <move.type.1
	sta [read.lo], y

	sty <nmiflag          ; Clear these, because we are done with them.
	sty <read.lo          ; Reset the NMI flag to zero so as to go to the
	sty <read.hi          ; first NMI next NMI.
	sty <move.type.1
	sty <move.type.2

	ldx <sound.flag       ; If the sound flag is activated, then do the "allow_beep" thing.
	bne allow_beep
	jmp final_write

allow_beep:                   ; I'll explain the check sound flag below.
	inx
	stx <sound.flag
	stx <check.sound.flag

final_write:                  ; Perform the final sprite write.  ^_^
	sty $2003
	lda #$07
	sta $4014

	jmp finish_nmi        ; Terminate, O final NMI!

send_me:                      ; If the NMI flag is not zero, then go here,
	lda <nmiflag          ; because it is not the first NMI.
	cmp #$02
	beq third_frame
	lda <nmiflag
	cmp #$03
	beq fourth_frame

	jmp update_sprites

nmi:                          ; Start the NMI code.
	pha                   ; Store the accumulator and the x and y registers in the stack to
	txa                   ; prevent them from being destroyed by the code here in case that this
	pha                   ; NMI interrupted RESET code.
	tya                   ; (Advised to do that by someone on nesdev.parodius.com
	pha                   ; "Membled" messageboards
	lda $2002             ; Set $2006 to $0000 each NMI to prevent the background display
	lda #$00              ; from being messed up.  Thanks (^_^) to the ever-helpful Memblers
	sta $2006             ; for his advice on that one.  Now the sprite's feet are in the
	sta $2006             ; middle of the tile in every emulator (and probably the real thing)!
	lda <nmiflag          ; If the NMI flag is anything but zero, then go to the handler.
	bne send_me           ; This was the result of the appendaging of the animation and sound codes.
	ldx #$01              ; Strobe the game controller.
	stx $4016
	dex
	stx $4016             ; Reset it.
	ldy #$08              ; Prepare to read it.

read_buttons:               ; Reading it.
	lda $4016             ; a.button is the first in the variable array,
	and #$01              ; so we just read it from there.
	sta <a.button, x      ; Fix it so we just have bit 0 of it.
	inx
	dey
	bne read_buttons

	lda #$00              ; Have accumulator=zero, so first writes are zero.
	ldx <up.button        ; Test if buttons have been pressed.
	bne up
	ldx <down.button
	bne down
	ldx <right.button
	bne right
	ldx <left.button
	bne left
	sta <y.move           ; If they have not.
	sta <x.move
	jmp continue

up:                         ; If the player pressed up.
	sta <x.move           ; Clear x.move.
	lda #$f8              ; Signal to decrement (not increment) 8
	sta <y.move           ; by taking advantage of the sign bit (number 7).
	lda #$fe              ; For the second and fourth NMI's,
	sta <move.type.1      ; it is to move backward (not forward two).
	lda #$fc              ; For the third NMI,
	sta <move.type.2      ; it is to move backward (not forward two).
	lda #$40              ; Since I'm moving up, it will not look like
	sta sprite.attributes ; the first foot is left.  So I flip it (bit attributes.6).
	lda #$01              ; Always do still tile first.
	sta sprite.tile
	jmp continue

down:                       ; Same as above, except that everything is opposite.
	sta <x.move
	sta sprite.attributes
	lda #$08
	sta <y.move
	lda #$02
	sta <move.type.1
	lda #$04
	sta <move.type.2
	lda #$01
	sta sprite.tile
	jmp continue

left:                       ; Same as above, except that things are suited for left-right.
	sta <y.move           ; The unflipped sprite in this faces left.
	sta sprite.attributes
	lda #$f8
	sta <x.move
	lda #$fe
	sta <move.type.1
	lda #$fc
	sta <move.type.2
	lda #$04
	sta sprite.tile
	jmp continue

right:                        ; You get my drift.
	sta <y.move
	lda #$08
	sta <x.move
	lda #$02
	sta <move.type.1
	lda #$04
	sta <move.type.2
	lda #$40
	sta sprite.attributes
	lda #$04
	sta sprite.tile

continue:
	lda #$00              ; This was to put in, to ensure that sound was disabled.
	sta $4015             ; Didn't know much about it, so I took that precaution.
	sta <address_hi       ; This was not designed to be a sound player, but I wanted
	lda <x.pos            ; to tinker with playing at least a tone when you get to the
	clc                   ; "finish" tile.
	adc <x.move           ; This batch of code tests whether there will be an obstruction
	sta <address_lo       ; in the characters path if he moves in whatever direction he
	lda <y.pos            ; was gonna.
	clc                   ; Ensure that address_hi is cleared, and store the x position
	adc <y.move           ; to address_lo.
	clc                   ; Also we calculate the proposed position to get it ready.
	adc #004              ; Add four to the y.move to get it in line with its true "spot."
	sta <temp             ; That was four off so the characters feet could be in the middle of the tile.
	ldx #$03              ; Store in temp, because we have a lot more work to do with it.
	ldy #$02              ; Prep vars for next two subroutines.

set_address_x:                ; Divide x by 8 (by shifting right 3).
	lsr <address_lo       ; This ensures that the x tile position (0-31) is set up properly
	dex                   ; for obstruction checking.
	bne set_address_x

set_address_y:                ; Get the y part ready.  It is spread over more than 2 bytes
	asl <temp             ; so I can OR it and get the true position in the mapdata,
	rol <address_hi       ; for which the sprite's movement is proposed.
	dey
	bne set_address_y

	lda <temp             ; Store in upper 3 bits of address_lo.
	ora <address_lo
	sta <address_lo

	clc                   ; Add the mapdata position to its true address,
	lda #low(mapdata)     ; so as to read from the spot the character is headed.
	adc <address_lo
	sta <address_lo
	lda #high(mapdata)
	adc <address_hi
	sta <address_hi

	lda [address_lo], y   ; With indirect addressing, I needed to put that useless y,
	sta <curtile          ; so think nothing of it.
	tax                   ; The cujmp finish_nmile part was part of testing, and I decided not to remove it.
	dex
	beq delete_move       ; "Delete" the move if the tile was an obstruction (ID #1)
	dex
	beq chk_flag          ; Do the sound if "finish" tile (ID #2)
	sty <check.sound.flag
	jmp finish

chk_flag:                   ; I had to do the check flag, so as to see if the second (or fourth)
	lda <check.sound.flag ; NMI "approved" the sound being played.
	beq snd_flag          ; This was in case the sprite stood still after moving
	jmp finish            ; into the "finish" tile, so the player would not have to move it
                            ; back out of the tile to actually play the sound.
snd_flag                    ; This turns on the sound flag.
	ldx #$01              ; This flag goes onto the second (or fourth) NMI, which then
	stx <sound.flag       ; "approves" it by sending a two back to the first NMI.
	jmp finish

delete_move:                ; If an obstruction was in the sprite's way.
	sty <y.move           ; It sets these "flags" to zero to indicate no movement.
	sty <x.move

finish:                     ; Finishing this.
	lda <y.pos            ; Might as well go ahead and store the positions
	clc                   ; in the "unofficial" memory locations.
	adc <y.move
	sta <y.pos
	lda <x.pos
	clc
	adc <x.move
	sta <x.pos

	inc <nmiflag           ; Signal second NMI coming up.

	ldx <sound.flag        ; Branch to beep if the second (or fourth) NMI approved
	cpx #$02               ; the tone being played.
	beq beep

	jmp finish_nmi         ; Terminate this first NMI.

beep:                          ; Do the sound channel crap finally!
	lda #$01                 ; Make that beep!
	sta $4015
	lda #$0f
	sta $4000
	lda #$00
	sta $4001
	lda #$10
	sta $4002
	lda #$12
	sta $4003
	lda #$00
	sta <sound.flag

finish_nmi:                    ; Terminate these NMI's.
	pla
	tay
	pla
	tax
	pla

	rti

brk:                            ; Now beginning the IRQ/BRK routine.
	rti                       ; Terminate the IRQ/BRK routine if it is ever needed.

paldata: .incbin "all.pal"      ; Include palette in the PRG-ROM.
mapdata: .incbin "bkg.map"      ; Including the maze in the PRG-ROM.

	.bank 1                 ; Do the vector tables.
	.org $FFFA

	.dw nmi
	.dw main
	.dw brk

	.bank 2                 ; Do the CHR-ROM.
 	.incbin "bkg.chr"       ; The background tiles for the "left half."
	.incbin "spr.chr"       ; The sprite tiles for the "right half."

; EOF!

