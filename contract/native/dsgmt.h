/**
 * @file    dsgmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _DSGMT_H
#define _DSGMT_H

#include "common.h"

#include "binaryen-c.h"

#define DSGMT_INIT_CAPACITY         10

#define dsgmt_occupy(dsgmt, l)      dsgmt_add(dsgmt, xcalloc(l), l)

typedef struct dsgmt_s {
    int cap;
    int size;
    uint32_t offset;

    uint32_t *lens;
    uint32_t *addrs;
    char **datas;
} dsgmt_t;

int dsgmt_add(dsgmt_t *dsgmt, void *ptr, uint32_t len);

static inline void
dsgmt_init(dsgmt_t *dsgmt)
{
    void *null = xcalloc(sizeof(void *));

    dsgmt->cap = DSGMT_INIT_CAPACITY;
    dsgmt->size = 0;
    dsgmt->offset = 0;

    dsgmt->lens = xmalloc(sizeof(uint32_t) * dsgmt->cap);
    dsgmt->addrs = xmalloc(sizeof(uint32_t) * dsgmt->cap);
    dsgmt->datas = xmalloc(sizeof(char *) * dsgmt->cap);

    dsgmt_add(dsgmt, null, sizeof(void *));
}

static inline dsgmt_t *
dsgmt_new(void)
{
    dsgmt_t *dsgmt = xmalloc(sizeof(dsgmt_t));

    dsgmt_init(dsgmt);

    return dsgmt;
}

#endif /* ! _DSGMT_H */
