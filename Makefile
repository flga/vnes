SRC=./cmd/vnes/*.go
BINARY=vnes

all: test build

test: 
	go test -v ./nes

run:
	go run $(SRC)

clean: clean-linux clean-windows
clean-linux: 
	rm -f $(BINARY) $(BINARY)-linux.zip
clean-windows: 
	rm -f $(BINARY).exe $(BINARY)-windows.zip

build: build-linux build-windows

build-linux: clean-linux
	CGO_ENABLED=1 \
	CC=gcc \
	GOOS=linux \
	GOARCH=amd64 \
	go build -o $(BINARY) $(SRC)

build-windows: clean-windows
	PKG_CONFIG_PATH="/usr/x86_64-w64-mingw32/lib/pkgconfig" \
	CGO_ENABLED="1" \
	CC="/usr/bin/x86_64-w64-mingw32-gcc" \
	GOOS=windows \
	GOARCH=amd64 \
	CGO_LDFLAGS="-Lportaudio_static" \
	go build -tags static -ldflags '-s -w -extldflags "-static"' -o $(BINARY).exe $(SRC)


dist-docker: clean
	docker build -t $(BINARY)-builder .
	docker create --name $(BINARY)-builder $(BINARY)-builder
	docker cp $(BINARY)-builder:/src/$(BINARY) $(BINARY)
	docker cp $(BINARY)-builder:/src/$(BINARY)-linux.zip $(BINARY)-linux.zip
	docker cp $(BINARY)-builder:/src/$(BINARY).exe $(BINARY).exe
	docker cp $(BINARY)-builder:/src/$(BINARY)-windows.zip $(BINARY)-windows.zip
	docker rm -fv $(BINARY)-builder

dist: dist-linux dist-windows

dist-linux: build-linux
	rm -f $(BINARY)-linux.zip
	zip -r $(BINARY)-linux.zip \
		./assets \
		./LICENSE \
		./README.md \
		./$(BINARY)

dist-windows: build-windows
	rm -f $(BINARY)-windows.zip
	zip -r $(BINARY)-windows.zip \
		./assets \
		./LICENSE \
		./README.md \
		./$(BINARY).exe
