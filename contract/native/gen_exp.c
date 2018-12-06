/**
 * @file    gen_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "gen_meta.h"

#include "gen_exp.h"

static BinaryenExpressionRef
exp_gen_id(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    ast_id_t *id = exp->id;
    meta_t *meta = &exp->meta;

    ASSERT(id != NULL);

    if (is_ref) {
        if (is_primitive_type(meta) && !is_array_type(meta))
            return BinaryenConst(gen->module, BinaryenLiteralInt32(id->idx));

        return BinaryenConst(gen->module, BinaryenLiteralInt32(id->offset));
    }
    else {
        if (is_primitive_type(meta) && !is_array_type(meta))
            return BinaryenGetLocal(gen->module, id->idx, meta_gen(gen, meta));

        return BinaryenLoad(gen->module, meta_size(meta), is_int_family(meta),
                            id->offset, 0, meta_gen(gen, meta),
                            BinaryenConst(gen->module, BinaryenLiteralInt32(id->addr)));
    }
}

static BinaryenExpressionRef
exp_gen_val(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    int addr;
    meta_t *meta = &exp->meta;
    value_t *val = &exp->u_lit.val;
    struct BinaryenLiteral value;

    switch (meta->type) {
    case TYPE_BOOL:
        ASSERT1(is_bool_type(meta), meta->type);
        value = BinaryenLiteralInt32(val_bool(val));
        break;

    case TYPE_BYTE:
    case TYPE_INT8:
    case TYPE_UINT8:
    case TYPE_INT16:
    case TYPE_UINT16:
    case TYPE_INT32:
    case TYPE_UINT32:
        ASSERT1(is_int_family(meta), meta->type);
        value = BinaryenLiteralInt32(val_i64(val));
        break;

    case TYPE_INT64:
    case TYPE_UINT64:
        value = BinaryenLiteralInt64(val_i64(val));
        break;

    case TYPE_FLOAT:
        value = BinaryenLiteralFloat32(val_f64(val));
        break;

    case TYPE_DOUBLE:
        value = BinaryenLiteralFloat64(val_f64(val));
        break;

    case TYPE_STRING:
        addr = dsgmt_add(gen->dsgmt, val_ptr(val), val_size(val) + 1);
        value = BinaryenLiteralInt32(addr);
        break;

    case TYPE_OBJECT:
    case TYPE_TUPLE:
        if (is_null_val(val)) {
            value = BinaryenLiteralInt32(0);
        }
        else {
            addr = dsgmt_add(gen->dsgmt, val_ptr(val), val_size(val));
            value = BinaryenLiteralInt32(addr);
        }
        break;

    default:
        ASSERT1(!"invalid value", meta->type);
    }

    return BinaryenConst(gen->module, value);
}

static BinaryenExpressionRef
exp_gen_array(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    ast_id_t *id = exp->id;
    meta_t *meta = &exp->meta;
    BinaryenExpressionRef idx_exp;

    ASSERT(id != NULL);

    if (is_ref) {
        if (is_array_type(&id->meta)) {
            BinaryenExpressionRef offset_exp;

            idx_exp = exp_gen(gen, exp->u_arr.idx_exp, false);

            if (BinaryenExpressionGetId(idx_exp) == BinaryenConstId())
                return BinaryenConst(gen->module, BinaryenLiteralInt32(
                    BinaryenConstGetValueI64(idx_exp) * ALIGN64(meta_size(meta))));

            offset_exp =
                BinaryenConst(gen->module, BinaryenLiteralInt32(meta_size(meta)));

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

            idx_exp = exp_gen(gen, exp->u_arr.idx_exp, false);
            addr_exp = BinaryenConst(gen->module, BinaryenLiteralInt32(id->addr));

            if (BinaryenExpressionGetId(idx_exp) == BinaryenConstId())
                offset = BinaryenConstGetValueI64(idx_exp) * meta_size(meta);

            return BinaryenLoad(gen->module, meta_size(meta), is_int_family(meta),
                                offset, 0, meta_gen(gen, meta), addr_exp);
        }
        else {
            ERROR(ERROR_NOT_SUPPORTED, &exp->pos);
            return NULL;
        }
    }
}

static BinaryenExpressionRef
exp_gen_cast(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_unary(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_op_bit(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_op_cmp(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_op_bool_cmp(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_binary(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    meta_t *l_meta = &exp->u_bin.l_exp->meta;
    BinaryenOp op;
    BinaryenExpressionRef l_exp, r_exp;


    l_exp = exp_gen(gen, exp->u_bin.l_exp, is_ref);
    r_exp = exp_gen(gen, exp->u_bin.r_exp, is_ref);

    switch (exp->u_bin.kind) {
    case OP_ADD:
        if (is_int64_type(l_meta) || is_uint64_type(l_meta))
            op = BinaryenAddInt64();
        else
            op = BinaryenAddInt32();
        break;

    case OP_SUB:
        if (is_int64_type(l_meta) || is_uint64_type(l_meta))
            op = BinaryenSubInt64();
        else
            op = BinaryenSubInt32();
        break;

    case OP_MUL:
        if (is_int64_type(l_meta) || is_uint64_type(l_meta))
            op = BinaryenMulInt64();
        else
            op = BinaryenMulInt32();
        break;

    case OP_DIV:
        if (is_int64_type(l_meta))
            op = BinaryenDivSInt64();
        else if (is_uint64_type(l_meta))
            op = BinaryenDivUInt64();
        else
            op = BinaryenDivSInt32();
        break;

    case OP_MOD:
        if (is_int64_type(l_meta))
            op = BinaryenRemSInt64();
        else if (is_uint64_type(l_meta))
            op = BinaryenRemUInt64();
        else
            op = BinaryenRemSInt32();
        break;

    case OP_BIT_AND:
    case OP_BIT_OR:
    case OP_BIT_XOR:
    case OP_RSHIFT:
    case OP_LSHIFT:
        return exp_gen_op_bit(gen, exp, is_ref);

    case OP_EQ:
    case OP_NE:
    case OP_LT:
    case OP_GT:
    case OP_LE:
    case OP_GE:
        return exp_gen_op_cmp(gen, exp, is_ref);

    case OP_AND:
    case OP_OR:
        return exp_gen_op_bool_cmp(gen, exp, is_ref);

    default:
        ASSERT1(!"invalid operator", exp->u_bin.kind);
    }

    return BinaryenBinary(gen->module, op, l_exp, r_exp);
}

static BinaryenExpressionRef
exp_gen_ternary(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_access(gen_t *gen, ast_exp_t *exp, bool is_ref)
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
exp_gen_call(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    return BinaryenNop(gen->module);
}

static BinaryenExpressionRef
exp_gen_sql(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    return NULL;
}

static BinaryenExpressionRef
exp_gen_tuple(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    return NULL;
}

BinaryenExpressionRef
exp_gen(gen_t *gen, ast_exp_t *exp, bool is_ref)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_NULL:
        return NULL;

    case EXP_REF:
        return exp_gen_id(gen, exp, is_ref);

    case EXP_LIT:
        return exp_gen_val(gen, exp, is_ref);

    case EXP_ARRAY:
        return exp_gen_array(gen, exp, is_ref);

    case EXP_CAST:
        return exp_gen_cast(gen, exp, is_ref);

    case EXP_UNARY:
        return exp_gen_unary(gen, exp, is_ref);

    case EXP_BINARY:
        return exp_gen_binary(gen, exp, is_ref);

    case EXP_TERNARY:
        return exp_gen_ternary(gen, exp, is_ref);

    case EXP_ACCESS:
        return exp_gen_access(gen, exp, is_ref);

    case EXP_CALL:
        return exp_gen_call(gen, exp, is_ref);

    case EXP_SQL:
        return exp_gen_sql(gen, exp, is_ref);

    case EXP_TUPLE:
        return exp_gen_tuple(gen, exp, is_ref);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return NULL;
}

/* end of gen_exp.c */
