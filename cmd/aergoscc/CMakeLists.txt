#
# @file    CMakeLists.txt
# @copyright defined in aergo/LICENSE.txt
#

find_package(LLVM REQUIRED CONFIG)

add_custom_target(aergoscc GOBIN=${BIN_DIR} CGO_LDFLAGS="-L${LLVM_LIBRARY_DIRS}" go install -a ${GCFLAGS} -ldflags \"-X github.com/aergoio/aergo/cmd/aergoscc/cmd.githash=`git describe --abbrev=0 --tags`\" ./cmd/aergoscc/...
    WORKING_DIRECTORY ${PROJECT_SOURCE_DIR}
    DEPENDS vendor binaryen asclvm ascl)
