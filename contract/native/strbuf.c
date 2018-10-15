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

void
strbuf_load(strbuf_t *sb, char *path)
{
    int n;
    FILE *fp = open_file(path, "r");
    char buf[STRBUF_INIT_SIZE];

    while ((n = fread(buf, 1, sizeof(buf), fp)) > 0) {
        strbuf_ncat(sb, buf, n);
    }

    if (!feof(fp))
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    fclose(fp);
}

void
strbuf_dump(strbuf_t *sb, char *path)
{
    int n;
    FILE *fp;

    fp = open_file(path, "w");

    n = fwrite(strbuf_str(sb), 1, strbuf_size(sb), fp);
    if (n < 0)
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    fclose(fp);
}

/* end of strbuf.c */
