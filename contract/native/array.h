/**
 * @file    array.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ARRAY_H
#define _ARRAY_H

#include "common.h"

#define ARRAY_INIT_SIZE                 4

#define is_empty_array(array)                                                  \
    ((array) == NULL ? true : (array)->idx == 0)

#define array_size(array)               ((array) == NULL ? 0 : (array)->idx)
#define array_item(array, idx, type)    ((type *)((array)->items[idx]))

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
array_clear(array_t *array)
{
    ASSERT(array != NULL);
    ASSERT(array->items != NULL);

    xfree(array->items);
    xfree(array);
}

static inline void
array_extend(array_t *array, int size)
{
    ASSERT(array != NULL);

    array->size += size;
    array->items = xrealloc(array->items, sizeof(void *) * array->size);
}

static inline void
array_join(array_t *dest, array_t *src)
{
    ASSERT(src != NULL);
    ASSERT(dest != NULL);

    if (src->idx + dest->idx > dest->size)
        array_extend(dest, src->size);

    memcpy(&dest->items[dest->idx], &src->items[0], sizeof(void *) * src->idx);
    dest->idx += src->idx;
}

static inline void
array_add_head(array_t *array, void *item)
{
    ASSERT(array != NULL);

    if (item == NULL)
        return;

    if (array->idx == array->size)
        array_extend(array, ARRAY_INIT_SIZE);

    memmove(&array->items[1], &array->items[0], sizeof(void *) * array->idx);

    array->items[0] = item;
    array->idx++;
}

static inline void
array_add_tail(array_t *array, void *item)
{
    ASSERT(array != NULL);

    if (item == NULL)
        return;

    if (array->idx == array->size)
        array_extend(array, ARRAY_INIT_SIZE);

    array->items[array->idx++] = item;
}

static inline void
array_sadd(array_t *array, void *item,
           int (*cmp_fn)(const void *, const void *))
{
    ASSERT(array != NULL);

    if (item == NULL)
        return;

    if (array->idx == array->size)
        array_extend(array, ARRAY_INIT_SIZE);

    if (cmp_fn == NULL) {
        array->items[array->idx++] = item;
    }
    else {
        int i;

        for (i = 0; i < array->idx; i++) {
            if (cmp_fn(item, array->items[i]) <= 0) {
                memmove(&array->items[i + 1], &array->items[i],
                        sizeof(void *) * (array->idx - i));
                array->items[i] = item;
                array->idx++;
                break;
            }
        }

        if (i == array->idx)
            array->items[array->idx++] = item;
    }
}

#endif /* ! _ARRAY_H */
