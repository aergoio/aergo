/**
 * @file    parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "prep.h"
#include "util.h"

#include "parser.h"

extern int yylex_init(void **);
extern int yylex_destroy(void *);

extern void yyset_in(FILE *, void *);
extern void yyset_extra(yyparam_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(yyparam_t *, void *);
extern int yydebug;

static void
yyparam_init(yyparam_t *param, char *path, strbuf_t *src)
{
    param->path = path;
    ASSERT(param->path != NULL);

    param->src = strbuf_text(src);
    param->len = strbuf_length(src);
    param->pos = 0;

    param->adj_token = 0;
    errpos_init(&param->adj_pos, path);

    strbuf_init(&param->buf);
}

void
parse(char *path, flag_t flag, strbuf_t *src)
{
    yyparam_t param;
    void *scanner;

    yyparam_init(&param, path, src);
    yylex_init(&scanner);

    yyset_extra(&param, scanner);

    if (flag_enabled(flag, FLAG_LEX_DUMP))
        yyset_debug(1, scanner);

    if (flag_enabled(flag, FLAG_YACC_DUMP))
        yydebug = 1;

    yyparse(&param, scanner);
    yylex_destroy(scanner);
}

/* end of parser.c */
