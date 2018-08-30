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

    subst->len = strbuf_get_len(&sb);
    subst->src = strbuf_get_str(&sb);

    append_directive(res, path, subst->line);
    subst->res = res;
}

void
preprocess(char *path, strbuf_t *res)
{
    subst_t subst;
    void *scanner;

    subst_init(&subst, path, res);

    yylex_init(&scanner);
    yyset_extra(&subst, scanner);

    // yyset_debug(1, scanner);

    yylex(scanner);
    yylex_destroy(scanner);
}

void
append_directive(strbuf_t *res, char *path, int line)
{
    char buf[PATH_MAX_LEN + 32];

    snprintf(buf, sizeof(buf), "#file \"%s\" %d\n", path, line);

    strbuf_append(res, buf, strlen(buf));
}

/* end of preprocess.c */
