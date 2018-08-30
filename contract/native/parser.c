/**
 * @file    parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "errors.h"
#include "strbuf.h"
#include "prep.h"

#include "parser.h"

extern int yylex_init(void **);
extern int yylex_destroy(void *);

extern void yyset_extra(scan_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(void *);
extern int yydebug;

static void
scan_init(scan_t *scan, strbuf_t *sb)
{
    scan->path = NULL;
    scan->file = NULL;

    scan->pos = 0;
    scan->len = strbuf_get_len(sb);
    scan->src = strbuf_get_str(sb);

    scan->errcnt = 0;

    scan->lloc.line = 1;
    scan->lloc.col = 1;
    scan->lloc.offset = 0;

    scan->offset = 0;
    scan->buf = NULL;
}

int
parse(char *path)
{
    scan_t scan;
    strbuf_t sb;
    void *scanner;

    strbuf_init(&sb);
    preprocess(path, &sb);

    scan_init(&scan, &sb);
    yylex_init(&scanner);
    yyset_extra(&scan, scanner);

    // yyset_debug(1, scanner);
    // yydebug = 1;

    yyparse(scanner);
    yylex_destroy(scanner);

    if (scan.errcnt > 0) {
        WARN(ERROR_PARSE_FAILED);
        return RC_ERROR;
    }

    return RC_OK;
}

/* end of parser.c */
