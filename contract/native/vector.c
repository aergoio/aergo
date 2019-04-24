/**
 * @file    vector.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "vector.h"

static void
vector_extend(vector_t *vect, int cap)
{
    ASSERT(vect != NULL);

    vect->cap += cap;
    vect->items = xrealloc(vect->items, sizeof(void *) * vect->cap);
}

void
vector_add(vector_t *vect, int idx, void *item)
{
    ASSERT(vect != NULL);
    ASSERT2(idx >= 0 && idx <= vect->size, idx, vect->size);

    if (item == NULL)
        return;

    if (vect->size == vect->cap)
        vector_extend(vect, VECTOR_INIT_CAPACITY);

    if (idx == vect->size) {
        vect->items[vect->size++] = item;
    }
    else {
        memmove(&vect->items[idx + 1], &vect->items[idx], sizeof(void *) * (vect->size - idx));

        vect->items[idx] = item;
        vect->size++;
    }
}

void
vector_sadd(vector_t *vect, void *item, int (*cmp_fn)(const void *, const void *))
{
    ASSERT(vect != NULL);

    if (item == NULL)
        return;

    if (vect->size == vect->cap)
        vector_extend(vect, VECTOR_INIT_CAPACITY);

    if (cmp_fn == NULL) {
        vect->items[vect->size++] = item;
    }
    else {
        int i;

        for (i = 0; i < vect->size; i++) {
            if (cmp_fn(item, vect->items[i]) <= 0) {
                memmove(&vect->items[i + 1], &vect->items[i], sizeof(void *) * (vect->size - i));
                vect->items[i] = item;
                vect->size++;
                break;
            }
        }

        if (i == vect->size)
            vect->items[vect->size++] = item;
    }
}

void
vector_join(vector_t *dest, int idx, vector_t *src)
{
    ASSERT(dest != NULL);
    ASSERT2(idx >= 0 && idx <= dest->size, idx, dest->size);

    if (src == NULL)
        return;

    if (src->size + dest->size > dest->cap)
        vector_extend(dest, src->cap);

    if (idx == dest->size) {
        memcpy(&dest->items[dest->size], &src->items[0], sizeof(void *) * src->size);
    }
    else {
        memmove(&dest->items[idx + src->size], &dest->items[idx],
                sizeof(void *) * (dest->size - idx));

        memcpy(&dest->items[idx], &src->items[0], sizeof(void *) * src->size);
    }

    dest->size += src->size;

    vector_reset(src);
}

void
vector_move(vector_t *vect, int from_idx, int to_idx)
{
    int i;
    void *item;

    ASSERT(vect != NULL);
    ASSERT2(from_idx >= 0 && to_idx >= 0, from_idx, to_idx);

    if (from_idx == to_idx)
        return;

    item = vect->items[from_idx];

    if (from_idx < to_idx) {
        for (i = from_idx; i < to_idx; i++) {
            vect->items[i] = vect->items[i + 1];
        }
    }
    else {
        for (i = from_idx; i > to_idx; i--) {
            vect->items[i] = vect->items[i - 1];
        }
    }

    vect->items[to_idx] = item;
}

void *
vector_del(vector_t *vect, int idx)
{
    void *item;

    if (idx < 0 || idx >= vect->size)
        return NULL;

    item = vect->items[idx];

    if (idx < vect->size - 1)
        memmove(&vect->items[idx], &vect->items[idx + 1], sizeof(void *) * vect->size - idx - 1);

    vect->size--;

    return item;
}

/* end of vector.c */
