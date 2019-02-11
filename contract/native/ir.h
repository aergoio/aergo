/**
 * @file    ir.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IR_H
#define _IR_H

#include "common.h"

#include "vector.h"
#include "ir_sgmt.h"

#ifndef _IR_MD_T
#define _IR_MD_T
typedef struct ir_md_s ir_md_t;
#endif /* ! _IR_MD_T */

typedef struct ir_s {
    vector_t mds;
} ir_t;

ir_t *ir_new(void);

void ir_add_md(ir_t *ir, ir_md_t *md);

#endif /* no _IR_H */
