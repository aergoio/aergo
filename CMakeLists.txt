#
# @file    CMakeLists.txt
# @copyright defined in aergo/LICENSE.txt
#

cmake_minimum_required(VERSION 3.0)

project(aergo NONE)

set(BIN_DIR ${CMAKE_CURRENT_LIST_DIR}/bin)
set(PROTO_DIR ${CMAKE_CURRENT_LIST_DIR}/aergo-protobuf/proto)

add_custom_target(build ALL DEPENDS vendor libtool aergocli aergosvr aergoluac)

add_custom_target(aergocli GOBIN=${BIN_DIR} go install -ldflags \"-X github.com/aergoio/aergo/cmd/aergocli/cmd.githash=`git rev-parse HEAD`\" ./cmd/aergocli/...
  DEPENDS vendor libtool)

add_custom_target(aergosvr GOBIN=${BIN_DIR} go install ./cmd/aergosvr/...
  DEPENDS vendor libtool)

add_custom_target(aergoluac GOBIN=${BIN_DIR} go install ./cmd/aergoluac/...
  DEPENDS vendor libtool)

add_custom_command(OUTPUT vendor
    COMMAND glide install
    DEPENDS glide.yaml glide.lock)

add_custom_target(deps DEPENDS vendor libtool)

add_custom_target(check go test -timeout 60s ./... DEPENDS build)
add_custom_target(cover-check go test -coverprofile c.out ./... DEPENDS build)

add_custom_target(distclean go clean
	COMMAND rm -rf ${BIN_DIR}/aergo*
	DEPENDS libtool-clean)

add_custom_target(realclean rm -rf `find . \\\( -name CMakeCache.txt -o -name CMakeFiles -o -name cmake_install.cmake \\\)`
	DEPENDS distclean)

add_custom_target(protoc
	COMMAND protoc -I/usr/local/include -I${PROTO_DIR} --go_out=plugins=grpc:$ENV{GOPATH}/src ${PROTO_DIR}/*.proto
	COMMAND go build ./types/...)

add_custom_target(protoclean rm -f types/*.pb.go)

add_subdirectory(libtool)
