/**
 * @file    ast_meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_META_H
#define _AST_META_H

#include "common.h"

#include "types.h"

#define meta_is_bool(meta)          ((meta)->type == TYPE_BOOL)

#define meta_is_integer(meta)                                                  \
    ((meta)->type >= TYPE_INT16 && (meta)->type <= TYPE_UINT64)

#define meta_is_float(meta)                                                    \
    ((meta)->type == TYPE_FLOAT || (meta)->type == TYPE_DOUBLE)

#define meta_is_string(meta)        ((meta)->type == TYPE_STRING)

#define meta_is_struct(meta)        ((meta)->type == TYPE_STRUCT)
#define meta_is_map(meta)           ((meta)->type == TYPE_MAP)
#define meta_is_ref(meta)           ((meta)->type == TYPE_REF)
#define meta_is_tuple(meta)         ((meta)->type == TYPE_TUPLE)

#define meta_is_const(meta)         (meta)->is_const
#define meta_is_local(meta)         (meta)->is_local
#define meta_is_array(meta)         (meta)->is_array

#define meta_is_primitive(meta)     ((meta)->type <= TYPE_PRIMITIVE)
#define meta_is_comparable(meta)    ((meta)->type <= TYPE_COMPARABLE)

#define meta_is_compatible(meta1, meta2)                                       \
    (meta_is_integer(meta1) ? meta_is_integer(meta2) :                         \
     (meta_is_float(meta1) ? meta_is_float(meta2) :                            \
      (meta_is_struct(meta1) ? meta_is_ref(meta2) || meta_is_struct(meta2) :   \
       (meta_is_map(meta1) ? meta_is_ref(meta2) || meta_equals(meta1, meta2) : \
        meta_equals(meta1, meta2)))))

#define meta_set_prim               ast_meta_set
#define meta_set_tuple(meta)        ast_meta_set((meta), TYPE_TUPLE)
#define meta_set_void(meta)         ast_meta_set((meta), TYPE_VOID)

#ifndef _AST_META_T
#define _AST_META_T
typedef struct ast_meta_s ast_meta_t;
#endif /* ! _AST_META_T */

typedef struct meta_map_s {
    type_t k_type;
    ast_meta_t *v_meta;
} meta_map_t;

struct ast_meta_s {
    type_t type;

    bool is_const;
    bool is_local;
    bool is_array;

    union {
        meta_map_t u_map;
    };
};

void ast_meta_dump(ast_meta_t *meta, int indent);

static inline void
ast_meta_init(ast_meta_t *meta)
{
    ASSERT(meta != NULL);

    memset(meta, 0x00, sizeof(ast_meta_t));
}

static inline void
ast_meta_set(ast_meta_t *meta, type_t type)
{
    ASSERT(meta != NULL);
    ASSERT1(type_is_valid(type), type);

    meta->type = type;
}

static inline void
meta_set_map(ast_meta_t *meta, type_t k_type, ast_meta_t *v_meta)
{
    ast_meta_set(meta, TYPE_MAP);

    meta->u_map.k_type = k_type;
    meta->u_map.v_meta = v_meta;
}

static inline bool
meta_equals(ast_meta_t *x, ast_meta_t *y)
{
    if (x->type != y->type)
        return false;

    if (x->type == TYPE_MAP) {
        if (x->u_map.k_type != y->u_map.k_type)
            return false;

        if (!meta_equals(x->u_map.v_meta, y->u_map.v_meta))
            return false;
    }

    return true;
}

#endif /* ! _AST_META_H */
