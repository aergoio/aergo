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
id_new_fn(char *name, modifier_t mod, array_t *param_ids, array_t *ret_ids,
          ast_blk_t *blk, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_FN, mod, name, pos);

    id->u_fn.param_ids = param_ids;
    id->u_fn.ret_ids = ret_ids;
    id->u_fn.blk = blk;

    if (id->u_fn.blk != NULL)
        id->u_fn.blk->kind = BLK_FN;

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
    else if (is_cont_id(id) && id->u_cont.blk != NULL)
        fld_ids = &id->u_cont.blk->ids;
    else
        return NULL;

    ASSERT(fld_ids != NULL);

    for (i = 0; i < array_size(fld_ids); i++) {
        ast_id_t *fld_id = array_get_id(fld_ids, i);

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
    ASSERT1(is_fn_id(id), id->kind);
    ASSERT(name != NULL);

    for (i = 0; i < array_size(id->u_fn.param_ids); i++) {
        ast_id_t *param_id = array_get_id(id->u_fn.param_ids, i);

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
        ast_id_t *id = array_get_id(ids, i);

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
        id_add(ids, idx + i, array_get_id(new_ids, i));
    }
}

void
ast_id_dump(ast_id_t *id, int indent)
{
}

/* end of ast_id.c */
