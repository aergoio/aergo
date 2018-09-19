cmake_minimum_required(VERSION 3.0)

project(libascl C)

set(CMAKE_BUILD_TYPE Debug)

set(LEX_SRC scanner.yy.c)
set(YACC_SRC grammar.tab.c)

find_package(BISON)
find_package(FLEX)

bison_target(parser grammar.y ${YACC_SRC} COMPILE_FLAGS -d)
flex_target(scanner scanner.l ${LEX_SRC})

add_flex_bison_dependency(scanner parser)
set_directory_properties(PROPERTIES 
    ADDITIONAL_MAKE_CLEAN_FILES "grammar.output") 

set(SRCS 
    compile.c 
    parser.c 
    prep.c
	ast_var.c 
    ast_struct.c 
    ast_blk.c 
    ast_stmt.c 
    ast_exp.c 
    ast_func.c
	error.c 
    strbuf.c 
    stack.c 
    util.c 
    xalloc.c
    ${YACC_SRC}
    ${LEX_SRC}
)

add_library(libascl STATIC ${SRCS})
set_target_properties(libascl PROPERTIES 
    ARCHIVE_OUTPUT_DIRECTORY "${CMAKE_SOURCE_DIR}/lib")

add_executable(parser_test tests/parser_test.c)
set_target_properties(parser_test PROPERTIES RUNTIME_OUTPUT_DIRECTORY "./tests")
target_include_directories(parser_test PRIVATE ${CMAKE_SOURCE_DIR})
target_link_libraries(parser_test libascl)

add_custom_target(tests DEPENDS parser_test libascl)
add_custom_command(TARGET tests
    COMMAND ./parser_test --silent `find ./parser -name '0*' | sort`
    WORKING_DIRECTORY "${CMAKE_SOURCE_DIR}/tests")
