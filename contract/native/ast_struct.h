/**
 * @file    ast_struct.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_STRUCT_H
#define _AST_STRUCT_H

#include "common.h"

#include "ast.h"

#ifndef _AST_STRUCT_T
#define _AST_STRUCT_T
typedef struct ast_struct_s ast_struct_t;
#endif /* ! _AST_STRUCT_T */

struct ast_struct_s {
    AST_NODE_DECL;

    char *name;
    list_t *field_l;
};

ast_struct_t *ast_struct_new(char *name, list_t *field_l, errpos_t *pos);

void ast_struct_dump(ast_struct_t *struc, int indent);

#endif /* ! _AST_STRUCT_H */
