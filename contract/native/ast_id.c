/**
 * @file    ast_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "ast_blk.h"
#include "ast_exp.h"

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

    id->idx = -1;

    return id;
}

ast_id_t *
id_new_var(char *name, modifier_t mod, src_pos_t *pos)
{
    return ast_id_new(ID_VAR, mod, name, pos);
}

ast_id_t *
id_new_struct(char *name, array_t *fld_ids, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_STRUCT, MOD_PRIVATE, name, pos);

    ASSERT(fld_ids != NULL);

    id->u_struc.fld_ids = fld_ids;

    return id;
}

ast_id_t *
id_new_enum(char *name, array_t *elem_ids, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_ENUM, MOD_PRIVATE, name, pos);

    ASSERT(elem_ids != NULL);

    id->u_enum.elem_ids = elem_ids;

    return id;
}

ast_id_t *
id_new_func(char *name, modifier_t mod, array_t *param_ids, array_t *ret_ids,
            ast_blk_t *blk, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_FUNC, mod, name, pos);

    id->u_func.param_ids = param_ids;
    id->u_func.ret_ids = ret_ids;
    id->u_func.blk = blk;

    if (id->u_func.blk != NULL)
        id->u_func.blk->kind = BLK_FUNC;

    return id;
}

ast_id_t *
id_new_contract(char *name, ast_blk_t *blk, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_CONTRACT, MOD_PUBLIC, name, pos);

    id->u_cont.blk = blk;

    if (id->u_cont.blk != NULL)
        id->u_cont.blk->kind = BLK_CONTRACT;

    return id;
}

ast_id_t *
id_new_label(char *name, ast_stmt_t *stmt, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_LABEL, MOD_PUBLIC, name, pos);

    id->u_label.stmt = stmt;

    return id;
}

ast_id_t *
id_search_name(ast_blk_t *blk, char *name, int num)
{
    int i;

    ASSERT(name != NULL);

    if (blk == NULL)
        return NULL;

    do {
        /* TODO: better to skip if it is equal to the current contract id */
        for (i = 0; i < array_size(&blk->ids); i++) {
            ast_id_t *id = array_get(&blk->ids, i, ast_id_t);

            if (!is_label_id(id) && id->num < num && strcmp(id->name, name) == 0)
                return id;
        }
    } while ((blk = blk->up) != NULL);

    return NULL;
}

ast_id_t *
id_search_fld(ast_id_t *id, char *name, bool is_self)
{
    int i;
    array_t *fld_ids = NULL;

    ASSERT(id != NULL);
    ASSERT(name != NULL);

    if (is_struct_id(id))
        fld_ids = id->u_struc.fld_ids;
    else if (is_enum_id(id))
        fld_ids = id->u_enum.elem_ids;
    else if (is_contract_id(id) && id->u_cont.blk != NULL)
        fld_ids = &id->u_cont.blk->ids;
    else
        return NULL;

    ASSERT(fld_ids != NULL);

    for (i = 0; i < array_size(fld_ids); i++) {
        ast_id_t *fld_id = array_get(fld_ids, i, ast_id_t);

        if ((is_self || is_public_id(fld_id)) && strcmp(fld_id->name, name) == 0)
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
        ast_id_t *param_id = array_get(id->u_func.param_ids, i, ast_id_t);

        if (strcmp(param_id->name, name) == 0)
            return param_id;
    }

    return NULL;
}

ast_id_t *
id_search_label(ast_blk_t *blk, char *name)
{
    int i;

    ASSERT(name != NULL);

    if (blk == NULL)
        return NULL;

    do {
        for (i = 0; i < array_size(&blk->ids); i++) {
            ast_id_t *id = array_get(&blk->ids, i, ast_id_t);

            if (is_label_id(id) && strcmp(id->name, name) == 0)
                return id;
        }
    } while ((blk = blk->up) != NULL);

    return NULL;
}

void
id_add(array_t *ids, int idx, ast_id_t *new_id)
{
    int i;

    if (new_id == NULL)
        return;

    for (i = 0; i < array_size(ids); i++) {
        ast_id_t *id = array_get(ids, i, ast_id_t);

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
        id_add(ids, idx + i, array_get(new_ids, i, ast_id_t));
    }
}

void
ast_id_dump(ast_id_t *id, int indent)
{
}

/* end of ast_id.c */
