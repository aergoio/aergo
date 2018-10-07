/**
 * @file    prep.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PREP_H
#define _PREP_H

#include "common.h"

#include "strbuf.h"

typedef struct scan_s {
    char *path;
    char *work_dir;

    int offset;
    strbuf_t in;

    trace_t trc;

    strbuf_t *out;
} scan_t;

void preprocess(char *path, flag_t flag, strbuf_t *out);

#endif /* ! _PREP_H */
