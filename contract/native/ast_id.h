/**
 * @file    ast_id.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_ID_H
#define _AST_ID_H

#include "common.h"

#include "location.h"
#include "ast_meta.h"

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif  /* _AST_ID_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif  /* _AST_EXP_T */

struct ast_id_s {
    char *name;
    ast_meta_t meta;
    ast_exp_t *init_exp;

    yypos_t pos;
};

#endif /* _AST_ID_H */
