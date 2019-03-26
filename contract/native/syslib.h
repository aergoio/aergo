/**
 * @file    syslib.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SYSLIB_H
#define _SYSLIB_H

#include "common.h"

#include "trans.h"
#include "gen.h"

#define SYSLIB_MODULE               "system"

#define SYS_FN(kind)                (&sys_fntab_[(kind)])

#ifndef _IR_ABI_T
#define _IR_ABI_T
typedef struct ir_abi_s ir_abi_t;
#endif /* ! _IR_ABI_T */

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

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

void syslib_load(ast_t *ast);

ir_abi_t *syslib_abi(sys_fn_t *sys_fn);

ast_exp_t *syslib_new_malloc(trans_t *trans, uint32_t size, src_pos_t *pos);
ast_exp_t *syslib_new_memcpy(trans_t *trans, ast_exp_t *dest_exp, ast_exp_t *src_exp,
                              uint32_t size, src_pos_t *pos);

BinaryenExpressionRef syslib_call_1(gen_t *gen, fn_kind_t kind,
                                    BinaryenExpressionRef argument);

BinaryenExpressionRef syslib_call_2(gen_t *gen, fn_kind_t kind,
                                    BinaryenExpressionRef left,
                                    BinaryenExpressionRef right);

#endif /* ! _SYSLIB_H */
