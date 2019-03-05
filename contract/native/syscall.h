/**
 * @file    syscall.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SYSCALL_H
#define _SYSCALL_H

#include "common.h"

#include "trans.h"

#define SYSCALL_MODULE              "system"

#ifndef _IR_ABI_T
#define _IR_ABI_T
typedef struct ir_abi_s ir_abi_t;
#endif /* ! _IR_ABI_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

ir_abi_t *syscall_abi(fn_kind_t kind);

ast_exp_t *syscall_new_malloc(trans_t *trans, uint32_t size, src_pos_t *pos);
ast_exp_t *syscall_new_memcpy(trans_t *trans, ast_exp_t *dest_exp, ast_exp_t *src_exp,
                              uint32_t size, src_pos_t *pos);

static inline char *
syscall_qname(fn_kind_t kind)
{
    switch (kind) {
    case FN_MALLOC:
        return SYSCALL_MODULE".malloc";

    case FN_MEMCPY:
        return SYSCALL_MODULE".memcpy";

    default:
        ASSERT1(!"invalid function", kind);
    }

    return NULL;
}

#endif /* ! _SYSCALL_H */
