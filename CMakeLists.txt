#
# @file    CMakeLists.txt
# @copyright defined in aergo/LICENSE.txt
#

cmake_minimum_required(VERSION 3.5)

project(aergo NONE)

set(BIN_DIR ${CMAKE_CURRENT_LIST_DIR}/bin)
set(PROTO_DIR ${CMAKE_CURRENT_LIST_DIR}/aergo-protobuf/proto)

# Set CGO environment variables for RocksDB/grocksdb builds
set(CGO_CFLAGS "-I${CMAKE_CURRENT_LIST_DIR}/libtool/include")
set(CGO_LDFLAGS "-L${CMAKE_CURRENT_LIST_DIR}/libtool/lib -lrocksdb -lstdc++ -lm -lz -lsnappy")

if(CMAKE_BUILD_TYPE MATCHES "Debug")
    set(GCFLAGS "-gcflags=\"-N" "-l\"")
    set(GFLAG -tags Debug)
endif()

# CMake command line option to accept GITHASH
option(GIT_TAG "Git hash value")
option(GIT_BRANCH "Git hash value")
option(GIT_REVISION "Git hash value")

# Determine the git hash value
if(GIT_TAG)
    set(GIT_TAG_CMD ${GIT_TAG})
else()
    execute_process(
        COMMAND git describe --tags
        WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
        OUTPUT_VARIABLE GIT_TAG_CMD
        OUTPUT_STRIP_TRAILING_WHITESPACE
    )
endif()
if(GIT_BRANCH)
    set(GIT_BRANCH_CMD ${GIT_BRANCH})
else()
    execute_process(
        COMMAND git rev-parse --symbolic-full-name --abbrev-ref HEAD
        WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
        OUTPUT_VARIABLE GIT_BRANCH_CMD
        OUTPUT_STRIP_TRAILING_WHITESPACE
    )
endif()
if(GIT_REVISION)
    set(GIT_REVISION_CMD ${GIT_REVISION})
else()
    execute_process(
        COMMAND git rev-parse --short HEAD
        WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
        OUTPUT_VARIABLE GIT_REVISION_CMD
        OUTPUT_STRIP_TRAILING_WHITESPACE
    )
endif()

# Pass the determined GIT_* value to makefile
configure_file(
    "GITINFO.txt.in"
    "${CMAKE_CURRENT_BINARY_DIR}/GITINFO.txt"
    @ONLY
)

add_custom_target(build ALL DEPENDS aergocli aergosvr aergoluac brick)

add_custom_target(aergocli CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -tags grocksdb_clean_link -ldflags \"-X github.com/aergoio/aergo/v2/cmd/aergocli/cmd.githash=${GIT_TAG_CMD}\" ./cmd/aergocli/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS libtool)

add_custom_target(aergosvr CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -tags grocksdb_clean_link -ldflags \"-X main.githash=${GIT_TAG_CMD} -X main.gitRevision=${GIT_REVISION_CMD} -X main.gitBranch=${GIT_BRANCH_CMD}\" ./cmd/aergosvr/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS libtool)

add_custom_target(polaris GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X main.githash=${GIT_TAG_CMD}\" ./cmd/polaris/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR})

add_custom_target(colaris GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X github.com/aergoio/aergo/v2/cmd/colaris/cmd.githash=${GIT_TAG_CMD}\" ./cmd/colaris/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR})

add_custom_target(aergoluac CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -tags grocksdb_clean_link -ldflags \"-X main.githash=${GIT_TAG_CMD}\" ./cmd/aergoluac/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS libtool)

add_custom_target(brick CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} ${GFLAG} -tags grocksdb_clean_link -ldflags \"-X 'github.com/aergoio/aergo/v2/cmd/brick/context.GitHash=${GIT_TAG_CMD}'
-X 'github.com/aergoio/aergo-lib/log.defaultConfStr=`cat ./cmd/brick/arglog.toml`'\"  ./cmd/brick/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS libtool)

add_custom_target(mpdumpdiag GO111MODULE=on GOBIN=${BIN_DIR} go install ${GCFLAGS} -ldflags \"-X main.githash=${GIT_TAG_CMD}\" ./tools/mpdumpdiag/...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR})

add_custom_target(deps DEPENDS libtool)

add_custom_target(check CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} GO111MODULE=on go test -timeout 600s ./...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS build)
add_custom_target(short-check CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} GO111MODULE=on go test -short -timeout 999s ./...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS build)
add_custom_target(cover-check CGO_CFLAGS=${CGO_CFLAGS} CGO_LDFLAGS=${CGO_LDFLAGS} GO111MODULE=on go test -timeout 600s -coverprofile build/c.out ./...
    WORKING_DIRECTORY ${CMAKE_CURRENT_LIST_DIR}
    DEPENDS build)

add_custom_target(distclean GO111MODULE=on go clean ..
	COMMAND rm -rf ${BIN_DIR}/aergo* ${BIN_DIR}/brick
	DEPENDS libtool-clean)

add_custom_target(protoc
	COMMAND protoc -I/usr/local/include -I${PROTO_DIR} --go_out=${CMAKE_CURRENT_LIST_DIR}/types --go_opt=paths=source_relative --go-grpc_out=${CMAKE_CURRENT_LIST_DIR}/types --go-grpc_opt=paths=source_relative ${PROTO_DIR}/*.proto
	COMMAND GO111MODULE=on go build ../types/...)

add_custom_target(protoclean rm -f ../types/*.pb.go)


add_subdirectory(libtool)
