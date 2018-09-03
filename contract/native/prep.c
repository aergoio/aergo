/**
 * @file    preprocess.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "strbuf.h"

#include "prep.h"

extern int yylex_init(void **);
extern int yylex_destroy(void *);

extern int yylex(void *);

extern void yyset_extra(subst_t *, void *);
extern void yyset_debug(int, void *);

static void
subst_init(subst_t *subst, char *path, strbuf_t *res)
{
    strbuf_t sb;

    subst->path = path;
    subst->line = 1;

    strbuf_init(&sb);
    read_file(path, &sb);

    subst->len = strbuf_length(&sb);
    subst->src = strbuf_text(&sb);

    append_directive(path, subst->line, res);
    subst->res = res;
}

int
preprocess(char *path, opt_t opt, strbuf_t *res)
{
    subst_t subst;
    void *scanner;

    subst_init(&subst, path, res);

    yylex_init(&scanner);
    yyset_extra(&subst, scanner);

    if (opt_enabled(opt, OPT_DEBUG))
        yyset_debug(1, scanner);

    yylex(scanner);
    yylex_destroy(scanner);

    if (!opt_enabled(opt, OPT_TEST))
        error_dump();

    return error_last();
}

void
append_directive(char *path, int line, strbuf_t *res)
{
    char buf[PATH_MAX_LEN + 32];

    snprintf(buf, sizeof(buf), "#file \"%s\" %d\n", path, line);

    strbuf_append(res, buf, strlen(buf));
}

/* end of preprocess.c */
