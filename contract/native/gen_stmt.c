/**
 * @file    gen_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "gen_exp.h"

#include "gen_stmt.h"

static void store_array(gen_t *gen, BinaryenExpressionRef var_addr, uint32_t var_offset,
                        meta_t *var_meta, BinaryenExpressionRef val_addr,
                        uint32_t val_offset, meta_t *val_meta);

static void
store_element(gen_t *gen, type_t type, BinaryenExpressionRef var_addr,
              uint32_t var_offset, BinaryenExpressionRef val_addr, uint32_t val_offset)
{
    BinaryenExpressionRef value;

    value = BinaryenLoad(gen->module, TYPE_SIZE(type), is_signed_type(type), val_offset,
                         0, type_gen(type), val_addr);

    instr_add(gen, BinaryenStore(gen->module, TYPE_SIZE(type), var_offset, 0, var_addr,
                                 value, type_gen(type)));
}

static void
store_struct(gen_t *gen, BinaryenExpressionRef var_addr, uint32_t var_offset,
             meta_t *var_meta, BinaryenExpressionRef val_addr, uint32_t val_offset,
             meta_t *val_meta)
{
    int i;

    ASSERT1(is_struct_meta(val_meta) || is_tuple_meta(val_meta), val_meta->type);

    for (i = 0; i < var_meta->elem_cnt; i++) {
        meta_t *elem_meta = var_meta->elems[i];

        if (is_array_meta(elem_meta))
            store_array(gen, var_addr, var_offset + elem_meta->rel_offset, elem_meta, 
                        val_addr, val_offset + elem_meta->rel_offset, val_meta->elems[i]);
        else if (is_struct_meta(elem_meta))
            store_struct(gen, var_addr, var_offset + elem_meta->rel_offset, elem_meta, 
                         val_addr, val_offset + elem_meta->rel_offset,
                         val_meta->elems[i]);
        else
            store_element(gen, elem_meta->type, var_addr, 
                          var_offset + elem_meta->rel_offset, val_addr,
                          val_offset + elem_meta->rel_offset);
    }
}

static void
store_array(gen_t *gen, BinaryenExpressionRef var_addr, uint32_t var_offset,
            meta_t *var_meta, BinaryenExpressionRef val_addr, uint32_t val_offset,
            meta_t *val_meta)
{
    int i, j;
    uint32_t unit_size = meta_unit(var_meta);

    ASSERT(var_meta->arr_dim > 0);
    ASSERT1(is_array_meta(val_meta) || is_tuple_meta(val_meta), val_meta->type);

    for (i = 0; i < var_meta->arr_dim; i++) {
        for (j = 0; j < var_meta->dim_sizes[i]; j++) {
            if (is_struct_meta(var_meta)) {
                if (is_array_meta(val_meta))
                    store_struct(gen, var_addr, var_offset, var_meta, val_addr, 
                                 val_offset, val_meta);
                else
                    store_struct(gen, var_addr, var_offset, var_meta, val_addr, 
                                 val_offset, val_meta->elems[i]);
            }
            else {
                store_element(gen, var_meta->type, var_addr, var_offset, val_addr,
                              val_offset);
            }

            var_offset += unit_size;
            val_offset += unit_size;
        }
    }
}

static BinaryenExpressionRef
stmt_gen_assign(gen_t *gen, ast_stmt_t *stmt)
{
    //uint32_t offset = 0;
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    ast_exp_t *r_exp = stmt->u_assign.r_exp;
    ast_id_t *id = l_exp->id;
    BinaryenExpressionRef address, value;

    if (id != NULL && is_map_meta(&id->meta))
        /* TODO: If the type of identifier is map,
         * lvalue and rvalue must be combined into a call expression */
        return NULL;

    value = exp_gen(gen, r_exp);
    //if (value == NULL || is_object_meta(&id->meta))
    if (value == NULL)
        return NULL;

#if 0
    if (id->is_param) {
        uint32_t bytes;

        if (is_return_id(id)) {
            address = BinaryenGetLocal(gen->module, id->idx, BinaryenTypeInt32());
            bytes = sizeof(int32_t);
        }
        else {
            address = BinaryenGetLocal(gen->module, id->idx, meta_gen(&id->meta));
            bytes = TYPE_SIZE(id->meta.type);
        }

        return BinaryenStore(gen->module, bytes, offset, 0, address, value,
                             meta_gen(&l_exp->meta));
    }
#endif

    if (is_global_exp(l_exp)) {
        ASSERT(l_exp->u_glob.name != NULL);
        return BinaryenSetGlobal(gen->module, l_exp->u_glob.name, value);
    }

    if (is_local_exp(l_exp)) {
        ASSERT(l_exp->u_local.idx >= 0);
        return BinaryenSetLocal(gen->module, l_exp->u_local.idx, value);
    }

    if (is_stack_exp(l_exp)) {
        ASSERT(l_exp->u_stk.base >= 0);
        ASSERT(l_exp->u_stk.addr >= 0);

        address = BinaryenGetLocal(gen->module, l_exp->u_stk.base, BinaryenTypeInt32());

        if (l_exp->u_stk.addr > 0)
            address = BinaryenBinary(gen->module, BinaryenAddInt32(), address,
                                     i32_gen(gen, l_exp->u_stk.addr));

        if (is_array_meta(&l_exp->meta)) {
            store_array(gen, address, l_exp->u_stk.offset, &l_exp->meta, value, 0, 
                        &r_exp->meta);
            return NULL;
        }

        if (is_struct_meta(&l_exp->meta)) {
            store_struct(gen, address, l_exp->u_stk.offset, &l_exp->meta, value, 0, 
                         &r_exp->meta);
            return NULL;
        }

        return BinaryenStore(gen->module, TYPE_SIZE(l_exp->u_stk.type),
                             l_exp->u_stk.offset, 0, address, value,
                             type_gen(l_exp->u_stk.type));
    }

    /*
    ASSERT(id != NULL);

    if (is_return_id(id)) {
        ASSERT(false);
        ASSERT(id->idx >= 0);
        return BinaryenStore(gen->module, TYPE_SIZE(l_exp->meta.type), 0, 0,
                             BinaryenGetLocal(gen->module, id->idx, BinaryenTypeInt32()),
                             value, meta_gen(&l_exp->meta));
    }
    */

    /* For an array whose index is a variable, we must dynamically determine the offset */
    ASSERT1(is_array_meta(&id->meta), id->meta.type);

    gen->is_lval = true;
    address = exp_gen(gen, l_exp);
    gen->is_lval = false;

    return BinaryenStore(gen->module, TYPE_SIZE(l_exp->meta.type), 0, 0, address,
                         value, meta_gen(&l_exp->meta));
}

static BinaryenExpressionRef
stmt_gen_return(gen_t *gen, ast_stmt_t *stmt)
{
    return BinaryenReturn(gen->module, exp_gen(gen, stmt->u_ret.arg_exp));
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
