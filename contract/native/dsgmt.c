/**
 * @file    dsgmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "dsgmt.h"

static void
dsgmt_extend(dsgmt_t *dsgmt)
{
    dsgmt->cap += DSGMT_INIT_CAPACITY;

    dsgmt->lens = xrealloc(dsgmt->lens, sizeof(uint32_t) * dsgmt->cap);
    dsgmt->addrs = xrealloc(dsgmt->addrs, sizeof(uint32_t) * dsgmt->cap);
    dsgmt->datas = xrealloc(dsgmt->datas, sizeof(char *) * dsgmt->cap);
}

int
dsgmt_add(dsgmt_t *dsgmt, void *ptr, uint32_t len)
{
    uint32_t offset = dsgmt->offset;

    if (dsgmt->size >= dsgmt->cap)
        dsgmt_extend(dsgmt);

    dsgmt->lens[dsgmt->size] = len;
    dsgmt->addrs[dsgmt->size] = offset;
    dsgmt->datas[dsgmt->size] = ptr;

    dsgmt->size++;
    dsgmt->offset += ALIGN64(len);

    return offset;
}

/* end of dsgmt.c */
