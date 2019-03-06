/**
 * @file    syscall.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SYSCALL_H
#define _SYSCALL_H

#include "common.h"

#include "trans.h"

#define SYSCALL_MODULE              "system"

#define SYS_FN(kind)                (&sys_fntab_[(kind)])

#ifndef _IR_ABI_T
#define _IR_ABI_T
typedef struct ir_abi_s ir_abi_t;
#endif /* ! _IR_ABI_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

typedef struct sys_fn_s {
    char *name;
    char *qname;

    int param_cnt;
    type_t params[4];

    type_t result;
} sys_fn_t;

extern sys_fn_t sys_fntab_[FN_MAX];

ir_abi_t *syscall_abi(fn_kind_t kind);

ast_exp_t *syscall_new_malloc(trans_t *trans, uint32_t size, src_pos_t *pos);
ast_exp_t *syscall_new_memcpy(trans_t *trans, ast_exp_t *dest_exp, ast_exp_t *src_exp,
                              uint32_t size, src_pos_t *pos);

#endif /* ! _SYSCALL_H */
