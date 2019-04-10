#!/bin/sh

LUAJIT=../../libtool/bin/luajit-2.1.0-beta3

make -C../.. libluajit-clean
make -C../.. DASM_XFLAGS='-D MEASURE' libluajit
$LUAJIT -joff op.lua

make -C../.. libluajit-clean
make -C../.. libluajit
go run -tags measure main.go | jq -r .message | tr -d \"
