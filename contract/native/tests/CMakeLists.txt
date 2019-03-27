add_executable(ascl_test ascl_test.c)
target_include_directories(ascl_test PRIVATE ${CMAKE_CURRENT_LIST_DIR}/.. ${TOOL_INCLUDE_DIR})

target_link_libraries(ascl_test ${TEST_LIBS})

add_custom_target(ascl-check ascl_test DEPENDS ${TEST_LIBS})
