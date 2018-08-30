/**
 * @file    preprocess.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PREPROCESS_H
#define _PREPROCESS_H

#include "common.h"

#ifndef _STRBUF_T
#define _STRBUF_T
typedef struct strbuf_s strbuf_t;
#endif  /* _STRBUF_T */

typedef struct subst_s {
    char *path;
    int line;

    int len;
    char *src;

    strbuf_t *res;
} subst_t;

void preprocess(char *path, strbuf_t *res);

void append_directive(strbuf_t *res, char *path, int line);

#endif /* no _PREPROCESS_H */
