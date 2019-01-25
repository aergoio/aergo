/**
 * @file    ir_abi.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IR_ABI_H
#define _IR_ABI_H

#include "common.h"

#include "ast_id.h"
#include "binaryen-c.h"

#ifndef _IR_ABI_T
#define _IR_ABI_T
typedef struct ir_abi_s ir_abi_t;
#endif /* ! _IR_ABI_T */

typedef struct ir_abi_s {
    char name[NAME_MAX_LEN + 1];

    /* parameter types (including return) */
    int param_cnt;
    BinaryenType *params;

    BinaryenType result;

    BinaryenFunctionTypeRef spec;
} ir_abi_t;

ir_abi_t *abi_lookup(vector_t *abis, ast_id_t *id);

#endif /* ! _IR_ABI_H */
