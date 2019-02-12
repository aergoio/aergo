/**
 * @file    ast_imp.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_IMP_H
#define _AST_IMP_H

#include "common.h"

#include "ast.h"

#define imp_add                     vector_add_last

#ifndef _AST_IMP_T
#define _AST_IMP_T
typedef struct ast_imp_s ast_imp_t;
#endif /* ! _AST_IMP_T */

struct ast_imp_s {
    char *path;

    AST_NODE_DECL;
};

ast_imp_t *imp_new(char *path, src_pos_t *pos);

#endif /* ! _AST_IMP_H */
