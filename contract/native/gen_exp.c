/**
 * @file    gen_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "gen_meta.h"
#include "gen_util.h"

#include "gen_exp.h"

static BinaryenExpressionRef
exp_gen_id(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    ast_id_t *id = exp->id;

    ASSERT(id != NULL);

    if (is_ref) {
        if (id->idx >= 0)
            return gen_i32(gen, id->idx);

        return gen_i32(gen, id->offset);
    }
    else {
        if (id->idx >= 0)
            return BinaryenGetLocal(gen->module, id->idx, meta_gen(gen, meta));

        return BinaryenLoad(gen->module, meta_size(meta), is_signed_type(meta),
                            id->offset, 0, meta_gen(gen, meta),
                            gen_i32(gen, id->meta.addr));
    }
}

static BinaryenExpressionRef
exp_gen_val(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    int addr;
    value_t *val = &exp->u_lit.val;

    switch (meta->type) {
    case TYPE_BOOL:
        ASSERT1(val->type == TYPE_BOOL, val->type);
        return gen_i32(gen, val_bool(val) ? 1 : 0);

    case TYPE_BYTE:
    case TYPE_INT8:
    case TYPE_UINT8:
    case TYPE_INT16:
    case TYPE_UINT16:
    case TYPE_INT32:
    case TYPE_UINT32:
        ASSERT1(val->type == TYPE_UINT64, val->type);
        return gen_i32(gen, val_i64(val));

    case TYPE_INT64:
    case TYPE_UINT64:
        ASSERT1(val->type == TYPE_UINT64, val->type);
        return gen_i64(gen, val_i64(val));

    case TYPE_FLOAT:
        ASSERT1(val->type == TYPE_DOUBLE, val->type);
        return gen_f32(gen, val_f64(val));

    case TYPE_DOUBLE:
        ASSERT1(val->type == TYPE_DOUBLE, val->type);
        return gen_f64(gen, val_f64(val));

    case TYPE_STRING:
        ASSERT1(val->type == TYPE_STRING, val->type);
        addr = dsgmt_add(gen->dsgmt, val_ptr(val), val_size(val) + 1);
        return gen_i32(gen, addr);

    case TYPE_OBJECT:
    case TYPE_TUPLE:
        ASSERT1(val->type == TYPE_OBJECT, val->type);
        if (is_null_val(val)) {
            return gen_i32(gen, 0);
        }
        else {
            addr = dsgmt_add(gen->dsgmt, val_ptr(val), val_size(val));
            return gen_i32(gen, addr);
        }
        break;

    default:
        ASSERT1(!"invalid value", meta->type);
    }

    return NULL;
}

static BinaryenExpressionRef
exp_gen_array(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    ast_id_t *id = exp->id;
    BinaryenExpressionRef idx_exp;

    ASSERT(id != NULL);

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
        if (is_array_type(&id->meta)) {
            uint32_t offset = id->offset;
            BinaryenExpressionRef addr_exp;

            idx_exp = exp_gen(gen, exp->u_arr.idx_exp, &exp->u_arr.idx_exp->meta, false);
            addr_exp = gen_i32(gen, id->meta.addr);

            if (BinaryenExpressionGetId(idx_exp) == BinaryenConstId())
                offset = BinaryenConstGetValueI64(idx_exp) * ALIGN64(meta_size(meta));

            return BinaryenLoad(gen->module, meta_size(meta), is_signed_type(meta),
                                offset, 0, meta_gen(gen, meta), addr_exp);
        }
        else {
            ERROR(ERROR_NOT_SUPPORTED, &exp->pos);
            return NULL;
        }
    }
}

static BinaryenExpressionRef
exp_gen_cast(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    meta_t *to_meta = &exp->u_cast.to_meta;
    BinaryenOp op;
    BinaryenExpressionRef val_exp;

    val_exp = exp_gen(gen, exp->u_cast.val_exp, &exp->meta, false);

    switch (meta->type) {
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
        ASSERT2(!"invalid conversion", meta->type, exp->u_cast.to_meta.type);
    }

    return BinaryenUnary(gen->module, op, val_exp);
}

static BinaryenExpressionRef
exp_gen_unary(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    ast_id_t *id;
    BinaryenExpressionRef val_exp, ref_exp;
    BinaryenExpressionRef add_exp;

    val_exp = exp_gen(gen, exp->u_un.val_exp, meta, is_ref);

    switch (exp->u_un.kind) {
    case OP_INC:
    case OP_DEC:
        /* XXX: handle postfix */
        id = exp->u_un.val_exp->id;
        ASSERT(id != NULL);

        if (id->idx >= 0)
            ref_exp = BinaryenGetLocal(gen->module, id->idx, meta_gen(gen, &id->meta));
        else
            ref_exp = BinaryenLoad(gen->module, meta_size(&id->meta),
                                   is_signed_type(&id->meta),
                                   BinaryenConstGetValueI32(val_exp), 0,
                                   meta_gen(gen, &id->meta), gen_i32(gen, id->meta.addr));

        if (is_int64_type(meta) || is_uint64_type(meta)) {
            add_exp = BinaryenBinary(gen->module, BinaryenAddInt64(), ref_exp,
                                     gen_i64(gen, 1));
        }
        else {
            add_exp = BinaryenBinary(gen->module, BinaryenAddInt32(), ref_exp,
                                     gen_i32(gen, 1));
        }

        if (is_ref) {
            if (id->idx >= 0)
                return BinaryenSetLocal(gen->module, id->idx, add_exp);
            else
                return BinaryenStore(gen->module, meta_size(meta),
                                     BinaryenConstGetValueI32(val_exp), 0,
                                     gen_i32(gen, id->meta.addr), add_exp,
                                     meta_gen(gen, meta));
        }
        else {
            return add_exp;
        }

    case OP_NEG:
        if (is_int64_type(meta) || is_uint64_type(meta))
            return BinaryenBinary(gen->module, BinaryenSubInt64(), gen_i64(gen, 0),
                                  val_exp);
        else if (is_float_type(meta))
            return BinaryenUnary(gen->module, BinaryenNegFloat32(), val_exp);
        else if (is_double_type(meta))
            return BinaryenUnary(gen->module, BinaryenNegFloat64(), val_exp);
        else
            return BinaryenBinary(gen->module, BinaryenSubInt32(), gen_i32(gen, 0),
                                  val_exp);

    case OP_NOT:
        return BinaryenSelect(gen->module,
                              BinaryenUnary(gen->module, BinaryenEqZInt32(), val_exp),
                              gen_i32(gen, 1), gen_i32(gen, 0));

    default:
        ASSERT1(!"invalid operator", exp->u_un.kind);
    }

    return NULL;
}

static BinaryenExpressionRef
exp_gen_binary(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    BinaryenOp op;
    BinaryenExpressionRef l_exp, r_exp;

    l_exp = exp_gen(gen, exp->u_bin.l_exp, meta, is_ref);
    r_exp = exp_gen(gen, exp->u_bin.r_exp, meta, is_ref);

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
exp_gen_ternary(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    return BinaryenSelect(gen->module, exp_gen(gen, exp->u_tern.pre_exp, meta, is_ref),
                          exp_gen(gen, exp->u_tern.in_exp, meta, is_ref),
                          exp_gen(gen, exp->u_tern.post_exp, meta, is_ref));
}

static BinaryenExpressionRef
exp_gen_access(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    ast_id_t *qual_id = exp->u_acc.id_exp->id;
    ast_id_t *fld_id = exp->id;
    meta_t *fld_meta = &exp->meta;

    if (is_func_id(fld_id))
        return NULL;

    ASSERT1(is_var_id(fld_id), fld_id->kind);

    if (is_ref) {
        ASSERT1(fld_id->idx < 0, fld_id);

        /* TODO: We need variable's address not struct.
         * But, this is little bit weird... T_T */
        exp->id = qual_id;

        return gen_i32(gen, fld_id->offset);
    }

    return BinaryenLoad(gen->module, meta_size(fld_meta), is_signed_type(fld_meta),
                        fld_id->offset, 0, meta_gen(gen, fld_meta),
                        gen_i32(gen, qual_id->meta.addr));
}

static BinaryenExpressionRef
exp_gen_call(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    int i, j = 0;
    int arg_cnt;
    ast_id_t *func_id = exp->id;
    array_t *param_ids = func_id->u_func.param_ids;
    array_t *ret_ids = func_id->u_func.ret_ids;
    array_t *param_exps = exp->u_call.param_exps;
    BinaryenExpressionRef *arg_exps;

    arg_cnt = array_size(param_exps) + array_size(ret_ids);
    arg_exps = xmalloc(sizeof(BinaryenExpressionRef) * arg_cnt);

    for (i = 0; i < array_size(param_exps); i++) {
        ast_exp_t *param_exp = array_get(param_exps, i, ast_exp_t);
        ast_id_t *param_id = array_get(param_ids, i, ast_id_t);

        arg_exps[j++] = exp_gen(gen, param_exp, &param_id->meta, false);
    }

    for (i = 0; i < array_size(ret_ids); i++) {
        ast_id_t *ret_id = array_get(ret_ids, i, ast_id_t);

        arg_exps[j++] = gen_i32(gen, ret_id->meta.addr);
    }

    return BinaryenCall(gen->module, func_id->name, arg_exps, arg_cnt,
                        BinaryenTypeNone());
}

static BinaryenExpressionRef
exp_gen_sql(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    return BinaryenNop(gen->module);
}

static BinaryenExpressionRef
exp_gen_tuple(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    ASSERT(!"invalid tuple expression");
    return NULL;
}

static BinaryenExpressionRef
exp_gen_init(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    int i;
    array_t *elem_exps = exp->u_init.exps;
    BinaryenExpressionRef val_exp;

    ASSERT1(is_array_type(meta) || is_struct_type(meta), meta->type);

    if (is_array_type(meta)) {
        int offset = 0;

        for (i = 0; i < array_size(elem_exps); i++) {
            ast_exp_t *elem_exp = array_get(elem_exps, i, ast_exp_t);

            val_exp = exp_gen(gen, elem_exp, meta, false);

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
            ast_exp_t *elem_exp = array_get(elem_exps, i, ast_exp_t);
            meta_t *elem_meta = meta->elems[i];

            val_exp = exp_gen(gen, elem_exp, elem_meta, false);

            gen_add_instr(gen, BinaryenStore(gen->module, meta_size(elem_meta),
                                             elem_meta->offset, 0,
                                             gen_i32(gen, meta->addr), val_exp,
                                             meta_gen(gen, elem_meta)));
        }
    }

    return NULL;
}

BinaryenExpressionRef
exp_gen(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_NULL:
        return NULL;

    case EXP_REF:
        return exp_gen_id(gen, exp, meta, is_ref);

    case EXP_LIT:
        return exp_gen_val(gen, exp, meta, is_ref);

    case EXP_ARRAY:
        return exp_gen_array(gen, exp, meta, is_ref);

    case EXP_CAST:
        return exp_gen_cast(gen, exp, meta, is_ref);

    case EXP_UNARY:
        return exp_gen_unary(gen, exp, meta, is_ref);

    case EXP_BINARY:
        return exp_gen_binary(gen, exp, meta, is_ref);

    case EXP_TERNARY:
        return exp_gen_ternary(gen, exp, meta, is_ref);

    case EXP_ACCESS:
        return exp_gen_access(gen, exp, meta, is_ref);

    case EXP_CALL:
        return exp_gen_call(gen, exp, meta, is_ref);

    case EXP_SQL:
        return exp_gen_sql(gen, exp, meta, is_ref);

    case EXP_TUPLE:
        return exp_gen_tuple(gen, exp, meta, is_ref);

    case EXP_INIT:
        return exp_gen_init(gen, exp, meta, is_ref);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return NULL;
}

/* end of gen_exp.c */
