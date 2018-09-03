/**
 * @file    parser.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PARSER_H
#define _PARSER_H

#include "common.h"

#include "compile.h"

typedef struct lloc_s {
    int line;
    int col;
    int offset;
} lloc_t;

typedef struct scan_s {
    char *path;
    char *file;

    char *src;
    int len;
    int pos;

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

int parse(char *src, int len, opt_t opt);

#endif /*_PARSER_H */
