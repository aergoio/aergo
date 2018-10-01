/**
 * @file    ast_meta.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_META_H
#define _AST_META_H

#include "common.h"

#include "ast_type.h"

#ifndef _AST_META_T
#define _AST_META_T
typedef struct ast_meta_s ast_meta_t;
#endif /* ! _AST_META_T */

typedef enum scope_e {
    SCOPE_GLOBAL    = 0,
    SCOPE_LOCAL,
    SCOPE_SHARED,
    SCOPE_MAX
} scope_t;

struct ast_meta_s {
    type_t type;
    scope_t scope;
    bool is_const;

    /* only for map */
    type_t k_type;
    ast_meta_t *v_meta;
};

void ast_meta_dump(ast_meta_t *meta, int indent);

static inline void
ast_meta_init(ast_meta_t *meta)
{
    memset(meta, 0x00, sizeof(ast_meta_t));
}

static inline void
ast_meta_set(ast_meta_t *meta, type_t type)
{
    ast_meta_init(meta);

    meta->type = type;
}

static inline void
ast_meta_set_map(ast_meta_t *meta, type_t k_type, ast_meta_t *v_meta)
{
    ast_meta_init(meta);

    meta->type = TYPE_MAP;
    meta->k_type = k_type;
    meta->v_meta = v_meta;
}

#endif /* ! _AST_META_H */
