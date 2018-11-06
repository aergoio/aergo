/**
 * @file    gen_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_exp.h"
#include "gen_blk.h"

#include "gen_stmt.h"

static void
stmt_gen_assign(gen_t *gen, ast_stmt_t *stmt)
{
}

static void
stmt_gen_if(gen_t *gen, ast_stmt_t *stmt)
{
}

static void
stmt_gen_loop(gen_t *gen, ast_stmt_t *stmt)
{
}

static void
stmt_gen_switch(gen_t *gen, ast_stmt_t *stmt)
{
}

static void
stmt_gen_return(gen_t *gen, ast_stmt_t *stmt)
{
}

static void
stmt_gen_jump(gen_t *gen, ast_stmt_t *stmt)
{
}

static void
stmt_gen_goto(gen_t *gen, ast_stmt_t *stmt)
{
}

static void
stmt_gen_ddl(gen_t *gen, ast_stmt_t *stmt)
{
}

static void
stmt_gen_blk(gen_t *gen, ast_stmt_t *stmt)
{
}

void
stmt_gen(gen_t *gen, ast_stmt_t *stmt)
{
    switch (stmt->kind) {
    case STMT_NULL:
        break;

    case STMT_EXP:
        exp_gen(gen, stmt->u_exp.exp);
        break;

    case STMT_ASSIGN:
        stmt_gen_assign(gen, stmt);
        break;

    case STMT_IF:
        stmt_gen_if(gen, stmt);
        break;

    case STMT_LOOP:
        stmt_gen_loop(gen, stmt);
        break;

    case STMT_SWITCH:
        stmt_gen_switch(gen, stmt);
        break;

    case STMT_RETURN:
        stmt_gen_return(gen, stmt);
        break;

    case STMT_CONTINUE:
        stmt_gen_jump(gen, stmt);
        break;

    case STMT_BREAK:
        /* TODO: because of switch statement, we will handle this later */
        break;

    case STMT_GOTO:
        stmt_gen_goto(gen, stmt);
        break;

    case STMT_DDL:
        stmt_gen_ddl(gen, stmt);
        break;

    case STMT_BLK:
        stmt_gen_blk(gen, stmt);
        break;

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }
}

/* end of gen_stmt.c */
