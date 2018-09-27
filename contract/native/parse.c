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
extern void yyset_extra(yyparam_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(yyparam_t *, void *);
extern int yydebug;

static void
yyparam_init(yyparam_t *penv, char *path, strbuf_t *src, ast_t **ast)
{
    penv->path = path;
    ASSERT(penv->path != NULL);

    penv->src = strbuf_text(src);
    penv->len = strbuf_length(src);
    penv->pos = 0;

    penv->ast = ast;

    penv->adj_token = 0;
    errpos_init(&penv->adj_pos, path);

    strbuf_init(&penv->buf);
}

void
parse(char *path, flag_t flag, strbuf_t *src, ast_t **ast)
{
    yyparam_t penv;
    void *scanner;

    yyparam_init(&penv, path, src, ast);
    yylex_init(&scanner);

    yyset_extra(&penv, scanner);

    if (flag_on(flag, FLAG_LEX_DUMP))
        yyset_debug(1, scanner);

    if (flag_on(flag, FLAG_YACC_DUMP))
        yydebug = 1;

    yyparse(&penv, scanner);
    yylex_destroy(scanner);
}

/* end of parser.c */
