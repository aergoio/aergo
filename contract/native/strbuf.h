/**
 * @file    strbuf.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _STRBUF_H
#define _STRBUF_H

#include "common.h"

#define STRBUF_INIT_SIZE            1024

#define strbuf_empty(sb)            ((sb)->offset == 0)
#define strbuf_length(sb)           ((sb)->offset)
#define strbuf_text(sb)             ((sb)->buf)

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

#endif /*_STRBUF_H */
