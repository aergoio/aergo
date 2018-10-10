/**
 * @file    parse.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PARSE_H
#define _PARSE_H

#include "common.h"

#include "strbuf.h"

#include "ast_id.h"
#include "ast_blk.h"
#include "ast_exp.h"
#include "ast_stmt.h"

typedef struct parse_s {
    char *path;
    flag_t flag;

    char *src;
    int len;
    int pos;

    /* abstract syntax tree */
    ast_t **ast;

    /* for error token */
    int adj_token;
    trace_t adj_pos;

    /* for string literal */
    strbuf_t buf;
} parse_t;

#define YYLTYPE             trace_t

#include "grammar.tab.h"

void parse(char *path, flag_t flag, strbuf_t *src, ast_t **ast);

#endif /* ! _PARSE_H */
