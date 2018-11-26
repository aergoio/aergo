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

#define array_add_first(array, item)    array_add((array), 0, (item))
#define array_add_last(array, item)     array_add((array), (array)->size, (item))

#define array_join_first(dest, src)     array_join((dest), 0, (src))
#define array_join_last(dest, src)      array_join((dest), (dest)->size, (src))

#define array_reset(array)              ((array)->size = 0)

typedef struct array_s {
    int cap;
    int size;
    void **items;
} array_t;

void array_add(array_t *array, int idx, void *item);
void array_sadd(array_t *array, void *item, int (*cmp_fn)(const void *, const void *));
void array_join(array_t *dest, int idx, array_t *src);

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
