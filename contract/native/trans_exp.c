/**
 * @file    trans_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_blk.h"
#include "ast_stmt.h"
#include "ir_bb.h"
#include "ir_fn.h"
#include "ir_sgmt.h"

#include "trans_exp.h"

static void copy_array(trans_t *trans, uint32_t base_idx, uint32_t rel_addr, int dim_idx,
                       meta_t *meta);

static void
exp_trans_lit(trans_t *trans, ast_exp_t *exp)
{
    int addr;
    value_t *val = &exp->u_lit.val;
    ir_sgmt_t *sgmt = &trans->ir->sgmt;

    switch (val->type) {
    case TYPE_BOOL:
    case TYPE_UINT64:
    case TYPE_DOUBLE:
        break;

    case TYPE_STRING:
        addr = sgmt_add_raw(sgmt, val_ptr(val), val_size(val) + 1);

        value_set_i64(val, addr);
        meta_set_uint32(&exp->meta);
        break;

    case TYPE_OBJECT:
        if (is_null_val(val))
            addr = sgmt_add_raw(sgmt, "\0\0\0\0", 4);
        else
            addr = sgmt_add_raw(sgmt, val_ptr(val), val_size(val));

        value_set_i64(val, addr);
        meta_set_uint32(&exp->meta);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
exp_trans_id(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;

    ASSERT(id != NULL);

    if (is_var_id(id)) {
        if (is_global_id(id)) {
            ASSERT1(id->meta.rel_offset == 0, id->meta.rel_offset);

            /* The global variable always refers to the memory */
            exp_set_memory(exp, id->meta.base_idx, id->meta.rel_addr, 0);
        }
        else {
            exp_set_register(exp, id->idx);
        }
    }
    else if (is_fn_id(id) || is_cont_id(id)) {
        /* In the case of a contract identifier, the "this" syntax is used */
        exp_set_register(exp, trans->fn->cont_idx);
        meta_set_uint32(&exp->meta);
    }
}

static void
exp_trans_array(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;
    ast_exp_t *id_exp = exp->u_arr.id_exp;
    ast_exp_t *idx_exp = exp->u_arr.idx_exp;

    exp_trans(trans, id_exp);
    exp_trans(trans, idx_exp);

    if (is_array_meta(&id->meta)) {
        uint32_t offset;

        /* In array expression, the offset is calculated as follows:
         *
         * Suppose that "int i[x][y][z]" is defined.
         *
         * First, when we access "i[a]", the formula for calculating the offset is
         * (a * sizeof(i[0])).
         *
         * Next, in the case of "i[a][b]",
         * (a * sizeof(i[0])) + (b * sizeof(i[0][0])).
         *
         * Finally, in the case of "i[a][b][c]",
         * (a * sizeof(i[0])) + (b * sizeof(i[0][0])) + (c * sizeof(int)). */

        if (!is_lit_exp(idx_exp))
            /* We must dynamically determine the address and offset */
            return;

        /* If "id_exp" is a call expression, it can be a memory expression */
        ASSERT1(is_memory_exp(id_exp) || is_register_exp(id_exp), id_exp->kind);

        /* The following meta_bytes() is stripped size of array */
        offset = val_i64(&idx_exp->u_lit.val) * meta_bytes(&exp->meta) +
            meta_align(&id->meta);

        if (is_memory_exp(id_exp))
            exp_set_memory(exp, id_exp->u_mem.base, id_exp->u_mem.addr,
                           id_exp->u_mem.offset + offset);
        else
            exp_set_memory(exp, id_exp->u_reg.idx, 0, offset);
    }
    else {
        /* TODO
         * int addr = fn_add_stack_var(trans->fn);
         * ast_exp_t *call_exp = exp_new_call("$map_get", &exp->pos);
         *
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         *
         * return <return address of call>; */
    }
}

static void
exp_trans_cast(trans_t *trans, ast_exp_t *exp)
{
    exp_trans(trans, exp->u_cast.val_exp);

    if (is_string_meta(&exp->meta) || is_string_meta(&exp->u_cast.to_meta)) {
        /* TODO
         * int addr = fn_add_stack_var(trans->fn);
         * ast_exp_t *call_exp = exp_new_call("$concat", &exp->pos);
         *
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         *
         * return <return address of call>; */
    }
}

static void
exp_trans_unary(trans_t *trans, ast_exp_t *exp)
{
    ast_exp_t *val_exp = exp->u_un.val_exp;
    ast_exp_t *var_exp, *bi_exp, *lit_exp;

    switch (exp->u_un.kind) {
    case OP_INC:
    case OP_DEC:
        /* Clone value expression because we have to transform it to "x op 1" */
        var_exp = exp_clone(val_exp);

        exp_trans(trans, var_exp);
        exp_trans(trans, val_exp);

        lit_exp = exp_new_lit_i64(1, &exp->pos);
        meta_copy(&lit_exp->meta, &val_exp->meta);

        bi_exp = exp_new_binary(exp->u_un.kind == OP_INC ? OP_ADD : OP_SUB, val_exp,
                                lit_exp, &exp->pos);
        meta_copy(&bi_exp->meta, &val_exp->meta);

        if (exp->u_un.is_prefix)
            bb_add_stmt(trans->bb, stmt_new_assign(var_exp, bi_exp, &exp->pos));
        else
            /* The postfix operator is added as a piggybacked statement since it must
             * be executed after the current statement is executed */
            bb_add_piggyback(trans->bb, stmt_new_assign(var_exp, bi_exp, &exp->pos));

        *exp = *val_exp;
        break;

    case OP_NEG:
    case OP_NOT:
        exp_trans(trans, val_exp);
        break;

    default:
        ASSERT1(!"invalid operator", exp->u_un.kind);
    }
}

static void
exp_trans_binary(trans_t *trans, ast_exp_t *exp)
{
    exp_trans(trans, exp->u_bin.l_exp);
    exp_trans(trans, exp->u_bin.r_exp);

    if (exp->u_bin.kind == OP_ADD && is_string_meta(&exp->meta)) {
        /* TODO
         * int addr = fn_add_stack();
         * ast_exp_t *call exp = exp_new_call("$concat", &exp->pos);
         * bb_add_stmt(trans->bb, stmt_new_exp(call_exp, &exp->pos));
         * return exp_new_memory(addr, offset, &exp->pos);
        */
    }
}

static void
exp_trans_ternary(trans_t *trans, ast_exp_t *exp)
{
    exp_trans(trans, exp->u_tern.pre_exp);
    exp_trans(trans, exp->u_tern.in_exp);
    exp_trans(trans, exp->u_tern.post_exp);

    if (is_lit_exp(exp->u_tern.pre_exp)) {
        /* Maybe we should do this in optimizer (if exists) */
        meta_t meta = exp->meta;

        if (val_bool(&exp->u_tern.pre_exp->u_lit.val))
            *exp = *exp->u_tern.in_exp;
        else
            *exp = *exp->u_tern.post_exp;

        meta_copy(&exp->meta, &meta);
    }
}

static void
exp_trans_access(trans_t *trans, ast_exp_t *exp)
{
    ast_exp_t *qual_exp = exp->u_acc.qual_exp;
    ast_id_t *fld_id = exp->id;

    exp_trans(trans, qual_exp);

    if (is_fn_id(fld_id)) {
        ASSERT1(is_register_exp(qual_exp), qual_exp->kind);
        return;
    }

    if (is_register_exp(qual_exp)) {
        /* The "rel_addr" of "fld_id" is greater than 0 when referring to a global
         * variable belonging to the contract register */
        exp_set_memory(exp, qual_exp->u_reg.idx, fld_id->meta.rel_addr,
                       fld_id->meta.rel_offset);
    }
    else if (is_memory_exp(qual_exp)) {
        /* It can be a memroy expression when referring to a global variable directly */
        ASSERT2(qual_exp->u_mem.offset == 0 || fld_id->meta.rel_offset == 0,
                qual_exp->u_mem.offset, fld_id->meta.rel_offset);
        ASSERT2(qual_exp->u_mem.base == trans->fn->cont_idx, qual_exp->u_mem.base,
                trans->fn->cont_idx);

        exp_set_memory(exp, qual_exp->u_mem.base, qual_exp->u_mem.addr,
                       qual_exp->u_mem.offset + fld_id->meta.rel_offset);
    }
    else {
        /* If qualifier is a function and returns an array or a struct, "qual_exp" can
         * be a binary expression (See exp_trans_call()) */
        ASSERT1(is_binary_exp(qual_exp), qual_exp->kind);
    }
}

static void
copy_val(trans_t *trans, uint32_t base_idx, uint32_t rel_addr, meta_t *meta)
{
    ast_exp_t *l_exp, *r_exp;

    l_exp = exp_new_memory(trans->fn->stack_idx, rel_addr, 0);
    meta_copy(&l_exp->meta, meta);

    r_exp = exp_new_memory(base_idx, rel_addr, 0);
    meta_copy(&r_exp->meta, meta);

    bb_add_stmt(trans->bb, stmt_new_assign(l_exp, r_exp, &null_pos_));
}

static void
copy_struct(trans_t *trans, uint32_t base_idx, uint32_t rel_addr, meta_t *meta)
{
    int i;

    ASSERT(meta->elem_cnt > 0);

    for (i = 0; i < meta->elem_cnt; i++) {
        meta_t *elem_meta = meta->elems[i];
        uint32_t offset = elem_meta->rel_offset;

        if (is_array_meta(elem_meta))
            copy_array(trans, base_idx, rel_addr + offset, 0, elem_meta);
        else if (is_struct_meta(elem_meta))
            copy_struct(trans, base_idx, rel_addr + offset, elem_meta);
        else
            copy_val(trans, base_idx, rel_addr + offset, elem_meta);
    }
}

static void
copy_array(trans_t *trans, uint32_t base_idx, uint32_t rel_addr, int dim_idx,
           meta_t *meta)
{
    int i;
    uint32_t offset = 0;
    uint32_t unit_size = meta_size(meta);
    meta_t hdr_meta;

    ASSERT(dim_idx >= 0);
    ASSERT(meta->dim_sizes[dim_idx] > 0);
    ASSERT(meta->arr_dim > 0);

    meta_set_uint32(&hdr_meta);

    for (i = 0; i < meta->dim_sizes[dim_idx]; i++) {
        if (dim_idx < meta->arr_dim - 1) {
            copy_val(trans, base_idx, rel_addr + offset, &hdr_meta);
            offset += meta_size(&hdr_meta);

            copy_array(trans, base_idx, rel_addr + offset, dim_idx + 1, meta);
        }
        else {
            if (is_struct_meta(meta))
                copy_struct(trans, base_idx, rel_addr + offset, meta);
            else
                copy_val(trans, base_idx, rel_addr + offset, meta);

            offset += unit_size;
        }
    }
}

static void
exp_trans_call(trans_t *trans, ast_exp_t *exp)
{
    int i;
    meta_t *meta = &exp->meta;
    ast_exp_t *id_exp = exp->u_call.id_exp;
    ast_id_t *fn_id = exp->id;
    ir_fn_t *fn = trans->fn;

    if (is_map_meta(meta))
        /* TODO */
        return;

    exp_trans(trans, id_exp);

    if (!is_ctor_id(exp->id)) {
        /* Since non-constructor functions are added the contract base address as a first
         * argument, we must also add the address as a call argument here */
        if (exp->u_call.param_exps == NULL)
            exp->u_call.param_exps = vector_new();

        if (is_access_exp(id_exp)) {
            ast_exp_t *qual_exp = id_exp->u_acc.qual_exp;

            ASSERT1(is_object_meta(&qual_exp->meta), qual_exp->meta.type);

            /* If the expression is of type "x.y()", pass "x" as the first argument */
            vector_add_first(exp->u_call.param_exps, qual_exp);
        }
        else {
            ast_exp_t *param_exp;

            ASSERT1(is_register_exp(id_exp), id_exp->kind);
            ASSERT1(trans->fn->cont_idx == 0, trans->fn->cont_idx);

            /* If the expression is of type "x()", pass my first parameter as the
             * first argument */
            param_exp = exp_new_register(0);
            meta_set_uint32(&param_exp->meta);

            vector_add_first(exp->u_call.param_exps, param_exp);
        }
    }

    vector_foreach(exp->u_call.param_exps, i) {
        exp_trans(trans, vector_get_exp(exp->u_call.param_exps, i));
    }

    if (fn_id->u_fn.ret_id != NULL) {
        uint32_t reg_idx = fn_add_register(fn, meta);
        ast_exp_t *reg_exp;

        reg_exp = exp_new_register(reg_idx);
        meta_copy(&reg_exp->meta, meta);

        /* We have to clone it because the call expression itself is transformed */
        bb_add_stmt(trans->bb, stmt_new_assign(reg_exp, exp_clone(exp), &exp->pos));

        if (is_array_meta(&fn_id->meta) || is_struct_meta(&fn_id->meta)) {
            /* If the return value is an array or struct, we must copy the value because
             * we do share memory space between the caller and the callee */
            if (trans->is_heap)
                fn_add_heap(fn, meta_bytes(meta), meta);
            else
                fn_add_stack(fn, meta_bytes(meta), meta);

            if (is_array_meta(&fn_id->meta))
                copy_array(trans, reg_idx, meta->rel_addr, 0, meta);
            else
                copy_struct(trans, reg_idx, meta->rel_addr, meta);

            meta_set_uint32(meta);

            if (meta->rel_addr > 0) {
                exp->kind = EXP_BINARY;
                exp->u_bin.kind = OP_ADD;
                exp->u_bin.l_exp = exp_new_register(meta->base_idx);
                exp->u_bin.r_exp = exp_new_lit_i64(meta->rel_addr, &exp->pos);
                meta_set_uint32(&exp->u_bin.l_exp->meta);
            }
            else {
                exp_set_register(exp, meta->base_idx);
            }
        }
        else {
            exp_set_register(exp, reg_idx);
        }
    }
    else {
        bb_add_stmt(trans->bb, stmt_new_exp(exp, &exp->pos));
    }
}

static void
exp_trans_sql(trans_t *trans, ast_exp_t *exp)
{
    /* TODO */
}

static void
exp_trans_tuple(trans_t *trans, ast_exp_t *exp)
{
    int i;
    vector_t *elem_exps = exp->u_tup.elem_exps;

    vector_foreach(elem_exps, i) {
        exp_trans(trans, vector_get_exp(elem_exps, i));
    }
}

static void
exp_trans_init(trans_t *trans, ast_exp_t *exp)
{
    int i;
    meta_t *meta = &exp->meta;
    vector_t *elem_exps = exp->u_init.elem_exps;

    ASSERT1(is_tuple_meta(meta), meta->type);

    if (!exp->u_init.is_aggr) {
        uint32_t size;

        if (is_array_meta(meta) && meta->arr_dim == meta->max_dim)
            size = sizeof(uint32_t);
        else
            size = meta_bytes(meta);

        if (trans->is_heap)
            fn_add_heap(trans->fn, size, meta);
        else
            fn_add_stack(trans->fn, size, meta);
    }

    vector_foreach(elem_exps, i) {
        exp_trans(trans, vector_get_exp(elem_exps, i));
    }

    if (exp->u_init.is_aggr) {
        int offset = 0;
        uint32_t size = meta_bytes(meta);
        char *raw = xcalloc(size);

        if (is_array_meta(meta)) {
            ASSERT(meta->dim_sizes[0] > 0);

            memcpy(raw + offset, &meta->dim_sizes[0], sizeof(int));
            offset += sizeof(uint32_t);
        }

        vector_foreach(elem_exps, i) {
            ast_exp_t *elem_exp = vector_get_exp(elem_exps, i);
            meta_t *elem_meta = &elem_exp->meta;
            void *val_ptr = value_ptr(&elem_exp->u_lit.val, elem_meta);
            uint32_t val_size = value_size(&elem_exp->u_lit.val, elem_meta);

            ASSERT2(val_size <= meta_bytes(elem_meta), val_size, meta_bytes(elem_meta));
            ASSERT3(offset + val_size <= size, offset, val_size, size);

            offset = ALIGN(offset, meta_align(elem_meta));

            memcpy(raw + offset, val_ptr, val_size);
            offset += meta_bytes(elem_meta);
        }

        ASSERT2(offset <= size, offset, size);

        exp_set_lit(exp, NULL);
        value_set_ptr(&exp->u_lit.val, raw, size);
    }
}

static void
exp_trans_alloc(trans_t *trans, ast_exp_t *exp)
{
    meta_t *meta = &exp->meta;

    if (trans->is_heap)
        fn_add_heap(trans->fn, meta_bytes(meta), meta);
    else
        fn_add_stack(trans->fn, meta_bytes(meta), meta);
}

void
exp_trans(trans_t *trans, ast_exp_t *exp)
{
    ASSERT(exp != NULL);

    switch (exp->kind) {
    case EXP_NULL:
        break;

    case EXP_LIT:
        exp_trans_lit(trans, exp);
        break;

    case EXP_ID:
        exp_trans_id(trans, exp);
        break;

    case EXP_ARRAY:
        exp_trans_array(trans, exp);
        break;

    case EXP_CAST:
        exp_trans_cast(trans, exp);
        break;

    case EXP_UNARY:
        exp_trans_unary(trans, exp);
        break;

    case EXP_BINARY:
        exp_trans_binary(trans, exp);
        break;

    case EXP_TERNARY:
        exp_trans_ternary(trans, exp);
        break;

    case EXP_ACCESS:
        exp_trans_access(trans, exp);
        break;

    case EXP_CALL:
        exp_trans_call(trans, exp);
        break;

    case EXP_SQL:
        exp_trans_sql(trans, exp);
        break;

    case EXP_TUPLE:
        exp_trans_tuple(trans, exp);
        break;

    case EXP_INIT:
        exp_trans_init(trans, exp);
        break;

    case EXP_ALLOC:
        exp_trans_alloc(trans, exp);
        break;

    case EXP_GLOBAL:
    case EXP_REGISTER:
    case EXP_MEMORY:
        break;

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }
}

/* end of trans_exp.c */
