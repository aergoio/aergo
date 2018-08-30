/**
 * @file    parser.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PARSER_H
#define _PARSER_H

#include "common.h"

typedef struct lloc_s {
    int line;
    int col;
    int offset;
} lloc_t;

typedef struct scan_s {
    char *path;
    char *file;

    int pos;
    int len;
    char *src;

    int errcnt;
    lloc_t lloc;     // source position

    /* temporary buffer for literal */
    int offset;
    char *buf;
} scan_t;

typedef struct yacc_s {
    char *src;
    void *scanner;
} yacc_t;

#define YYLTYPE             lloc_t

#include "grammar.tab.h"

int parse(char *path);

#endif /* no _PARSER_H */
