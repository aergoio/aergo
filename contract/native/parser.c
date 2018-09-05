/**
 * @file    parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "strbuf.h"
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

    yypos_init(&param->lloc.first);
    yypos_init(&param->lloc.last);
}

int
parse(char *path, opt_t opt, strbuf_t *src)
{
    yyparam_t param;
    void *scanner;

    yyparam_init(&param, path, src);
    yylex_init(&scanner);

    yyset_extra(&param, scanner);

    if (opt_enabled(opt, OPT_DEBUG)) {
        yyset_debug(1, scanner);
        yydebug = 1;
    }

    yyparse(&param, scanner);
    yylex_destroy(scanner);

    if (!opt_enabled(opt, OPT_TEST))
        error_dump();

    return error_last();
}

/* end of parser.c */
