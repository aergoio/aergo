/**
 * @file    trans_exp.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_blk.h"
#include "ast_stmt.h"
#include "ir_md.h"
#include "ir_abi.h"
#include "ir_fn.h"
#include "ir_bb.h"
#include "ir_sgmt.h"
#include "trans_stmt.h"
#include "syslib.h"

#include "trans_exp.h"

static void
exp_trans_lit(trans_t *trans, ast_exp_t *exp)
{
    value_t *val = &exp->u_lit.val;
    meta_t *meta = &exp->meta;
    ir_md_t *md = trans->md;

    ASSERT(md != NULL);

    switch (val->type) {
    case TYPE_BOOL:
    case TYPE_UINT128:
    case TYPE_DOUBLE:
        break;

    case TYPE_STRING:
        value_set_int(val, sgmt_add_raw(&md->sgmt, val_ptr(val), val_size(val) + 1));
        meta_set_uint32(meta);
        break;

    case TYPE_OBJECT:
        if (is_null_val(val))
            value_init_int(val);
        else
            value_set_int(val, sgmt_add_raw(&md->sgmt, val_ptr(val), val_size(val)));
        meta_set_uint32(meta);
        break;

    default:
        ASSERT1(!"invalid value", val->type);
    }
}

static void
exp_trans_id(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;
    ir_fn_t *fn = trans->fn;

    ASSERT(id != NULL);

    if (is_var_id(id)) {
        if (is_global_id(id)) {
            ASSERT1(id->meta.rel_offset == 0, id->meta.rel_offset);

            /* The global variable always refers to the memory. */
            exp_set_mem(exp, id->meta.base_idx, id->meta.rel_addr, 0);
        }
        else {
            exp_set_reg(exp, id->idx);
        }
    }
    else if (is_cont_id(id)) {
        /* In the case of a contract identifier, the "this" syntax is used */
        exp_set_reg(exp, fn->cont_idx);
        meta_set_uint32(&exp->meta);
    }
}

static void
exp_trans_array(trans_t *trans, ast_exp_t *exp)
{
    ast_id_t *id = exp->id;
    ast_exp_t *id_exp = exp->u_arr.id_exp;
    ast_exp_t *idx_exp = exp->u_arr.idx_exp;

    ASSERT(id != NULL);

    exp_trans(trans, id_exp);
    exp_trans(trans, idx_exp);

    if (is_array_meta(&id->meta)) {
        int i;
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

        for (i = 0; i < id->meta.arr_dim; i++) {
            if (id->meta.dim_sizes[i] == -1)
                /* When the user uses the expression like "int v[][] = f()", the size of the array
                 * is determined dynamically. */
                return;
        }

        ASSERT1(is_mem_exp(id_exp) || is_reg_exp(id_exp), id_exp->kind);

        /* The following meta_bytes() is stripped size of array */
        offset = val_i64(&idx_exp->u_lit.val) * meta_bytes(&exp->meta) + meta_align(&id->meta);

        if (is_mem_exp(id_exp))
            exp_set_mem(exp, id_exp->u_mem.base, id_exp->u_mem.addr, id_exp->u_mem.offset + offset);
        else
            exp_set_mem(exp, id_exp->u_reg.idx, 0, offset);
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

        lit_exp = exp_new_lit_int(1, &exp->pos);
        meta_copy(&lit_exp->meta, &val_exp->meta);

        bi_exp = exp_new_binary(exp->u_un.kind == OP_INC ? OP_ADD : OP_SUB, val_exp, lit_exp,
                                &exp->pos);
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
        ASSERT1(is_reg_exp(qual_exp), qual_exp->kind);
        return;
    }

    if (is_reg_exp(qual_exp)) {
        /* The "rel_addr" of "fld_id" is greater than 0 when referring to a global variable
         * belonging to the contract register */
        exp_set_mem(exp, qual_exp->u_reg.idx, fld_id->meta.rel_addr, fld_id->meta.rel_offset);
    }
    else if (is_mem_exp(qual_exp)) {
        /* It can be a memroy expression when referring to a global variable directly */
        ASSERT2(qual_exp->u_mem.offset == 0 || fld_id->meta.rel_offset == 0,
                qual_exp->u_mem.offset, fld_id->meta.rel_offset);

        exp_set_mem(exp, qual_exp->u_mem.base, qual_exp->u_mem.addr,
                    qual_exp->u_mem.offset + fld_id->meta.rel_offset);
    }
    else {
        /* If qualifier is a function and returns an array or a struct, "qual_exp" can be a binary
         * expression (See exp_trans_call()) */
        ASSERT1(is_binary_exp(qual_exp), qual_exp->kind);
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

    exp_trans(trans, id_exp);

    if (fn_id->up != trans->id)
        md_add_imp(trans->md, abi_new(fn_id));

    if (is_ctor_id(fn_id) || is_system_id(fn_id)) {
        /* The constructor does not change the parameter, it always returns address. */
        vector_foreach(exp->u_call.param_exps, i) {
            exp_trans(trans, vector_get_exp(exp->u_call.param_exps, i));
        }
        return;
    }

    /* Since non-constructor functions are added the contract base address as a first parameter,
     * we must also add the address as a call argument here. */
    if (exp->u_call.param_exps == NULL)
        exp->u_call.param_exps = vector_new();

    if (is_access_exp(id_exp)) {
        ast_exp_t *qual_exp = id_exp->u_acc.qual_exp;

        ASSERT1(is_reg_exp(qual_exp), qual_exp->kind);
        ASSERT1(is_object_meta(&qual_exp->meta), qual_exp->meta.type);

        /* If the expression is of type "x.y()", pass "x" as the first argument. */
        vector_add_first(exp->u_call.param_exps, qual_exp);
    }
    else {
        ast_exp_t *param_exp;

        ASSERT1(fn->cont_idx == 0, fn->cont_idx);

        /* If the expression is of type "x()", pass my first parameter as the first argument. */
        param_exp = exp_new_reg(0);
        meta_set_uint32(&param_exp->meta);

        vector_add_first(exp->u_call.param_exps, param_exp);
    }

    vector_foreach(exp->u_call.param_exps, i) {
        exp_trans(trans, vector_get_exp(exp->u_call.param_exps, i));
    }

    if (fn_id->u_fn.ret_id != NULL) {
        uint32_t reg_idx = fn_add_register(fn, meta);
        ast_exp_t *reg_exp;

        reg_exp = exp_new_reg(reg_idx);
        meta_copy(&reg_exp->meta, meta);

        /* We have to clone it because the call expression itself is transformed */
        bb_add_stmt(trans->bb, stmt_new_assign(reg_exp, exp_clone(exp), &exp->pos));

        if (is_array_meta(&fn_id->meta) || is_struct_meta(&fn_id->meta)) {
            uint32_t mem_idx;
            uint32_t size = meta_bytes(meta);
            ast_exp_t *l_exp, *r_exp;
            ast_exp_t *addr_exp, *cpy_exp;

            /* If the return value is an array or struct, we must copy the value because we do
             * share memory space between the caller and the callee */
            if (trans->is_global) {
                mem_idx = fn_add_register(trans->fn, meta);

                l_exp = exp_new_reg(mem_idx);
                meta_set_uint32(&l_exp->meta);

                r_exp = syslib_new_malloc(trans, size, &exp->pos);

                bb_add_stmt(trans->bb, stmt_new_assign(l_exp, r_exp, &exp->pos));

                addr_exp = l_exp;

                exp_set_reg(exp, mem_idx);
            }
            else {
                fn_add_stack(fn, size, meta);
                mem_idx = meta->base_idx;

                l_exp = exp_new_reg(mem_idx);
                meta_set_uint32(&l_exp->meta);

                r_exp = exp_new_lit_int(meta->rel_addr, &exp->pos);
                meta_set_uint32(&r_exp->meta);

                addr_exp = exp_new_binary(OP_ADD, l_exp, r_exp, &exp->pos);

                exp_set_mem(exp, mem_idx, meta->rel_addr, 0);
            }

            cpy_exp = syslib_new_memcpy(trans, addr_exp, reg_exp, size, &exp->pos);

            bb_add_stmt(trans->bb, stmt_new_exp(cpy_exp, &exp->pos));
        }
        else {
            exp_set_reg(exp, reg_idx);
        }
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

    if (exp->u_init.is_aggr) {
        uint32_t offset = 0;
        uint32_t size = meta_bytes(meta);
        char *raw = xcalloc(size);

        vector_foreach(elem_exps, i) {
            exp_trans(trans, vector_get_exp(elem_exps, i));
        }

        if (is_array_meta(meta)) {
            ASSERT(meta->dim_sizes[0] > 0);

            memcpy(raw + offset, &meta->dim_sizes[0], sizeof(int));
            offset += sizeof(uint32_t);
        }

        vector_foreach(elem_exps, i) {
            uint32_t val_size;
            ast_exp_t *elem_exp = vector_get_exp(elem_exps, i);
            meta_t *elem_meta = &elem_exp->meta;

            offset = ALIGN(offset, meta_align(elem_meta));

            val_size = value_serialize(&elem_exp->u_lit.val, raw + offset, elem_meta);
            ASSERT2(val_size <= meta_bytes(elem_meta), val_size, meta_bytes(elem_meta));
            ASSERT3(offset + val_size <= size, offset, val_size, size);

            offset += meta_bytes(elem_meta);
        }

        ASSERT2(offset <= size, offset, size);

        exp_set_lit(exp, NULL);
        value_set_ptr(&exp->u_lit.val, raw, size);
    }
    else {
        uint32_t size;

        if (is_array_meta(meta) && meta->arr_dim == meta->max_dim)
            size = sizeof(uint32_t);
        else
            size = meta_bytes(meta);

        if (trans->is_global) {
            uint32_t reg_idx = meta->base_idx;
            uint32_t offset = meta->rel_offset;
            ast_exp_t *l_exp, *r_exp;

            if (exp->u_init.is_outmost) {
                reg_idx = fn_add_register(trans->fn, meta);

                l_exp = exp_new_reg(reg_idx);
                meta_set_uint32(&l_exp->meta);

                r_exp = syslib_new_malloc(trans, meta_bytes(meta), &exp->pos);

                bb_add_stmt(trans->bb, stmt_new_assign(l_exp, r_exp, &exp->pos));

                exp_set_reg(exp, reg_idx);
            }

            if (is_array_meta(meta)) {
                ASSERT(meta->dim_sizes[0] > 0);

                l_exp = exp_new_mem(reg_idx, 0, offset);
                meta_set_uint32(&l_exp->meta);

                r_exp = exp_new_lit_int(meta->dim_sizes[0], &exp->pos);
                meta_set_uint32(&r_exp->meta);

                bb_add_stmt(trans->bb, stmt_new_assign(l_exp, r_exp, &exp->pos));
                offset += sizeof(uint32_t);
            }

            vector_foreach(elem_exps, i) {
                ast_exp_t *elem_exp = vector_get_exp(elem_exps, i);
                meta_t *elem_meta = &elem_exp->meta;

                offset = ALIGN(offset, meta_align(elem_meta));

                elem_meta->base_idx = reg_idx;
                elem_meta->rel_offset = offset;

                exp_trans(trans, elem_exp);

                if (!is_init_exp(elem_exp)) {
                    l_exp = exp_new_mem(reg_idx, 0, offset);
                    meta_copy(&l_exp->meta, elem_meta);

                    bb_add_stmt(trans->bb, stmt_new_assign(l_exp, elem_exp, &exp->pos));
                }

                offset += meta_bytes(elem_meta);
            }
        }
        else {
            vector_foreach(elem_exps, i) {
                exp_trans(trans, vector_get_exp(elem_exps, i));
            }

            fn_add_stack(trans->fn, size, meta);
        }
    }
}

static void
exp_trans_alloc(trans_t *trans, ast_exp_t *exp)
{
    meta_t *meta = &exp->meta;

    if (trans->is_global) {
        uint32_t reg_idx = fn_add_register(trans->fn, meta);
        ast_exp_t *l_exp, *r_exp;

        l_exp = exp_new_reg(reg_idx);
        meta_set_uint32(&l_exp->meta);

        r_exp = syslib_new_malloc(trans, meta_bytes(meta), &exp->pos);

        bb_add_stmt(trans->bb, stmt_new_assign(l_exp, r_exp, &exp->pos));

        if (is_array_meta(meta)) {
            ASSERT(meta->dim_sizes[0] > 0);

            l_exp = exp_new_mem(reg_idx, 0, 0);
            meta_set_uint32(&l_exp->meta);

            r_exp = exp_new_lit_int(meta->dim_sizes[0], &exp->pos);
            meta_set_uint32(&r_exp->meta);

            bb_add_stmt(trans->bb, stmt_new_assign(l_exp, r_exp, &exp->pos));
        }

        exp_set_reg(exp, reg_idx);
    }
    else {
        fn_add_stack(trans->fn, meta_bytes(meta), meta);
    }
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
    case EXP_REG:
    case EXP_MEM:
        break;

    default:
        ASSERT1(!"invalid expression", exp->kind);
    }
}

/* end of trans_exp.c */
