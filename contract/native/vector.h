/**
 * @file    vector.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _VECTOR_H
#define _VECTOR_H

#include "common.h"

#define VECTOR_INIT_CAPACITY            4

#define is_empty_vector(vect)           ((vect) == NULL ? true : (vect)->size == 0)
#define vector_size(vect)               ((vect) == NULL ? 0 : (vect)->size)

#define vector_get(vect, idx, type)     ((type *)((vect)->items[idx]))

#define vector_get_id(vect, idx)        vector_get(vect, idx, ast_id_t)
#define vector_get_exp(vect, idx)       vector_get(vect, idx, ast_exp_t)
#define vector_get_stmt(vect, idx)      vector_get(vect, idx, ast_stmt_t)
#define vector_get_md(vect, idx)        vector_get(vect, idx, ir_md_t)
#define vector_get_abi(vect, idx)       vector_get(vect, idx, ir_abi_t)
#define vector_get_fn(vect, idx)        vector_get(vect, idx, ir_fn_t)
#define vector_get_bb(vect, idx)        vector_get(vect, idx, ir_bb_t)
#define vector_get_br(vect, idx)        vector_get(vect, idx, ir_br_t)

#define vector_get_first(vect, type)                                                               \
    (vector_size(vect) > 0 ? (type *)((vect)->items[0]) : NULL)
#define vector_get_last(vect, type)                                                                \
    (vector_size(vect) > 0 ? (type *)((vect)->items[(vect)->size - 1]) : NULL)

#define vector_add_first(vect, item)    vector_add((vect), 0, (item))
#define vector_add_last(vect, item)     vector_add((vect), (vect)->size, (item))

#define vector_join_first(dest, src)    vector_join((dest), 0, (src))
#define vector_join_last(dest, src)     vector_join((dest), (dest)->size, (src))

#define vector_reset(vect)              ((vect)->size = 0)

#define vector_foreach(vect, i)         for ((i) = 0; (i) < vector_size(vect); (i)++)

#ifndef _VECTOR_T
#define _VECTOR_T
typedef struct vector_s vector_t;
#endif /* ! _VECTOR_T */

struct vector_s {
    int cap;
    int size;
    void **items;
};

void vector_add(vector_t *vect, int idx, void *item);
void vector_sadd(vector_t *vect, void *item, int (*cmp_fn)(const void *, const void *));
void vector_join(vector_t *dest, int idx, vector_t *src);

void vector_move(vector_t *vect, int from_idx, int to_idx);

void *vector_del(vector_t *vect, int idx);

static inline void
vector_init(vector_t *vect)
{
    ASSERT(vect != NULL);

    vect->cap = VECTOR_INIT_CAPACITY;
    vect->size = 0;
    vect->items = xmalloc(sizeof(void *) * vect->cap);
}

static inline vector_t *
vector_new(void)
{
	vector_t *vect = xmalloc(sizeof(vector_t));

    vector_init(vect);

	return vect;
}

static inline void
vector_set(vector_t *vect, int idx, void *item)
{
    ASSERT2(idx < vect->size, idx, vect->size);

    vect->items[idx] = item;
}

static inline void
vector_clear(vector_t *vect)
{
    ASSERT(vect != NULL);
    ASSERT(vect->items != NULL);

    xfree(vect->items);
    xfree(vect);
}

#endif /* ! _VECTOR_H */
