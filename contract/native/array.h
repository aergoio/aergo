/**
 * @file    array.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ARRAY_H
#define _ARRAY_H

#include "common.h"

#define ARRAY_INIT_CAPACITY             4

#define is_empty_array(array)           ((array) == NULL ? true : (array)->size == 0)
#define array_size(array)               ((array) == NULL ? 0 : (array)->size)

#define array_get(array, idx, type)     ((type *)((array)->items[idx]))

#define array_get_id(array, idx)        array_get(array, idx, ast_id_t)
#define array_get_exp(array, idx)       array_get(array, idx, ast_exp_t)
#define array_get_stmt(array, idx)      array_get(array, idx, ast_stmt_t)
#define array_get_abi(array, idx)       array_get(array, idx, ir_abi_t)
#define array_get_fn(array, idx)        array_get(array, idx, ir_fn_t)
#define array_get_bb(array, idx)        array_get(array, idx, ir_bb_t)
#define array_get_br(array, idx)        array_get(array, idx, ir_br_t)

#define array_get_first(array, type)                                                     \
    (array_size(array) > 0 ? (type *)((array)->items[0]) : NULL)
#define array_get_last(array, type)                                                      \
    (array_size(array) > 0 ? (type *)((array)->items[(array)->size - 1]) : NULL)

#define array_add_first(array, item)    array_add((array), 0, (item))
#define array_add_last(array, item)     array_add((array), (array)->size, (item))

#define array_join_first(dest, src)     array_join((dest), 0, (src))
#define array_join_last(dest, src)      array_join((dest), (dest)->size, (src))

#define array_reset(array)              ((array)->size = 0)

#define array_foreach(array, i)         for ((i) = 0; (i) < array_size(array); (i)++)

typedef struct array_s {
    int cap;
    int size;
    void **items;
} array_t;

void array_add(array_t *array, int idx, void *item);
void array_sadd(array_t *array, void *item, int (*cmp_fn)(const void *, const void *));
void array_join(array_t *dest, int idx, array_t *src);

void array_move(array_t *array, int from_idx, int to_idx);

void *array_del(array_t *array, int idx);

static inline void
array_init(array_t *array)
{
    ASSERT(array != NULL);

    array->cap = ARRAY_INIT_CAPACITY;
    array->size = 0;
    array->items = xmalloc(sizeof(void *) * array->cap);
}

static inline array_t *
array_new(void)
{
	array_t *array = xmalloc(sizeof(array_t));

    array_init(array);

	return array;
}

static inline void
array_set(array_t *array, int idx, void *item)
{
    ASSERT2(idx < array->size, idx, array->size);

    array->items[idx] = item;
}

static inline void
array_clear(array_t *array)
{
    ASSERT(array != NULL);
    ASSERT(array->items != NULL);

    xfree(array->items);
    xfree(array);
}

#endif /* ! _ARRAY_H */
