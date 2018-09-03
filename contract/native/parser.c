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

extern void yyset_extra(scan_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(void *);
extern int yydebug;

static void
scan_init(scan_t *scan, char *src, int len)
{
    scan->path = NULL;
    scan->file = NULL;

    scan->src = src;
    scan->len = len;
    scan->pos = 0;

    scan->lloc.line = 1;
    scan->lloc.col = 1;
    scan->lloc.offset = 0;

    scan->offset = 0;
    scan->buf = NULL;
}

int
parse(char *src, int len, opt_t opt)
{
    scan_t scan;
    void *scanner;

    scan_init(&scan, src, len);
    yylex_init(&scanner);
    yyset_extra(&scan, scanner);

    if (opt_enabled(opt, OPT_DEBUG)) {
        yyset_debug(1, scanner);
        yydebug = 1;
    }

    yyparse(scanner);
    yylex_destroy(scanner);

    if (!opt_enabled(opt, OPT_TEST))
        error_dump();

    return error_last();
}

/* end of parser.c */
