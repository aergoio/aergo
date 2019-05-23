#
# @file     Makefile
# @copyright defined in aergo/LICENSE.txt
#

.SUFFIXES:

CMAKE_CMD ?= cmake

BUILD_DIR := build
BUILD_FILE := $(BUILD_DIR)/Makefile

ifeq ($(OS),Windows_NT)
    ifneq ($(filter MINGW%,$(shell uname 2>/dev/null || echo Unknown)),)
	    MAKE_FLAG := -D CMAKE_MAKE_PROGRAM=mingw32-make.exe
    endif
endif

BUILD_RULES := \
	deps \
	aergocli aergosvr aergoluac polaris colaris brick \
	libtool libtool-clean \
	libluajit liblmdb libgmp \
	libluajit-clean liblmdb-clean libgmp-clean \
	check cover-check \
	distclean \
	protoc protoclean

.PHONY: all release debug clean $(BUILD_RULES)

all: $(BUILD_FILE)
	@$(MAKE) --no-print-directory -C $(BUILD_DIR)

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

$(BUILD_FILE): $(BUILD_DIR)
	@if ! [ -f $(BUILD_FILE) ]; then \
		cd $(BUILD_DIR) && $(CMAKE_CMD) -G "Unix Makefiles" -D CMAKE_BUILD_TYPE="Release" $(MAKE_FLAG) ..; \
	fi

release: $(BUILD_DIR)
	cd $(BUILD_DIR) && $(CMAKE_CMD) -G "Unix Makefiles" -D CMAKE_BUILD_TYPE="Release" $(MAKE_FLAG) ..
	@$(MAKE) --no-print-directory -C $(BUILD_DIR)

debug: $(BUILD_DIR)
	@cd $(BUILD_DIR) && $(CMAKE_CMD) -G "Unix Makefiles" -D CMAKE_BUILD_TYPE="Debug" $(MAKE_FLAG) ..
	@$(MAKE) --no-print-directory -C $(BUILD_DIR)

clean:
	@$(MAKE) --no-print-directory -C $(BUILD_DIR) distclean

realclean: clean
	@rm -rf $(BUILD_DIR)

$(BUILD_RULES): $(BUILD_FILE)
	@$(MAKE) --no-print-directory -C $(BUILD_DIR) $@
