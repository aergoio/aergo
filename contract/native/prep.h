/**
 * @file    preprocess.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PREPROCESS_H
#define _PREPROCESS_H

#include "common.h"

#include "location.h"

#define SCAN_BUF_SIZE       8192

#ifndef _STRBUF_T
#define _STRBUF_T
typedef struct strbuf_s strbuf_t;
#endif  /* _STRBUF_T */

typedef struct scan_s {
    char *path;
    FILE *fp;

    yypos_t loc;

    int buf_len;
    int buf_pos;
    char buf[SCAN_BUF_SIZE];

    strbuf_t *out;
} scan_t;

void preprocess(char *path, strbuf_t *out);

#endif /*_PREPROCESS_H */
