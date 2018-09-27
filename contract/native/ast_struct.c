/**
 * @file    ast_struct.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_struct.h"

ast_struct_t *
ast_struct_new(char *name, list_t *field_l, errpos_t *pos)
{
    ast_struct_t *struc = xmalloc(sizeof(ast_struct_t));

    ast_node_init(struc, pos);

    struc->name = name;
    struc->field_l = field_l;

    return struc;
}

void
ast_struct_dump(ast_struct_t *struc, int indent)
{
}

/* end of ast_struct.c */
