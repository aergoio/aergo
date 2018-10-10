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
#define is_if_stmt(stmt)            ((stmt)->kind == STMT_IF)
#define is_loop_stmt(stmt)          ((stmt)->kind == STMT_LOOP)
#define is_switch_stmt(stmt)        ((stmt)->kind == STMT_SWITCH)
#define is_case_stmt(stmt)          ((stmt)->kind == STMT_CASE)
#define is_cont_stmt(stmt)          ((stmt)->kind == STMT_CONTINUE)
#define is_break_stmt(stmt)         ((stmt)->kind == STMT_BREAK)
#define is_return_stmt(stmt)        ((stmt)->kind == STMT_RETURN)
#define is_goto_stmt(stmt)          ((stmt)->kind == STMT_GOTO)
#define is_ddl_stmt(stmt)           ((stmt)->kind == STMT_DDL)
#define is_blk_stmt(stmt)           ((stmt)->kind == STMT_BLK)

#define stmt_add_first              array_add_first
#define stmt_add_last               array_add_last

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

#ifndef _AST_STMT_T
#define _AST_STMT_T
typedef struct ast_stmt_s ast_stmt_t;
#endif /* ! _AST_STMT_T */

typedef enum stmt_kind_e {
    STMT_NULL           = 0,
    STMT_EXP,
    STMT_IF,
    STMT_LOOP,
    STMT_SWITCH,
    STMT_CASE,
    STMT_CONTINUE,
    STMT_BREAK,
    STMT_RETURN,
    STMT_GOTO,
    STMT_DDL,
    STMT_BLK,
    STMT_MAX
} stmt_kind_t;

typedef struct stmt_exp_s {
    ast_exp_t *exp;
} stmt_exp_t;

typedef struct stmt_if_s {
    ast_exp_t *cond_exp;
    ast_blk_t *if_blk;
    ast_blk_t *else_blk;
    array_t elif_stmts;
} stmt_if_t;

typedef enum loop_kind_e {
    LOOP_FOR            = 0,
    LOOP_EACH,
    LOOP_MAX
} loop_kind_t;

typedef struct stmt_loop_s {
    loop_kind_t kind;

    array_t *init_ids;
    ast_exp_t *init_exp;
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

typedef enum ddl_kind_e {
    DDL_CREATE_TBL      = 0,
    DDL_DROP_TBL,
    DDL_CREATE_IDX,
    DDL_DROP_IDX,
    DDL_MAX
} ddl_kind_t;

typedef struct stmt_ddl_s {
    ddl_kind_t kind;
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

ast_stmt_t *stmt_null_new(src_pos_t *pos);
ast_stmt_t *stmt_exp_new(ast_exp_t *exp, src_pos_t *pos);
ast_stmt_t *stmt_if_new(ast_exp_t *cond_exp, ast_blk_t *if_blk, src_pos_t *pos);
ast_stmt_t *stmt_loop_new(loop_kind_t kind, ast_exp_t *cond_exp, 
                          ast_exp_t *loop_exp, ast_blk_t *blk, src_pos_t *pos);
ast_stmt_t *stmt_switch_new(ast_exp_t *cond_exp, array_t *case_stmts,
                            src_pos_t *pos);
ast_stmt_t *stmt_case_new(ast_exp_t *val_exp, array_t *stmts, src_pos_t *pos);
ast_stmt_t *stmt_return_new(ast_exp_t *arg_exp, src_pos_t *pos);
ast_stmt_t *stmt_goto_new(char *label, src_pos_t *pos);
ast_stmt_t *stmt_jump_new(stmt_kind_t kind, src_pos_t *pos);
ast_stmt_t *stmt_ddl_new(ddl_kind_t kind, char *ddl, src_pos_t *pos);
ast_stmt_t *stmt_blk_new(ast_blk_t *blk, src_pos_t *pos);

void ast_stmt_dump(ast_stmt_t *stmt, int indent);

#endif /* ! _AST_STMT_H */
