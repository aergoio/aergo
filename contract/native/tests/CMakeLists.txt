cmake_minimum_required(VERSION 3.0)

project(ascl_tests C)

set(CMAKE_BUILD_TYPE Debug)

add_executable(parser_test parser_test.c)
target_include_directories(parser_test PRIVATE ${CMAKE_SOURCE_DIR})

target_link_libraries(parser_test ${TEST_LIBS})

add_custom_target(tests DEPENDS parser_test)
add_custom_command(TARGET tests
    COMMAND ./parser_test `find ./parser -name '0*' | sort`)
