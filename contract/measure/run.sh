#!/bin/sh

LUAJIT=../../libtool/bin/luajit-2.1.0-beta3

make -C../.. libluajit-clean
make -C../.. DASM_XFLAGS='-D MEASURE' libluajit
echo ""
echo "---------------------------------------------"
echo "Measure execution of op.lua"
$LUAJIT -joff op.lua
echo "---------------------------------------------"
echo ""

make -C../.. libluajit-clean
make -C../.. libluajit
echo "---------------------------------------------"
go run -tags measure main.go 2>&1 | jq -r .message | tr -d \"
