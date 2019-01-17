/**
 * @file    ir_sgmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_sgmt.h"

char *null = "\0\0\0\0\0\0\0\0";

static void
sgmt_extend(ir_sgmt_t *sgmt)
{
    sgmt->cap += SGMT_INIT_CAPACITY;

    sgmt->lens = xrealloc(sgmt->lens, sizeof(uint32_t) * sgmt->cap);
    sgmt->addrs = xrealloc(sgmt->addrs, sizeof(uint32_t) * sgmt->cap);
    sgmt->datas = xrealloc(sgmt->datas, sizeof(char *) * sgmt->cap);
}

/*
int
sgmt_add_global(ir_sgmt_t *sgmt, type_t type)
{
    int len = TYPE_SIZE(type);
    uint32_t addr;

    if (sgmt->size >= sgmt->cap)
        sgmt_extend(sgmt);

    sgmt->offset = ALIGN(sgmt->offset, TYPE_BYTE(type));
    addr = sgmt->offset;

    sgmt->lens[sgmt->size] = len;
    sgmt->addrs[sgmt->size] = addr;
    sgmt->datas[sgmt->size] = null;

    sgmt->size++;
    sgmt->offset += len;

    return addr;
}
*/

static int
sgmt_lookup(ir_sgmt_t *sgmt, void *ptr, uint32_t len)
{
    int i;

    for (i = 0; i < sgmt->size; i++) {
        if (sgmt->lens[i] == len && memcmp(sgmt->datas[i], ptr, len) == 0)
            return sgmt->addrs[i];
    }

    return -1;
}

int
sgmt_add_raw(ir_sgmt_t *sgmt, void *ptr, uint32_t len)
{
    int addr;

    ASSERT(ptr != NULL);
    ASSERT(len > 0);

    addr = sgmt_lookup(sgmt, ptr, len);
    if (addr >= 0)
        return addr;

    if (sgmt->size >= sgmt->cap)
        sgmt_extend(sgmt);

    /* TODO: Apply proper alignment */
    if (len > 4)
        sgmt->offset = ALIGN64(sgmt->offset);
    else
        sgmt->offset = ALIGN32(sgmt->offset);

    addr = sgmt->offset;

    sgmt->lens[sgmt->size] = len;
    sgmt->addrs[sgmt->size] = addr;
    sgmt->datas[sgmt->size] = ptr;

    sgmt->size++;
    sgmt->offset += len;

    return addr;
}

/* end of ir_sgmt.c */
