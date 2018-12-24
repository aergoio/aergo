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
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    ast_exp_t *r_exp = stmt->u_assign.r_exp;
    ast_id_t *id;
    meta_t *meta;
    BinaryenExpressionRef var_exp, val_exp;

    if (is_tuple_exp(l_exp)) {
        ASSERT(!is_tuple_exp(r_exp));
        ERROR(ERROR_NOT_SUPPORTED, &stmt->pos);
        return NULL;
    }

    meta = &l_exp->meta;

    var_exp = exp_gen(gen, l_exp);

    ASSERT2(BinaryenExpressionGetId(var_exp) == BinaryenConstId(),
            BinaryenExpressionGetId(var_exp), BinaryenConstId());

    val_exp = exp_gen(gen, r_exp);

    id = l_exp->id;
    ASSERT(id != NULL);

    /* XXX */
#if 0
    if (id->idx >= 0)
        /* var_exp == index of variable */
        return BinaryenSetLocal(gen->module, BinaryenConstGetValueI32(var_exp), val_exp);
    else
        /* var_exp == offset of variable */
        return BinaryenStore(gen->module, meta_size(meta),
                             BinaryenConstGetValueI32(var_exp), 0,
                             gen_i32(gen, id->addr), val_exp, meta_gen(gen, meta));
#endif
    return NULL;
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

            /* XXX
            gen_add_instr(gen,
                BinaryenStore(gen->module, meta_size(elem_meta), elem_meta->offset, 0,
                    BinaryenGetLocal(gen->module, ret_id->idx, BinaryenTypeInt32()),
                    value, meta_gen(gen, elem_meta)));
                    */
        }

        return NULL;
    }

    arg_meta = &arg_exp->meta;
    value = exp_gen(gen, arg_exp);
    ret_id = array_get_id(ret_ids, 0);

    /* XXX
    return BinaryenStore(gen->module, meta_size(arg_meta), arg_meta->offset, 0,
                         BinaryenGetLocal(gen->module, ret_id->idx, BinaryenTypeInt32()),
                         value, meta_gen(gen, arg_meta));
                         */
    return NULL;
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
