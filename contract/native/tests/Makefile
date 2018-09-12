#
# @file    Makefile
# @copyright defined in aergo/LICENSE.txt
#

TARGET = parser_test

LIB_DIR = ../lib
DEP_DIR = ../.deps

CC = gcc
CPP = cpp
CFLAGS += -g -Wall
LDFLAGS += -L$(LIB_DIR)
LIBS = $(LIB_DIR)/libaergoscc.a
LINKFLAGS = $(patsubst $(LIB_DIR)/lib%.a,-l%,$(LIBS)) -lm
INCLUDES = -I..

SRCS = parser_test.c
DEPS = $(DEP_DIR)/$(SRCS:%.c=%.d)

all: $(TARGET)

$(TARGET): $(SRCS) $(LIBS)
	$(CC) $(CFLAGS) $(INCLUDES) -o $@ $(SRCS) $(LDFLAGS) $(LINKFLAGS)

$(LIBS):
	@$(MAKE) -C ..

CASES = $(shell find ./parser -name '0*' | sort)

test: $(TARGET)
	@./$(TARGET) --silent $(CASES)

dump: $(TARGET)
	@./$(TARGET) --lex-dump --yacc-dump $(CASES)

debug: $(TARGET)
	@./$(TARGET) $(CASES)

clean:
	$(RM) *.o $(TARGET)

.PHONY: clean test debug

-include $(DEPS)

$(DEP_DIR)/%.d: %.c
	@$(CPP) $(CFLAGS) $(INCLUDES) $< -MM -MT $(TARGET) >$@

.PRECIOUS: $(DEPS)
