/**
 * @file    gen_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "gen_exp.h"
#include "gen_util.h"

#include "gen_stmt.h"

/*
static void store_array(gen_t *gen, BinaryenExpressionRef var_addr,
                        BinaryenExpressionRef val_addr, uint32_t offset, meta_t *meta);
                        */

static BinaryenExpressionRef
stmt_gen_assign(gen_t *gen, ast_stmt_t *stmt)
{
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    ast_exp_t *r_exp = stmt->u_assign.r_exp;
    ast_id_t *id = l_exp->id;
    BinaryenExpressionRef address, value;

    if (id != NULL && is_map_meta(&id->meta))
        /* TODO: If the type of identifier is map,
         * lvalue and rvalue must be combined into a call expression */
        return NULL;

    value = exp_gen(gen, r_exp);
    if (value == NULL)
        return NULL;

    if (is_global_exp(l_exp))
        return BinaryenSetGlobal(gen->module, l_exp->u_glob.name, value);

    if (is_local_exp(l_exp))
        return BinaryenSetLocal(gen->module, l_exp->u_local.idx, value);

    gen->is_lval = true;
    address = exp_gen(gen, l_exp);
    gen->is_lval = false;

    if (is_stack_exp(l_exp)) {
        type_t type = l_exp->u_stk.type;

        if (is_array_meta(&l_exp->meta))
            type = TYPE_UINT32;

        return BinaryenStore(gen->module, TYPE_BYTE(type), 0, 0, address, value,
                             type_gen(type));
    }

    /* For an array whose index is a variable, we must dynamically determine the offset */
    ASSERT1(is_array_meta(&id->meta), id->meta.type);

    return BinaryenStore(gen->module, TYPE_BYTE(l_exp->meta.type), 0, 0, address,
                         value, meta_gen(&l_exp->meta));
}

#if 0
static void
store_elem(gen_t *gen, BinaryenExpressionRef address, BinaryenExpressionRef value,
           uint32_t offset, type_t type)
{
    /*
    BinaryenExpressionRef value;

    value = BinaryenLoad(gen->module, TYPE_BYTE(type), is_signed_type(type), offset, 0,
                         type_gen(type), val_addr);
                         */
    if (value == NULL)
        return;

    instr_add(gen, BinaryenStore(gen->module, TYPE_BYTE(type), offset, 0, address,
                                 value, type_gen(type)));
}

static void
store_struct(gen_t *gen, BinaryenExpressionRef address, BinaryenExpressionRef value,
             uint32_t offset, meta_t *meta)
{
    int i;

    ASSERT(meta->elem_cnt > 0);

    for (i = 0; i < meta->elem_cnt; i++) {
        meta_t *elem_meta = meta->elems[i];

        if (is_array_meta(elem_meta))
            store_array(gen, address, value, offset + elem_meta->rel_offset, elem_meta);
        else if (is_struct_meta(elem_meta))
            store_struct(gen, address, value, offset + elem_meta->rel_offset, elem_meta);
        else
            store_elem(gen, address, value, offset + elem_meta->rel_offset,
                       elem_meta->type);
    }
}

static void
store_array(gen_t *gen, BinaryenExpressionRef address, BinaryenExpressionRef value,
            uint32_t offset, meta_t *meta)
{
    int i, j;
    uint32_t unit_size = meta_unit(meta);

    ASSERT(meta->arr_dim > 0);

    for (i = 0; i < meta->arr_dim; i++) {
        ASSERT(meta->dim_sizes[i] > 0);

        for (j = 0; j < meta->dim_sizes[i]; j++) {
            if (is_struct_meta(meta))
                store_struct(gen, address, value, offset, meta);
            else
                store_elem(gen, address, value, offset, meta->type);

            offset += unit_size;
        }
    }
}

static void
return_value(gen_t *gen, ast_id_t *id, ast_exp_t *exp)
{
    BinaryenExpressionRef address, value;

    ASSERT(id->idx >= 0);

    address = BinaryenGetLocal(gen->module, id->idx, BinaryenTypeInt32());

    if (is_init_exp(exp))
        exp->meta.base_idx = id->idx;

    value = exp_gen(gen, exp);

    if (is_array_meta(&exp->meta))
        store_array(gen, address, value, 0, &exp->meta);
    else if (is_struct_meta(&exp->meta) || is_tuple_meta(&exp->meta))
        store_struct(gen, address, value, 0, &exp->meta);
    else
        store_elem(gen, address, value, 0, exp->meta.type);
}
#endif

static BinaryenExpressionRef
stmt_gen_return(gen_t *gen, ast_stmt_t *stmt)
{
    return BinaryenReturn(gen->module, exp_gen(gen, stmt->u_ret.arg_exp));

    // TODO multiple return values
#if 0
    ast_exp_t *arg_exp = stmt->u_ret.arg_exp;

    if (arg_exp != NULL) {
        ast_id_t *ret_id = stmt->u_ret.ret_id;

        if (is_ctor_id(ret_id->up))
            return BinaryenReturn(gen->module, exp_gen(gen, stmt->u_ret.arg_exp));

        if (is_tuple_exp(arg_exp)) {
            int i;
            array_t *elem_exps = arg_exp->u_tup.elem_exps;
            array_t *elem_ids = ret_id->u_tup.elem_ids;

            ASSERT1(is_tuple_id(ret_id), ret_id->kind);
            ASSERT2(array_size(elem_exps) == array_size(elem_ids), array_size(elem_exps),
                    array_size(elem_ids));

            array_foreach(elem_exps, i) {
                return_value(gen, array_get_id(elem_ids, i), array_get_exp(elem_exps, i));
            }
        }
        else {
            return_value(gen, ret_id, arg_exp);
        }
    }

    return NULL;
#endif
}

static BinaryenExpressionRef
stmt_gen_ddl(gen_t *gen, ast_stmt_t *stmt)
{
    /* TODO */
    return NULL;
}

BinaryenExpressionRef
stmt_gen(gen_t *gen, ast_stmt_t *stmt)
{
    switch (stmt->kind) {
    case STMT_EXP:
        return exp_gen(gen, stmt->u_exp.exp);

    case STMT_ASSIGN:
        return stmt_gen_assign(gen, stmt);

    case STMT_RETURN:
        return stmt_gen_return(gen, stmt);

    case STMT_DDL:
        return stmt_gen_ddl(gen, stmt);

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }

    return NULL;
}

/* end of gen_stmt.c */
