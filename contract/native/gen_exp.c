/**
 * @file    gen_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ir_abi.h"
#include "ir_md.h"
#include "gen_util.h"
#include "syslib.h"

#include "gen_exp.h"

static BinaryenExpressionRef
exp_gen_lit(gen_t *gen, ast_exp_t *exp)
{
    value_t *val = &exp->u_lit.val;
    meta_t *meta = &exp->meta;
    ir_md_t *md = gen->md;

    switch (val->type) {
    case TYPE_BOOL:
        return i32_gen(gen, val_bool(val) ? 1 : 0);

    case TYPE_BYTE:
        return i32_gen(gen, val_byte(val));

    case TYPE_INT128:
        if (is_int128_meta(meta)) {
            char *z_str;
            BinaryenExpressionRef argument;

            if (value_fits_i32(val))
                return syslib_gen(gen, FN_MPZ_SET_I32, 1, i32_gen(gen, val_i64(val)));

            if (value_fits_i64(val))
                return syslib_gen(gen, FN_MPZ_SET_I64, 1, i64_gen(gen, val_i64(val)));

            z_str = mpz_get_str(NULL, 10, val_mpz(val));
            ASSERT(z_str != NULL && z_str[0] != '\0');

            argument = i32_gen(gen, sgmt_add_str(&md->sgmt, z_str));

            return syslib_call_1(gen, FN_MPZ_SET_STR, argument);
        }

        if (is_int64_meta(meta))
            return i64_gen(gen, val_i64(val));

        return i32_gen(gen, val_i64(val));

    case TYPE_DOUBLE:
        if (is_double_meta(meta))
            return f64_gen(gen, val_f64(val));

        return f32_gen(gen, val_f64(val));

    case TYPE_OBJECT:
        return i32_gen(gen, sgmt_add_raw(&md->sgmt, val_ptr(val), val_size(val)));

    default:
        ASSERT2(!"invalid value", val->type, meta->type);
    }

    return NULL;
}

static BinaryenExpressionRef
exp_gen_array(gen_t *gen, ast_exp_t *exp)
{
    ast_exp_t *id_exp = exp->u_arr.id_exp;
    ast_exp_t *idx_exp = exp->u_arr.idx_exp;
    ast_id_t *id = exp->id;
    meta_t *meta = &exp->meta;
    BinaryenExpressionRef address;

    address = exp_gen(gen, id_exp);

    if (is_array_meta(&id->meta)) {
        uint32_t offset = 0;
        fn_kind_t kind = FN_ARR_GET_I32;

        ASSERT2(meta->arr_dim < meta->max_dim, meta->arr_dim, meta->max_dim);

        if (is_lit_exp(idx_exp)) {
            if (is_array_meta(meta)) {
                if (meta->dim_sizes[0] == -1) {
                    if (is_int64_meta(meta))
                        kind = FN_ARR_GET_I64;

                    address = syslib_gen(gen, kind, 3, address, i32_gen(gen, meta->arr_dim),
                                         exp_gen(gen, idx_exp));
                }
                else {
                    /* The total size of the subdimensions is required. */
                    offset = val_i64(&idx_exp->u_lit.val) * meta_memsz(meta) + meta_align(meta);
                }
            }
            else {
                offset = val_i64(&idx_exp->u_lit.val) * meta_regsz(meta) + meta_align(meta);
            }
        }
        else {
            if (is_int64_meta(meta))
                kind = FN_ARR_GET_I64;

            address = syslib_gen(gen, kind, 3, address, i32_gen(gen, meta->arr_dim),
                                 exp_gen(gen, idx_exp));
        }

        ASSERT1(offset % meta_align(meta) == 0, offset);

        if (gen->is_lval || is_array_meta(meta))
            /* Even if it is not an lvalue, it should return address when accessing the
             * intermediate element of a multi-dimensional array. */
            /* TODO: Need to change BinaryenBinary() to return structure */
            return BinaryenBinary(gen->module, BinaryenAddInt32(), address, i32_gen(gen, offset));

        return BinaryenLoad(gen->module, meta_iosz(meta), is_signed_meta(meta), offset, 0,
                            meta_gen(meta), address);
    }
    else if (is_string_meta(&id->meta)) {
        /* XXX */
        ASSERT(!gen->is_lval);

        return syslib_gen(gen, FN_CHAR_GET, 2, address, exp_gen(gen, idx_exp));
    }

    ERROR(ERROR_NOT_SUPPORTED, &exp->pos);

    return NULL;
}

static BinaryenExpressionRef
exp_gen_cast(gen_t *gen, ast_exp_t *exp)
{
    ast_exp_t *val_exp = exp->u_cast.val_exp;
    meta_t *from_meta = &val_exp->meta;
    meta_t *to_meta = &exp->meta;
    ir_md_t *md = gen->md;
    BinaryenOp op = 0;
    BinaryenExpressionRef value;

    value = exp_gen(gen, val_exp);

    switch (from_meta->type) {
    case TYPE_BOOL:
        ASSERT1(is_string_meta(to_meta), to_meta->type);
        return BinaryenSelect(gen->module, value, i32_gen(gen, sgmt_add_str(&md->sgmt, "true")),
                              i32_gen(gen, sgmt_add_str(&md->sgmt, "false")));

    case TYPE_BYTE:
        if (is_string_meta(to_meta))
            return syslib_gen(gen, FN_CTOA, 1, value);
        /* fall through */

    case TYPE_INT8:
    case TYPE_INT16:
    case TYPE_INT32:
        if (is_string_meta(to_meta))
            return syslib_gen(gen, FN_ITOA32, 1, value);

        if (is_int128_meta(to_meta))
            return syslib_gen(gen, FN_MPZ_SET_I32, 1, value);

        if (is_float_meta(to_meta))
            op = BinaryenConvertSInt32ToFloat32();
        else if (is_double_meta(to_meta))
            op = BinaryenConvertSInt32ToFloat64();
        else if (is_int64_meta(to_meta))
            op = BinaryenExtendSInt32();
        else
            return value;
        break;

    case TYPE_INT64:
        if (is_string_meta(to_meta))
            return syslib_gen(gen, FN_ITOA64, 1, value);

        if (is_int128_meta(to_meta))
            return syslib_gen(gen, FN_MPZ_SET_I64, 1, value);

        if (is_float_meta(to_meta))
            op = BinaryenConvertSInt64ToFloat32();
        else if (is_double_meta(to_meta))
            op = BinaryenConvertSInt64ToFloat64();
        else if (!is_int64_meta(to_meta))
            op = BinaryenWrapInt64();
        else
            return value;
        break;

    case TYPE_INT128:
        if (is_string_meta(to_meta))
            return syslib_gen(gen, FN_MPZ_GET_STR, 1, value);

        if (is_int64_meta(to_meta))
            return syslib_call_1(gen, FN_MPZ_GET_I64, value);

        return syslib_call_1(gen, FN_MPZ_GET_I32, value);

    case TYPE_FLOAT:
        if (is_int64_meta(to_meta))
            op = BinaryenTruncSFloat32ToInt64();
        else if (is_integer_meta(to_meta))
            op = BinaryenTruncSFloat32ToInt32();
        else if (is_double_meta(to_meta))
            op = BinaryenPromoteFloat32();
        else
            return value;
        break;

    case TYPE_DOUBLE:
        if (is_int64_meta(to_meta))
            op = BinaryenTruncSFloat64ToInt64();
        else if (is_integer_meta(to_meta))
            op = BinaryenTruncSFloat64ToInt32();
        else if (is_float_meta(to_meta))
            op = BinaryenDemoteFloat64();
        else
            return value;
        break;

    case TYPE_STRING:
        if (is_int64_meta(to_meta))
            return syslib_gen(gen, FN_ATOI64, 1, value);

        if (is_int128_meta(to_meta))
            return syslib_gen(gen, FN_MPZ_SET_STR, 1, value);

        return syslib_gen(gen, FN_ATOI32, 1, value);

    default:
        ASSERT2(!"invalid conversion", from_meta->type, to_meta->type);
    }

    return BinaryenUnary(gen->module, op, value);
}

static BinaryenExpressionRef
exp_gen_unary(gen_t *gen, ast_exp_t *exp)
{
    meta_t *meta = &exp->meta;
    BinaryenExpressionRef value;

    value = exp_gen(gen, exp->u_un.val_exp);

    switch (exp->u_un.kind) {
    case OP_NEG:
        if (is_int128_meta(meta))
            return syslib_call_1(gen, FN_MPZ_NEG, value);

        if (is_int64_meta(meta))
            return BinaryenBinary(gen->module, BinaryenSubInt64(), i64_gen(gen, 0), value);

        if (is_float_meta(meta))
            return BinaryenUnary(gen->module, BinaryenNegFloat32(), value);

        if (is_double_meta(meta))
            return BinaryenUnary(gen->module, BinaryenNegFloat64(), value);

        return BinaryenBinary(gen->module, BinaryenSubInt32(), i32_gen(gen, 0), value);

    case OP_NOT:
        return BinaryenUnary(gen->module, BinaryenEqZInt32(), value);

    default:
        ASSERT1(!"invalid operator", exp->u_un.kind);
    }

    return NULL;
}

static BinaryenExpressionRef
exp_gen_op_arith(gen_t *gen, ast_exp_t *exp, meta_t *meta)
{
    BinaryenOp op;
    BinaryenExpressionRef left, right;

    left = exp_gen(gen, exp->u_bin.l_exp);
    right = exp_gen(gen, exp->u_bin.r_exp);

    switch (exp->u_bin.kind) {
    case OP_ADD:
        if (is_string_meta(meta))
            return syslib_call_2(gen, FN_STRCAT, left, right);

        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_ADD, left, right);

        if (is_int64_meta(meta))
            op = BinaryenAddInt64();
        else if (is_float_meta(meta))
            op = BinaryenAddFloat32();
        else if (is_double_meta(meta))
            op = BinaryenAddFloat64();
        else
            op = BinaryenAddInt32();
        break;

    case OP_SUB:
        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_SUB, left, right);

        if (is_int64_meta(meta))
            op = BinaryenSubInt64();
        else if (is_float_meta(meta))
            op = BinaryenSubFloat32();
        else if (is_double_meta(meta))
            op = BinaryenSubFloat64();
        else
            op = BinaryenSubInt32();
        break;

    case OP_MUL:
        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_MUL, left, right);

        if (is_int64_meta(meta))
            op = BinaryenMulInt64();
        else if (is_float_meta(meta))
            op = BinaryenMulFloat32();
        else if (is_double_meta(meta))
            op = BinaryenMulFloat64();
        else
            op = BinaryenMulInt32();
        break;

    case OP_DIV:
        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_DIV, left, right);

        if (is_int64_meta(meta))
            op = BinaryenDivSInt64();
        else if (is_float_meta(meta))
            op = BinaryenDivFloat32();
        else if (is_double_meta(meta))
            op = BinaryenDivFloat64();
        else
            op = BinaryenDivSInt32();
        break;

    case OP_MOD:
        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_MOD, left, right);

        if (is_int64_meta(meta))
            op = BinaryenRemSInt64();
        else
            op = BinaryenRemSInt32();
        break;

    case OP_BIT_AND:
        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_AND, left, right);

        if (is_int64_meta(meta))
            op = BinaryenAndInt64();
        else
            op = BinaryenAndInt32();
        break;

    case OP_BIT_OR:
        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_OR, left, right);

        if (is_int64_meta(meta))
            op = BinaryenOrInt64();
        else
            op = BinaryenOrInt32();
        break;

    case OP_BIT_XOR:
        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_XOR, left, right);

        if (is_int64_meta(meta))
            op = BinaryenXorInt64();
        else
            op = BinaryenXorInt32();
        break;

    case OP_RSHIFT:
        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_RSHIFT, left, right);

        if (is_int64_meta(meta))
            op = BinaryenShrSInt64();
        else
            op = BinaryenShrSInt32();
        break;

    case OP_LSHIFT:
        if (is_int128_meta(meta))
            return syslib_call_2(gen, FN_MPZ_LSHIFT, left, right);

        if (is_int64_meta(meta))
            op = BinaryenShlInt64();
        else
            op = BinaryenShlInt32();
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_bin.kind);
    }

    return BinaryenBinary(gen->module, op, left, right);
}

static BinaryenExpressionRef
exp_gen_op_cmp(gen_t *gen, ast_exp_t *exp, meta_t *meta)
{
    BinaryenOp op;
    BinaryenExpressionRef left, right;

    left = exp_gen(gen, exp->u_bin.l_exp);
    right = exp_gen(gen, exp->u_bin.r_exp);

    switch (exp->u_bin.kind) {
    case OP_AND:
        ASSERT(!is_int64_meta(meta));
        op = BinaryenAndInt32();
        break;

    case OP_OR:
        ASSERT(!is_int64_meta(meta));
        op = BinaryenOrInt32();
        break;

    case OP_EQ:
        if (is_int128_meta(meta)) {
            left = syslib_call_2(gen, FN_MPZ_CMP, left, right);
            right = i32_gen(gen, 0);
        }

        if (is_int64_meta(meta))
            op = BinaryenEqInt64();
        else if (is_float_meta(meta))
            op = BinaryenEqFloat32();
        else if (is_double_meta(meta))
            op = BinaryenEqFloat64();
        else
            op = BinaryenEqInt32();
        break;

    case OP_NE:
        if (is_int128_meta(meta)) {
            left = syslib_call_2(gen, FN_MPZ_CMP, left, right);
            right = i32_gen(gen, 0);
        }

        if (is_int64_meta(meta))
            op = BinaryenNeInt64();
        else if (is_float_meta(meta))
            op = BinaryenNeFloat32();
        else if (is_double_meta(meta))
            op = BinaryenNeFloat64();
        else
            op = BinaryenNeInt32();
        break;

    case OP_LT:
        if (is_int128_meta(meta)) {
            left = syslib_call_2(gen, FN_MPZ_CMP, left, right);
            right = i32_gen(gen, 0);
        }

        if (is_int64_meta(meta))
            op = BinaryenLtSInt64();
        else if (is_float_meta(meta))
            op = BinaryenLtFloat32();
        else if (is_double_meta(meta))
            op = BinaryenLtFloat64();
        else
            op = BinaryenLtSInt32();
        break;

    case OP_GT:
        if (is_int128_meta(meta)) {
            left = syslib_call_2(gen, FN_MPZ_CMP, left, right);
            right = i32_gen(gen, 0);
        }

        if (is_int64_meta(meta))
            op = BinaryenGtSInt64();
        else if (is_float_meta(meta))
            op = BinaryenGtFloat32();
        else if (is_double_meta(meta))
            op = BinaryenGtFloat64();
        else
            op = BinaryenGtSInt32();
        break;

    case OP_LE:
        if (is_int128_meta(meta)) {
            left = syslib_call_2(gen, FN_MPZ_CMP, left, right);
            right = i32_gen(gen, 0);
        }

        if (is_int64_meta(meta))
            op = BinaryenLeSInt64();
        else if (is_float_meta(meta))
            op = BinaryenLeFloat32();
        else if (is_double_meta(meta))
            op = BinaryenLeFloat64();
        else
            op = BinaryenLeSInt32();
        break;

    case OP_GE:
        if (is_int128_meta(meta)) {
            left = syslib_call_2(gen, FN_MPZ_CMP, left, right);
            right = i32_gen(gen, 0);
        }

        if (is_int64_meta(meta))
            op = BinaryenGeSInt64();
        else if (is_float_meta(meta))
            op = BinaryenGeFloat32();
        else if (is_double_meta(meta))
            op = BinaryenGeFloat64();
        else
            op = BinaryenGeSInt32();
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_bin.kind);
    }

    if (is_string_meta(meta)) {
        left = syslib_call_2(gen, FN_STRCMP, left, right);
        right = i32_gen(gen, 0);
    }

    return BinaryenBinary(gen->module, op, left, right);
}

static BinaryenExpressionRef
exp_gen_binary(gen_t *gen, ast_exp_t *exp)
{
    switch (exp->u_bin.kind) {
    case OP_ADD:
    case OP_SUB:
    case OP_MUL:
    case OP_DIV:
    case OP_MOD:
    case OP_BIT_AND:
    case OP_BIT_OR:
    case OP_BIT_XOR:
    case OP_RSHIFT:
    case OP_LSHIFT:
        return exp_gen_op_arith(gen, exp, &exp->meta);

    case OP_AND:
    case OP_OR:
    case OP_EQ:
    case OP_NE:
    case OP_LT:
    case OP_GT:
    case OP_LE:
    case OP_GE:
        return exp_gen_op_cmp(gen, exp, &exp->u_bin.l_exp->meta);

    default:
        ASSERT1(!"invalid operator", exp->u_bin.kind);
    }

    return NULL;
}

static BinaryenExpressionRef
exp_gen_ternary(gen_t *gen, ast_exp_t *exp)
{
    return BinaryenSelect(gen->module, exp_gen(gen, exp->u_tern.pre_exp),
                          exp_gen(gen, exp->u_tern.in_exp), exp_gen(gen, exp->u_tern.post_exp));
}

static BinaryenExpressionRef
exp_gen_access(gen_t *gen, ast_exp_t *exp)
{
    meta_t *meta = &exp->meta;
    BinaryenExpressionRef address;

    address = exp_gen(gen, exp->u_acc.qual_exp);

    if (is_fn_id(exp->id))
        return address;

    if (gen->is_lval)
        return BinaryenBinary(gen->module, BinaryenAddInt32(), address,
                              i32_gen(gen, meta->rel_offset));

    return BinaryenLoad(gen->module, meta_iosz(meta), is_signed_meta(meta), meta->rel_offset, 0,
                        meta_gen(meta), address);
}

static BinaryenExpressionRef
exp_gen_call(gen_t *gen, ast_exp_t *exp)
{
    int i;
    BinaryenIndex arg_cnt;
    BinaryenExpressionRef *arguments;

    ASSERT(exp->u_call.qname != NULL);

    arg_cnt = vector_size(exp->u_call.arg_exps);
    arguments = xmalloc(sizeof(BinaryenExpressionRef) * arg_cnt);

    vector_foreach(exp->u_call.arg_exps, i) {
        arguments[i] = exp_gen(gen, vector_get_exp(exp->u_call.arg_exps, i));
    }

    return BinaryenCall(gen->module, exp->u_call.qname, arguments, arg_cnt, meta_gen(&exp->meta));
}

static BinaryenExpressionRef
exp_gen_sql(gen_t *gen, ast_exp_t *exp)
{
    /* TODO */
    return i32_gen(gen, 0);
}

static BinaryenExpressionRef
exp_gen_global(gen_t *gen, ast_exp_t *exp)
{
    ASSERT(exp->u_glob.name != NULL);

    return BinaryenGetGlobal(gen->module, exp->u_glob.name, BinaryenTypeInt32());
}

static BinaryenExpressionRef
exp_gen_reg(gen_t *gen, ast_exp_t *exp)
{
    return BinaryenGetLocal(gen->module, exp->meta.base_idx, meta_gen(&exp->meta));
}

static BinaryenExpressionRef
exp_gen_mem(gen_t *gen, ast_exp_t *exp)
{
    uint32_t offset = exp->meta.rel_addr + exp->meta.rel_offset;
    meta_t *meta = &exp->meta;
    BinaryenExpressionRef address;

    address = BinaryenGetLocal(gen->module, meta->base_idx, BinaryenTypeInt32());

    if (is_address_meta(meta)) {
        /* Since "address" is literally an address value, it must be of type I32. */
        ASSERT1(BinaryenExpressionGetType(address) == BinaryenTypeInt32(),
                BinaryenExpressionGetType(address));

        return BinaryenLoad(gen->module, ADDR_SIZE, 0, offset, 0, BinaryenTypeInt32(), address);
    }

    if (gen->is_lval)
        return BinaryenBinary(gen->module, BinaryenAddInt32(), address, i32_gen(gen, offset));

    return BinaryenLoad(gen->module, meta_iosz(meta), is_signed_meta(meta), offset, 0,
                        meta_gen(meta), address);
}

BinaryenExpressionRef
exp_gen(gen_t *gen, ast_exp_t *exp)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_LIT:
        return exp_gen_lit(gen, exp);

    case EXP_ARRAY:
        return exp_gen_array(gen, exp);

    case EXP_CAST:
        return exp_gen_cast(gen, exp);

    case EXP_UNARY:
        return exp_gen_unary(gen, exp);

    case EXP_BINARY:
        return exp_gen_binary(gen, exp);

    case EXP_TERNARY:
        return exp_gen_ternary(gen, exp);

    case EXP_ACCESS:
        return exp_gen_access(gen, exp);

    case EXP_CALL:
        return exp_gen_call(gen, exp);

    case EXP_SQL:
        return exp_gen_sql(gen, exp);

    case EXP_GLOBAL:
        return exp_gen_global(gen, exp);

    case EXP_REG:
        return exp_gen_reg(gen, exp);

    case EXP_MEM:
        return exp_gen_mem(gen, exp);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return NULL;
}

/* end of gen_exp.c */
