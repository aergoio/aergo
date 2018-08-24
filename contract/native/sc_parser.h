/**
 * @file    sc_parser.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SC_PARSER_H
#define _SC_PARSER_H

#include "sc_common.h"

typedef struct sc_lloc_s {
    int line;
    int column;
    int offset;
} sc_lloc_t;

typedef struct sc_yyextra_s {
    char *path;
    char file[SC_PATH_MAX_LEN];

    int errcnt;
    sc_lloc_t lloc;     // source position

    /* temporary buffer for literal */
    int offset;
    char *buf;
} sc_yyextra_t;

int sc_parse(char *path);

#endif /* no _SC_PARSER_H */
