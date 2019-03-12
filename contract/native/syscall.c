/**
 * @file    syscall.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "binaryen-c.h"
#include "ast_id.h"
#include "ast_exp.h"
#include "ir_abi.h"
#include "ir_md.h"
#include "gen_util.h"

#include "syscall.h"

sys_fn_t sys_fntab_[FN_MAX] = {
    { "malloc", SYSCALL_MODULE".malloc", 1, { TYPE_UINT32 }, TYPE_UINT32 },
    { "memcpy", SYSCALL_MODULE".memcpy", 3,
        { TYPE_UINT32, TYPE_UINT32, TYPE_UINT32 }, TYPE_VOID },
    { "strcat", SYSCALL_MODULE".strcat", 2, { TYPE_UINT32, TYPE_UINT32 }, TYPE_UINT32 },
    { "strcmp", SYSCALL_MODULE".strcmp", 2, { TYPE_UINT32, TYPE_UINT32 }, TYPE_UINT32 },
    { "atoi32", SYSCALL_MODULE".atoi32", 1, { TYPE_UINT32 }, TYPE_UINT32 },
    { "atoi64", SYSCALL_MODULE".atoi64", 1, { TYPE_UINT32 }, TYPE_UINT64 },
    { "atof32", SYSCALL_MODULE".atof32", 1, { TYPE_UINT32 }, TYPE_FLOAT },
    { "atof64", SYSCALL_MODULE".atof64", 1, { TYPE_UINT32 }, TYPE_DOUBLE },
    { "itoa32", SYSCALL_MODULE".itoa32", 1, { TYPE_UINT32 }, TYPE_UINT32 },
    { "itoa64", SYSCALL_MODULE".itoa64", 1, { TYPE_UINT64 }, TYPE_UINT32 },
    { "ftoa32", SYSCALL_MODULE".ftoa32", 1, { TYPE_FLOAT }, TYPE_UINT32 },
    { "ftoa64", SYSCALL_MODULE".ftoa64", 1, { TYPE_DOUBLE }, TYPE_UINT32 },
};

ast_id_t *
syscall_load(void)
{
#if 0
    char *src = " \
library system { \
    func abs(int8 v) int8 : abs_i8; \
    func abs(int16 v) int16 : abs_i16; \
    func abs(int32 v) int32 : abs_i32; \
    func abs(int64 v) int64 : abs_i64; \
} \
";
#endif
    return NULL;
}

ir_abi_t *
syscall_abi(fn_kind_t kind)
{
    int i;
    sys_fn_t *sys_fn;
    ir_abi_t *abi = xcalloc(sizeof(ir_abi_t));

    ASSERT1(kind >= 0 && kind < FN_MAX, kind);

    sys_fn = &sys_fntab_[kind];

    abi->module = SYSCALL_MODULE;
    abi->name = sys_fn->name;

    abi->param_cnt = sys_fn->param_cnt;
    abi->params = xmalloc(sizeof(BinaryenType) * abi->param_cnt);

    for (i = 0; i < abi->param_cnt; i++) {
        abi->params[i] = type_gen(sys_fn->params[i]);
    }

    abi->result = type_gen(sys_fn->result);

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

    res_exp->u_call.qname = sys_fntab_[FN_MALLOC].qname;
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

    res_exp->u_call.qname = sys_fntab_[FN_MEMCPY].qname;
    meta_set_void(&res_exp->meta);

    md_add_imp(trans->md, syscall_abi(FN_MEMCPY));

    return res_exp;
}

/* end of syscall.c */
