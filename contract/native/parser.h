/**
 * @file    parser.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PARSER_H
#define _PARSER_H

#include "common.h"

#include "strbuf.h"
#include "list.h"

#include "ast_meta.h"
#include "ast_var.h"
#include "ast_blk.h"
#include "ast_exp.h"
#include "ast_stmt.h"
#include "ast_struct.h"
#include "ast_func.h"

typedef struct yyparam_s {
    char *path;

    char *src;
    int len;
    int pos;

    list_t *blk_l;

    int adj_token;
    errpos_t adj_pos;

    strbuf_t buf;
} yyparam_t;

#define YYLTYPE             errpos_t

#include "grammar.tab.h"

void parse(char *path, flag_t flag, strbuf_t *src, list_t *blk_l);

#endif /* ! _PARSER_H */
