/**
 * @file    ast_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "ast_blk.h"

#include "ast_id.h"

ast_id_t *
ast_id_new(id_kind_t kind, modifier_t mod, errpos_t *pos)
{
    ast_id_t *id = xcalloc(sizeof(ast_id_t));

    ast_node_init(id, pos);

    id->kind = kind;
    id->mod = mod;

    return id;
}

ast_id_t *
id_var_new(ast_exp_t *type_exp, ast_exp_t *id_exp, ast_exp_t *init_exp,
           errpos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_VAR, MOD_GLOBAL, pos);

    ASSERT(type_exp != NULL);
    ASSERT(id_exp != NULL);

    ast_node_init(id, pos);

    id->u_var.type_exp = type_exp;
    id->u_var.id_exp = id_exp;
    id->u_var.init_exp = init_exp;

    return id;
}

ast_id_t *
id_struct_new(char *name, array_t *fld_ids, errpos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_STRUCT, MOD_GLOBAL, pos);

    ASSERT(name != NULL);
    ASSERT(fld_ids != NULL);

    ast_node_init(id, pos);

    id->name = name;
    id->u_st.fld_ids = fld_ids;

    return id;
}

ast_id_t *
id_func_new(char *name, modifier_t mod, array_t *param_ids, array_t *ret_exps,
            ast_blk_t *blk, errpos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_FUNC, mod, pos);

    ASSERT(name != NULL);

    ast_node_init(id, pos);

    id->name = name;
    id->u_func.param_ids = param_ids;
    id->u_func.ret_exps = ret_exps;
    id->u_func.blk = blk;

    return id;
}

ast_id_t *
id_contract_new(char *name, ast_blk_t *blk, errpos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_CONTRACT, MOD_GLOBAL, pos);

    ASSERT(name != NULL);

    ast_node_init(id, pos);

    id->name = name;
    id->u_cont.blk = blk;

    return id;
}

ast_id_t *
ast_id_search(ast_blk_t *blk, int num, char *name)
{
    int i;

    if (blk == NULL)
        return NULL;

    do {
        // XXX: need to check siblings of root blk
        for (i = 0; i < array_size(&blk->ids); i++) {
            ast_id_t *id = array_item(&blk->ids, i, ast_id_t);

            ASSERT(id->name != NULL);
            ASSERT2(id->num != num, id->num, num);

            if (id->num < num && strcmp(id->name, name) == 0)
                return id;
        }
    } while ((blk = blk->up) != NULL);
}

void
ast_id_dump(ast_id_t *id, int indent)
{
}

/* end of ast_id.c */
