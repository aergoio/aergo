/**
 * @file    ast_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_exp.h"
#include "ast_blk.h"

#include "ast_stmt.h"

static ast_stmt_t *
ast_stmt_new(stmt_kind_t kind, src_pos_t *pos)
{
    ast_stmt_t *stmt = xcalloc(sizeof(ast_stmt_t));

    ast_node_init(stmt, *pos);

    stmt->kind = kind;

    return stmt;
}

ast_stmt_t *
stmt_new_null(src_pos_t *pos)
{
    return ast_stmt_new(STMT_NULL, pos);
}

ast_stmt_t *
stmt_new_exp(ast_exp_t *exp, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_EXP, pos);

    stmt->u_exp.exp = exp;

    return stmt;
}

ast_stmt_t *
stmt_new_assign(ast_exp_t *l_exp, ast_exp_t *r_exp, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_ASSIGN, pos);

    stmt->u_assign.l_exp = l_exp;
    stmt->u_assign.r_exp = r_exp;

    return stmt;
}

ast_stmt_t *
stmt_new_if(ast_exp_t *cond_exp, ast_blk_t *if_blk, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_IF, pos);

    stmt->u_if.cond_exp = cond_exp;
    stmt->u_if.if_blk = if_blk;
    stmt->u_if.else_blk = NULL;
    array_init(&stmt->u_if.elif_stmts);

    return stmt;
}

ast_stmt_t *
stmt_new_loop(loop_kind_t kind, ast_exp_t *cond_exp, ast_exp_t *loop_exp,
              ast_blk_t *blk, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_LOOP, pos);

    stmt->u_loop.kind = kind;
    stmt->u_loop.cond_exp = cond_exp;
    stmt->u_loop.loop_exp = loop_exp;
    stmt->u_loop.blk = blk;

    if (stmt->u_loop.blk != NULL)
        stmt->u_loop.blk->kind = BLK_LOOP;

    return stmt;
}

ast_stmt_t *
stmt_new_switch(ast_exp_t *cond_exp, ast_blk_t *blk, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_SWITCH, pos);

    stmt->u_sw.cond_exp = cond_exp;
    stmt->u_sw.blk = blk;

    return stmt;
}

ast_stmt_t *
stmt_new_case(ast_exp_t *val_exp, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_CASE, pos);

    stmt->u_case.val_exp = val_exp;

    return stmt;
}

ast_stmt_t *
stmt_new_return(ast_exp_t *arg_exp, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_RETURN, pos);

    stmt->u_ret.arg_exp = arg_exp;

    return stmt;
}

ast_stmt_t *
stmt_new_goto(char *label, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_GOTO, pos);

    stmt->u_goto.label = label;

    return stmt;
}

ast_stmt_t *
stmt_new_jump(stmt_kind_t kind, ast_exp_t *cond_exp, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(kind, pos);

    stmt->u_jump.cond_exp = cond_exp;

    return stmt;
}

ast_stmt_t *
stmt_new_ddl(char *ddl, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_DDL, pos);

    stmt->u_ddl.ddl = ddl;

    return stmt;
}

ast_stmt_t *
stmt_new_blk(ast_blk_t *blk, src_pos_t *pos)
{
    ast_stmt_t *stmt = ast_stmt_new(STMT_BLK, pos);

    stmt->u_blk.blk = blk;

    return stmt;
}

ast_stmt_t *
stmt_make_assign(ast_id_t *var_id, ast_exp_t *val_exp)
{
    ast_exp_t *var_exp;

    if (is_tuple_id(var_id)) {
        int i;
        array_t *elem_exps = array_new();
        ast_exp_t *id_exp;

        /* Since the number of elements in "val_exp" may be smaller than
         * the number of elements in "var_id", it is made as a tuple expression
         * for asymmetry assignment processing */
        array_foreach(var_id->u_tup.elem_ids, i) {
            ast_id_t *elem_id = array_get_id(var_id->u_tup.elem_ids, i);

            id_exp = exp_new_id(elem_id->name, &elem_id->pos);

            id_exp->id = elem_id;
            meta_copy(&id_exp->meta, &elem_id->meta);

            array_add_last(elem_exps, id_exp);
        }

        var_exp = exp_new_tuple(elem_exps, &val_exp->pos);
    }
    /*
    else if (is_return_id(var_id)) {
        var_exp = exp_new_local(TYPE_INT32, var_id->idx);
    }
    */
    else {
        var_exp = exp_new_id(var_id->name, &var_id->pos);

        var_exp->id = var_id;
        meta_copy(&var_exp->meta, &var_id->meta);
    }

    return stmt_new_assign(var_exp, val_exp, &val_exp->pos);
}

/* end of ast_stmt.c */
