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

    ASSERT(path != NULL);
    ASSERT(mode != NULL);

    fp = fopen(path, mode);
    if (fp == NULL)
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    return fp;
}

void
close_file(FILE *fp)
{
    ASSERT(fp != NULL);

    // ignore error
    fclose(fp);
}

void
write_file(char *path, char *str, int len)
{
    int n;
    FILE *fp = open_file(path, "w+");

    n = fwrite(str, len, 1, fp);
    if (n == 0)
        FATAL(ERROR_FILE_IO, path, strerror(errno));

    close_file(fp);
}

char *
strtrim(char *str, char *ptn)
{
    int i;
    int str_len;
    char *ptr = str;

    ASSERT(str != NULL);
    ASSERT(ptn != NULL);

    str_len = strlen(str);

    for (i = 0; i < str_len; i++) {
        if (strchr(ptn, str[i]) == NULL)
            break;

        ptr++;
    }

    str_len = strlen(ptr);

    for (i = str_len - 1; i >= 0; i--) {
        if (strchr(ptn, ptr[i]) == NULL)
            break;

        ptr[i] = '\0';
    }

    return ptr;
}

void
strset(char *buf, char ch, int size)
{
    ASSERT(buf != NULL);

    memset(buf, ch, size);
    buf[size] = '\0';
}

/* end of util.c */
