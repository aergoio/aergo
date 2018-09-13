/**
 * @file    ast_stmt.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_STMT_H
#define _AST_STMT_H

#include "common.h"

#include "location.h"
#include "list.h"

#define STMT_DECL                                                              \
    yypos_t pos

#ifndef _AST_BLK_T
#define _AST_BLK_T
typedef struct ast_blk_s ast_blk_t;
#endif  /* _AST_BLK_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif  /* _AST_EXP_T */

typedef enum stmt_type_e {
    STMT_EXPR       = 0,
    STMT_IF,
    STMT_FOR,
    STMT_SWITCH,
    STMT_CONTINUE,
    STMT_BREAK,
    STMT_RETURN,
    STMT_DDL,
    STMT_BLK,
    STMT_MAX
} stmt_type_t;

typedef struct stmt_exp_s {
    STMT_DECL;
    ast_exp_t *exp;
} stmt_exp_t;

typedef struct stmt_if_s {
    STMT_DECL;
    ast_exp_t *exp;
    ast_blk_t *blk;
    list_t else_l;
} stmt_if_t;

typedef struct stmt_for_s {
    STMT_DECL;
    ast_exp_t *init_exp;
    ast_exp_t *check_exp;
    ast_exp_t *inc_exp;
    ast_blk_t *blk;
} stmt_for_t;

typedef struct stmt_switch_s {
    STMT_DECL;
    ast_exp_t *exp;
    list_t case_l;
    ast_blk_t *dflt;
} stmt_switch_t;

typedef struct stmt_return_s {
    STMT_DECL;
    list_t exp_l;
} stmt_return_t;

typedef struct stmt_ddl_s {
    STMT_DECL;
    char *ddl;
} stmt_ddl_t;

typedef struct stmt_blk_s {
    STMT_DECL;
    ast_blk_t *blk;
} stmt_blk_t;

typedef struct ast_stmt_s {
    STMT_DECL;
    stmt_type_t type;
} ast_stmt_t;

#endif /* _AST_STMT_H */
