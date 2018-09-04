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

typedef struct prep_param_s {
    char *file;
    FILE *fp;

    int line;

    strbuf_t *res;
} prep_param_t;

void preprocess(char *file, strbuf_t *res);

#endif /*_PREPROCESS_H */
