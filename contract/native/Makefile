#
# @file		Makefile
# @copyright defined in aergo/LICENSE.txt
#

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
LIB_DIR = lib
DEP_DIR = .deps

TARGET = ../../bin/aergoscc
LIBRARY = $(LIB_DIR)/libaergoscc.a

SRCS = compile.c parser.c grammar.tab.c scanner.yy.c prep.c \
	error.c strbuf.c stack.c util.c xalloc.c

OBJS = $(SRCS:%.c=$(OBJ_DIR)/%.o)
DEPS = $(SRCS:%.c=$(DEP_DIR)/%.d)

SUBDIRS = . tests

all: $(OBJ_DIR) $(LIB_DIR) $(DEP_DIR) all-recursive

$(OBJ_DIR) $(LIB_DIR) $(DEP_DIR):
	@$(MKDIR_P) $@

all-recursive clean-recursive:
	@target=`echo "$@" | sed s/-recursive//`; \
	for subdir in $(SUBDIRS); do \
		if [ "$$subdir" = "." ]; then \
			local_target="$$target-local"; \
		else \
			local_target="$$target"; \
		fi; \
		(cd $$subdir && $(MAKE) $$local_target); \
		if [ $$? != 0 ]; then exit 1; fi; \
	done

all-local: $(LIBRARY) 

.SUFFIXES: .c .o

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
	$(RM) -r *.yy.* *.tab.* *.output $(OBJ_DIR) $(LIB_DIR) $(TARGET)

test: $(LIBRARY)
	@$(MAKE) -C tests $(MAKECMDGOALS)

-include $(DEPS)

$(DEP_DIR)/%.d: %.c
	@$(CPP) $(CFLAGS) $< -MM -MT $(@:.d=.o) >$@

.PHONY: all-recursive clean-recursive clean-local clean
