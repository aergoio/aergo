set(LIB_DIR ${CMAKE_CURRENT_SOURCE_DIR})

set(LUAJIT_SRC_DIR ${CMAKE_CURRENT_SOURCE_DIR}/src/luajit/src)
set(LUAJIT_BUILD_DIR ${LUAJIT_SRC_DIR}/..)

if(NOT EXISTS ${LUAJIT_SRC_DIR})
    execute_process(COMMAND git submodule update --init --force)
endif()

set(LUAJIT_GIT_FILE ${CMAKE_CURRENT_SOURCE_DIR}/src/luajit/.git)

if(CMAKE_BUILD_TYPE MATCHES "Debug")
    execute_process(COMMAND git --git-dir=${LUAJIT_GIT_FILE} checkout origin/debugable)
    execute_process(COMMAND git submodule update)
else()
    execute_process(COMMAND git --git-dir=${LUAJIT_GIT_FILE} checkout origin/master)
    execute_process(COMMAND git submodule update)
endif()

set(LMDB_SRC_DIR ${CMAKE_CURRENT_SOURCE_DIR}/src/lmdb)
set(LMDB_BUILD_DIR ${LMDB_SRC_DIR}/libraries/liblmdb)

if(NOT EXISTS ${LMDB_SRC_DIR})
    execute_process(COMMAND git clone https://github.com/lmdb/lmdb.git ${LMDB_SRC_DIR})
endif()

add_custom_target(libtool DEPENDS libluajit liblmdb)

add_custom_target(libluajit $(MAKE) PREFIX=${LIB_DIR} all install
	WORKING_DIRECTORY ${LUAJIT_BUILD_DIR})
add_custom_target(liblmdb $(MAKE) prefix=${LIB_DIR} all install
	WORKING_DIRECTORY ${LMDB_BUILD_DIR})

add_custom_target(libtool-clean 
    COMMAND rm -rf ${LIB_DIR}/lib/*
    COMMAND rm -rf ${LIB_DIR}/include/*
    COMMAND rm -rf ${LIB_DIR}/bin/*
    COMMAND rm -rf ${LIB_DIR}/share/*
    DEPENDS libluajit-clean liblmdb-clean)

add_custom_target(libluajit-clean $(MAKE) clean WORKING_DIRECTORY ${LUAJIT_BUILD_DIR})
add_custom_target(liblmdb-clean $(MAKE) clean WORKING_DIRECTORY ${LMDB_BUILD_DIR})
