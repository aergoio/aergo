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
dsgmt_add_raw(dsgmt_t *dsgmt, uint32_t len, char *raw)
{
    uint32_t offset = dsgmt->offset;

    if (dsgmt->size >= dsgmt->cap)
        dsgmt_extend(dsgmt);

    dsgmt->lens[dsgmt->size] = len;
    dsgmt->addrs[dsgmt->size] = offset;
    dsgmt->datas[dsgmt->size] = raw;

    dsgmt->size++;
    dsgmt->offset += ALIGN64(len);

    return offset;
}

int
dsgmt_add_string(dsgmt_t *dsgmt, char *str)
{
    int i;
    uint32_t len = strlen(str) + 1;  /* including '\0' */

    for (i = 0; i < dsgmt->size; i++) {
        if (dsgmt->lens[i] == len && strcmp(dsgmt->datas[i], str) == 0)
            return dsgmt->addrs[i];
    }

    return dsgmt_add_raw(dsgmt, strlen(str) + 1, str);
}

char *
dsgmt_raw(dsgmt_t *dsgmt, uint32_t addr)
{
    int i;

    for (i = 0; i < dsgmt->size; i++) {
        if (dsgmt->addrs[i] == addr)
            return dsgmt->datas[i];
    }

    ASSERT1(!"invalid address", addr);

    return NULL;
}

/* end of dsgmt.c */
