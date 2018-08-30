/**
 * @file    util.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "strbuf.h"

#include "util.h"

void
read_file(char *path, strbuf_t *sb)
{
    int n;
    FILE *fp;
    char buf[STRBUF_INIT_SIZE];

    fp = fopen(path, "r");
    if (fp == NULL)
        FATAL(ERROR_FILE_OPEN_FAILED, path, strerror(errno));

    while ((n = fread(buf, 1, sizeof(buf), fp)) > 0) {
        strbuf_append(sb, buf, n);
    }

    if (!feof(fp))
        FATAL(ERROR_FILE_READ_FAILED, strerror(errno));

    fclose(fp);
}

/* end of util.c */
