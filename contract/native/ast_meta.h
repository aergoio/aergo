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

typedef enum scope_e {
    SCOPE_GLOBAL    = 0,
    SCOPE_LOCAL,
    SCOPE_SHARED,
    SCOPE_MAX
} scope_t;

typedef struct meta_struct_s {
    int fld_cnt;
    ast_meta_t **fld_metas;
} meta_struct_t;

typedef struct meta_map_s {
    type_t k_type;
    ast_meta_t *v_meta;
} meta_map_t;

struct ast_meta_s {
    type_t type;
    bool is_const;
    bool is_local;

    union {
        meta_struct_t u_st;
        meta_map_t u_map;
    };
};

void meta_set_struct(ast_meta_t *meta, list_t *field_l);
void meta_set_map(ast_meta_t *meta, type_t k_type, ast_meta_t *v_meta);

void ast_meta_dump(ast_meta_t *meta, int indent);

static inline void
ast_meta_init(ast_meta_t *meta, type_t type)
{
    ASSERT(meta != NULL);

    memset(meta, 0x00, sizeof(ast_meta_t));
    meta->type = type;
}

#endif /* ! _AST_META_H */
