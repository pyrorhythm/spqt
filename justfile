cgo_cxxflags := "-std=c++17"

default: build

build PATH="./tmp/spqt":
    CGO_CXXFLAGS="{{ cgo_cxxflags }}" go build -ldflags '-s -w' -o {{ PATH }} ./cmd

open:
    mkdir -p "/tmp/spqt"
    CGO_CXXFLAGS="{{ cgo_cxxflags }}" go build -ldflags '-s -w' -o "/tmp/spqt/spqt_binary" ./cmd
    open "/tmp/spqt/spqt_binary"
    rm "/tmp/spqt/spqt_binary"
