/**
 * @file    strbuf.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _STRBUF_H
#define _STRBUF_H

#include "common.h"

#define STRBUF_INIT_SIZE            1024

#ifndef _STRBUF_T
#define _STRBUF_T
typedef struct strbuf_s strbuf_t;
#endif  /* _STRBUF_T */

struct strbuf_s {
    int size;
    int offset;
    char *buf;
}; 

void strbuf_init(strbuf_t *sb);
void strbuf_reset(strbuf_t *sb);

void strbuf_append(strbuf_t *sb, char *str, int str_len);

void strbuf_copy(strbuf_t *src, strbuf_t *dest);

static inline int
strbuf_get_len(strbuf_t *sb)
{
    return sb->offset;
}

static inline char *
strbuf_get_str(strbuf_t *sb)
{
    return sb->buf;
}

#endif /* no _STRBUF_H */
