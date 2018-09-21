cmake_minimum_required(VERSION 3.0)

project(ascl_tests C)

set(CMAKE_BUILD_TYPE Debug)

add_executable(compile_test compile_test.c)
target_include_directories(compile_test PRIVATE ${CMAKE_SOURCE_DIR})

target_link_libraries(compile_test ${TEST_LIBS})

add_custom_target(check DEPENDS compile_test)
add_custom_command(TARGET check COMMAND compile_test)
