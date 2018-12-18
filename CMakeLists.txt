#
# @file    CMakeLists.txt
# @copyright defined in aergo/LICENSE.txt
#

cmake_minimum_required(VERSION 3.0)

project(aergo NONE)

set(BIN_DIR ${CMAKE_CURRENT_LIST_DIR}/bin)
set(PROTO_DIR ${CMAKE_CURRENT_LIST_DIR}/aergo-protobuf/proto)

if(CMAKE_BUILD_TYPE MATCHES "Debug")
    set(GCFLAGS "-gcflags=\"-N" "-l\"")
endif()

add_custom_target(build ALL DEPENDS aergocli aergosvr aergoluac brick)

add_custom_target(aergocli GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X github.com/aergoio/aergo/cmd/aergocli/cmd.githash=`git rev-parse HEAD`\" ./cmd/aergocli/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS vendor libtool)

add_custom_target(aergosvr GOBIN=${BIN_DIR} go install ${GCFLAGS} ./cmd/aergosvr/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS vendor libtool)

add_custom_target(aergomap GOBIN=${BIN_DIR} go install ${GCFLAGS} ./cmd/aergomap/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS vendor libtool)

add_custom_target(aergoluac GOBIN=${BIN_DIR} go install ${GCFLAGS} ./cmd/aergoluac/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS vendor libtool)

add_custom_target(brick GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X 'github.com/aergoio/aergo/cmd/brick/context.GitHash=`git describe --tags`' 
-X 'github.com/aergoio/aergo/vendor/github.com/aergoio/aergo-lib/log.defaultConfStr=`cat ./cmd/brick/arglog.toml`'\"  ./cmd/brick/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS vendor libtool)

set(VENDOR ${CMAKE_CURRENT_LIST_DIR}/vendor)
add_custom_command(OUTPUT ${VENDOR}
    COMMAND glide install
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS glide.yaml glide.lock)
add_custom_target(vendor DEPENDS ${VENDOR})

add_custom_target(deps DEPENDS vendor libtool)

add_custom_target(check go test -timeout 60s ./...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS build)
add_custom_target(cover-check go test -coverprofile c.out ./...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS build)

add_custom_target(distclean go clean ..
	COMMAND rm -rf ${BIN_DIR}/aergo* ${BIN_DIR}/brick
	DEPENDS libtool-clean)

add_custom_target(protoc
	COMMAND protoc -I/usr/local/include -I${PROTO_DIR} --go_out=plugins=grpc:$ENV{GOPATH}/src ${PROTO_DIR}/*.proto
	COMMAND go build ../types/...)

add_custom_target(protoclean rm -f ../types/*.pb.go)

add_subdirectory(libtool)
