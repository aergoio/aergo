/**
 * @file    ast_stmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_STMT_H
#define _AST_STMT_H

#include "common.h"

#include "ast.h"

#define is_null_stmt(stmt)          ((stmt)->kind == STMT_NULL)
#define is_exp_stmt(stmt)           ((stmt)->kind == STMT_EXP)
#define is_assign_stmt(stmt)        ((stmt)->kind == STMT_ASSIGN)
#define is_if_stmt(stmt)            ((stmt)->kind == STMT_IF)
#define is_loop_stmt(stmt)          ((stmt)->kind == STMT_LOOP)
#define is_switch_stmt(stmt)        ((stmt)->kind == STMT_SWITCH)
#define is_case_stmt(stmt)          ((stmt)->kind == STMT_CASE)
#define is_continue_stmt(stmt)      ((stmt)->kind == STMT_CONTINUE)
#define is_break_stmt(stmt)         ((stmt)->kind == STMT_BREAK)
#define is_return_stmt(stmt)        ((stmt)->kind == STMT_RETURN)
#define is_goto_stmt(stmt)          ((stmt)->kind == STMT_GOTO)
#define is_ddl_stmt(stmt)           ((stmt)->kind == STMT_DDL)
#define is_blk_stmt(stmt)           ((stmt)->kind == STMT_BLK)

#define stmt_add_first              array_add_first
#define stmt_add_last               array_add_last

#ifndef _AST_STMT_T
#define _AST_STMT_T
typedef struct ast_stmt_s ast_stmt_t;
#endif /* ! _AST_STMT_T */

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

typedef struct stmt_exp_s {
    ast_exp_t *exp;
} stmt_exp_t;

typedef struct stmt_assign_s {
    ast_exp_t *l_exp;
    ast_exp_t *r_exp;
} stmt_assign_t;

typedef struct stmt_if_s {
    ast_exp_t *cond_exp;
    ast_blk_t *if_blk;
    ast_blk_t *else_blk;
    array_t elif_stmts;
} stmt_if_t;

typedef struct stmt_loop_s {
    loop_kind_t kind;

    array_t *init_ids;
    ast_stmt_t *init_stmt;
    ast_exp_t *cond_exp;
    ast_exp_t *loop_exp;

    ast_blk_t *blk;
} stmt_loop_t;

typedef struct stmt_switch_s {
    ast_exp_t *cond_exp;
    array_t *case_stmts;
} stmt_switch_t;

typedef struct stmt_case_s {
    ast_exp_t *val_exp;
    array_t *stmts;
} stmt_case_t;

typedef struct stmt_return_s {
    ast_exp_t *arg_exp;
} stmt_return_t;

typedef struct stmt_goto_s {
    char *label;
} stmt_goto_t;

typedef struct stmt_ddl_s {
    char *ddl;
} stmt_ddl_t;

typedef struct stmt_blk_s {
    ast_blk_t *blk;
} stmt_blk_t;

struct ast_stmt_s {
    AST_NODE_DECL;

    stmt_kind_t kind;
    char *label;

    union {
        stmt_exp_t u_exp;
        stmt_assign_t u_assign;
        stmt_if_t u_if;
        stmt_loop_t u_loop;
        stmt_switch_t u_sw;
        stmt_case_t u_case;
        stmt_return_t u_ret;
        stmt_goto_t u_goto;
        stmt_ddl_t u_ddl;
        stmt_blk_t u_blk;
    };
};

ast_stmt_t *stmt_new_null(src_pos_t *pos);
ast_stmt_t *stmt_new_exp(ast_exp_t *exp, src_pos_t *pos);
ast_stmt_t *stmt_new_assign(ast_exp_t *l_exp, ast_exp_t *r_exp, src_pos_t *pos);
ast_stmt_t *stmt_new_if(ast_exp_t *cond_exp, ast_blk_t *if_blk, src_pos_t *pos);
ast_stmt_t *stmt_new_loop(loop_kind_t kind, ast_exp_t *cond_exp, ast_exp_t *loop_exp, 
                          ast_blk_t *blk, src_pos_t *pos);
ast_stmt_t *stmt_new_switch(ast_exp_t *cond_exp, array_t *case_stmts, src_pos_t *pos);
ast_stmt_t *stmt_new_case(ast_exp_t *val_exp, array_t *stmts, src_pos_t *pos);
ast_stmt_t *stmt_new_return(ast_exp_t *arg_exp, src_pos_t *pos);
ast_stmt_t *stmt_new_goto(char *label, src_pos_t *pos);
ast_stmt_t *stmt_new_jump(stmt_kind_t kind, src_pos_t *pos);
ast_stmt_t *stmt_new_ddl(char *ddl, src_pos_t *pos);
ast_stmt_t *stmt_new_blk(ast_blk_t *blk, src_pos_t *pos);

void ast_stmt_dump(ast_stmt_t *stmt, int indent);

#endif /* ! _AST_STMT_H */
