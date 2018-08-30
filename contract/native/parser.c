/**
 * @file    parser.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "throw.h"
#include "xutil.h"

#include "parser.h"

extern int yylex_init(void **);
extern int yylex_destroy(void *);

extern void yyset_in(FILE *, void *);
extern void yyset_extra(scan_t *, void *);
extern void yyset_debug(int, void *);

extern int yyparse(void *);
extern int yydebug;

static void
scan_init(scan_t *scan, char *path)
{
    char *delim;

    scan->path = path;

    delim = strrchr(path, PATH_DELIM);
    strcpy(scan->file, delim == NULL ? path : delim + 1);

    scan->errcnt = 0;

    scan->lloc.line = 1;
    scan->lloc.column = 1;
    scan->lloc.offset = 0;

    scan->offset = 0;
    scan->buf = NULL;
}

static void
yacc_init(yacc_t *yacc)
{
    yylex_init(&yacc->scanner);
}

int
parse(char *path)
{
    FILE *fp;
    scan_t scan;
    yacc_t yacc;

    fp = xfopen(path, "r");
    ASSERT(fp != NULL);

    scan_init(&scan, path);
    yacc_init(&yacc);

    yyset_in(fp, yacc.scanner);
    yyset_extra(&scan, yacc.scanner);

    // yyset_debug(1, yacc.scanner);
    // yydebug = 1;

    yyparse(yacc.scanner);
    yylex_destroy(yacc.scanner);

    xfclose(fp);

    if (scan.errcnt > 0) {
        WARN(ERROR_PARSE_FAILED);
        return RC_ERROR;
    }

    return RC_OK;
}

/* end of parser.c */
