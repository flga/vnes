#include <iostream>
#include <stdio.h>
#include <algorithm>
#include <malloc.h>
#include <io.h>
#include <conio.h>

using namespace std;

int hiread(unsigned char *contents, int offset, int length) {
	for(int i = 0, k = 0; i < length; i++)
		k += unsigned int(contents[offset+i]) << (8 * i);
	return k;
}

int main(void) {
	unsigned char filename[280], outputfile[280], *filecontents, *outputcontents;
	FILE *file, *output;
	int filehandle, outputhandle, filelen, outputlen, bytebuffer,
		paintoffset, width, height, bitcount, pixelslen, j, wastedpixelsperline;

	cout << "BMP2MAP\n";
	cout << "Copyright (C) 2003 Justin Patrick Butler\n";
	cout << "256-color Bitmap file:\n";
	cin >> filename;
	cout << "output Map file:\n";
	cin >> outputfile;
	
	file = fopen((const char *) filename, "r");
	filehandle = fileno(file);
	filelen = (int) filelength(filehandle);
	filecontents = (unsigned char *) malloc(filelen);
	read(filehandle, (void *) filecontents, (unsigned int) filelen);
	fclose(file);

	paintoffset = hiread(filecontents, 0xa, 4);
	width = hiread(filecontents, 0x12, 4);
	height = hiread(filecontents, 0x16, 4);
	bitcount = hiread(filecontents, 0x1c, 2);
	bytebuffer = hiread(filecontents, 0x22, 4);
	if(bitcount != 8) {cout << "needs to be 256 colors!\n"; free((void *) filecontents); return 0;}
	pixelslen = bytebuffer;
	outputlen = height * width;
	wastedpixelsperline = (bytebuffer - (width * height)) / height;
	outputcontents = (unsigned char *) malloc(outputlen);
	j = 0;
	for(int h = 1, offset = (pixelslen - width + paintoffset - wastedpixelsperline); h <= height;
	h++, offset -= (width + wastedpixelsperline)){
			for(int w = 0; w < width; w++) {
				outputcontents[j] = (unsigned char) filecontents[offset+w]; j++;
			}
			w = 0;
	}
	free((void *) filecontents);
	for(int i = 0; i < outputlen; i++) {
		if(outputcontents[i]==(unsigned char) 0xff) outputcontents[i] = (unsigned char) 0x00;
		else if(outputcontents[i]==(unsigned char) 0x00) outputcontents[i] = (unsigned char) 0x01;
		else outputcontents[i] = (unsigned char) 0x02;}
	output = fopen((const char *) outputfile, "w");
	outputhandle = fileno(output);
	write(outputhandle, (const void *) outputcontents, outputlen);
	fclose(output);
	free((void *) outputcontents);
	cout << "done!\n";
	cout << "Press any key to close this window\n";
	while (!getch()) ;
	return 1;
}
