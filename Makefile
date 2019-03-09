# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=vnes
BINARY_WIN=vnes.exe

all: test build
build: 
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/vnes/*.go
test: 
	$(GOTEST) -v ./nes
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_WIN)
run:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/vnes/*.go
	./$(BINARY_NAME)

build-windows:
	CGO_ENABLED="1" CC="/usr/bin/x86_64-w64-mingw32-gcc" GOOS="windows" \
	CGO_LDFLAGS="-lmingw32 -lSDL2" CGO_CFLAGS="-D_REENTRANT" \
	$(GOBUILD) -x -o $(BINARY_WIN) ./cmd/vnes/*.go