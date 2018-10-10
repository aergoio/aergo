/**
 * @file    array.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ARRAY_H
#define _ARRAY_H

#include "common.h"

#define ARRAY_INIT_SIZE                 4

#define is_empty_array(array)                                                  \
    ((array) == NULL ? true : (array)->cnt == 0)

#define array_size(array)               ((array) == NULL ? 0 : (array)->cnt)
#define array_item(array, idx, type)    ((type *)((array)->items[idx]))

#define array_add_first(array, item)    array_add((array), 0, (item))
#define array_add_last(array, item)     array_add((array), (array)->cnt, (item))

#define array_join_first(dest, src)     array_join((dest), 0, (src))
#define array_join_last(dest, src)      array_join((dest), (dest)->cnt, (src))

#define array_reset(array)              ((array)->cnt = 0)

typedef struct array_s {
    int size;
    int cnt;
    void **items;
} array_t;

static inline void
array_init(array_t *array)
{
    array->size = ARRAY_INIT_SIZE;
    array->cnt = 0;
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
array_add(array_t *array, int idx, void *item)
{
    ASSERT(array != NULL);
    ASSERT2(idx >= 0 && idx <= array->cnt, idx, array->cnt);

    if (item == NULL)
        return;

    if (array->cnt == array->size)
        array_extend(array, ARRAY_INIT_SIZE);

    if (idx == array->cnt) {
        array->items[array->cnt++] = item;
    }
    else {
        memmove(&array->items[idx + 1], &array->items[idx], 
                sizeof(void *) * array->cnt);

        array->items[idx] = item;
        array->cnt++;
    }
}

static inline void
array_sadd(array_t *array, void *item,
           int (*cmp_fn)(const void *, const void *))
{
    ASSERT(array != NULL);

    if (item == NULL)
        return;

    if (array->cnt == array->size)
        array_extend(array, ARRAY_INIT_SIZE);

    if (cmp_fn == NULL) {
        array->items[array->cnt++] = item;
    }
    else {
        int i;

        for (i = 0; i < array->cnt; i++) {
            if (cmp_fn(item, array->items[i]) <= 0) {
                memmove(&array->items[i + 1], &array->items[i],
                        sizeof(void *) * (array->cnt - i));
                array->items[i] = item;
                array->cnt++;
                break;
            }
        }

        if (i == array->cnt)
            array->items[array->cnt++] = item;
    }
}

static inline void
array_join(array_t *dest, int idx, array_t *src)
{
    ASSERT(src != NULL);
    ASSERT(dest != NULL);
    ASSERT2(idx >= 0 && idx <= dest->cnt, idx, dest->cnt);

    if (src->cnt + dest->cnt > dest->size)
        array_extend(dest, src->size);

    if (idx == dest->cnt) {
        memcpy(&dest->items[dest->cnt], &src->items[0], sizeof(void *) * src->cnt);
    }
    else {
        memmove(&dest->items[idx + src->cnt], &dest->items[idx],
                sizeof(void *) * (dest->cnt - idx));

        memcpy(&dest->items[idx], &src->items[0], sizeof(void *) * src->cnt);
    }

    dest->cnt += src->cnt;
}

#endif /* ! _ARRAY_H */
