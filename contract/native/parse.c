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
parse_init(parse_t *ctx, char *path, strbuf_t *src, ast_t **ast)
{
    ctx->path = path;
    ASSERT(ctx->path != NULL);

    ctx->src = strbuf_text(src);
    ctx->len = strbuf_length(src);
    ctx->pos = 0;

    ctx->ast = ast;

    ctx->adj_token = 0;
    errpos_init(&ctx->adj_pos, path);

    strbuf_init(&ctx->buf);
}

void
parse(char *path, flag_t flag, strbuf_t *src, ast_t **ast)
{
    parse_t ctx;
    void *scanner;

    parse_init(&ctx, path, src, ast);
    yylex_init(&scanner);

    yyset_extra(&ctx, scanner);

    if (flag_on(flag, FLAG_LEX_DUMP))
        yyset_debug(1, scanner);

    if (flag_on(flag, FLAG_YACC_DUMP))
        yydebug = 1;

    yyparse(&ctx, scanner);
    yylex_destroy(scanner);
}

/* end of parser.c */
