/**
 * @file    ast_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "ast_blk.h"

#include "ast_id.h"

static ast_id_t *
ast_id_new(id_kind_t kind, modifier_t mod, char *name, src_pos_t *pos)
{
    ast_id_t *id = xcalloc(sizeof(ast_id_t));

    ASSERT(name != NULL);

    ast_node_init(id, pos);

    id->kind = kind;
    id->mod = mod;
    id->name = name;

    meta_init(&id->meta, &id->pos);

    return id;
}

ast_id_t *
id_var_new(char *name, src_pos_t *pos)
{
    return ast_id_new(ID_VAR, MOD_GLOBAL, name, pos);
}

ast_id_t *
id_struct_new(char *name, array_t *fld_ids, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_STRUCT, MOD_GLOBAL, name, pos);

    ASSERT(fld_ids != NULL);

    id->u_st.fld_ids = fld_ids;

    return id;
}

ast_id_t *
id_func_new(char *name, modifier_t mod, array_t *param_ids, array_t *ret_exps,
            ast_blk_t *blk, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_FUNC, mod, name, pos);

    id->u_func.param_ids = param_ids;
    id->u_func.ret_exps = ret_exps;
    id->u_func.blk = blk;

    return id;
}

ast_id_t *
id_contract_new(char *name, ast_blk_t *blk, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_CONTRACT, MOD_GLOBAL, name, pos);

    id->u_cont.blk = blk;

    return id;
}

ast_id_t *
id_search_name(ast_blk_t *blk, int num, char *name)
{
    int i;

    ASSERT(name != NULL);

    if (blk == NULL)
        return NULL;

    do {
        for (i = 0; i < array_size(&blk->ids); i++) {
            ast_id_t *id = array_item(&blk->ids, i, ast_id_t);

            if (id->num < num && strcmp(id->name, name) == 0)
                return id;
        }
    } while ((blk = blk->up) != NULL);

    return NULL;
}

ast_id_t *
id_search_fld(ast_id_t *id, char *name)
{
    int i;
    array_t *fld_ids = NULL;

    ASSERT(id != NULL);
    ASSERT(name != NULL);

    if (is_struct_id(id))
        fld_ids = id->u_st.fld_ids;
    else if (is_contract_id(id) && id->u_cont.blk != NULL)
        fld_ids = &id->u_cont.blk->ids;
    else
        return NULL;

    ASSERT(fld_ids != NULL);

    for (i = 0; i < array_size(fld_ids); i++) {
        ast_id_t *fld_id = array_item(fld_ids, i, ast_id_t);

        if (!is_local_id(fld_id) && strcmp(fld_id->name, name) == 0)
            return fld_id;
    }

    return NULL;
}

ast_id_t *
id_search_param(ast_id_t *id, char *name)
{
    int i;

    ASSERT(id != NULL);
    ASSERT1(is_func_id(id), id->kind);
    ASSERT(name != NULL);

    for (i = 0; i < array_size(id->u_func.param_ids); i++) {
        ast_id_t *param_id = array_item(id->u_func.param_ids, i, ast_id_t);

        if (strcmp(param_id->name, name) == 0)
            return param_id;
    }

    return NULL;
}

void
id_add(array_t *ids, int idx, ast_id_t *new_id)
{
    int i;

    if (new_id == NULL)
        return;

    for (i = 0; i < array_size(ids); i++) {
        ast_id_t *id = array_item(ids, i, ast_id_t);

        if (strcmp(id->name, new_id->name) == 0) {
            ERROR(ERROR_DUPLICATED_ID, &new_id->pos, new_id->name);
            return;
        }
    }

    array_add(ids, idx, new_id);
}

void
id_join(array_t *ids, int idx, array_t *new_ids)
{
    int i;

    if (new_ids == NULL)
        return;

    for (i = 0; i < array_size(new_ids); i++) {
        id_add(ids, idx + i, array_item(new_ids, i, ast_id_t));
    }
}

void
ast_id_dump(ast_id_t *id, int indent)
{
}

/* end of ast_id.c */
