#
# @file    Makefile
# @copyright defined in aergo/LICENSE.txt
#

TARGET = AllTests

CC = gcc
CFLAGS += -g
LDFLAGS = -L../build
LIBS = -laergoscc -lm
INCLUDES = -I. -I..

GEN_SRC = AllTests.c
SOURCES = $(GEN_SRC) test_parser.c CuTest.c 

all: $(TARGET)

$(GEN_SRC): make-tests.sh
	@./make-tests.sh > AllTests.c

$(TARGET): $(SOURCES)
	$(CC) $(CFLAGS) $(INCLUDES) -o $@ $^ $(LDFLAGS) $(LIBS)

test: $(TARGET)
	@./$(TARGET)

clean:
	$(RM) *.o $(GEN_SRC) $(TARGET)
