FROM golang:1.12-stretch

# get dev deps
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade -yy \
    && DEBIAN_FRONTEND=noninteractive apt-get install -yy \
        libsdl2-dev \
        mingw-w64 \
        portaudio19-dev \
        zip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* /build


# download portaudio src and compile it
RUN cd / \
    && git clone https://git.assembla.com/portaudio.git \
    && cd portaudio \
    && ./configure --host=x86_64-w64-mingw32  --target=x86_64-w64-mingw32 --prefix=/usr/x86_64-w64-mingw32/ --with-winapi=wmme,directx --enable-static --disable-shared \
    && make \
    && make install

# download sdl mingw and install it
RUN mkdir /sdl && cd /sdl \
    && wget https://www.libsdl.org/release/SDL2-devel-2.0.9-mingw.tar.gz \
    && tar zxf SDL2-devel-2.0.9-mingw.tar.gz \
    && cp -r SDL2-2.0.9/x86_64-w64-mingw32/* /usr/x86_64-w64-mingw32/


COPY go.mod go.sum /src/
WORKDIR /src/

RUN go mod download

COPY assets assets/
COPY cmd cmd/
COPY nes nes/
COPY Makefile LICENSE README.md ./

RUN make dist
