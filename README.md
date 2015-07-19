# A simple shiftindicator app for iRacing

This is a simple cli program that plays a sound when you need to upshift in
iRacing. I created it to work on Linux but it also could be made to work on Mac
and Windows.

It consists of three main files:

- the binary `shiftindicator`
- an audio file
- the `shiftindicator.yml` configuration file

## Installation

``` sh
sudo apt-get install libopenal-dev build-essentials gcc-mingw-w64-i686

go get -d -u github.com/leonb/wineshm-go
cd `go env GOPATH`/src/github.com/leonb/wineshm-go/
make

go get -d -u github.com/leonb/irsdk-go
cd `go env GOPATH`/src/github.com/leonb/irsdk-go/
make

go get -u github.com/codegangsta/cli
go get -u golang.org/x/mobile/exp/audio
go get -u github.com/leonb/shiftindicator
```

## Running the program

If you installed the program with the help of `go get` you could just run
`shiftindicator` in your terminal (and your `$PATH` is setup right). If you
downloaded the binary package, just cd to the directory where you extracted the
binary and run `./shiftindicator`.

## Customisations

The configuratio file `shiftindicator.yml` is search for in the following
places:

- `$XDG_CONFIG_HOME` (usually `~/.config/`)
- the current directory (`.`)
- the directory where the binary is placed

The following things can be customised:

- volume of soundfile played (0 - 1.0)
- the soundfile played (has to be a `.wav` file)
- the minimum time between beeps (when your rpm's are hovering over the optimal
  shiftpoint)
- the shiftpoints per car

## Information on shift points

- http://members.iracing.com/jforum/posts/downloadAttach/2214716.page
- http://members.iracing.com/jforum/posts/list/2022116.page
- http://www.iracing.com/iracingnews/iracing-news/the-power-of-skip

## Inspiration

- http://nspace.hu/soundshift/
