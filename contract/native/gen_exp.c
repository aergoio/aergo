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
        if (is_primitive_type(meta) && !is_array_type(meta))
            return gen_i32(gen, id->idx);

        return gen_i32(gen, id->offset);
    }
    else {
        if (is_primitive_type(meta) && !is_array_type(meta))
            return BinaryenGetLocal(gen->module, id->idx, meta_gen(gen, meta));

        return BinaryenLoad(gen->module, meta_size(meta), is_signed_type(meta),
                            id->offset, 0, meta_gen(gen, meta), gen_i32(gen, id->addr));
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
    }
}

static BinaryenExpressionRef
exp_gen_cast(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    return NULL;
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
                                   meta_gen(gen, &id->meta), gen_i32(gen, id->addr));

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
                                     gen_i32(gen, id->addr), add_exp, 
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
    /*
    struct {
        int x;
        string y;
    };
    struct a;
    struct b;
        // a, b should has address of memory

    a.x = 1;
        if (a is struct) {
            value = BinaryenConst(module, BinaryenLiteralInt32(1));
            BinaryenStore(module, size, offset of x, align, base address of a, value, type);
        }
        else if (a is contract) {
        }

    a.x = p1;
        value = BinaryenGetLocal(module, index of p1);
        BinaryenStore(module, size, offset of x, align, base address of a, value, type);

    a.y = "abc";
        value = address of "abc";
        BinaryenStore(module, size, offset of y, align, base address of a, value, type);

    a.y = b.y;
        value = BinaryenLoad(module, size, signed, offset of y, align, type, base address of b);
        BinaryenStore(module, size, offset of x, align, base address of a, value, type);
    */
    return NULL;
}

static BinaryenExpressionRef
exp_gen_call(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    return BinaryenNop(gen->module);
}

static BinaryenExpressionRef
exp_gen_sql(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_tuple(gen_t *gen, ast_exp_t *exp, meta_t *meta, bool is_ref)
{
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

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return NULL;
}

/* end of gen_exp.c */
