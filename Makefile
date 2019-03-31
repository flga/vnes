SRC=./cmd/vnes/*.go
BINARY=vnes

all: test build

test: 
	go test -v ./nes

run:
	go run $(SRC)

clean: clean-linux clean-windows
clean-linux: 
	rm -f $(BINARY)
clean-windows: 
	rm -f $(BINARY).exe

generate:
	go generate $(SRC)

build: build-linux build-windows

build-linux: clean-linux generate
	CGO_ENABLED=1 \
	CC=gcc \
	GOOS=linux \
	GOARCH=amd64 \
	go build -o $(BINARY) $(SRC)

build-windows: clean-windows generate
	PKG_CONFIG_PATH="/usr/x86_64-w64-mingw32/lib/pkgconfig" \
	CGO_ENABLED="1" \
	CC="/usr/bin/x86_64-w64-mingw32-gcc" \
	GOOS=windows \
	GOARCH=amd64 \
	CGO_LDFLAGS="-Lportaudio_static" \
	go build -tags static -ldflags '-s -w -extldflags "-static"' -o $(BINARY).exe $(SRC)


docker-build:
	docker build -t $(BINARY)-builder .
	docker run --name $(BINARY)-builder $(BINARY)-builder make build
	docker cp $(BINARY)-builder:/src/$(BINARY) $(BINARY)
	docker cp $(BINARY)-builder:/src/$(BINARY).exe $(BINARY).exe
	docker rm -fv $(BINARY)-builder

docker-build-linux: clean-linux
	docker build -t $(BINARY)-builder .
	docker run --name $(BINARY)-builder $(BINARY)-builder make build-linux
	docker cp $(BINARY)-builder:/src/$(BINARY) $(BINARY)
	docker rm -fv $(BINARY)-builder

docker-build-windows: clean-windows
	docker build -t $(BINARY)-builder .
	docker run --name $(BINARY)-builder $(BINARY)-builder make build-windows
	docker cp $(BINARY)-builder:/src/$(BINARY).exe $(BINARY).exe
	docker rm -fv $(BINARY)-builder