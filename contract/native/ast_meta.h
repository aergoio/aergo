/**
 * @file    ast_meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_META_H
#define _AST_META_H

#include "common.h"

#include "ast_type.h"

#define ast_meta_set_prim           ast_meta_init

#ifndef _AST_META_T
#define _AST_META_T
typedef struct ast_meta_s ast_meta_t;
#endif /* ! _AST_META_T */

typedef struct meta_tuple_s {
    int size;
    ast_meta_t **metas;
} meta_tuple_t;

typedef struct meta_map_s {
    type_t k_type;
    ast_meta_t *v_meta;
} meta_map_t;

struct ast_meta_s {
    type_t type;
    bool is_const;
    bool is_local;

    union {
        meta_tuple_t u_tup;
        meta_map_t u_map;
    };
};

void meta_set_map(ast_meta_t *meta, type_t k_type, ast_meta_t *v_meta);
void meta_set_tuple(ast_meta_t *meta, list_t *field_l);

void ast_meta_dump(ast_meta_t *meta, int indent);

static inline void
ast_meta_init(ast_meta_t *meta, type_t type)
{
    ASSERT(meta != NULL);

    memset(meta, 0x00, sizeof(ast_meta_t));
    meta->type = type;
}

static inline bool
ast_meta_equals(ast_meta_t *x, ast_meta_t *y)
{
    if (x->type != y->type)
        return false;

    if (type_is_map(x->type)) {
        if (x->u_map.k_type != y->u_map.k_type)
            return false;

        if (!ast_meta_equals(x->u_map.v_meta, y->u_map.v_meta))
            return false;
    }
    else if (type_is_tuple(x->type)) {
        int i;

        if (x->u_tup.size != y->u_tup.size)
            return false;

        for (i = 0; i < x->u_tup.size; i++) {
            if (!ast_meta_equals(x->u_tup.metas[i], y->u_tup.metas[i]))
                return false;
        }
    }

    return true;
}

#endif /* ! _AST_META_H */
