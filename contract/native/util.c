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
        FATAL(ERROR_FILE_IO_FAILED, path, strerror(errno));

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
        FATAL(ERROR_FILE_IO_FAILED, strerror(errno));

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
        FATAL(ERROR_FILE_IO_FAILED, strerror(errno));

    fclose(fp);
}

char *
make_trace(char *file, yylloc_t *lloc)
{
    int i, j;
    int nread;
    int buf_size;
    char *buf;
    FILE *fp = open_file(file, "r");

    if (fseek(fp, lloc->first.offset, SEEK_SET) < 0)
        FATAL(ERROR_FILE_IO_FAILED, strerror(errno));

    buf_size = max(lloc->first.col * 2, STRBUF_INIT_SIZE);
    buf = malloc(buf_size);

    nread = fread(buf, buf_size, 1, fp);
    if (nread <= 0 && !feof(fp))
        FATAL(ERROR_FILE_IO_FAILED, strerror(errno));

    for (i = 0; i < nread; i++) {
        if (buf[i] == '\n' || buf[i] == '\r')
            break;
    }

    for (j = 0; j < lloc->first.col - 1; j++) {
        buf[i + j] = ' ';
    }

    strcpy(buf + i + j, ANSI_GREEN"^"ANSI_NONE);

    fclose(fp);

    return buf;
}

/* end of util.c */
