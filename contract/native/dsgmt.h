/**
 * @file    dsgmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _DSGMT_H
#define _DSGMT_H

#include "common.h"

#include "binaryen-c.h"

#define DSGMT_INIT_CAPACITY     10

typedef struct dsgmt_s {
    int cap;
    int size;
    uint32_t offset;

    uint32_t *lens;
    uint32_t *addrs;
    char **datas;
} dsgmt_t;

int dsgmt_add_raw(dsgmt_t *dsgmt, uint32_t len, char *raw);
int dsgmt_add_string(dsgmt_t *dsgmt, char *str);

char *dsgmt_raw(dsgmt_t *dsgmt, uint32_t addr);

static inline void
dsgmt_init(dsgmt_t *dsgmt)
{
    dsgmt->cap = DSGMT_INIT_CAPACITY;
    dsgmt->size = 0;
    dsgmt->offset = 0;

    dsgmt->lens = xmalloc(sizeof(uint32_t) * dsgmt->cap);
    dsgmt->addrs = xmalloc(sizeof(uint32_t) * dsgmt->cap);
    dsgmt->datas = xmalloc(sizeof(char *) * dsgmt->cap);
}

static inline dsgmt_t *
dsgmt_new(void)
{
    dsgmt_t *dsgmt = xmalloc(sizeof(dsgmt_t));

    dsgmt_init(dsgmt);

    return dsgmt;
}

#endif /* ! _DSGMT_H */
