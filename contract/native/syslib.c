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
#undef sc_def
#define sc_def(src)     #src
#include "syslib.sc"
;

sys_fn_t sys_fntab_[FN_MAX] = {
    { "__udf", NULL, 0, { TYPE_NONE }, TYPE_NONE },
    { "__ctor", NULL, 0, { TYPE_NONE }, TYPE_NONE },
#undef fn_def
#define fn_def(kind, name, ...)                                                                    \
    { name, SYSLIB_MODULE"."name, __VA_ARGS__ },
#include "fn.list"
};

void
syslib_load(ast_t *ast)
{
    flag_t flag = { FLAG_NONE, NULL, 0, 0 };
    iobuf_t src;

    iobuf_init(&src, "system_library");
    iobuf_set(&src, strlen(lib_src), lib_src);

    parse(&src, flag, ast);
}

ir_abi_t *
syslib_abi(sys_fn_t *sys_fn)
{
    int i;
    ir_abi_t *abi = xcalloc(sizeof(ir_abi_t));

    ASSERT(sys_fn != NULL);

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

BinaryenExpressionRef
syslib_call(gen_t *gen, fn_kind_t kind, int argc, ...)
{
    int i;
    va_list vargs;
    sys_fn_t *sys_fn = SYS_FN(kind);
    BinaryenExpressionRef arguments[SYSLIB_MAX_ARGS];

    ASSERT1(argc <= SYSLIB_MAX_ARGS, argc);
    ASSERT3(sys_fn->param_cnt == argc, kind, sys_fn->param_cnt, argc);

    va_start(vargs, argc);

    for (i = 0; i < argc; i++) {
        arguments[i] = va_arg(vargs, BinaryenExpressionRef);
    }

    va_end(vargs);

    md_add_abi(gen->md, syslib_abi(sys_fn));

    return BinaryenCall(gen->module, sys_fn->qname, arguments, argc, type_gen(sys_fn->result));
}

void
syslib_gen(gen_t *gen, fn_kind_t kind)
{
    BinaryenType params[2] = { BinaryenTypeInt32(), BinaryenTypeInt32() };
    BinaryenType locals[2] = { BinaryenTypeInt32(), BinaryenTypeInt32() };
    BinaryenExpressionRef arguments[2];
    BinaryenExpressionRef children[4];
    BinaryenFunctionTypeRef type;
    sys_fn_t *sys_fn;

    arguments[0] = BinaryenGetGlobal(gen->module, "__STACK_TOP", BinaryenTypeInt32());
    arguments[1] = BinaryenGetLocal(gen->module, 1, BinaryenTypeInt32());

    sys_fn = SYS_FN(FN_ALIGN);
    md_add_abi(gen->md, syslib_abi(sys_fn));

    children[0] = BinaryenSetLocal(gen->module, 2,
        BinaryenCall(gen->module, sys_fn->qname, arguments, 2, BinaryenTypeInt32()));

    children[1] = BinaryenSetGlobal(gen->module, "__STACK_TOP",
        BinaryenBinary(gen->module, BinaryenAddInt32(),
                       BinaryenGetLocal(gen->module, 2, BinaryenTypeInt32()),
                       BinaryenGetLocal(gen->module, 0, BinaryenTypeInt32())));

    sys_fn = SYS_FN(FN_STACK_OVF);
    md_add_abi(gen->md, syslib_abi(sys_fn));

    children[2] = BinaryenIf(gen->module,
        BinaryenBinary(gen->module, BinaryenGeSInt32(),
                       BinaryenGetGlobal(gen->module, "__STACK_TOP", BinaryenTypeInt32()),
                       BinaryenGetGlobal(gen->module, "__STACK_MAX", BinaryenTypeInt32())),
        BinaryenCall(gen->module, sys_fn->qname, NULL, 0, BinaryenTypeNone()), NULL);

    children[3] =
        BinaryenReturn(gen->module, BinaryenGetLocal(gen->module, 2, BinaryenTypeInt32()));

    type = BinaryenAddFunctionType(gen->module, NULL, BinaryenTypeInt32(), params, 2);

    BinaryenAddFunction(gen->module, SYS_FN(kind)->qname, type, locals, 2,
                        BinaryenBlock(gen->module, NULL, children, 4, BinaryenTypeInt32()));
}

ast_exp_t *
syslib_make_alloc(fn_kind_t kind, uint32_t size, uint32_t align, src_pos_t *pos)
{
    ast_exp_t *exp = exp_new_call(kind, NULL, NULL, pos);
    ast_exp_t *arg_exp;

    exp->u_call.arg_exps = vector_new();

    arg_exp = exp_new_lit_int(size, pos);
    meta_set_int32(&arg_exp->meta);
    exp_add(exp->u_call.arg_exps, arg_exp);

    arg_exp = exp_new_lit_int(align, pos);
    meta_set_int32(&arg_exp->meta);
    exp_add(exp->u_call.arg_exps, arg_exp);

    meta_set_int32(&exp->meta);

    return exp;
}

ast_exp_t *
syslib_make_memcpy(ast_exp_t *dest_exp, ast_exp_t *src_exp, uint32_t size, src_pos_t *pos)
{
    ast_exp_t *exp = exp_new_call(FN_MEMCPY, NULL, NULL, pos);
    ast_exp_t *arg_exp;

    exp->u_call.arg_exps = vector_new();

    exp_add(exp->u_call.arg_exps, dest_exp);
    exp_add(exp->u_call.arg_exps, src_exp);

    arg_exp = exp_new_lit_int(size, pos);
    meta_set_int32(&arg_exp->meta);
    exp_add(exp->u_call.arg_exps, arg_exp);

    meta_set_void(&exp->meta);

    return exp;
}

ast_exp_t *
syslib_make_memset(ast_exp_t *addr_exp, uint32_t val, uint32_t size, src_pos_t *pos)
{
    ast_exp_t *exp = exp_new_call(FN_MEMSET, NULL, NULL, pos);
    ast_exp_t *arg_exp;

    exp->u_call.arg_exps = vector_new();

    exp_add(exp->u_call.arg_exps, addr_exp);

    arg_exp = exp_new_lit_int(val, pos);
    meta_set_int32(&arg_exp->meta);
    exp_add(exp->u_call.arg_exps, arg_exp);

    arg_exp = exp_new_lit_int(size, pos);
    meta_set_int32(&arg_exp->meta);
    exp_add(exp->u_call.arg_exps, arg_exp);

    meta_set_void(&exp->meta);

    return exp;
}

ast_exp_t *
syslib_make_strlen(ast_exp_t *addr_exp, src_pos_t *pos)
{
    ast_exp_t *exp = exp_new_call(FN_STRLEN, NULL, NULL, pos);

    exp->u_call.arg_exps = vector_new();

    exp_add(exp->u_call.arg_exps, addr_exp);

    meta_set_int32(&exp->meta);

    return exp;
}

ast_exp_t *
syslib_make_char_set(ast_exp_t *addr_exp, ast_exp_t *idx_exp, ast_exp_t *val_exp, src_pos_t *pos)
{
    ast_exp_t *exp = exp_new_call(FN_CHAR_SET, NULL, NULL, pos);

    exp->u_call.arg_exps = vector_new();

    exp_add(exp->u_call.arg_exps, addr_exp);
    exp_add(exp->u_call.arg_exps, idx_exp);
    exp_add(exp->u_call.arg_exps, val_exp);

    meta_set_void(&exp->meta);

    return exp;
}

ast_exp_t *
syslib_make_map_new(meta_t *key_meta, meta_t *val_meta, src_pos_t *pos)
{
    fn_kind_t kind;
    ast_exp_t *exp;

    if (is_int64_meta(key_meta) && is_int64_meta(val_meta))
        kind = FN_MAP_NEW_I64_I64;
    else if (is_int64_meta(key_meta))
        kind = FN_MAP_NEW_I64_I32;
    else if (is_int64_meta(val_meta))
        kind = FN_MAP_NEW_I32_I64;
    else
        kind = FN_MAP_NEW_I32_I32;

    exp = exp_new_call(kind, NULL, NULL, pos);
    meta_set(&exp->meta, TYPE_MAP);

    return exp;
}

ast_exp_t *
syslib_make_map_put(ast_exp_t *addr_exp, ast_exp_t *key_exp, ast_exp_t *val_exp, src_pos_t *pos)
{
    fn_kind_t kind;
    ast_exp_t *exp;
    vector_t *arg_exps = vector_new();

    exp_add(arg_exps, addr_exp);
    exp_add(arg_exps, key_exp);
    exp_add(arg_exps, val_exp);

    if (is_int64_meta(&key_exp->meta) && is_int64_meta(&val_exp->meta))
        kind = FN_MAP_PUT_I64_I64;
    else if (is_int64_meta(&key_exp->meta))
        kind = FN_MAP_PUT_I64_I32;
    else if (is_int64_meta(&val_exp->meta))
        kind = FN_MAP_PUT_I32_I64;
    else
        kind = FN_MAP_PUT_I32_I32;

    exp = exp_new_call(kind, NULL, arg_exps, pos);
    meta_set_void(&exp->meta);

    return exp;
}

/* end of syslib.c */
