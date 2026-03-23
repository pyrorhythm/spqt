cgo_cxxflags := "-std=c++17"

default: build

build PATH="./tmp/spqt":
    CGO_CXXFLAGS="{{cgo_cxxflags}}" go build -ldflags '-s -w' -o {{PATH}} ./cmd
