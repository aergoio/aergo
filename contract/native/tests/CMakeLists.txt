add_executable(compile_test compile_test.c)
target_include_directories(compile_test PRIVATE ${CMAKE_CURRENT_LIST_DIR}/..)

target_link_libraries(compile_test ${TEST_LIBS})

add_custom_target(ascl-check compile_test DEPENDS ${TEST_LIBS})
