/**
 * @file    ast_id.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"
#include "ast_blk.h"
#include "ast_exp.h"
#include "ast_stmt.h"

#include "ast_id.h"

static ast_id_t *
ast_id_new(id_kind_t kind, modifier_t mod, char *name, src_pos_t *pos)
{
    ast_id_t *id = xcalloc(sizeof(ast_id_t));

    ast_node_init(id, *pos);

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
id_new_param(char *name, ast_exp_t *type_exp, src_pos_t *pos)
{
    static int param_num = 1;
    ast_id_t *id;

    if (name == NULL) {
        char gen_name[NAME_MAX_LEN + 1];

        snprintf(gen_name, sizeof(gen_name), "param$%d", param_num++);

        id = id_new_var(xstrdup(gen_name), MOD_PRIVATE, pos);
    }
    else {
        id = id_new_var(name, MOD_PRIVATE, pos);
    }

    id->u_var.is_param = true;
    id->u_var.type_exp = type_exp;

    return id;
}

ast_id_t *
id_new_struct(char *name, vector_t *fld_ids, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_STRUCT, MOD_PRIVATE, name, pos);

    ASSERT(fld_ids != NULL);

    id->u_struc.fld_ids = fld_ids;

    return id;
}

ast_id_t *
id_new_enum(char *name, vector_t *elem_ids, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_ENUM, MOD_PRIVATE, name, pos);

    ASSERT(elem_ids != NULL);

    id->u_enum.elem_ids = elem_ids;

    return id;
}

ast_id_t *
id_new_func(char *name, modifier_t mod, vector_t *param_ids, ast_id_t *ret_id, ast_blk_t *blk,
            src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_FN, mod, name, pos);

    id->u_fn.param_ids = param_ids;
    id->u_fn.ret_id = ret_id;
    id->u_fn.blk = blk;

    return id;
}

ast_id_t *
id_new_ctor(char *name, vector_t *param_ids, ast_blk_t *blk, src_pos_t *pos)
{
    ast_exp_t *type_exp = exp_new_type(TYPE_NONE, pos);

    type_exp->u_type.name = name;

    return id_new_func(name, MOD_PUBLIC | MOD_CTOR, param_ids,
                       id_new_param(name, type_exp, pos), blk, pos);
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
id_new_library(char *name, ast_blk_t *blk, src_pos_t *pos)
{
    ast_id_t *id = ast_id_new(ID_LIB, MOD_PUBLIC, name, pos);

    id->u_lib.blk = blk;

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

    id->u_tup.elem_ids = vector_new();

    return id;
}

ast_id_t *
id_new_tmp_var(char *name)
{
    return ast_id_new(ID_VAR, MOD_PRIVATE, name, &null_pos_);
}

ast_id_t *
id_search_fld(ast_id_t *id, char *name, bool is_self)
{
    int i;
    vector_t *fld_ids = NULL;

    ASSERT(id != NULL);
    ASSERT(name != NULL);

    if (is_struct_id(id))
        fld_ids = id->u_struc.fld_ids;
    else if (is_enum_id(id))
        fld_ids = id->u_enum.elem_ids;
    else if (is_cont_id(id) && id->u_cont.blk != NULL)
        fld_ids = &id->u_cont.blk->ids;
    else if (is_itf_id(id))
        fld_ids = &id->u_itf.blk->ids;
    else
        return NULL;

    ASSERT(fld_ids != NULL);

    vector_foreach(fld_ids, i) {
        ast_id_t *fld_id = vector_get_id(fld_ids, i);

        if ((is_self || is_itf_id(id) || is_public_id(fld_id)) && strcmp(fld_id->name, name) == 0)
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

    vector_foreach(id->u_fn.param_ids, i) {
        ast_id_t *param_id = vector_get_id(id->u_fn.param_ids, i);

        if (strcmp(param_id->name, name) == 0)
            return param_id;
    }

    return NULL;
}

static bool
check_dup(vector_t *ids, ast_id_t *new_id)
{
    int i, j;

    vector_foreach(ids, i) {
        ast_id_t *id = vector_get_id(ids, i);

        if (is_tuple_id(id)) {
            vector_foreach(id->u_tup.elem_ids, j) {
                ast_id_t *var_id = vector_get_id(id->u_tup.elem_ids, j);

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
id_add(vector_t *ids, ast_id_t *new_id)
{
    int i;

    if (new_id == NULL)
        return;

    if (is_tuple_id(new_id)) {
        vector_foreach(new_id->u_tup.elem_ids, i) {
            check_dup(ids, vector_get_id(new_id->u_tup.elem_ids, i));
        }
    }
    else if (!check_dup(ids, new_id)) {
        return;
    }

    if (is_ctor_id(new_id))
        vector_add_first(ids, new_id);
    else
        vector_add_last(ids, new_id);
}

void
id_join(vector_t *ids, vector_t *new_ids)
{
    int i;

    if (new_ids == NULL)
        return;

    vector_foreach(new_ids, i) {
        id_add(ids, vector_get_id(new_ids, i));
    }
}

vector_t *
id_strip(ast_id_t *id)
{
    int i;
    vector_t *ids = vector_new();

    ASSERT1(is_tuple_id(id), id->kind);
    ASSERT(id->u_tup.dflt_exp == NULL);

    vector_foreach(id->u_tup.elem_ids, i) {
        ast_id_t *var_id = vector_get_id(id->u_tup.elem_ids, i);

        var_id->mod = id->mod;
        var_id->u_var.type_exp = id->u_tup.type_exp;

        vector_add_last(ids, var_id);
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

    if (vector_size(x->u_fn.param_ids) != vector_size(y->u_fn.param_ids))
        return false;

    vector_foreach(x->u_fn.param_ids, i) {
        ast_id_t *x_param = vector_get_id(x->u_fn.param_ids, i);
        ast_id_t *y_param = vector_get_id(y->u_fn.param_ids, i);

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
