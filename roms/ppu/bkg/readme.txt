BKG Graphics Test
Copyright (C) 2003 Justin Patrick "Beneficii" Butler
E-mail: beneficii@yahoo.com
beneficii on the nesdev.parodius.com "Membled" message boards

Hi!  I see that you are interested in my test NES file.  Here is the
list of the files that should have been included in your bkg.zip package:

all.pal -- The game's palette (to be assembled with bkg.asm)
bkg.asm -- The commented source code
bkg.chr -- The screen tiles
bkg.map -- The maze
bkg.nes -- The finished game, which you could play with any emulator
bmp2map.exe -- The program that converts 256-color bitmaps into files
		compatible with my maze format
bmp2map.cpp -- The source code for the last
map.bmp -- The BMP version of bkg.map
spr.chr -- The sprite tiles
readme.txt -- This file

All files were authored by and are copyrighted (2003) to me.  You may
download this .zip package for your personal use, or you may put it on
your website provided that I am given full credit for the works.  If you
want to do anything profitable with them, then talk to me first, and we'll
see if we can reach an understanding.

Explanations of files:

all.pal:

The game's palette, with each byte being one member of the palette.  The
first $10 bytes are the background palette, while the second $10 bytes
are the sprite palette.  It is done like that, because that is the exact
order, in which it is loaded into register $2007 starting at VRAM $3f00.

bkg.asm:

This is the game's commented source code.  It is meant to be assembled
using Marat Fazhulyan (sp?)'s NESASM (NES Assembler).  I shouldn't have
to explain anything else.  The rest of the info is in the file.

bkg.chr:

This contains the tiles for the screen/background.

bkg.map:

This contains the data for the maze, loaded just as it is into VRAM $2000
for $400 bytes via $2007 in the program.

bkg.nes:

The mumbo jumbo.  The thing that is actually playable in the emulator.
Theoretically if you can upload it to a cartridge, then it would play
on a real Nintendo.  I think it could, because when I made it, I assumed
that the emulators would not be forgiven, and it's run perfectly on every
emulator on which I tried it.

bmp2map.exe:

This is the program that can convert 256-color BMPs to the MAP format I used
for the maze.  It is quite limited, however, accepting only three colors.
This was designed to be used with Bitmaps created with Microsoft Paint using
it's default palette.  The color white is used for walkable regions, the color
black is used for obstructions, and the color plain red (just below the darker
red about third column from the left) is used for the "finished" tile.  When you
make it, have it be 32x32 pixels, with the two lowest lines left blank for the
attribute table coding (for which the program doesn't really support except for
it being blank).  Try leaving the top and the third lowest lines blank too, as
they will do you no good.  Each dot on the bitmap is one tile, and the sprite's
starting point (as hard-coded) is at tile 2 left and 27 down.  On the program's
screen, put the input and output files in that respective order.

bmp2map.cpp:

This is BMP2MAP's source code.  It's rather bulky I agree, but I hope that perhaps
one day I can create a much more flexible and smoother program.

map.bmp:

This is an example of the maze before being converted into that .map format.  It
should give you a good idea of how to set one up.

spr.chr:

This is the CHR-ROM for the sprite.

readme.chr:

I think that's self-explanatory!  ^_^

Later!
N00bin'
=8-F