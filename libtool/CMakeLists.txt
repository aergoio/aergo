set(LIB_DIR ${CMAKE_CURRENT_SOURCE_DIR})

set(LUAJIT_SRC_DIR ${CMAKE_CURRENT_SOURCE_DIR}/src/luajit/src)
set(LUAJIT_BUILD_DIR ${LUAJIT_SRC_DIR}/..)

if(NOT EXISTS ${LUAJIT_SRC_DIR})
    execute_process(COMMAND git submodule update --init --force)
endif()

set(LUAJIT_GIT_FILE ${CMAKE_CURRENT_SOURCE_DIR}/src/luajit/.git)

if(CMAKE_BUILD_TYPE MATCHES "Debug")
	set(CFLAGS "CFLAGS=\"-DLJ_ENABLE_DEBUG\"")
endif()

set(LMDB_SRC_DIR ${CMAKE_CURRENT_SOURCE_DIR}/src/lmdb)
set(LMDB_BUILD_DIR ${LMDB_SRC_DIR}/libraries/liblmdb)

if(NOT EXISTS ${LMDB_SRC_DIR})
    execute_process(COMMAND git clone https://github.com/lmdb/lmdb.git ${LMDB_SRC_DIR})
endif()

set(GMP_VERSION "6.3.0" CACHE STRING "Override libgmp version" )
if (NOT "$ENV{GMP_VERSION}" STREQUAL "")
    message(STATUS "Overriding libgmp version to $ENV{GMP_VERSION} by environment variable")
    set(GMP_VERSION "$ENV{GMP_VERSION}" )
endif()
message(STATUS "Using libgmp version ${GMP_VERSION}")

# for libgmp-6.3.0-aergo
set(GMP_HASH f02ece9197c05b07eaca77bcb18448bcc4e38b0b3ce77202accfc1cc9f338012)

set(GMP_NAME gmp-${GMP_VERSION})
set(GMP_ARCHIVE gmp-6.3.0-aergo.tar.bz2)
set(GMP_PATCH_FILE ${GMP_NAME}.patch)
set(GMP_SRC_DIR ${CMAKE_CURRENT_SOURCE_DIR}/src/${GMP_NAME})
set(GMP_BUILD_DIR ${CMAKE_CURRENT_BINARY_DIR}/build/${GMP_NAME})

include(ExternalProject)

ExternalProject_Add(libgmp
   SOURCE_DIR ${GMP_SRC_DIR}
   BINARY_DIR ${GMP_BUILD_DIR}
   #--Download step--------------
   DOWNLOAD_DIR ${CMAKE_CURRENT_SOURCE_DIR}/download
   URL          ${CMAKE_CURRENT_SOURCE_DIR}/download/${GMP_ARCHIVE}
   URL_HASH     SHA256=${GMP_HASH}
   TMP_DIR      ${GMP_BUILD_DIR}
   STAMP_DIR    ${GMP_BUILD_DIR}
   CONFIGURE_COMMAND ${GMP_SRC_DIR}/configure --prefix=${LIB_DIR} --disable-static --enable-shared > /dev/null
   PREFIX ${GMP_SRC_DIR}
   BUILD_COMMAND $(MAKE) all install > make.log
   BUILD_ALWAYS 1)

add_custom_target(libtool DEPENDS libluajit liblmdb libgmp)

add_custom_target(libluajit $(MAKE) PREFIX=${LIB_DIR} ${CFLAGS} all install
	WORKING_DIRECTORY ${LUAJIT_BUILD_DIR})
add_custom_target(liblmdb $(MAKE) prefix=${LIB_DIR} all install
	WORKING_DIRECTORY ${LMDB_BUILD_DIR})

add_custom_target(libtool-clean 
    COMMAND rm -rf ${LIB_DIR}/lib/*
    COMMAND rm -rf ${LIB_DIR}/include/*
    COMMAND rm -rf ${LIB_DIR}/bin/*
    COMMAND rm -rf ${LIB_DIR}/share/*
    DEPENDS libluajit-clean liblmdb-clean libgmp-clean)

add_custom_target(libluajit-clean $(MAKE) clean WORKING_DIRECTORY ${LUAJIT_BUILD_DIR})
add_custom_target(liblmdb-clean $(MAKE) clean WORKING_DIRECTORY ${LMDB_BUILD_DIR})
add_custom_target(libgmp-clean $(MAKE) clean WORKING_DIRECTORY ${GMP_BUILD_DIR})
