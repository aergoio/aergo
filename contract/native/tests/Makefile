#
# @file    Makefile
# @copyright defined in aergo/LICENSE.txt
#

TARGET = AllTests

CC = gcc
CFLAGS += -g
INCLUDES = -I. -I..

GEN_SRC = AllTests.c
SOURCES = $(GEN_SRC) test_parser.c CuTest.c \
	../sc_parser.c ../sc_scanner.yy.c ../sc_grammar.tab.c

all: $(TARGET)

$(GEN_SRC): make-tests.sh
	@./make-tests.sh > AllTests.c

$(TARGET): $(SOURCES)
	$(CC) $(CFLAGS) $(INCLUDES) -o $@ $^

test: $(TARGET)
	@./$(TARGET)

clean:
	$(RM) *.o $(GEN_SRC) $(TARGET)
