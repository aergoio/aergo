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

    ast_node_init(id, pos);

    id->kind = kind;
    id->mod = mod;
    id->name = name;

    meta_init(&id->meta, &id->pos);

    id->idx = -1;
    id->addr = -1;

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
id_new_return(char *name, meta_t *type_meta, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_RETURN, MOD_PRIVATE, name, pos);

    id->u_ret.type_meta = type_meta;

    return id;
}

ast_id_t *
id_new_func(char *name, modifier_t mod, array_t *param_ids, ast_id_t *ret_id,
            ast_blk_t *blk, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_FN, mod, name, pos);

    id->u_fn.param_ids = param_ids;
    id->u_fn.ret_id = ret_id;
    id->u_fn.blk = blk;

    if (id->u_fn.blk != NULL)
        id->u_fn.blk->kind = BLK_FN;

    return id;
}

ast_id_t *
id_new_ctor(char *name, array_t *param_ids, ast_blk_t *blk, src_pos_t *pos)
{
    return id_new_func(name, MOD_PUBLIC | MOD_CTOR, param_ids, NULL, blk, pos);
}

ast_id_t *
id_new_contract(char *name, ast_exp_t *impl_exp, ast_blk_t *blk, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_CONT, MOD_PUBLIC, name, pos);

    ASSERT1(is_cont_blk(blk), blk->kind);

    id->u_cont.impl_exp = impl_exp;
    id->u_cont.blk = blk;

    return id;
}

ast_id_t *
id_new_interface(char *name, ast_blk_t *blk, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_ITF, MOD_PUBLIC, name, pos);

    id->u_itf.blk = blk;

    return id;
}

ast_id_t *
id_new_label(char *name, ast_stmt_t *stmt, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_LABEL, MOD_PRIVATE, name, pos);

    id->u_lab.stmt = stmt;

    return id;
}

ast_id_t *
id_new_tuple(src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_TUPLE, MOD_PRIVATE, NULL, pos);

    id->u_tup.elem_ids = array_new();

    return id;
}

ast_id_t *
lookup_array(array_t *fld_ids, char *name, bool is_self)
{
    int i;

    ASSERT(fld_ids != NULL);

    array_foreach(fld_ids, i) {
        ast_id_t *fld_id = array_get_id(fld_ids, i);

        if ((is_self || is_public_id(fld_id)) && strcmp(fld_id->name, name) == 0)
            return fld_id;
    }

    return NULL;
}

ast_id_t *
id_lookup_fld(ast_id_t *id, char *name, bool is_self)
{
    array_t *fld_ids = NULL;

    ASSERT(id != NULL);
    ASSERT(name != NULL);

    if (is_cont_id(id)) {
        ast_id_t *fld_id;

        if (id->u_cont.blk == NULL)
            return NULL;

        fld_id = lookup_array(&id->u_cont.blk->ids, name, is_self);
        if (fld_id != NULL)
            return fld_id;

        return lookup_array(&id->u_cont.blk->fns, name, is_self);
    }

    if (is_struct_id(id))
        fld_ids = id->u_struc.fld_ids;
    else if (is_enum_id(id))
        fld_ids = id->u_enum.elem_ids;
    else if (is_itf_id(id))
        fld_ids = &id->u_itf.blk->fns;
    else
        return NULL;

    return lookup_array(fld_ids, name, is_self);
}

ast_id_t *
id_lookup_param(ast_id_t *id, char *name)
{
    int i;

    ASSERT(id != NULL);
    ASSERT1(is_fn_id(id), id->kind);
    ASSERT(name != NULL);

    array_foreach(id->u_fn.param_ids, i) {
        ast_id_t *param_id = array_get_id(id->u_fn.param_ids, i);

        if (strcmp(param_id->name, name) == 0)
            return param_id;
    }

    return NULL;
}

static bool
check_dup(array_t *ids, ast_id_t *new_id)
{
    int i, j;

    array_foreach(ids, i) {
        ast_id_t *id = array_get_id(ids, i);

        if (is_tuple_id(id)) {
            array_foreach(id->u_tup.elem_ids, j) {
                ast_id_t *var_id = array_get_id(id->u_tup.elem_ids, j);

                if (strcmp(var_id->name, new_id->name) == 0) {
                    ERROR(ERROR_DUPLICATED_ID, &new_id->pos, new_id->name);
                    return false;
                }
            }
        }
        else if (strcmp(id->name, new_id->name) == 0) {
            ERROR(ERROR_DUPLICATED_ID, &new_id->pos, new_id->name);
            return false;
        }
    }

    return true;
}

void
id_add(array_t *ids, ast_id_t *new_id)
{
    int i;

    if (new_id == NULL)
        return;

    if (is_tuple_id(new_id)) {
        array_foreach(new_id->u_tup.elem_ids, i) {
            check_dup(ids, array_get_id(new_id->u_tup.elem_ids, i));
        }
    }
    else if (!check_dup(ids, new_id)) {
        return;
    }

    array_add_last(ids, new_id);
}

void
id_join(array_t *ids, array_t *new_ids)
{
    int i;

    if (new_ids == NULL)
        return;

    array_foreach(new_ids, i) {
        id_add(ids, array_get_id(new_ids, i));
    }
}

array_t *
id_strip(ast_id_t *id)
{
    int i;
    array_t *ids = array_new();

    ASSERT1(is_tuple_id(id), id->kind);
    ASSERT(id->u_tup.dflt_exp == NULL);

    array_foreach(id->u_tup.elem_ids, i) {
        ast_id_t *var_id = array_get_id(id->u_tup.elem_ids, i);

        var_id->mod = id->mod;
        var_id->u_var.type_meta = id->u_tup.type_meta;

        array_add_last(ids, var_id);
    }

    return ids;
}

bool
id_cmp(ast_id_t *x, ast_id_t *y)
{
    int i;

    ASSERT1(is_fn_id(x), x->kind);
    ASSERT1(is_fn_id(y), y->kind);

    if (x->mod != y->mod)
        return false;

    if (array_size(x->u_fn.param_ids) != array_size(y->u_fn.param_ids))
        return false;

    array_foreach(x->u_fn.param_ids, i) {
        ast_id_t *x_param = array_get_id(x->u_fn.param_ids, i);
        ast_id_t *y_param = array_get_id(y->u_fn.param_ids, i);

        if (!meta_cmp(&x_param->meta, &y_param->meta)) {
            error_pop();
            return false;
        }
    }

    if (!meta_cmp(&x->meta, &y->meta)) {
        error_pop();
        return false;
    }

    return true;
}

/* end of ast_id.c */