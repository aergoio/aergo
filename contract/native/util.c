/**
 * @file    util.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "strbuf.h"

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

    n = fwrite(strbuf_text(sb), 1, strbuf_length(sb), fp);
    if (n == 0)
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    fclose(fp);
}

char *
trim_str(char *str)
{
    int i;
    int str_len = strlen(str);
    char *ptr = str;

    for (i = 0; i < str_len; i++) {
        if (!isspace(str[i]))
            break;

        ptr++;
    }

    str_len = strlen(ptr);

    for (i = str_len - 1; i >= 0; i--) {
        if (!isspace(ptr[i]))
            break;

        ptr[i] = '\0';
    }

    return ptr;
}

/* end of util.c */
