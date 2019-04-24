#
# @file    CMakeLists.txt
# @copyright defined in aergo/LICENSE.txt
#

cmake_minimum_required(VERSION 3.0)

project(libascl C CXX)

find_package(BISON QUIET)
find_package(FLEX QUIET)

if(BISON_FOUND AND FLEX_FOUND)
    bison_target(PARSER grammar.y ${CMAKE_CURRENT_SOURCE_DIR}/grammar.tab.c COMPILE_FLAGS -d)
    flex_target(SCANNER scanner.l ${CMAKE_CURRENT_SOURCE_DIR}/scanner.yy.c)

    add_flex_bison_dependency(SCANNER PARSER)
    set_directory_properties(PROPERTIES ADDITIONAL_MAKE_CLEAN_FILES "grammar.output")
endif()

set(ascl_sources
    compile.c
    prep.c
    parse.c
    ast.c
    ast_id.c
    ast_blk.c
    ast_stmt.c
    ast_exp.c
    check.c
    check_id.c
    check_blk.c
    check_stmt.c
    check_exp.c
    ir.c
    ir_md.c
    ir_fn.c
    ir_abi.c
    ir_bb.c
    ir_sgmt.c
    trans.c
    trans_id.c
    trans_blk.c
    trans_stmt.c
    trans_exp.c
    gen.c
    gen_md.c
    gen_fn.c
    gen_bb.c
    gen_stmt.c
    gen_exp.c
    error.c
    strbuf.c
    iobuf.c
    vector.c
    stack.c
    src_pos.c
    enum.c
    meta.c
    value.c
    util.c
    xalloc.c
    xassert.c
    syslib.c
    grammar.tab.c
    scanner.yy.c
)

set(ASCL_TEST_LIBS ascl CACHE STRING "ASCL test libraries")
set(ASCL_INCLUDE_DIR ${CMAKE_CURRENT_SOURCE_DIR} CACHE STRING "ASCL include directories")

link_directories(${TOOL_LIB_DIR})

add_library(ascl STATIC ${ascl_sources})
target_include_directories(ascl PRIVATE ${TOOL_INCLUDE_DIR})

add_dependencies(ascl binaryen libgmp)
target_link_libraries(ascl binaryen stdc++ m
    ${TOOL_LIB_DIR}/${CMAKE_SHARED_LIBRARY_PREFIX}gmp${CMAKE_SHARED_LIBRARY_SUFFIX})

set_target_properties(ascl PROPERTIES
    ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/lib
    COMPILE_FLAGS "-Wall -Wextra -Wno-unused-parameter")

add_subdirectory(tests)
