/**
 * @file    ast_meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_META_H
#define _AST_META_H

#include "common.h"

#include "type.h"

#define is_bool_meta(meta)          ((meta)->type == TYPE_BOOL)

#define is_integer_meta(meta)                                                  \
    ((meta)->type >= TYPE_INT8 && (meta)->type <= TYPE_UINT64)

#define is_float_meta(meta)                                                    \
    ((meta)->type == TYPE_FLOAT || (meta)->type == TYPE_DOUBLE)

#define is_string_meta(meta)        ((meta)->type == TYPE_STRING)

#define is_struct_meta(meta)        ((meta)->type == TYPE_STRUCT)
#define is_map_meta(meta)           ((meta)->type == TYPE_MAP)
#define is_ref_meta(meta)           ((meta)->type == TYPE_REF)
#define is_tuple_meta(meta)         ((meta)->type == TYPE_TUPLE)

#define is_const_meta(meta)         (meta)->is_const
#define is_array_meta(meta)         (meta)->is_array
#define is_dynamic_meta(meta)       (meta)->is_dynamic

#define is_primitive_meta(meta)     ((meta)->type <= TYPE_PRIMITIVE)
#define is_comparable_meta(meta)    ((meta)->type <= TYPE_COMPARABLE)

#define is_compatible_meta(meta1, meta2)                                       \
    (is_integer_meta(meta1) ? is_integer_meta(meta2) :                         \
     (is_float_meta(meta1) ? is_float_meta(meta2) :                            \
      (is_struct_meta(meta1) ? is_ref_meta(meta2) || is_struct_meta(meta2) :   \
       (is_map_meta(meta1) ? is_ref_meta(meta2) || meta_equals(meta1, meta2) : \
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
    bool is_array;
    bool is_dynamic;

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
    ASSERT1(is_valid_type(type), type);

    meta->type = type;
}

static inline void
meta_set_literal(ast_meta_t *meta, type_t type)
{
    ast_meta_set(meta, type);

    meta->is_dynamic = true;
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
