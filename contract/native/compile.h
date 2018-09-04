/**
 * @file    compile.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _COMPILE_H
#define _COMPILE_H

#include "common.h"

#include "strbuf.h"

typedef struct yypos_s {
    int line;
    int col;
    int offset;
} yypos_t;

typedef struct yylloc_s {
    yypos_t first;
    yypos_t last;
} yylloc_t;

typedef struct yyparam_s {
    char *path;
    FILE *fp;

    yylloc_t lloc;

    strbuf_t buf;
} yyparam_t;

int compile(char *path, opt_t opt);

void yyparam_init(yyparam_t *param, char *path);
char *yyparam_trace(yyparam_t *param);

#endif /* _COMPILE_H */
