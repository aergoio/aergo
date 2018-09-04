#
# @file    Makefile
# @copyright defined in aergo/LICENSE.txt
#

TARGET = parser_test

OBJ_DIR = ../build
DEP_DIR = ../.deps

CC = gcc
CPP = cpp
CFLAGS += -g -Wall
LDFLAGS += -L$(OBJ_DIR)
LIBS = $(OBJ_DIR)/libaergoscc.a
LINKFLAGS = $(patsubst $(OBJ_DIR)/lib%.a,-l%,$(LIBS)) -lm
INCLUDES = -I..

SRCS = parser_test.c
DEPS = $(DEP_DIR)/$(SRCS:%.c=%.d)

all: $(TARGET)

$(TARGET): $(SRCS) $(LIBS)
	$(CC) $(CFLAGS) $(INCLUDES) -o $@ $(SRCS) $(LDFLAGS) $(LINKFLAGS)

$(LIBS):
	@$(MAKE) -C ..

test: $(TARGET)
	@./$(TARGET)

debug: $(TARGET)
	@./$(TARGET) --debug

clean:
	$(RM) *.o $(TARGET)

.PHONY: clean test debug

-include $(DEPS)

$(DEP_DIR)/%.d: %.c
	@$(CPP) $(CFLAGS) $(INCLUDES) $< -MM -MT $(TARGET) >$@

.PRECIOUS: $(DEPS)
