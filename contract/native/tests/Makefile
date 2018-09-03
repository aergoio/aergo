#
# @file    Makefile
# @copyright defined in aergo/LICENSE.txt
#

TARGET = parser_test

CC = gcc
CFLAGS += -g
LDFLAGS = -L../build
LIBS = -laergoscc -lm
INCLUDES = -I. -I..

SOURCES = parser_test.c

all: $(TARGET)

$(TARGET): $(SOURCES)
	$(CC) $(CFLAGS) $(INCLUDES) -o $@ $^ $(LDFLAGS) $(LIBS)

test: $(TARGET)
	@./$(TARGET)

clean:
	$(RM) *.o $(TARGET)
