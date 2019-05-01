/**
 * @file    gen_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "ast_blk.h"
#include "ir_md.h"
#include "gen_exp.h"
#include "gen_util.h"
#include "syslib.h"

#include "gen_stmt.h"

static BinaryenExpressionRef
stmt_gen_exp(gen_t *gen, ast_stmt_t *stmt)
{
    ast_exp_t *exp = stmt->u_exp.exp;

    if (is_call_exp(exp) && !is_void_meta(&exp->meta))
        return BinaryenDrop(gen->module, exp_gen(gen, exp, NULL));

    return exp_gen(gen, exp, NULL);
}

static BinaryenExpressionRef
stmt_gen_assign(gen_t *gen, ast_stmt_t *stmt)
{
    BinaryenExpressionRef value, statement;

    if (stmt->u_assign.l_exp->id != NULL && is_map_meta(&stmt->u_assign.l_exp->id->meta))
        /* TODO: If the type of identifier is map, lvalue and rvalue must be combined
         * into a call expression */
        return NULL;

    value = exp_gen(gen, stmt->u_assign.r_exp, NULL);
    if (value == NULL)
        return NULL;

    gen->is_lval = true;

    statement = exp_gen(gen, stmt->u_assign.l_exp, value);

    gen->is_lval = false;

    return statement;
}

static BinaryenExpressionRef
stmt_gen_if(gen_t *gen, ast_stmt_t *stmt)
{
    ast_exp_t *cond_exp = stmt->u_if.cond_exp;
    ast_blk_t *if_blk = stmt->u_if.if_blk;

    /* All user-defined "if" statements are transformed into basic blocks, so stmt_gen_if() is for
     * internal use only. (see stmt_check_if()) */

    ASSERT(cond_exp != NULL);
    ASSERT(if_blk != NULL);
    ASSERT1(is_empty_vector(&if_blk->ids), vector_size(&if_blk->ids));
    ASSERT1(vector_size(&if_blk->stmts) == 1, vector_size(&if_blk->stmts));
    ASSERT(stmt->u_if.else_blk == NULL);
    ASSERT1(is_empty_vector(&stmt->u_if.elif_stmts), vector_size(&stmt->u_if.elif_stmts));

    return BinaryenIf(gen->module, exp_gen(gen, cond_exp, NULL),
                      stmt_gen(gen, vector_get_stmt(&if_blk->stmts, 0)), NULL);
}

static BinaryenExpressionRef
stmt_gen_ddl(gen_t *gen, ast_stmt_t *stmt)
{
    /* TODO */
    return NULL;
}

static BinaryenExpressionRef
stmt_gen_pragma(gen_t *gen, ast_stmt_t *stmt)
{
    ast_exp_t *val_exp = stmt->u_pragma.val_exp;
    ir_md_t *md = gen->md;
    BinaryenExpressionRef condition, description;

    condition = exp_gen(gen, val_exp, NULL);

    if (stmt->u_pragma.desc_exp != NULL)
        description = exp_gen(gen, stmt->u_pragma.desc_exp, NULL);
    else
        description = i32_gen(gen, 0);

    return syslib_gen(gen, FN_ASSERT, 6, condition,
                      i32_gen(gen, sgmt_add_str(&md->sgmt, stmt->u_pragma.val_str)), description,
                      i32_gen(gen, val_exp->pos.first_line), i32_gen(gen, val_exp->pos.first_col),
                      i32_gen(gen, val_exp->pos.first_offset));
}

BinaryenExpressionRef
stmt_gen(gen_t *gen, ast_stmt_t *stmt)
{
    switch (stmt->kind) {
    case STMT_EXP:
        return stmt_gen_exp(gen, stmt);

    case STMT_ASSIGN:
        return stmt_gen_assign(gen, stmt);

    case STMT_IF:
        return stmt_gen_if(gen, stmt);

    case STMT_RETURN:
        return BinaryenReturn(gen->module, exp_gen(gen, stmt->u_ret.arg_exp, NULL));

    case STMT_DDL:
        return stmt_gen_ddl(gen, stmt);

    case STMT_PRAGMA:
        return stmt_gen_pragma(gen, stmt);

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }

    return NULL;
}

/* end of gen_stmt.c */
