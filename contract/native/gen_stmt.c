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
    ast_id_t *id;
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    BinaryenExpressionRef value;

    id = l_exp->id;
    ASSERT(id != NULL);

    value = exp_gen(gen, stmt->u_assign.r_exp);

    if (is_id_ref_exp(l_exp))
        return BinaryenSetGlobal(gen->module, id->name, value);

    if (is_local_ref_exp(l_exp))
        return BinaryenSetLocal(gen->module, id->idx, value);

    ASSERT1(is_stack_ref_exp(l_exp), l_exp->kind);

    return BinaryenStore(gen->module, meta_size(&id->meta), id->offset, 0,
                         gen_i32(gen, id->addr), value, meta_gen(gen, &id->meta));
}

static BinaryenExpressionRef
stmt_gen_return(gen_t *gen, ast_stmt_t *stmt)
{
    meta_t *arg_meta;
    ast_exp_t *arg_exp = stmt->u_ret.arg_exp;
    ast_id_t *ret_id;
    array_t *ret_ids = stmt->u_ret.ret_ids;
    BinaryenExpressionRef value;

    if (arg_exp == NULL)
        return BinaryenReturn(gen->module, NULL);

    if (is_tuple_exp(arg_exp)) {
        int i;
        array_t *elem_exps = arg_exp->u_tup.exps;

        for (i = 0; i < array_size(elem_exps); i++) {
            ast_exp_t *elem_exp = array_get_exp(elem_exps, i);
            meta_t *elem_meta = &elem_exp->meta;

            value = exp_gen(gen, elem_exp);
            ret_id = array_get_id(ret_ids, i);

            gen_add_instr(gen,
                BinaryenStore(gen->module, meta_size(elem_meta), ret_id->offset, 0,
                    BinaryenGetLocal(gen->module, ret_id->idx, BinaryenTypeInt32()),
                    value, meta_gen(gen, elem_meta)));
        }

        return NULL;
    }

    arg_meta = &arg_exp->meta;
    value = exp_gen(gen, arg_exp);
    ret_id = array_get_id(ret_ids, 0);

    return BinaryenStore(gen->module, meta_size(arg_meta), ret_id->offset, 0,
                         BinaryenGetLocal(gen->module, ret_id->idx, BinaryenTypeInt32()),
                         value, meta_gen(gen, arg_meta));
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
