/**
 * @file    array.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ARRAY_H
#define _ARRAY_H

#include "common.h"

#define ARRAY_INIT_CAPACITY         10

#define array_size(arr)             ((arr)->size)
#define array_items(arr)            ((arr)->items)

#define array_init(arr, type)                                                                      \
    do {                                                                                           \
        ASSERT((arr) != NULL);                                                                     \
        ASSERT1(sizeof(type) % 4 == 0, sizeof(type));                                              \
        (arr)->cap = ARRAY_INIT_CAPACITY;                                                          \
        (arr)->size = 0;                                                                           \
        (arr)->unit = sizeof(type);                                                                \
        (arr)->items = xmalloc((arr)->unit * (arr)->cap);                                          \
    } while (0)

#define array_add(arr, item, type)                                                                 \
    do {                                                                                           \
        ASSERT2(sizeof(type) == (arr)->unit, sizeof(type), (arr)->unit);                           \
        if ((arr)->size == (arr)->cap) {                                                           \
            (arr)->cap += ARRAY_INIT_CAPACITY;                                                     \
            (arr)->items = xrealloc((arr)->items, (arr)->unit * (arr)->cap);                       \
        }                                                                                          \
        ((type *)((arr)->items))[(arr)->size++] = (item);                                          \
    } while (0)

#define array_reset(arr)            ((arr)->size = 0)

#ifndef _ARRAY_T
#define _ARRAY_T
typedef struct array_s array_t;
#endif /* ! _ARRAY_T */

struct array_s {
    int cap;
    int size;
    uint unit;
    void *items;
};

#endif /* ! _ARRAY_H */
