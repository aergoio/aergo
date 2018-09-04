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
extern void yyset_extra(parse_param_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(parse_param_t *, void *);
extern int yydebug;

static void
yypos_init(yypos_t *pos)
{
    pos->line = 1;
    pos->col = 1;
    pos->offset = 0;
}

static void
parse_param_init(parse_param_t *param, char *file, strbuf_t *src)
{
    param->file = file;
    ASSERT(param->file != NULL);

    param->src = strbuf_text(src);
    param->len = strbuf_length(src);
    param->pos = 0;

    yypos_init(&param->lloc.first);
    yypos_init(&param->lloc.last);
}

int
parse(char *file, opt_t opt, strbuf_t *src)
{
    parse_param_t param;
    void *scanner;

    parse_param_init(&param, file, src);
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
