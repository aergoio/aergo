/**
 * @file    dsgmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _DSGMT_H
#define _DSGMT_H

#include "common.h"

#include "binaryen-c.h"

#define DSGMT_INIT_CAPACITY         10

#define dsgmt_occupy(dsgmt, mdl, l) dsgmt_add((dsgmt), (mdl), NULL, (l))

typedef struct dsgmt_s {
    int cap;
    int size;
    uint32_t offset;

    BinaryenIndex *lens;
    BinaryenExpressionRef *addrs;
    char **datas;
} dsgmt_t;

int dsgmt_add(dsgmt_t *dsgmt, BinaryenModuleRef module, void *ptr, uint32_t len);

static inline void
dsgmt_init(dsgmt_t *dsgmt)
{
    dsgmt->cap = DSGMT_INIT_CAPACITY;
    dsgmt->size = 0;
    dsgmt->offset = 0;

    dsgmt->lens = xmalloc(sizeof(BinaryenIndex) * dsgmt->cap);
    dsgmt->addrs = xmalloc(sizeof(BinaryenExpressionRef) * dsgmt->cap);
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
