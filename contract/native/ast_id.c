/**
 * @file    ast_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "ast_blk.h"

#include "ast_id.h"

ast_id_t *
ast_id_new(id_kind_t kind, modifier_t mod, char *name, errpos_t *pos)
{
    ast_id_t *id = xcalloc(sizeof(ast_id_t));

    ASSERT(name != NULL);

    ast_node_init(id, pos);

    id->kind = kind;
    id->mod = mod;
    id->name = name;

    return id;
}

ast_id_t *
id_var_new(char *name, errpos_t *pos)
{
    return ast_id_new(ID_VAR, MOD_GLOBAL, name, pos);
}

ast_id_t *
id_struct_new(char *name, array_t *fld_ids, errpos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_STRUCT, MOD_GLOBAL, name, pos);

    ASSERT(fld_ids != NULL);

    id->u_st.fld_ids = fld_ids;

    return id;
}

ast_id_t *
id_func_new(char *name, modifier_t mod, array_t *param_ids, array_t *ret_exps,
            ast_blk_t *blk, errpos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_FUNC, mod, name, pos);

    id->u_func.param_ids = param_ids;
    id->u_func.ret_exps = ret_exps;
    id->u_func.blk = blk;

    return id;
}

ast_id_t *
id_contract_new(char *name, ast_blk_t *blk, errpos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_CONTRACT, MOD_GLOBAL, name, pos);

    id->u_cont.blk = blk;

    return id;
}

ast_id_t *
ast_id_search_fld(ast_id_t *id, int num, char *name)
{
    array_t *fld_ids = NULL;

    // XXX: search function

    if (id->kind == ID_STRUCT)
        fld_ids = id->u_st.fld_ids;
    else if (id->kind == ID_CONTRACT && id->u_cont.blk != NULL)
        fld_ids = &id->u_cont.blk->ids;

    if (fld_ids != NULL) {
        int i;

        for (i = 0; i < array_size(fld_ids); i++) {
            ast_id_t *fld_id = array_item(fld_ids, i, ast_id_t);

            ASSERT(fld_id->name != NULL);
            ASSERT2(fld_id->num != num, fld_id->num, num);

            if (fld_id->num < num && strcmp(fld_id->name, name) == 0)
                return fld_id;
        }
    }

    return NULL;
}

ast_id_t *
ast_id_search_blk(ast_blk_t *blk, int num, char *name)
{
    int i;

    if (blk == NULL)
        return NULL;

    do {
        for (i = 0; i < array_size(&blk->ids); i++) {
            ast_id_t *id = array_item(&blk->ids, i, ast_id_t);

            ASSERT(id->name != NULL);
            ASSERT2(id->num != num, id->num, num);

            if (id->num < num && strcmp(id->name, name) == 0)
                return id;
        }
    } while ((blk = blk->up) != NULL);

    return NULL;
}

void
ast_id_add(array_t *ids, ast_id_t *new_id)
{
    int i;

    for (i = 0; i < array_size(ids); i++) {
        ast_id_t *id = array_item(ids, i, ast_id_t);

        if (strcmp(id->name, new_id->name) == 0) {
            ERROR(ERROR_DUPLICATED_ID, &new_id->pos, new_id->name);
            return;
        }
    }

    array_add(ids, new_id);
}

void
ast_id_merge(array_t *ids, array_t *new_ids)
{
    int i;

    for (i = 0; i < array_size(new_ids); i++) {
        ast_id_add(ids, array_item(new_ids, i, ast_id_t));
    }
}

void
ast_id_dump(ast_id_t *id, int indent)
{
}

/* end of ast_id.c */
