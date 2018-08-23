#
# @file		Makefile
# @copyright defined in aergo/LICENSE.txt
#

TARGET = ../../bin/ascc

CC = gcc
CFLAGS = -I.
LEX = flex

OBJ_DIR = build

SRCS = ascc.c sc_parser.c sc_scanner.yy.c
OBJS = $(SRCS:%.c=$(OBJ_DIR)/%.o)

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) -g -o $@ $^

sc_scanner.yy.c: sc_scanner.l
	$(LEX) -o $@ --header-file=sc_scanner.yy.h $^

$(OBJ_DIR)/%.o: %.c
	@if [ ! -d $(OBJ_DIR) ]; then mkdir -p $(OBJ_DIR); fi
	$(CC) -c -g -o $@ $(CFLAGS) $(DEFS) $<

clean:
	rm -rf build $(TARGET)

depend: .depend

.depend: $(SRCS)
	$(CC) $(CFLAGS) -MM $^ > .depend;

include .depend
