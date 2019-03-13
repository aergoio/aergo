/**
 * @file    strbuf.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "util.h"

#include "strbuf.h"

void
strbuf_init(strbuf_t *sb)
{
    sb->size = STRBUF_INIT_SIZE;
    sb->offset = 0;
    sb->buf = xmalloc(sb->size + 1);
    sb->buf[0] = '\0';
}

void
strbuf_reset(strbuf_t *sb)
{
    sb->offset = 0;
    sb->buf[0] = '\0';
}

void
strbuf_ncat(strbuf_t *sb, char *str, int n)
{
    if (sb->offset + n > sb->size) {
        /* double up */
        sb->size += MAX(sb->size, n);
        sb->buf = xrealloc(sb->buf, sb->size + 1);
    }

    memcpy(sb->buf + sb->offset, str, n);

    sb->offset += n;
    sb->buf[sb->offset] = '\0';
}

void
strbuf_trunc(strbuf_t *sb, int n)
{
    sb->offset -= n;
    sb->buf[sb->offset] = '\0';
}

void
strbuf_copy(strbuf_t *src, strbuf_t *dest)
{
    strbuf_reset(dest);
    strbuf_ncat(dest, src->buf, src->offset);
}

/* end of strbuf.c */
