#!/bin/bash
go generate

echo "compile each binary"

if [ -z "$TRAVIS_BUILD_DIR" ]
then
	TRAVIS_BUILD_DIR=$PWD
fi

echo $TRAVIS_BUILD_DIR

for GOOS in darwin linux windows; do
  for GOARCH in 386 amd64; do
    echo "Building $GOOS-$GOARCH"
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    export CGO_ENABLED=0
    go build -o $TRAVIS_BUILD_DIR/bin/goHashStore-$GOOS-$GOARCH -ldflags "-X main.Version=`cat VERSION`"
  done
done
mv bin/goHashStore-darwin-386 bin/goHashStore-darwin-386.bin
mv bin/goHashStore-darwin-amd64 bin/goHashStore-darwin-amd64.bin
mv bin/goHashStore-linux-386 bin/goHashStore-linux-386.bin
mv bin/goHashStore-linux-amd64 bin/goHashStore-linux-amd64.bin
mv bin/goHashStore-windows-386 bin/goHashStore-windows-386.exe
mv bin/goHashStore-windows-amd64 bin/goHashStore-windows-amd64.exe