/**
 * @file    hash.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _HASH_H
#define _HASH_H

#include "common.h"

#include "array.h"

#define HASH_BUCKET_SIZE            64

typedef struct hash_elem_s {
    char *key;
    void *item;
} hash_elem_t;

typedef struct hash_s {
    array_t buckets[HASH_BUCKET_SIZE];
} hash_t;

void hash_add(hash_t *hash, char *key, void *val);

void *hash_lookup(hash_t *hash, char *key);

static inline void
hash_init(hash_t *hash)
{
    int i;

    ASSERT(hash != NULL);

    for (i = 0; i < HASH_BUCKET_SIZE; i++) {
        array_init(&hash->buckets[i]);
    }
}

static inline hash_t *
hash_new(void)
{
	hash_t *hash = xmalloc(sizeof(hash_t));

    hash_init(hash);

	return hash;
}

#endif /* ! _HASH_H */
