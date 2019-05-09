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

.PHONY: all release debug clean

all: $(BUILD_FILE)
	@$(MAKE) --no-print-directory -C $(BUILD_DIR)

$(BUILD_FILE):
	@mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && $(CMAKE_CMD) -G "Unix Makefiles" -D CMAKE_BUILD_TYPE="Release" $(MAKE_FLAG) ..

release:
	@mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && $(CMAKE_CMD) -G "Unix Makefiles" -D CMAKE_BUILD_TYPE="Release" $(MAKE_FLAG) ..
	@$(MAKE) --no-print-directory -C $(BUILD_DIR)

debug:
	@mkdir -p $(BUILD_DIR)
	cd $(BUILD_DIR) && $(CMAKE_CMD) -G "Unix Makefiles" -D CMAKE_BUILD_TYPE="Debug" $(MAKE_FLAG) ..
	@$(MAKE) --no-print-directory -C $(BUILD_DIR)

clean:
	@$(MAKE) --no-print-directory -C $(BUILD_DIR) distclean

realclean: clean
	@rm -rf $(BUILD_DIR)

%: $(BUILD_DIR)
	@$(MAKE) --no-print-directory -C $(BUILD_DIR) $(MAKECMDGOALS)
