/**
 * @file    ir.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IR_H
#define _IR_H

#include "common.h"

#include "array.h"
#include "ast_id.h"
#include "binaryen-c.h"

typedef struct ir_s {
    array_t globals;
    array_t fns;
} ir_t;

ir_t *ir_new(void);

void ir_add_global(ir_t *ir, ast_id_t *id);

#endif /* no _IR_H */
