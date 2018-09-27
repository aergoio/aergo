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

char *
strtrim(char *str)
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

void
strset(char *buf, char ch, int size)
{
    memset(buf, ch, size);
    buf[size] = '\0';
}

/* end of util.c */
