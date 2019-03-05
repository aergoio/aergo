/**
 * @file    syscall.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "binaryen-c.h"
#include "ast_exp.h"
#include "ir_abi.h"
#include "ir_md.h"

#include "syscall.h"

ir_abi_t *
syscall_abi(fn_kind_t kind)
{
    ir_abi_t *abi = xcalloc(sizeof(ir_abi_t));

    abi->module = SYSCALL_MODULE;

    switch (kind) {
    case FN_MALLOC:
        abi->name = FN_NAME(FN_MALLOC);
        abi->param_cnt = 1;
        abi->params = xmalloc(sizeof(BinaryenType));
        abi->params[0] = BinaryenTypeInt32();
        abi->result = BinaryenTypeInt32();
        break;

    case FN_MEMCPY:
        abi->name = FN_NAME(FN_MEMCPY);
        abi->param_cnt = 3;
        abi->params = xmalloc(sizeof(BinaryenType) * abi->param_cnt);
        abi->params[0] = BinaryenTypeInt32();
        abi->params[1] = BinaryenTypeInt32();
        abi->params[2] = BinaryenTypeInt32();
        abi->result = BinaryenTypeNone();
        break;

    default:
        ASSERT1(!"invalid function", kind);
    }

    return abi;
}

ast_exp_t *
syscall_new_malloc(trans_t *trans, uint32_t size, src_pos_t *pos)
{
    ast_exp_t *res_exp;
    ast_exp_t *param_exp;
    vector_t *param_exps = vector_new();

    param_exp = exp_new_lit_i64(size, pos);
    meta_set_uint32(&param_exp->meta);

    exp_add(param_exps, param_exp);

    res_exp = exp_new_call(false, NULL, param_exps, pos);

    res_exp->u_call.qname = syscall_qname(FN_MALLOC);
    meta_set_uint32(&res_exp->meta);

    md_add_imp(trans->md, syscall_abi(FN_MALLOC));

    return res_exp;
}

ast_exp_t *
syscall_new_memcpy(trans_t *trans, ast_exp_t *dest_exp, ast_exp_t *src_exp,
                   uint32_t size, src_pos_t *pos)
{
    ast_exp_t *res_exp;
    ast_exp_t *param_exp;
    vector_t *param_exps = vector_new();

    exp_add(param_exps, dest_exp);
    exp_add(param_exps, src_exp);

    param_exp = exp_new_lit_i64(size, pos);
    meta_set_uint32(&param_exp->meta);

    exp_add(param_exps, param_exp);

    res_exp = exp_new_call(false, NULL, param_exps, pos);

    res_exp->u_call.qname = syscall_qname(FN_MEMCPY);
    meta_set_void(&res_exp->meta);

    md_add_imp(trans->md, syscall_abi(FN_MEMCPY));

    return res_exp;
}

/* end of syscall.c */
