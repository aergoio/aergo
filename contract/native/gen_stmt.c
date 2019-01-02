/**
 * @file    gen_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "gen_exp.h"
#include "gen_util.h"

#include "gen_stmt.h"

static BinaryenExpressionRef
stmt_gen_assign(gen_t *gen, ast_stmt_t *stmt)
{
    uint32_t offset;
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    BinaryenExpressionRef address, value;

    ASSERT(l_exp->id != NULL);

    if (is_map_type(&l_exp->id->meta))
        /* TODO: If the type of identifier is map, 
         * lvalue and rvlaue must be combined into a call expression */
        return NULL;

    value = exp_gen(gen, stmt->u_assign.r_exp);

    if (is_local_ref_exp(l_exp))
        return BinaryenSetLocal(gen->module, l_exp->u_lo.idx, value);

    if (is_id_ref_exp(l_exp)) {
        ASSERT(false);
        //ASSERT1(is_global_id(id), id->scope);
        return BinaryenSetGlobal(gen->module, "", value);
    }

    if (is_stack_ref_exp(l_exp)) {
        address = gen_i32(gen, l_exp->u_st.addr);
        offset = l_exp->u_st.offset;
    }
    else {
        gen->is_lval = true;
        address = exp_gen(gen, l_exp);
        offset = 0;
        gen->is_lval = false;
    }

    return BinaryenStore(gen->module, TYPE_SIZE(l_exp->meta.type), offset, 0, address, 
                         value, meta_gen(&l_exp->meta));
}

static BinaryenExpressionRef
stmt_gen_return(gen_t *gen, ast_stmt_t *stmt)
{
    meta_t *arg_meta;
    ast_exp_t *arg_exp = stmt->u_ret.arg_exp;
    ast_id_t *ret_id = stmt->u_ret.ret_id;
    BinaryenExpressionRef value;

    if (arg_exp == NULL)
        return BinaryenReturn(gen->module, NULL);

    if (is_tuple_exp(arg_exp)) {
        int i;
        array_t *elem_exps = arg_exp->u_tup.exps;

        ASSERT1(is_tuple_id(ret_id), ret_id->kind);

        array_foreach(elem_exps, i) {
            ast_exp_t *elem_exp = array_get_exp(elem_exps, i);
            meta_t *elem_meta = &elem_exp->meta;
            ast_id_t *var_id;

            value = exp_gen(gen, elem_exp);
            var_id = array_get_id(&ret_id->u_tup.var_ids, i);

            ASSERT(var_id->idx >= 0);
            ASSERT(var_id->offset == 0);

            gen_add_instr(gen,
                BinaryenStore(gen->module, meta_size(elem_meta), var_id->offset, 0,
                    BinaryenGetLocal(gen->module, var_id->idx, BinaryenTypeInt32()),
                    value, meta_gen(elem_meta)));
        }

        return NULL;
    }

    arg_meta = &arg_exp->meta;
    value = exp_gen(gen, arg_exp);

    ASSERT(ret_id->idx >= 0);
    ASSERT(ret_id->offset == 0);

    return BinaryenStore(gen->module, meta_size(arg_meta), ret_id->offset, 0,
                         BinaryenGetLocal(gen->module, ret_id->idx, BinaryenTypeInt32()),
                         value, meta_gen(arg_meta));
}

static BinaryenExpressionRef
stmt_gen_ddl(gen_t *gen, ast_stmt_t *stmt)
{
    return NULL;
}

BinaryenExpressionRef
stmt_gen(gen_t *gen, ast_stmt_t *stmt)
{
    switch (stmt->kind) {
    case STMT_NULL:
        return BinaryenNop(gen->module);

    case STMT_EXP:
        return exp_gen(gen, stmt->u_exp.exp);

    case STMT_ASSIGN:
        return stmt_gen_assign(gen, stmt);

    case STMT_RETURN:
        return stmt_gen_return(gen, stmt);

    case STMT_DDL:
        return stmt_gen_ddl(gen, stmt);

    case STMT_IF:
    case STMT_LOOP:
    case STMT_SWITCH:
    case STMT_BLK:
    case STMT_CONTINUE:
    case STMT_BREAK:
    case STMT_GOTO:
    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }

    return NULL;
}

/* end of gen_stmt.c */
