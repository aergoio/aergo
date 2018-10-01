/**
 * @file    array.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ARRAY_H
#define _ARRAY_H

#include "common.h"

#define ARRAY_INIT_SIZE                 4

#define array_empty(array)              ((array)->idx == 0)
#define array_size(array)               (array)->idx
#define array_item(array, idx, type)    (type *)((array)->items[idx])

#define array_reset(array)              ((array)->idx = 0)

typedef struct array_s {
    int size;
    int idx;
    void **items;
} array_t;

static inline void
array_init(array_t *array)
{
    array->size = ARRAY_INIT_SIZE;
    array->idx = 0;
    array->items = xmalloc(sizeof(void *) * array->size);
}

static inline array_t *
array_new(void)
{
	array_t *array = xmalloc(sizeof(array_t));

    array_init(array);

	return array;
}

static inline void
array_add(array_t *array, void *item)
{
    if (item == NULL)
        return;

    if (array->idx == array->size) {
        array->size += ARRAY_INIT_SIZE;
        array->items = xrealloc(array->items, sizeof(void *) * array->size);
    }

    array->items[array->idx++] = item;
}

static inline void
array_join(array_t *dest, array_t *src)
{
    ASSERT(src != NULL);
    ASSERT(dest != NULL);

    if (src->idx + dest->idx > dest->size) {
        dest->size += src->size;
        dest->items = xrealloc(dest->items, sizeof(void *) * dest->size);
    }

    memcpy(&dest->items[dest->idx], &src->items[0], sizeof(void *) * src->idx);
    dest->idx += src->idx;
}

#endif /* ! _ARRAY_H */
