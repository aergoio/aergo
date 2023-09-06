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
    set(GFLAG -tags Debug)
endif()

add_custom_target(build ALL DEPENDS aergocli aergosvr aergoluac brick)

add_custom_target(aergocli GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X github.com/aergoio/aergo/v2/cmd/aergocli/cmd.githash=`git describe --tags`\" ./cmd/aergocli/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS libtool)

add_custom_target(aergosvr GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X main.githash=`git describe --tags` -X main.gitRevision=`git rev-parse --short HEAD` -X main.gitBranch=`git rev-parse --symbolic-full-name --abbrev-ref HEAD`\" ./cmd/aergosvr/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS libtool)

add_custom_target(polaris GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X main.githash=`git describe --tags`\" ./cmd/polaris/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR})

add_custom_target(colaris GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X github.com/aergoio/aergo/v2/cmd/colaris/cmd.githash=`git describe --tags`\" ./cmd/colaris/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR})

add_custom_target(aergoluac GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X main.githash=`git describe --tags`\" ./cmd/aergoluac/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS libtool)

add_custom_target(brick GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} ${GFLAG} -ldflags \"-X 'github.com/aergoio/aergo/v2/cmd/brick/context.GitHash=`git describe --tags`'
-X 'github.com/aergoio/aergo-lib/log.defaultConfStr=`cat ./cmd/brick/arglog.toml`'\"  ./cmd/brick/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS libtool)

add_custom_target(deps DEPENDS libtool)

add_custom_target(check GO111MODULE=on go test -timeout 600s ./...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS build)
add_custom_target(cover-check GO111MODULE=on go test -timeout 600s -coverprofile c.out ./...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS build)

add_custom_target(distclean GO111MODULE=on go clean ..
	COMMAND rm -rf ${BIN_DIR}/aergo* ${BIN_DIR}/brick
	DEPENDS libtool-clean)

add_custom_target(protoc
	COMMAND protoc -I/usr/local/include -I${PROTO_DIR} --go_out=plugins=grpc:${CMAKE_CURRENT_LIST_DIR}/types ${PROTO_DIR}/*.proto
	COMMAND GO111MODULE=on go build ../types/...)

add_custom_target(protoclean rm -f ../types/*.pb.go)


add_subdirectory(libtool)
