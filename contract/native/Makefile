#
# @file		Makefile
# @copyright defined in aergo/LICENSE.txt
#

TARGET = ../../bin/aergoscc

CC = gcc
CFLAGS += -g
INCLUDES = -I.
LEX = flex
YACC = bison

OBJDIR = build
SUBDIRS = . tests

SOURCES = aergoscc.c sc_parser.c sc_grammar.tab.c sc_scanner.yy.c 
OBJS = $(SOURCES:%.c=$(OBJDIR)/%.o)

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

all-local: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(CFLAGS) -o $@ $^

sc_scanner.yy.c: sc_scanner.l # sc_throw.h sc_util.h sc_parser.h
	$(LEX) -o $@ --header-file=sc_scanner.yy.h $^

sc_grammar.tab.c sc_grammar.tab.h: sc_grammar.y 
	$(YACC) -o $@ -d $^

$(OBJDIR)/%.o: %.c
	@if [ ! -d $(OBJDIR) ]; then mkdir -p $(OBJDIR); fi
	$(CC) $(CFLAGS) -c -o $@ $(INCLUDES) $(DEFS) $<

clean: clean-recursive

clean-local:
	$(RM) -r build *.yy.* *.tab.* *.output $(TARGET)

test:
	@$(MAKE) -C tests $(MAKECMDGOALS)

deps .deps: $(SOURCES)
	@$(RM) .deps; \
	for src in $(SOURCES); do \
		obj=`echo $$src | sed 's/\(.*\)\.c/${OBJDIR}\/\1.o/g'`; \
		$(CC) $(INCLUDES) -MM $$src -MT $$obj >> .deps; \
	done;

include .deps

.PHONY: all-recursive clean-recursive
