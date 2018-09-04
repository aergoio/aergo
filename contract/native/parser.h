/**
 * @file    parser.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PARSER_H
#define _PARSER_H

#include "common.h"

#ifndef _STRBUF_T
#define _STRBUF_T
typedef struct strbuf_s strbuf_t;
#endif  /* _STRBUF_T */

#ifndef _YYLLOC_T
#define _YYLLOC_T
typedef struct yylloc_s yylloc_t;
#endif /* _YYLLOC_T */

typedef struct yypos_s {
    int line;
    int col;
    int offset;
} yypos_t;

struct yylloc_s {
    yypos_t first;
    yypos_t last;
};

typedef struct parse_param_s {
    char *file;

    char *src;
    int len;
    int pos;

    yylloc_t lloc;
} parse_param_t;

#define YYLTYPE             yylloc_t

#include "grammar.tab.h"

int parse(char *file, opt_t opt, strbuf_t *src);

#endif /*_PARSER_H */
