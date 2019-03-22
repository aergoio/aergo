/**
 * @file    syslib.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "binaryen-c.h"
#include "ast_id.h"
#include "ast_exp.h"
#include "parse.h"
#include "ir_abi.h"
#include "ir_md.h"
#include "gen_util.h"
#include "iobuf.h"

#include "syslib.h"

char *lib_src =
"library system {\n"
"    func malloc(uint32 size) uint32 : \"malloc\";\n"
"    func memcpy(uint32 dest, uint32 src, uint32 size) uint32 : \"memcpy\";\n"

"    func strcat(uint32 s1, uint32 s2) uint32 : \"strcat\";\n"
"    func strcmp(uint32 s1, uint32 s2) int32 : \"strcmp\";\n"

"    func to_int32(uint32 str) int32 : \"atoi32\";\n"
"    func to_int64(uint32 str) int64 : \"atoi64\";\n"

"    //func to_float(uint32 str) float : \"atof32\";\n"
"    //func to_double(uint32 str) double : \"atof64\";\n"

"    func to_char(uint32 i) uint32 : \"itoa32\";\n"
"    func to_char(uint64 i) uint32 : \"itoa64\";\n"
"    //func to_char(float f) uint32 : \"ftoa32\";\n"
"    //func to_char(double f) uint32 : \"ftoa64\";\n"

"    func abs(int64 i) int64 : \"abs_i64\";\n"
"    func abs(int32 i) int32 : \"abs_i32\";\n"
"    func abs(int16 i) int16 : \"abs_i16\";\n"
"    func abs(int8 i) int8 : \"abs_i8\";\n"
"    //func abs(float f) float : \"abs_f32\";\n"
"    //func abs(double f) double : \"abs_f64\";\n"
"}";

sys_fn_t sys_fntab_[FN_MAX] = {
    { "malloc", SYSLIB_MODULE".malloc", 1, { TYPE_UINT32 }, TYPE_UINT32 },
    { "memcpy", SYSLIB_MODULE".memcpy", 3,
        { TYPE_UINT32, TYPE_UINT32, TYPE_UINT32 }, TYPE_VOID },
    { "strcat", SYSLIB_MODULE".strcat", 2, { TYPE_UINT32, TYPE_UINT32 }, TYPE_UINT32 },
    { "strcmp", SYSLIB_MODULE".strcmp", 2, { TYPE_UINT32, TYPE_UINT32 }, TYPE_UINT32 },
    { "atoi32", SYSLIB_MODULE".atoi32", 1, { TYPE_UINT32 }, TYPE_UINT32 },
    { "atoi64", SYSLIB_MODULE".atoi64", 1, { TYPE_UINT32 }, TYPE_UINT64 },
    { "atof32", SYSLIB_MODULE".atof32", 1, { TYPE_UINT32 }, TYPE_FLOAT },
    { "atof64", SYSLIB_MODULE".atof64", 1, { TYPE_UINT32 }, TYPE_DOUBLE },
    { "itoa32", SYSLIB_MODULE".itoa32", 1, { TYPE_UINT32 }, TYPE_UINT32 },
    { "itoa64", SYSLIB_MODULE".itoa64", 1, { TYPE_UINT64 }, TYPE_UINT32 },
    { "ftoa32", SYSLIB_MODULE".ftoa32", 1, { TYPE_FLOAT }, TYPE_UINT32 },
    { "ftoa64", SYSLIB_MODULE".ftoa64", 1, { TYPE_DOUBLE }, TYPE_UINT32 },
};

void
syslib_load(ast_t *ast)
{
    flag_t flag = { FLAG_NONE, 0, 0 };
    iobuf_t src;

    iobuf_init(&src, "system_library");
    iobuf_set(&src, strlen(lib_src), lib_src);

    parse(&src, flag, ast);
}

ir_abi_t *
syslib_abi(fn_kind_t kind)
{
    int i;
    sys_fn_t *sys_fn;
    ir_abi_t *abi = xcalloc(sizeof(ir_abi_t));

    ASSERT1(kind >= 0 && kind < FN_MAX, kind);

    sys_fn = &sys_fntab_[kind];

    abi->module = SYSLIB_MODULE;
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
syslib_new_malloc(trans_t *trans, uint32_t size, src_pos_t *pos)
{
    ast_exp_t *res_exp;
    ast_exp_t *param_exp;
    vector_t *param_exps = vector_new();

    param_exp = exp_new_lit_int(size, pos);
    meta_set_uint32(&param_exp->meta);

    exp_add(param_exps, param_exp);

    res_exp = exp_new_call(false, NULL, param_exps, pos);

    res_exp->u_call.qname = sys_fntab_[FN_MALLOC].qname;
    meta_set_uint32(&res_exp->meta);

    md_add_imp(trans->md, syslib_abi(FN_MALLOC));

    return res_exp;
}

ast_exp_t *
syslib_new_memcpy(trans_t *trans, ast_exp_t *dest_exp, ast_exp_t *src_exp,
                   uint32_t size, src_pos_t *pos)
{
    ast_exp_t *res_exp;
    ast_exp_t *param_exp;
    vector_t *param_exps = vector_new();

    exp_add(param_exps, dest_exp);
    exp_add(param_exps, src_exp);

    param_exp = exp_new_lit_int(size, pos);
    meta_set_uint32(&param_exp->meta);

    exp_add(param_exps, param_exp);

    res_exp = exp_new_call(false, NULL, param_exps, pos);

    res_exp->u_call.qname = sys_fntab_[FN_MEMCPY].qname;
    meta_set_void(&res_exp->meta);

    md_add_imp(trans->md, syslib_abi(FN_MEMCPY));

    return res_exp;
}

/* end of syslib.c */
