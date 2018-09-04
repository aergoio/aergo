/**
 * @file    preprocess.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "strbuf.h"

#include "prep.h"

/*
extern int prep_yylex_init(void **);
extern int prep_yylex_destroy(void *);

extern int prep_yylex(void *);

extern void prep_yyset_in(FILE *, void *);
extern void prep_yyset_extra(prep_param_t *, void *);
extern void prep_yyset_debug(int, void *);

static void
prep_param_init(prep_param_t *param, char *file, strbuf_t *res)
{
    param->file = file;
    ASSERT(param->file != NULL);

    param->fp = open_file(param->file, "r");
    param->line = 1;

    param->res = res;
}

int
preprocess(char *file, opt_t opt, strbuf_t *res)
{
    prep_param_t param;
    void *scanner;

    prep_param_init(&param, file, res);
    prep_yylex_init(&scanner);

    prep_yyset_in(param.fp, scanner);
    prep_yyset_extra(&param, scanner);

    if (opt_enabled(opt, OPT_DEBUG))
        prep_yyset_debug(1, scanner);

    prep_yylex(scanner);
    prep_yylex_destroy(scanner);

    ASSERT(error_empty());

    return NO_ERROR;
}
*/

typedef struct scan_s {
    FILE *fp;
    int len;
    int pos;
    int size;
    char buf[8192];
    strbuf_t *res;
} scan_t;

static void
scan_init(scan_t *scan, char *file, strbuf_t *res)
{
    scan->fp = open_file(file, "r");
    scan->len = 0;
    scan->pos = 0;
    scan->size = 8192;
    scan->buf[0] = '\0';
    scan->res = res;
}

static bool
scan_next(scan_t *scan)
{
    if (scan->pos == scan->len) {
        scan->len = fread(scan->buf, 1, scan->size, scan->fp);
        if (scan->len == 0)
            return false;

        scan->pos = 0;
    }

    return true;
}

static char
scan_get(scan_t *scan)
{
    if (!scan_next(scan))
        return EOF;

    return scan->buf[scan->pos++];
}

static void
scan_put(scan_t *scan)
{
    scan->pos--;
}

static char
scan_peek(scan_t *scan, int offset)
{
    if (!scan_next(scan))
        return EOF;

    return scan->buf[scan->pos + offset];
}

static void
copy_comment(scan_t *scan)
{
}

static void
copy_literal(scan_t *scan)
{
}

static void
copy_char

void
preprocess(char *file, strbuf_t *res)
{
    char c;
    scan_t scan;

    scan_init(&scan, file, res);

    while ((c = scan_get(scan)) != EOF) {
        if (c == '/') {
            char n = scan_peek(scan);
            if (n == '*' || n == '/')
                copy_comment(scan);
        }
        else if (c == '"') {
            copy_literal(scan);
        }
        else if (isspace(c)) {
            copy_whitespace(scan);
        }
        else if (c == 'i') {
            check_import(scan);
        }
    }
}

/* end of preprocess.c */
