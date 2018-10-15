/**
 * @file    strbuf.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _STRBUF_H
#define _STRBUF_H

#include "common.h"

#define STRBUF_INIT_SIZE            1024

#define is_empty_strbuf(sb)         ((sb)->offset == 0)

#define strbuf_size(sb)             ((sb)->offset)
#define strbuf_str(sb)              ((sb)->buf)

#define strbuf_cat(sb, str)         strbuf_ncat((sb), (str), strlen(str))

#ifndef _STRBUF_T
#define _STRBUF_T
typedef struct strbuf_s strbuf_t;
#endif /* ! _STRBUF_T */

struct strbuf_s {
    int size;
    int offset;
    char *buf;
};

void strbuf_init(strbuf_t *sb);
void strbuf_reset(strbuf_t *sb);

void strbuf_ncat(strbuf_t *sb, char *str, int n);
void strbuf_trunc(strbuf_t *sb, int n);

void strbuf_copy(strbuf_t *src, strbuf_t *dest);

void strbuf_load(strbuf_t *sb, char *path);
void strbuf_dump(strbuf_t *sb, char *path);

static inline char
strbuf_char(strbuf_t *sb, int idx) 
{
    ASSERT(sb != NULL);
    ASSERT2(idx < sb->offset, idx, sb->offset);

    return sb->buf[idx];
}

#endif /* ! _STRBUF_H */
