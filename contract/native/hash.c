/**
 * @file    hash.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "hash.h"

static inline uint32_t
hash_fn(const char *key, int len)
{
	/* Murmurhash3 (ref, https://en.wikipedia.org/wiki/MurmurHash) */
	uint32_t c1 = 0xcc9e2d51;
	uint32_t c2 = 0x1b873593;
	uint32_t r1 = 15;
	uint32_t r2 = 13;
	uint32_t m = 5;
	uint32_t n = 0xe6546b64;
	uint32_t h = 0;
	uint32_t k = 0;
	uint8_t *d = (uint8_t *)key;
	const uint32_t *chunks = NULL;
	const uint8_t *tail = NULL;
	int i = 0;
	int l = len / 4;

	chunks = (const uint32_t *)(d + l * 4);
	tail = (const uint8_t *)(d + l * 4);

	// for each 4 byte chunk of `key'
	for (i = -l; i != 0; ++i) {
		k = chunks[i];

		k *= c1;
		k = (k << r1) | (k >> (32 - r1));
		k *= c2;

		h ^= k;
		h = (h << r2) | (h >> (32 - r2));
		h = h * m + n;
	}

	k = 0;

	// remainder
    switch (len & 3) {
    case 3: k ^= (tail[2] << 16);
    case 2: k ^= (tail[1] << 8);
    case 1:
        k ^= tail[0];
        k *= c1;
        k = (k << r1) | (k >> (32 - r1));
        k *= c2;
        h ^= k;
    }

	h ^= len;

	h ^= (h >> 16);
	h *= 0x85ebca6b;
	h ^= (h >> 13);
	h *= 0xc2b2ae35;
	h ^= (h >> 16);

	return h;
}

void
hash_add(hash_t *hash, char *key, void *item)
{
    uint32_t hashval = hash_fn(key, strlen(key));
    int bucket_idx = hashval % HASH_BUCKET_SIZE;
    hash_elem_t *elem = xmalloc(sizeof(hash_elem_t));

    elem->key = key;
    elem->item = item;

    hashval = hash_fn(key, strlen(key));
    bucket_idx = hashval % HASH_BUCKET_SIZE;

    array_add_last(&hash->buckets[bucket_idx], elem);
}

void *
hash_lookup(hash_t *hash, char *key)
{
    int i;
    uint32_t hashval = hash_fn(key, strlen(key));
    array_t *bucket = &hash->buckets[hashval % HASH_BUCKET_SIZE];

    array_foreach(bucket, i) {
        hash_elem_t *elem = array_get(bucket, i, hash_elem_t);

        if (strcmp(elem->key, key) == 0)
            return elem->item;
    }

    return NULL;
}

/* end of hash.c */
