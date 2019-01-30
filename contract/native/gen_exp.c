/**
 * @file    gen_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ir_abi.h"
#include "gen_util.h"

#include "gen_exp.h"

static BinaryenExpressionRef
exp_gen_lit(gen_t *gen, ast_exp_t *exp)
{
    value_t *val = &exp->u_lit.val;
    meta_t *meta = &exp->meta;

    switch (val->type) {
    case TYPE_BOOL:
        return i32_gen(gen, val_bool(val) ? 1 : 0);

    case TYPE_UINT64:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            return i64_gen(gen, val_i64(val));

        return i32_gen(gen, val_i64(val));

    case TYPE_DOUBLE:
        if (is_double_meta(meta))
            return f64_gen(gen, val_f64(val));

        return f32_gen(gen, val_f64(val));

    case TYPE_OBJECT:
        return i32_gen(gen, sgmt_add_raw(&gen->ir->sgmt, val_ptr(val), val_size(val)));

    default:
        ASSERT2(!"invalid value", val->type, meta->type);
    }

    return NULL;
}

static BinaryenExpressionRef
exp_gen_array(gen_t *gen, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;
    meta_t *meta = &exp->meta;

    /* This function is used when the offset value needs to be computed dynamically */

    if (is_array_meta(&id->meta)) {
        ast_exp_t *id_exp = exp->u_arr.id_exp;
        ast_exp_t *idx_exp = exp->u_arr.idx_exp;
        BinaryenExpressionRef base, address, offset, index;

        ASSERT(!is_lit_exp(idx_exp));

        /* Because BinaryenLoad() takes an offset as uint32_t and "idx_exp" is a
         * register or memory expression, we do not know the offset value, so we add
         * the offset to the address and use BinaryenLoad(). See exp_trans_array() for
         * the following formula. */

        base = exp_gen(gen, id_exp);
        index = exp_gen(gen, idx_exp);

        if (is_int64_meta(&idx_exp->meta) || is_uint64_meta(&idx_exp->meta))
            /* TODO: need to check range of index in semantic checker */
            index = BinaryenUnary(gen->module, BinaryenWrapInt64(), index);

        offset = BinaryenBinary(gen->module, BinaryenMulInt32(), index,
                                i32_gen(gen, meta_bytes(meta)));

        if (id->meta.rel_addr > 0)
            offset = BinaryenBinary(gen->module, BinaryenAddInt32(), offset,
                                    i32_gen(gen, id->meta.rel_addr +
                                            meta_align(&id->meta)));
        else
            offset = BinaryenBinary(gen->module, BinaryenAddInt32(), offset,
                                    i32_gen(gen, meta_align(&id->meta)));

        address = BinaryenBinary(gen->module, BinaryenAddInt32(), base, offset);

        /* XXX: change is_array_meta() to assertion */
        if (gen->is_lval || is_array_meta(meta))
            return address;

        return BinaryenLoad(gen->module, TYPE_BYTE(meta->type), is_signed_meta(meta),
                            0, 0, meta_gen(meta), address);
    }

    ERROR(ERROR_NOT_SUPPORTED, &exp->pos);

    return NULL;
}

static BinaryenExpressionRef
exp_gen_cast(gen_t *gen, ast_exp_t *exp)
{
    ast_exp_t *val_exp = exp->u_cast.val_exp;
    meta_t *to_meta = &exp->meta;
    BinaryenOp op = 0;
    BinaryenExpressionRef value;

    value = exp_gen(gen, val_exp);

    switch (val_exp->meta.type) {
    case TYPE_INT8:
    case TYPE_INT16:
    case TYPE_INT32:
        if (is_float_meta(to_meta))
            op = BinaryenConvertSInt32ToFloat32();
        else if (is_double_meta(to_meta))
            op = BinaryenConvertSInt32ToFloat64();
        else if (is_int64_meta(to_meta) || is_uint64_meta(to_meta))
            op = BinaryenExtendSInt32();
        else if (is_string_meta(to_meta))
            /* TOOD */
            return NULL;
        else
            return value;
        break;

    case TYPE_BOOL:
    case TYPE_BYTE:
    case TYPE_UINT8:
    case TYPE_UINT16:
    case TYPE_UINT32:
        if (is_float_meta(to_meta))
            op = BinaryenConvertUInt32ToFloat32();
        else if (is_double_meta(to_meta))
            op = BinaryenConvertUInt32ToFloat64();
        else if (is_int64_meta(to_meta) || is_uint64_meta(to_meta))
            op = BinaryenExtendUInt32();
        else if (is_string_meta(to_meta))
            /* TOOD */
            return NULL;
        else
            return value;
        break;

    case TYPE_INT64:
        if (is_float_meta(to_meta))
            op = BinaryenConvertSInt64ToFloat32();
        else if (is_double_meta(to_meta))
            op = BinaryenConvertSInt64ToFloat64();
        else if (!is_int64_meta(to_meta) && !is_uint64_meta(to_meta))
            op = BinaryenWrapInt64();
        else if (is_string_meta(to_meta))
            /* TOOD */
            return NULL;
        else
            return value;
        break;

    case TYPE_UINT64:
        if (is_float_meta(to_meta))
            op = BinaryenConvertUInt64ToFloat32();
        else if (is_double_meta(to_meta))
            op = BinaryenConvertUInt64ToFloat64();
        else if (!is_int64_meta(to_meta) && !is_uint64_meta(to_meta))
            op = BinaryenWrapInt64();
        else if (is_string_meta(to_meta))
            /* TOOD */
            return NULL;
        else
            return value;
        break;

    case TYPE_FLOAT:
        if (is_int64_meta(to_meta))
            op = BinaryenTruncSFloat32ToInt64();
        else if (is_uint64_meta(to_meta))
            op = BinaryenTruncUFloat32ToInt64();
        else if (is_signed_meta(to_meta))
            op = BinaryenTruncSFloat32ToInt32();
        else if (is_bool_meta(to_meta) || is_unsigned_meta(to_meta))
            op = BinaryenTruncUFloat32ToInt32();
        else if (is_double_meta(to_meta))
            op = BinaryenPromoteFloat32();
        else if (is_string_meta(to_meta))
            /* TOOD */
            return NULL;
        else
            return value;
        break;

    case TYPE_DOUBLE:
        if (is_int64_meta(to_meta))
            op = BinaryenTruncSFloat64ToInt64();
        else if (is_uint64_meta(to_meta))
            op = BinaryenTruncUFloat64ToInt64();
        else if (is_signed_meta(to_meta))
            op = BinaryenTruncSFloat64ToInt32();
        else if (is_bool_meta(to_meta) || is_unsigned_meta(to_meta))
            op = BinaryenTruncUFloat64ToInt32();
        else if (is_float_meta(to_meta))
            op = BinaryenDemoteFloat64();
        else if (is_string_meta(to_meta))
            /* TOOD */
            return NULL;
        else
            return value;
        break;

    case TYPE_STRING:
        /* TODO */
        return NULL;

    default:
        ASSERT2(!"invalid conversion", exp->meta.type, to_meta->type);
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
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            return BinaryenBinary(gen->module, BinaryenSubInt64(), i64_gen(gen, 0),
                                  value);
        else if (is_float_meta(meta))
            return BinaryenUnary(gen->module, BinaryenNegFloat32(), value);
        else if (is_double_meta(meta))
            return BinaryenUnary(gen->module, BinaryenNegFloat64(), value);
        else
            return BinaryenBinary(gen->module, BinaryenSubInt32(), i32_gen(gen, 0),
                                  value);

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
        if (is_string_meta(meta)) {
            /* XXX
            BinaryenExpressionRef args[2] = { l_exp, r_exp };

            return BinaryenCall(gen->module, xstrdup("concat$"), args, 2,
                                BinaryenTypeInt32());
                                */
            return i32_gen(gen, 0);
        }

        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenAddInt64();
        else if (is_float_meta(meta))
            op = BinaryenAddFloat32();
        else if (is_double_meta(meta))
            op = BinaryenAddFloat64();
        else
            op = BinaryenAddInt32();
        break;

    case OP_SUB:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenSubInt64();
        else if (is_float_meta(meta))
            op = BinaryenSubFloat32();
        else if (is_double_meta(meta))
            op = BinaryenSubFloat64();
        else
            op = BinaryenSubInt32();
        break;

    case OP_MUL:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenMulInt64();
        else if (is_float_meta(meta))
            op = BinaryenMulFloat32();
        else if (is_double_meta(meta))
            op = BinaryenMulFloat64();
        else
            op = BinaryenMulInt32();
        break;

    case OP_DIV:
        if (is_int64_meta(meta))
            op = BinaryenDivSInt64();
        else if (is_uint64_meta(meta))
            op = BinaryenDivUInt64();
        else if (is_float_meta(meta))
            op = BinaryenDivFloat32();
        else if (is_double_meta(meta))
            op = BinaryenDivFloat64();
        else if (is_signed_meta(meta))
            op = BinaryenDivSInt32();
        else
            op = BinaryenDivUInt32();
        break;

    case OP_MOD:
        if (is_int64_meta(meta))
            op = BinaryenRemSInt64();
        else if (is_uint64_meta(meta))
            op = BinaryenRemUInt64();
        else if (is_signed_meta(meta))
            op = BinaryenRemSInt32();
        else
            op = BinaryenRemUInt32();
        break;

    case OP_BIT_AND:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenAndInt64();
        else
            op = BinaryenAndInt32();
        break;

    case OP_BIT_OR:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenOrInt64();
        else
            op = BinaryenOrInt32();
        break;

    case OP_BIT_XOR:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenXorInt64();
        else
            op = BinaryenXorInt32();
        break;

    case OP_RSHIFT:
        if (is_int64_meta(meta))
            op = BinaryenShrSInt64();
        else if (is_uint64_meta(meta))
            op = BinaryenShrUInt64();
        else if (is_signed_meta(meta))
            op = BinaryenShrSInt32();
        else
            op = BinaryenShrUInt32();
        break;

    case OP_LSHIFT:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
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

    if (is_string_meta(meta)) {
        /* XXX */
        return i32_gen(gen, 0);
    }

    switch (exp->u_bin.kind) {
    case OP_AND:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenAndInt64();
        else
            op = BinaryenAndInt32();
        break;

    case OP_OR:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenOrInt64();
        else
            op = BinaryenOrInt32();
        break;

    case OP_EQ:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenEqInt64();
        else if (is_float_meta(meta))
            op = BinaryenEqFloat32();
        else if (is_double_meta(meta))
            op = BinaryenEqFloat64();
        else
            op = BinaryenEqInt32();
        break;

    case OP_NE:
        if (is_int64_meta(meta) || is_uint64_meta(meta))
            op = BinaryenNeInt64();
        else if (is_float_meta(meta))
            op = BinaryenNeFloat32();
        else if (is_double_meta(meta))
            op = BinaryenNeFloat64();
        else
            op = BinaryenNeInt32();
        break;

    case OP_LT:
        if (is_int64_meta(meta))
            op = BinaryenLtSInt64();
        else if (is_uint64_meta(meta))
            op = BinaryenLtUInt64();
        else if (is_float_meta(meta))
            op = BinaryenLtFloat32();
        else if (is_double_meta(meta))
            op = BinaryenLtFloat64();
        else if (is_signed_meta(meta))
            op = BinaryenLtSInt32();
        else
            op = BinaryenLtUInt32();
        break;

    case OP_GT:
        if (is_int64_meta(meta))
            op = BinaryenGtSInt64();
        else if (is_uint64_meta(meta))
            op = BinaryenGtUInt64();
        else if (is_float_meta(meta))
            op = BinaryenGtFloat32();
        else if (is_double_meta(meta))
            op = BinaryenGtFloat64();
        else if (is_signed_meta(meta))
            op = BinaryenGtSInt32();
        else
            op = BinaryenGtUInt32();
        break;

    case OP_LE:
        if (is_int64_meta(meta))
            op = BinaryenLeSInt64();
        else if (is_uint64_meta(meta))
            op = BinaryenLeUInt64();
        else if (is_float_meta(meta))
            op = BinaryenLeFloat32();
        else if (is_double_meta(meta))
            op = BinaryenLeFloat64();
        else if (is_signed_meta(meta))
            op = BinaryenLeSInt32();
        else
            op = BinaryenLeUInt32();
        break;

    case OP_GE:
        if (is_int64_meta(meta))
            op = BinaryenGeSInt64();
        else if (is_uint64_meta(meta))
            op = BinaryenGeUInt64();
        else if (is_float_meta(meta))
            op = BinaryenGeFloat32();
        else if (is_double_meta(meta))
            op = BinaryenGeFloat64();
        else if (is_signed_meta(meta))
            op = BinaryenGeSInt32();
        else
            op = BinaryenGeUInt32();
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_bin.kind);
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
                          exp_gen(gen, exp->u_tern.in_exp),
                          exp_gen(gen, exp->u_tern.post_exp));
}

static BinaryenExpressionRef
exp_gen_access(gen_t *gen, ast_exp_t *exp)
{
    ast_id_t *fld_id = exp->id;
    meta_t *meta = &exp->meta;
    ast_exp_t *qual_exp = exp->u_acc.qual_exp;
    BinaryenExpressionRef address;

    /* If qualifier is a function and returns an array or a struct, "qual_exp" can be
     * a binary expression. Otherwise all are register expressions */

    ASSERT1(is_register_exp(qual_exp) || is_binary_exp(qual_exp), qual_exp->kind);

    address = exp_gen(gen, qual_exp);

    if (is_fn_id(fld_id))
        return address;

    if (gen->is_lval)
        return BinaryenBinary(gen->module, BinaryenAddInt32(), address,
                              i32_gen(gen, fld_id->meta.rel_offset));

    return BinaryenLoad(gen->module, TYPE_BYTE(meta->type), is_signed_meta(meta),
                        fld_id->meta.rel_offset, 0, meta_gen(meta), address);
}

static BinaryenExpressionRef
exp_gen_call(gen_t *gen, ast_exp_t *exp)
{
    int i, j = 0;
    ast_id_t *id = exp->id;
    ir_abi_t *abi = id->abi;
    BinaryenExpressionRef index;
    BinaryenExpressionRef *arguments;

    if (is_map_meta(&exp->meta))
        /* TODO */
        return i32_gen(gen, 0);

    ASSERT(abi != NULL);
    ASSERT(id->idx >= 0);

    arguments = xmalloc(sizeof(BinaryenExpressionRef) * abi->param_cnt);

    vector_foreach(exp->u_call.param_exps, i) {
        arguments[j++] = exp_gen(gen, vector_get_exp(exp->u_call.param_exps, i));
    }

    if (is_ctor_id(id))
        /* The constructor is called with an absolute index */
        return BinaryenCallIndirect(gen->module, i32_gen(gen, id->idx), arguments,
                                    abi->param_cnt, abi->name);

    index = BinaryenBinary(gen->module, BinaryenAddInt32(),
                           BinaryenLoad(gen->module, sizeof(int32_t), 1, 0, 0,
                                        BinaryenTypeInt32(),
                                        exp_gen(gen, exp->u_call.id_exp)),
                           i32_gen(gen, id->idx));

    return BinaryenCallIndirect(gen->module, index, arguments, abi->param_cnt, abi->name);
}

static BinaryenExpressionRef
exp_gen_sql(gen_t *gen, ast_exp_t *exp)
{
    /* TODO */
    return i32_gen(gen, 0);
}

static BinaryenExpressionRef
exp_gen_init(gen_t *gen, ast_exp_t *exp)
{
    int i;
    uint32_t offset = 0;
    meta_t *meta = &exp->meta;
    vector_t *elem_exps = exp->u_init.elem_exps;
    BinaryenExpressionRef address, value;

    if (is_map_meta(meta)) {
        /* elem_exps is the vector of key-value pair */
        BinaryenExpressionRef args[2];

        vector_foreach(elem_exps, i) {
            ast_exp_t *elem_exp = vector_get_exp(elem_exps, i);

            ASSERT1(is_tuple_exp(elem_exp), elem_exp->kind);

            args[0] = exp_gen(gen, vector_get_exp(elem_exp->u_tup.elem_exps, 0));
            args[1] = exp_gen(gen, vector_get_exp(elem_exp->u_tup.elem_exps, 1));

            instr_add(gen, BinaryenCall(gen->module, xstrdup("map-put$"), args, 2,
                                        BinaryenTypeNone()));
        }

        return NULL;
    }

    /* The heap or stack memory is already allocated in exp_trans_init() */
    address = BinaryenGetLocal(gen->module, exp->meta.base_idx, BinaryenTypeInt32());

    if (meta->rel_addr > 0)
        address = BinaryenBinary(gen->module, BinaryenAddInt32(), address,
                                 i32_gen(gen, meta->rel_addr));

    if (is_array_meta(meta)) {
        instr_add(gen, BinaryenStore(gen->module, sizeof(uint32_t), offset, 0, address,
                                     i32_gen(gen, meta->dim_sizes[0]),
                                     BinaryenTypeInt32()));
        offset += meta_align(meta);
    }

    vector_foreach(elem_exps, i) {
        ast_exp_t *elem_exp = vector_get_exp(elem_exps, i);
        meta_t *elem_meta = &elem_exp->meta;

        offset = ALIGN(offset, meta_align(elem_meta));

        if (is_init_exp(elem_exp)) {
            elem_exp->meta.base_idx = exp->meta.base_idx;
            elem_exp->meta.rel_addr = exp->meta.rel_addr + offset;

            exp_gen(gen, elem_exp);
        }
        else {
            value = exp_gen(gen, elem_exp);
            ASSERT(value != NULL);

            instr_add(gen, BinaryenStore(gen->module, TYPE_BYTE(elem_meta->type), offset,
                                         0, address, value, meta_gen(elem_meta)));
        }

        offset += meta_bytes(elem_meta);
    }

    return address;
}

static BinaryenExpressionRef
exp_gen_alloc(gen_t *gen, ast_exp_t *exp)
{
    BinaryenExpressionRef address;

    address = BinaryenGetLocal(gen->module, exp->meta.base_idx, BinaryenTypeInt32());

    if (exp->meta.rel_addr > 0)
        address = BinaryenBinary(gen->module, BinaryenAddInt32(), address,
                                 i32_gen(gen, exp->meta.rel_addr));

    return address;
}

static BinaryenExpressionRef
exp_gen_global(gen_t *gen, ast_exp_t *exp)
{
    ASSERT(exp->u_glob.name != NULL);

    return BinaryenGetGlobal(gen->module, exp->u_glob.name, BinaryenTypeInt32());
}

static BinaryenExpressionRef
exp_gen_register(gen_t *gen, ast_exp_t *exp)
{
    return BinaryenGetLocal(gen->module, exp->u_reg.idx, meta_gen(&exp->meta));
}

static BinaryenExpressionRef
exp_gen_memory(gen_t *gen, ast_exp_t *exp)
{
    meta_t *meta = &exp->meta;
    BinaryenExpressionRef address;

    address = BinaryenGetLocal(gen->module, exp->u_mem.base, BinaryenTypeInt32());

    if (exp->u_mem.addr > 0)
        address = BinaryenBinary(gen->module, BinaryenAddInt32(), address,
                                 i32_gen(gen, exp->u_mem.addr));

    if (gen->is_lval || is_array_meta(meta) || is_object_meta(meta)) {
        if (exp->u_mem.offset > 0)
            return BinaryenBinary(gen->module, BinaryenAddInt32(), address,
                                  i32_gen(gen, exp->u_mem.offset));

        return address;
    }

    return BinaryenLoad(gen->module, TYPE_BYTE(meta->type), is_signed_type(meta->type),
                        exp->u_mem.offset, 0, meta_gen(meta), address);
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

    case EXP_INIT:
        return exp_gen_init(gen, exp);

    case EXP_ALLOC:
        return exp_gen_alloc(gen, exp);

    case EXP_GLOBAL:
        return exp_gen_global(gen, exp);

    case EXP_REGISTER:
        return exp_gen_register(gen, exp);

    case EXP_MEMORY:
        return exp_gen_memory(gen, exp);

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }

    return NULL;
}

/* end of gen_exp.c */
