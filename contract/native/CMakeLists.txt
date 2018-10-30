cmake_minimum_required(VERSION 3.0)

project(libascl C)

set(CMAKE_BUILD_TYPE Debug)

set(LEX_SRC scanner.yy.c)
set(YACC_SRC grammar.tab.c)

find_package(BISON REQUIRED)
find_package(FLEX REQUIRED)

bison_target(parser grammar.y ${YACC_SRC} COMPILE_FLAGS -d)
flex_target(scanner scanner.l ${LEX_SRC})

add_flex_bison_dependency(scanner parser)
set_directory_properties(PROPERTIES
    ADDITIONAL_MAKE_CLEAN_FILES "grammar.output")

set(SRCS
    compile.c
    parse.c
    prep.c
    check.c
    gen.c
    ast.c
    ast_id.c
    ast_blk.c
    ast_stmt.c
    ast_exp.c
    check_id.c
    check_blk.c
    check_stmt.c
    check_exp.c
    assert.c
    error.c
    strbuf.c
    array.c
    stack.c
    src_pos.c
    enum.c
    meta.c
    value.c
    util.c
    xalloc.c
    ${YACC_SRC}
    ${LEX_SRC}
)

set(TEST_LIBS ascl CACHE STRING "this is a main library")

add_library(ascl STATIC ${SRCS})
set_target_properties(ascl PROPERTIES
    ARCHIVE_OUTPUT_DIRECTORY "${CMAKE_SOURCE_DIR}/lib"
    COMPILE_FLAGS "-Wall")

add_custom_target(realclean rm -rf `find . \\\( -name CMakeCache.txt -o -name CMakeFiles -o -name cmake_install.cmake \\\)`)

add_subdirectory(tests)
