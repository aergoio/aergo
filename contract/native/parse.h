/**
 * @file    parse.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PARSE_H
#define _PARSE_H

#include "common.h"

#include "flag.h"
#include "strbuf.h"

#include "ast_id.h"
#include "ast_blk.h"
#include "ast_exp.h"
#include "ast_stmt.h"

typedef struct parse_s {
    char *path;
    flag_t flag;

    /* input source information */
    char *src;
    int len;
    int pos;

    /* abstract syntax tree */
    ast_t *ast;

    /* for handling error token */
    int adj_token;
    src_pos_t adj_pos;

    /* string literal buffer */
    strbuf_t buf;

    /* vector of label identifier */
    vector_t labels;
} parse_t;

#define YYLTYPE             src_pos_t

#include "grammar.tab.h"

void parse(char *path, flag_t flag, ast_t *ast);

#endif /* ! _PARSE_H */
