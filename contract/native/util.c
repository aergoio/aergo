/**
 * @file    util.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "strbuf.h"
#include "parser.h"

#include "util.h"

FILE *
open_file(char *path, char *mode)
{
    FILE *fp;
   
    fp = fopen(path, mode);
    if (fp == NULL)
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    return fp;
}

void
close_file(FILE *fp)
{
    // ignore error
    fclose(fp);
}

void
read_file(char *path, strbuf_t *sb)
{
    int n;
    FILE *fp;
    char buf[STRBUF_INIT_SIZE];

    fp = open_file(path, "r");

    while ((n = fread(buf, 1, sizeof(buf), fp)) > 0) {
        strbuf_append(sb, buf, n);
    }

    if (!feof(fp))
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    fclose(fp);
}

void
write_file(char *path, strbuf_t *sb)
{
    int n;
    FILE *fp;

    fp = open_file(path, "w");

    n = fwrite(strbuf_text(sb), strbuf_length(sb), 1, fp);
    if (n == 0)
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    fclose(fp);
}

char *
make_trace(char *path, yylloc_t *lloc)
{
#define TRACE_LINE_MAX      80
    int i, j;
    int nread;
    int tok_len;
    int adj_offset = lloc->first.offset;
    int adj_col = lloc->first.col;
    FILE *fp = open_file(path, "r");
    char *buf;

    ASSERT(adj_offset >= 0);
    ASSERT(adj_col > 0);

    tok_len = MIN(lloc->last.offset - lloc->first.offset, TRACE_LINE_MAX - 1);
    ASSERT(tok_len >= 0);

    if (adj_col + tok_len > TRACE_LINE_MAX) {
        adj_col = TRACE_LINE_MAX - tok_len;
        adj_offset += lloc->first.col - adj_col;
    }
    
    if (fseek(fp, adj_offset, SEEK_SET) < 0)
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    buf = xmalloc(TRACE_LINE_MAX * 3);

    nread = fread(buf, 1, TRACE_LINE_MAX, fp);
    if (nread <= 0 && !feof(fp))
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    for (i = 0; i < nread; i++) {
        if (buf[i] == '\n' || buf[i] == '\r')
            break;
    }
    buf[i++] = '\n';

    for (j = 0; j < adj_col - 1; j++) {
        buf[i++] = ' ';
    }

    strcpy(buf + i, ANSI_GREEN"^"ANSI_NONE);

    close_file(fp);

    return buf;
}

/* end of util.c */
