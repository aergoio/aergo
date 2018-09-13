/**
 * @file    ast_exp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_EXP_H
#define _AST_EXP_H

#include "common.h"

#include "location.h"
#include "list.h"
#include "ast_meta.h"

#define EXP_DECL                                                               \
    yypos_t pos

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif  /* _AST_EXP_T */

#ifndef _AST_VAR_T
#define _AST_VAR_T
typedef struct ast_var_s ast_var_t;
#endif  /* _AST_VAR_T */

typedef enum exp_type_e {
    EXP_OP          = 0,
    EXP_UNARY,
    EXP_SQL,
    EXP_COND,
    EXP_ID,
    EXP_LIT,
    EXP_NEW,
    EXP_MAX
} exp_type_t;

typedef enum op_e {
    OP_ASSIGN       = 0,
    OP_AND,
    OP_OR,
    OP_BIT_AND,
    OP_BIT_OR,
    OP_BIT_XOR,
    OP_EQ,
    OP_NE,
    OP_LT,
    OP_GT,
    OP_LE,
    OP_GE,
    OP_RSHIFT,
    OP_LSHIFT,
    OP_INC,
    OP_DEC,
    OP_NOT,
    OP_MAX
} op_t;

typedef enum sql_kind_e {
    SQL_QUERY       = 0,
    SQL_INSERT,
    SQL_UPDATE,
    SQL_DELETE,
    SQL_MAX
} sql_kind_t;

typedef struct exp_op_s {
    EXP_DECL;
    op_t op;
    ast_exp_t *left;
    ast_exp_t *right;
} exp_op_t;

typedef struct exp_id_s {
    EXP_DECL;
    ast_var_t *var;
} exp_id_t;

typedef struct exp_sql_s {
    EXP_DECL;
    sql_kind_t kind;
    char *sql;
} exp_sql_t;

typedef struct exp_lit_s {
    EXP_DECL;
    char *val;
    exp_meta_t meta;
} exp_lit_t;

typedef struct exp_new_s {
    EXP_DECL;
    list_t arg_l;
} exp_new_t;

struct ast_exp_s {
    EXP_DECL;
    exp_type_t type;
};

#endif /* _AST_EXP_H */
