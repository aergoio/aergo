#
# @file		Makefile
# @copyright defined in aergo/LICENSE.txt
#

TARGET = ../../bin/aergoscc
LIBRARY = $(OBJ_DIR)/libaergoscc.a

CC = cc
CPP = cpp
AR = ar
RANLIB = ranlib
CFLAGS += -g -Wall
LDFLAGS = -Lbuild
INCLUDES = -I.
LIBS = -laergoscc -lm
LEX = flex
YACC = bison
MKDIR_P = mkdir -p

OBJ_DIR = build
SUBDIRS = . tests

SRCS = prep.c parser.c grammar.tab.c scanner.yy.c errors.c
OBJS = $(SRCS:%.c=$(OBJ_DIR)/%.o)
DEPS = $(OBJS:%.o=%.d)

RECURSIVE_TARGETS = all-recursive clean-recursive

all: all-recursive

all-recursive clean-recursive:
	@target=`echo "$@" | sed s/-recursive//`; \
	for subdir in $(SUBDIRS); do \
		if [ "$$subdir" = "." ]; then \
			local_target="$$target-local"; \
		else \
			local_target="$$target"; \
		fi; \
		(cd $$subdir && $(MAKE) $$local_target); \
	done

all-local: $(LIBRARY) $(TARGET)

.SUFFIXES: .c .o

$(TARGET): $(OBJ_DIR)/aergoscc.o
	$(CC) $(CFLAGS) -o $@ $^ $(LDFLAGS) $(LIBS)

$(LIBRARY): $(OBJS)
	$(AR) rcs $@ $^
	$(RANLIB) $@

scanner.yy.c: scanner.l grammar.tab.h
	$(LEX) -o $@ $^

grammar.tab.c: grammar.y
	$(YACC) -o $@ -d $^

$(OBJ_DIR)/%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $(INCLUDES) $<

clean: clean-recursive

clean-local:
	$(RM) -r build *.yy.* *.tab.* *.output $(TARGET)

test:
	@$(MAKE) -C tests $(MAKECMDGOALS)

-include $(DEPS)

$(OBJ_DIR)/%.d: %.c
	@if [ ! -d $(OBJ_DIR) ]; then $(MKDIR_P) $(OBJ_DIR); fi
	@$(CPP) $(CFLAGS) $< -MM -MT $(@:.d=.o) >$@

.PHONY: all-recursive clean-recursive clean-local clean
