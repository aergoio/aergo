/**
 * @file    parser.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PARSER_H
#define _PARSER_H

#include "common.h"

#include "compile.h"

#ifndef _STRBUF_T
#define _STRBUF_T
typedef struct strbuf_s strbuf_t;
#endif  /* _STRBUF_T */

typedef struct yyparam_s {
    char *path;

    char *src;
    int len;
    int pos;

    yylloc_t lloc;
} yyparam_t;

#define YYLTYPE             yylloc_t

#include "grammar.tab.h"

int parse(char *path, opt_t opt, strbuf_t *src);

#endif /*_PARSER_H */
