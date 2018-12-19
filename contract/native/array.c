/**
 * @file    array.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "array.h"

static void
array_extend(array_t *array, int cap)
{
    ASSERT(array != NULL);

    array->cap += cap;
    array->items = xrealloc(array->items, sizeof(void *) * array->cap);
}

void
array_add(array_t *array, int idx, void *item)
{
    ASSERT(array != NULL);
    ASSERT2(idx >= 0 && idx <= array->size, idx, array->size);

    if (item == NULL)
        return;

    if (array->size == array->cap)
        array_extend(array, ARRAY_INIT_CAPACITY);

    if (idx == array->size) {
        array->items[array->size++] = item;
    }
    else {
        memmove(&array->items[idx + 1], &array->items[idx],
                sizeof(void *) * array->size);

        array->items[idx] = item;
        array->size++;
    }
}

void
array_sadd(array_t *array, void *item, int (*cmp_fn)(const void *, const void *))
{
    ASSERT(array != NULL);

    if (item == NULL)
        return;

    if (array->size == array->cap)
        array_extend(array, ARRAY_INIT_CAPACITY);

    if (cmp_fn == NULL) {
        array->items[array->size++] = item;
    }
    else {
        int i;

        for (i = 0; i < array->size; i++) {
            if (cmp_fn(item, array->items[i]) <= 0) {
                memmove(&array->items[i + 1], &array->items[i],
                        sizeof(void *) * (array->size - i));
                array->items[i] = item;
                array->size++;
                break;
            }
        }

        if (i == array->size)
            array->items[array->size++] = item;
    }
}

void
array_join(array_t *dest, int idx, array_t *src)
{
    ASSERT(dest != NULL);
    ASSERT2(idx >= 0 && idx <= dest->size, idx, dest->size);

    if (src == NULL)
        return;

    if (src->size + dest->size > dest->cap)
        array_extend(dest, src->cap);

    if (idx == dest->size) {
        memcpy(&dest->items[dest->size], &src->items[0], sizeof(void *) * src->size);
    }
    else {
        memmove(&dest->items[idx + src->size], &dest->items[idx],
                sizeof(void *) * (dest->size - idx));

        memcpy(&dest->items[idx], &src->items[0], sizeof(void *) * src->size);
    }

    dest->size += src->size;

    array_reset(src);
}

void *
array_del(array_t *array, int idx)
{
    void *item;

    if (idx < 0 || idx >= array->size)
        return NULL;

    item = array->items[idx];

    if (idx < array->size - 1)
        memmove(&array->items[idx], &array->items[idx + 1], 
                sizeof(void *) * array->size - idx - 1);

    array->size--;

    return item;
}

/* end of array.c */
