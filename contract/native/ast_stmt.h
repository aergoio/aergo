/**
 * @file    ast_stmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_STMT_H
#define _AST_STMT_H

#include "common.h"

#include "ast.h"

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif /* ! _AST_BLK_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

#ifndef _AST_STMT_T
#define _AST_STMT_T
typedef struct ast_stmt_s ast_stmt_t;
#endif /* ! _AST_STMT_T */

typedef enum stmt_kind_e {
    STMT_NULL           = 0,
    STMT_EXP,
    STMT_IF,
    STMT_FOR,
    STMT_SWITCH,
    STMT_CASE,
    STMT_CONTINUE,
    STMT_BREAK,
    STMT_RETURN,
    STMT_DDL,
    STMT_BLK,
    STMT_MAX
} stmt_kind_t;

typedef struct stmt_exp_s {
    ast_exp_t *exp;
} stmt_exp_t;

typedef struct stmt_if_s {
    ast_exp_t *cmp_exp;
    ast_blk_t *if_blk;
    ast_blk_t *else_blk;
    list_t elsif_l;
} stmt_if_t;

typedef struct stmt_for_s {
    list_t *init_l;
    ast_exp_t *init_exp;
    ast_exp_t *check_exp;
    ast_exp_t *inc_exp;
    ast_blk_t *blk;
} stmt_for_t;

typedef struct stmt_switch_s {
    ast_exp_t *cmp_exp;
    list_t *case_l;
} stmt_switch_t;

typedef struct stmt_case_s {
    ast_exp_t *cmp_exp;
    list_t *stmt_l;
} stmt_case_t;

typedef struct stmt_return_s {
    ast_exp_t *exp;
} stmt_return_t;

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

    union {
        stmt_exp_t u_exp;
        stmt_if_t u_if;
        stmt_for_t u_for;
        stmt_switch_t u_sw;
        stmt_case_t u_case;
        stmt_return_t u_ret;
        stmt_ddl_t u_ddl;
        stmt_blk_t u_blk;
    };
};

ast_stmt_t *ast_stmt_new(stmt_kind_t kind, errpos_t *pos);

ast_stmt_t *stmt_exp_new(ast_exp_t *exp, errpos_t *pos);
ast_stmt_t *stmt_if_new(ast_exp_t *cmp_exp, ast_blk_t *if_blk, errpos_t *pos);
ast_stmt_t *stmt_for_new(ast_exp_t *init_exp, ast_exp_t *check_exp,
                         ast_exp_t *inc_exp, ast_blk_t *blk, errpos_t *pos);
ast_stmt_t *stmt_switch_new(ast_exp_t *cmp_exp, list_t *case_l, errpos_t *pos);
ast_stmt_t *stmt_case_new(ast_exp_t *cmp_exp, list_t *stmt_l, errpos_t *pos);
ast_stmt_t *stmt_return_new(ast_exp_t *exp, errpos_t *pos);
ast_stmt_t *stmt_ddl_new(ddl_kind_t kind, char *ddl, errpos_t *pos);
ast_stmt_t *stmt_blk_new(ast_blk_t *blk, errpos_t *pos);

void ast_stmt_dump(ast_stmt_t *stmt);

#endif /* ! _AST_STMT_H */
