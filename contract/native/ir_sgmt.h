/**
 * @file    ir_sgmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IR_SGMT_H
#define _IR_SGMT_H

#include "common.h"
#include "binaryen-c.h"

#define SGMT_INIT_CAPACITY          10

#ifndef _IR_SGMT_T
#define _IR_SGMT_T
typedef struct ir_sgmt_s ir_sgmt_t;
#endif /* ! _IR_SGMT_T */

/* Save data elements as an array to eliminate duplicate values. */
struct ir_sgmt_s {
    int cap;
    int size;
    BinaryenIndex offset;

    BinaryenIndex *lens;
    BinaryenIndex *addrs;
    char **datas;
};

int sgmt_add_str(ir_sgmt_t *sgmt, char *str);
int sgmt_add_raw(ir_sgmt_t *sgmt, void *ptr, int len);

static inline void
sgmt_init(ir_sgmt_t *sgmt)
{
    sgmt->cap = SGMT_INIT_CAPACITY;
    sgmt->size = 0;
    /* Reserve the first 4 bytes for a null value. */
    sgmt->offset = 4;

    sgmt->lens = xmalloc(sizeof(BinaryenIndex) * sgmt->cap);
    sgmt->addrs = xmalloc(sizeof(BinaryenIndex) * sgmt->cap);
    sgmt->datas = xmalloc(sizeof(char *) * sgmt->cap);
}

static inline ir_sgmt_t *
sgmt_new(void)
{
    ir_sgmt_t *sgmt = xcalloc(sizeof(ir_sgmt_t));

    sgmt_init(sgmt);

    return sgmt;
}

#endif /* ! _IR_SGMT_H */
