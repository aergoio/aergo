/**
 * @file    gen_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "gen_util.h"

#include "gen_exp.h"

static BinaryenExpressionRef
exp_gen_id_ref(gen_t *gen, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;

    ASSERT1(is_global_id(id), id->scope);
    ASSERT1(id->offset == 0, id->offset);

    return BinaryenLoad(gen->module, meta_size(&id->meta), is_signed_type(&id->meta),
                        0, 0, meta_gen(gen, &id->meta), gen_i32(gen, id->addr));
}

static BinaryenExpressionRef
exp_gen_local_ref(gen_t *gen, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;

    ASSERT1(is_local_id(id), id->scope);

    return BinaryenGetLocal(gen->module, id->idx, meta_gen(gen, &id->meta));
}

static BinaryenExpressionRef
exp_gen_stack_ref(gen_t *gen, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;

    ASSERT1(is_stack_id(id), id->scope);

    return BinaryenLoad(gen->module, meta_size(&id->meta), is_signed_type(&id->meta),
                        id->offset, 0, meta_gen(gen, &id->meta), gen_i32(gen, id->addr));
}

static BinaryenExpressionRef
exp_gen_lit(gen_t *gen, ast_exp_t *exp)
{
    int addr;
    value_t *val = &exp->u_lit.val;
    meta_t *meta = &exp->meta;

    switch (val->type) {
    case TYPE_BOOL:
        return gen_i32(gen, val_bool(val) ? 1 : 0);

    case TYPE_UINT64:
        if (is_int64_type(meta) || is_uint64_type(meta))
            return gen_i64(gen, val_i64(val));

        return gen_i32(gen, val_i64(val));

    case TYPE_DOUBLE:
        if (is_double_type(meta))
            return gen_f64(gen, val_f64(val));

        return gen_f32(gen, val_f64(val));

    case TYPE_STRING:
        addr = dsgmt_add(gen->dsgmt, gen->module, val_ptr(val), val_size(val) + 1);
        return gen_i32(gen, addr);

    case TYPE_OBJECT:
        if (is_null_val(val))
            return gen_i32(gen, 0);

        addr = dsgmt_add(gen->dsgmt, gen->module, val_ptr(val), val_size(val));
        return gen_i32(gen, addr);

    default:
        ASSERT1(!"invalid value", meta->type);
    }

    return NULL;
}

#if 0
static BinaryenExpressionRef
exp_gen_array(gen_t *gen, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;
    meta_t *meta = &id->meta;
    BinaryenExpressionRef idx_exp;

        /* XXX
         * we need id address + offset
         * offset = meta_size() * index value */
    /*
    if (is_ref) {
        if (is_array_type(&id->meta)) {
            BinaryenExpressionRef offset_exp;

            idx_exp = exp_gen(gen, exp->u_arr.idx_exp, &exp->u_arr.idx_exp->meta, false);

            if (BinaryenExpressionGetId(idx_exp) == BinaryenConstId())
                return gen_i32(gen,
                    BinaryenConstGetValueI64(idx_exp) * ALIGN64(meta_size(meta)));

            offset_exp = gen_i32(gen, meta_size(meta));

            return BinaryenBinary(gen->module, BinaryenMulInt32(), idx_exp, offset_exp);
        }
        else {
            ERROR(ERROR_NOT_SUPPORTED, &exp->pos);
            return NULL;
        }
    }
    else {
        */
        if (is_array_type(meta)) {
            uint32_t offset = id->offset;
            BinaryenExpressionRef addr_exp;

            idx_exp = exp_gen(gen, exp->u_arr.idx_exp);
            addr_exp = gen_i32(gen, id->addr);

            if (BinaryenExpressionGetId(idx_exp) == BinaryenConstId())
                offset = BinaryenConstGetValueI64(idx_exp) * ALIGN64(meta_size(meta));

            return BinaryenLoad(gen->module, meta_size(meta), is_signed_type(meta),
                                offset, 0, meta_gen(gen, meta), addr_exp);
        }
        else {
            ERROR(ERROR_NOT_SUPPORTED, &exp->pos);
            return NULL;
        }
    //}
}
#endif

static BinaryenExpressionRef
exp_gen_cast(gen_t *gen, ast_exp_t *exp)
{
    meta_t *to_meta = &exp->u_cast.to_meta;
	/* XXX */
    BinaryenOp op;
    BinaryenExpressionRef val_exp;

    val_exp = exp_gen(gen, exp->u_cast.val_exp);

    switch (exp->meta.type) {
    case TYPE_INT8:
    case TYPE_INT16:
    case TYPE_INT32:
        if (is_float_type(to_meta))
            op = BinaryenConvertSInt32ToFloat32();
        else if (is_double_type(to_meta))
            op = BinaryenConvertSInt32ToFloat64();
        else if (is_int64_type(to_meta) || is_uint64_type(to_meta))
            op = BinaryenExtendSInt32();
        break;

    case TYPE_BOOL:
    case TYPE_BYTE:
    case TYPE_UINT8:
    case TYPE_UINT16:
    case TYPE_UINT32:
        if (is_float_type(to_meta))
            op = BinaryenConvertUInt32ToFloat32();
        else if (is_double_type(to_meta))
            op = BinaryenConvertUInt32ToFloat64();
        else if (is_int64_type(to_meta) || is_uint64_type(to_meta))
            op = BinaryenExtendUInt32();
        break;

    case TYPE_INT64:
        if (is_float_type(to_meta))
            op = BinaryenConvertSInt64ToFloat32();
        else if (is_double_type(to_meta))
            op = BinaryenConvertSInt64ToFloat64();
        else if (!is_int64_type(to_meta) && !is_uint64_type(to_meta))
            op = BinaryenWrapInt64();
        break;

    case TYPE_UINT64:
        if (is_float_type(to_meta))
            op = BinaryenConvertUInt64ToFloat32();
        else if (is_double_type(to_meta))
            op = BinaryenConvertUInt64ToFloat64();
        else if (!is_int64_type(to_meta) && !is_uint64_type(to_meta))
            op = BinaryenWrapInt64();
        break;

    case TYPE_FLOAT:
        if (is_int64_type(to_meta))
            op = BinaryenTruncSFloat32ToInt64();
        else if (is_uint64_type(to_meta))
            op = BinaryenTruncUFloat32ToInt64();
        else if (is_signed_type(to_meta))
            op = BinaryenTruncSFloat32ToInt32();
        else if (is_unsigned_type(to_meta))
            op = BinaryenTruncUFloat32ToInt32();
        else if (is_double_type(to_meta))
            op = BinaryenPromoteFloat32();
        break;

    case TYPE_DOUBLE:
        if (is_int64_type(to_meta))
            op = BinaryenTruncSFloat64ToInt64();
        else if (is_uint64_type(to_meta))
            op = BinaryenTruncUFloat64ToInt64();
        else if (is_signed_type(to_meta))
            op = BinaryenTruncSFloat64ToInt32();
        else if (is_unsigned_type(to_meta))
            op = BinaryenTruncUFloat64ToInt32();
        else if (is_float_type(to_meta))
            op = BinaryenPromoteFloat32();
        break;

    case TYPE_STRING:
        break;

    default:
        ASSERT2(!"invalid conversion", exp->meta.type, to_meta->type);
    }

    return BinaryenUnary(gen->module, op, val_exp);
}

static BinaryenExpressionRef
exp_gen_unary(gen_t *gen, ast_exp_t *exp)
{
    meta_t *meta = &exp->meta;
    BinaryenExpressionRef value;

    value = exp_gen(gen, exp->u_un.val_exp);

    switch (exp->u_un.kind) {
    case OP_NEG:
        if (is_int64_type(meta) || is_uint64_type(meta))
            return BinaryenBinary(gen->module, BinaryenSubInt64(), gen_i64(gen, 0),
                                  value);
        else if (is_float_type(meta))
            return BinaryenUnary(gen->module, BinaryenNegFloat32(), value);
        else if (is_double_type(meta))
            return BinaryenUnary(gen->module, BinaryenNegFloat64(), value);
        else
            return BinaryenBinary(gen->module, BinaryenSubInt32(), gen_i32(gen, 0),
                                  value);

    case OP_NOT:
        return BinaryenUnary(gen->module, BinaryenEqZInt32(), value);

    default:
        ASSERT1(!"invalid operator", exp->u_un.kind);
    }

    return NULL;
}

static BinaryenExpressionRef
exp_gen_binary(gen_t *gen, ast_exp_t *exp)
{
    meta_t *meta = &exp->meta;
    BinaryenOp op;
    BinaryenExpressionRef l_exp, r_exp;

    l_exp = exp_gen(gen, exp->u_bin.l_exp);
    r_exp = exp_gen(gen, exp->u_bin.r_exp);

    switch (exp->u_bin.kind) {
    case OP_ADD:
        if (is_string_type(meta)) {
            ERROR(ERROR_NOT_SUPPORTED, &exp->pos);
            return NULL;
        }

        if (is_int64_type(meta) || is_uint64_type(meta))
            op = BinaryenAddInt64();
        else if (is_float_type(meta))
            op = BinaryenAddFloat32();
        else if (is_double_type(meta))
            op = BinaryenAddFloat64();
        else
            op = BinaryenAddInt32();
        break;

    case OP_SUB:
        if (is_int64_type(meta) || is_uint64_type(meta))
            op = BinaryenSubInt64();
        else if (is_float_type(meta))
            op = BinaryenSubFloat32();
        else if (is_double_type(meta))
            op = BinaryenSubFloat64();
        else
            op = BinaryenSubInt32();
        break;

    case OP_MUL:
        if (is_int64_type(meta) || is_uint64_type(meta))
            op = BinaryenMulInt64();
        else if (is_float_type(meta))
            op = BinaryenMulFloat32();
        else if (is_double_type(meta))
            op = BinaryenMulFloat64();
        else
            op = BinaryenMulInt32();
        break;

    case OP_DIV:
        if (is_int64_type(meta))
            op = BinaryenDivSInt64();
        else if (is_uint64_type(meta))
            op = BinaryenDivUInt64();
        else if (is_float_type(meta))
            op = BinaryenDivFloat32();
        else if (is_double_type(meta))
            op = BinaryenDivFloat64();
        else if (is_signed_type(meta))
            op = BinaryenDivSInt32();
        else
            op = BinaryenDivUInt32();
        break;

    case OP_MOD:
        if (is_int64_type(meta))
            op = BinaryenRemSInt64();
        else if (is_uint64_type(meta))
            op = BinaryenRemUInt64();
        else if (is_signed_type(meta))
            op = BinaryenRemSInt32();
        else
            op = BinaryenRemUInt32();
        break;

    case OP_BIT_AND:
        if (is_int64_type(meta) || is_uint64_type(meta))
            op = BinaryenAndInt64();
        else
            op = BinaryenAndInt32();
        break;

    case OP_BIT_OR:
        if (is_int64_type(meta) || is_uint64_type(meta))
            op = BinaryenOrInt64();
        else
            op = BinaryenOrInt32();
        break;

    case OP_BIT_XOR:
        if (is_int64_type(meta) || is_uint64_type(meta))
            op = BinaryenXorInt64();
        else
            op = BinaryenXorInt32();
        break;

    case OP_RSHIFT:
        if (is_int64_type(meta))
            op = BinaryenShrSInt64();
        else if (is_uint64_type(meta))
            op = BinaryenShrUInt64();
        else if (is_signed_type(meta))
            op = BinaryenShrSInt32();
        else
            op = BinaryenShrUInt32();
        break;

    case OP_LSHIFT:
        if (is_int64_type(meta) || is_uint64_type(meta))
            op = BinaryenShlInt64();
        else
            op = BinaryenShlInt32();
        break;

    case OP_EQ:
        if (is_int64_type(meta) || is_uint64_type(meta))
            op = BinaryenEqInt64();
        else if (is_float_type(meta))
            op = BinaryenEqFloat32();
        else if (is_double_type(meta))
            op = BinaryenEqFloat64();
        else
            op = BinaryenEqInt32();
        break;

    case OP_NE:
        if (is_int64_type(meta) || is_uint64_type(meta))
            op = BinaryenNeInt64();
        else if (is_float_type(meta))
            op = BinaryenNeFloat32();
        else if (is_double_type(meta))
            op = BinaryenNeFloat64();
        else
            op = BinaryenNeInt32();
        break;

    case OP_LT:
        if (is_int64_type(meta))
            op = BinaryenLtSInt64();
        else if (is_uint64_type(meta))
            op = BinaryenLtUInt64();
        else if (is_float_type(meta))
            op = BinaryenLtFloat32();
        else if (is_double_type(meta))
            op = BinaryenLtFloat64();
        else if (is_signed_type(meta))
            op = BinaryenLtSInt32();
        else
            op = BinaryenLtUInt32();
        break;

    case OP_GT:
        if (is_int64_type(meta))
            op = BinaryenGtSInt64();
        else if (is_uint64_type(meta))
            op = BinaryenGtUInt64();
        else if (is_float_type(meta))
            op = BinaryenGtFloat32();
        else if (is_double_type(meta))
            op = BinaryenGtFloat64();
        else if (is_signed_type(meta))
            op = BinaryenGtSInt32();
        else
            op = BinaryenGtUInt32();
        break;

    case OP_LE:
        if (is_int64_type(meta))
            op = BinaryenLeSInt64();
        else if (is_uint64_type(meta))
            op = BinaryenLeUInt64();
        else if (is_float_type(meta))
            op = BinaryenLeFloat32();
        else if (is_double_type(meta))
            op = BinaryenLeFloat64();
        else if (is_signed_type(meta))
            op = BinaryenLeSInt32();
        else
            op = BinaryenLeUInt32();
        break;

    case OP_GE:
        if (is_int64_type(meta))
            op = BinaryenGeSInt64();
        else if (is_uint64_type(meta))
            op = BinaryenGeUInt64();
        else if (is_float_type(meta))
            op = BinaryenGeFloat32();
        else if (is_double_type(meta))
            op = BinaryenGeFloat64();
        else if (is_signed_type(meta))
            op = BinaryenGeSInt32();
        else
            op = BinaryenGeUInt32();
        break;

    case OP_AND:
        break;

    case OP_OR:
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_bin.kind);
    }

    return BinaryenBinary(gen->module, op, l_exp, r_exp);
}

static BinaryenExpressionRef
exp_gen_ternary(gen_t *gen, ast_exp_t *exp)
{
    return BinaryenSelect(gen->module, exp_gen(gen, exp->u_tern.pre_exp),
                          exp_gen(gen, exp->u_tern.in_exp),
                          exp_gen(gen, exp->u_tern.post_exp));
}

#if 0
static BinaryenExpressionRef
exp_gen_access(gen_t *gen, ast_exp_t *exp)
{
    ast_id_t *qual_id = exp->u_acc.id_exp->id;
    ast_id_t *fld_id = exp->id;
    meta_t *fld_meta = &exp->meta;

    if (is_fn_id(fld_id))
        /* nothing to do */
        return NULL;

    ASSERT1(is_var_id(fld_id), fld_id->kind);

    /* XXX:
     * we need qual_id address + offset
     * offset = fld_id's offset */
#if 0
    if (is_ref) {
        ASSERT1(fld_id->idx < 0, fld_id);

        /* TODO: We need variable's address not struct.
         * But, this is little bit weird... T_T */
        exp->id = qual_id;

        return gen_i32(gen, fld_id->offset);
    }
#endif

    return BinaryenLoad(gen->module, meta_size(fld_meta), is_signed_type(fld_meta),
                        fld_id->offset, 0, meta_gen(gen, fld_meta),
                        gen_i32(gen, qual_id->addr));
}
#endif

static BinaryenExpressionRef
exp_gen_call(gen_t *gen, ast_exp_t *exp)
{
    int i, j = 0;
    int arg_cnt;
    ast_id_t *fn_id = exp->id;
    array_t *ret_ids = fn_id->u_fn.ret_ids;
    array_t *param_exps = exp->u_call.param_exps;
    BinaryenExpressionRef *arg_exps;

    arg_cnt = array_size(param_exps) + array_size(ret_ids);
    arg_exps = xmalloc(sizeof(BinaryenExpressionRef) * arg_cnt);

    for (i = 0; i < array_size(param_exps); i++) {
        arg_exps[j++] = exp_gen(gen, array_get_exp(param_exps, i));
    }

    for (i = 0; i < array_size(ret_ids); i++) {
        arg_exps[j++] = gen_i32(gen, array_get_id(ret_ids, i)->addr);
    }

    return BinaryenCall(gen->module, fn_id->name, arg_exps, arg_cnt, BinaryenTypeNone());
}

static BinaryenExpressionRef
exp_gen_sql(gen_t *gen, ast_exp_t *exp)
{
    return BinaryenNop(gen->module);
}

static BinaryenExpressionRef
exp_gen_init(gen_t *gen, ast_exp_t *exp)
{
    /*
    int i;
    meta_t *meta = &exp->meta;
    array_t *elem_exps = exp->u_init.exps;
    BinaryenExpressionRef val_exp;
    */

    ASSERT(false);
    /*
    ASSERT1(is_array_type(meta) || is_struct_type(meta), meta->type);

    if (is_array_type(meta)) {
        int offset = 0;

        for (i = 0; i < array_size(elem_exps); i++) {
            ast_exp_t *elem_exp = array_get_exp(elem_exps, i);

            val_exp = exp_gen(gen, elem_exp);

            gen_add_instr(gen, BinaryenStore(gen->module, meta_size(meta), offset, 0,
                                             gen_i32(gen, meta->addr), val_exp,
                                             meta_gen(gen, meta)));

            offset += ALIGN64(meta_size(meta));
        }
    }
    else {
        ASSERT2(array_size(elem_exps) == meta->elem_cnt,
                array_size(elem_exps), meta->elem_cnt);

        for (i = 0; i < array_size(elem_exps); i++) {
            ast_exp_t *elem_exp = array_get_exp(elem_exps, i);
            //meta_t *elem_meta = meta->elems[i];

            val_exp = exp_gen(gen, elem_exp);

            gen_add_instr(gen, BinaryenStore(gen->module, meta_size(elem_meta),
                                             elem_meta->offset, 0,
                                             gen_i32(gen, meta->addr), val_exp,
                                             meta_gen(gen, elem_meta)));
        }
    }
    */

    return NULL;
}

BinaryenExpressionRef
exp_gen(gen_t *gen, ast_exp_t *exp)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_NULL:
        return NULL;

    case EXP_ID_REF:
        return exp_gen_id_ref(gen, exp);

    case EXP_LOCAL_REF:
        return exp_gen_local_ref(gen, exp);

    case EXP_STACK_REF:
        return exp_gen_stack_ref(gen, exp);

    case EXP_LIT:
        return exp_gen_lit(gen, exp);

        /*
    case EXP_ARRAY:
        return exp_gen_array(gen, exp);
        */

    case EXP_CAST:
        return exp_gen_cast(gen, exp);

    case EXP_UNARY:
        return exp_gen_unary(gen, exp);

    case EXP_BINARY:
        return exp_gen_binary(gen, exp);

    case EXP_TERNARY:
        return exp_gen_ternary(gen, exp);

        /*
    case EXP_ACCESS:
        return exp_gen_access(gen, exp);
        */

    case EXP_CALL:
        return exp_gen_call(gen, exp);

    case EXP_SQL:
        return exp_gen_sql(gen, exp);

    case EXP_INIT:
        return exp_gen_init(gen, exp);

    case EXP_TUPLE:
    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return NULL;
}

/* end of gen_exp.c */
