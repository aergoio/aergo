/**
 * @file    parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "prep.h"
#include "util.h"

#include "parse.h"

extern int yylex_init(void **);
extern int yylex_destroy(void *);

extern void yyset_in(FILE *, void *);
extern void yyset_extra(parse_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(parse_t *, void *);
extern int yydebug;

static void
parse_init(parse_t *parse, char *path, strbuf_t *src, ast_t **ast)
{
    parse->path = path;
    ASSERT(parse->path != NULL);

    parse->src = strbuf_text(src);
    parse->len = strbuf_length(src);
    parse->pos = 0;

    parse->ast = ast;
    parse->blk = NULL;

    parse->adj_token = 0;
    trace_init(&parse->adj_pos, parse->src, path);

    strbuf_init(&parse->buf);
}

void
parse(char *path, flag_t flag, strbuf_t *src, ast_t **ast)
{
    parse_t parse;
    void *scanner;

    parse_init(&parse, path, src, ast);
    yylex_init(&scanner);

    yyset_extra(&parse, scanner);

    if (flag_on(flag, FLAG_LEX_DUMP))
        yyset_debug(1, scanner);

    if (flag_on(flag, FLAG_YACC_DUMP))
        yydebug = 1;

    yyparse(&parse, scanner);
    yylex_destroy(scanner);
}

/* end of parser.c */
