/**
 * @file    dsgmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _DSGMT_H
#define _DSGMT_H

#include "common.h"

#define DSGMT_INIT_CAPACITY     10

typedef struct dsgmt_s {
    int cap;
    int size;
    int offset;

    int *lens;
    int *addrs;
    char **datas;
} dsgmt_t;

static inline void
dsgmt_init(dsgmt_t *dsgmt)
{
    dsgmt->cap = DSGMT_INIT_CAPACITY;
    dsgmt->size = 0;
    dsgmt->offset = 0;

    dsgmt->lens = xmalloc(sizeof(int) * dsgmt->cap);
    dsgmt->addrs = xmalloc(sizeof(int) * dsgmt->cap);
    dsgmt->datas = xmalloc(sizeof(char *) * dsgmt->cap);
}

static inline dsgmt_t *
dsgmt_new(void)
{
    dsgmt_t *dsgmt = xmalloc(sizeof(dsgmt_t));

    dsgmt_init(dsgmt);

    return dsgmt;
}

static inline void
dsgmt_extend(dsgmt_t *dsgmt)
{
    dsgmt->cap += DSGMT_INIT_CAPACITY;

    dsgmt->lens = xrealloc(dsgmt->lens, sizeof(int) * dsgmt->cap);
    dsgmt->addrs = xrealloc(dsgmt->addrs, sizeof(int) * dsgmt->cap);
    dsgmt->datas = xrealloc(dsgmt->datas, sizeof(char *) * dsgmt->cap);
}

static inline int
dsgmt_add_raw(dsgmt_t *dsgmt, int len, char *raw)
{
    int offset = dsgmt->offset;

    if (dsgmt->size >= dsgmt->cap)
        dsgmt_extend(dsgmt);

    dsgmt->lens[dsgmt->size] = len;
    dsgmt->addrs[dsgmt->size] = offset;
    dsgmt->datas[dsgmt->size] = raw;

    dsgmt->size++;
    dsgmt->offset += ALIGN64(len);

    return offset;
}

static inline int
dsgmt_add_string(dsgmt_t *dsgmt, char *str)
{
    int i;
    int len = strlen(str) + 1;  /* including '\0' */

    for (i = 0; i < dsgmt->size; i++) {
        if (dsgmt->lens[i] == len && strcmp(dsgmt->datas[i], str) == 0)
            return dsgmt->addrs[i];
    }

    return dsgmt_add_raw(dsgmt, strlen(str) + 1, str);
}

static inline char *
dsgmt_raw(dsgmt_t *dsgmt, int addr)
{
    int i;

    for (i = 0; i < dsgmt->size; i++) {
        if (dsgmt->addrs[i] == addr)
            return dsgmt->datas[i];
    }

    ASSERT1(!"invalid address", addr);

    return NULL;
}

#endif /* ! _DSGMT_H */
