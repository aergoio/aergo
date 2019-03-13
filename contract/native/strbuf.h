/**
 * @file    strbuf.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _STRBUF_H
#define _STRBUF_H

#include "common.h"

#define STRBUF_INIT_SIZE            256

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

#endif /* ! _STRBUF_H */
