/**
 * @file    gen_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "gen_exp.h"
#include "gen_blk.h"
#include "gen_meta.h"

#include "gen_stmt.h"

static BinaryenExpressionRef
stmt_gen_assign(gen_t *gen, ast_stmt_t *stmt)
{
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    ast_id_t *id = l_exp->id;
    BinaryenExpressionRef value;

    ASSERT(id != NULL);

    value = exp_gen(gen, stmt->u_assign.r_exp);

    if (is_primitive_type(&id->meta) && !is_array_type(&id->meta))
        return BinaryenSetLocal(gen->module, 0, value);
    else
        return BinaryenStore(gen->module, meta_size(&id->meta), id->offset, 0,
                             BinaryenConst(gen->module, BinaryenLiteralInt32(id->addr)),
                             value, meta_gen(gen, &id->meta));
}

static BinaryenExpressionRef
stmt_gen_if(gen_t *gen, ast_stmt_t *stmt)
{
    return NULL;
}

static BinaryenExpressionRef
stmt_gen_loop(gen_t *gen, ast_stmt_t *stmt)
{
    return NULL;
}

static BinaryenExpressionRef
stmt_gen_switch(gen_t *gen, ast_stmt_t *stmt)
{
    return NULL;
}

static BinaryenExpressionRef
stmt_gen_return(gen_t *gen, ast_stmt_t *stmt)
{
    return BinaryenNop(gen->module);
}

static BinaryenExpressionRef
stmt_gen_jump(gen_t *gen, ast_stmt_t *stmt)
{
    return NULL;
}

static BinaryenExpressionRef
stmt_gen_goto(gen_t *gen, ast_stmt_t *stmt)
{
    return NULL;
}

static BinaryenExpressionRef
stmt_gen_ddl(gen_t *gen, ast_stmt_t *stmt)
{
    return NULL;
}

static BinaryenExpressionRef
stmt_gen_blk(gen_t *gen, ast_stmt_t *stmt)
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

    case STMT_IF:
        return stmt_gen_if(gen, stmt);

    case STMT_LOOP:
        return stmt_gen_loop(gen, stmt);

    case STMT_SWITCH:
        return stmt_gen_switch(gen, stmt);

    case STMT_RETURN:
        return stmt_gen_return(gen, stmt);

    case STMT_CONTINUE:
        return stmt_gen_jump(gen, stmt);

    case STMT_BREAK:
        /* TODO: because of switch statement, we will handle this later */
        return NULL;

    case STMT_GOTO:
        return stmt_gen_goto(gen, stmt);

    case STMT_DDL:
        return stmt_gen_ddl(gen, stmt);

    case STMT_BLK:
        return stmt_gen_blk(gen, stmt);

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }

    return NULL;
}

/* end of gen_stmt.c */
