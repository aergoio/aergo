cmake_minimum_required(VERSION 3.0)

project(libascl C)

find_package(BISON REQUIRED)
find_package(FLEX REQUIRED)

bison_target(PARSER grammar.y ${CMAKE_CURRENT_SOURCE_DIR}/grammar.tab.c COMPILE_FLAGS -d)
flex_target(SCANNER scanner.l ${CMAKE_CURRENT_SOURCE_DIR}/scanner.yy.c)

add_flex_bison_dependency(SCANNER PARSER)
set_directory_properties(PROPERTIES ADDITIONAL_MAKE_CLEAN_FILES "grammar.output")

set(libascl_SRCS
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
    gen_id.c
    gen_blk.c
    gen_stmt.c
    gen_exp.c
    gen_meta.c
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
    ${BISON_PARSER_OUTPUTS}
    ${FLEX_SCANNER_OUTPUTS}
)

set(TEST_LIBS ascl CACHE STRING "ascl library")

add_library(ascl STATIC ${libascl_SRCS})
target_include_directories(ascl PRIVATE ${binaryen_SOURCE_DIR}/src)

link_directories(${binaryen_SOURCE_DIR}/lib)
target_link_libraries(ascl binaryen stdc++ m)

set_target_properties(ascl PROPERTIES ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/lib COMPILE_FLAGS "-Wall -Werror -Wextra -Wno-unused-parameter")

add_subdirectory(tests)
