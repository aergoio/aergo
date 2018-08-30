/**
 * @file    xutil.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _XUTIL_H
#define _XUTIL_H

#include "common.h"

#include "errors.h"

#define xfclose                 fclose

static inline FILE *
xfopen(char *path, char *mode)
{
    FILE *fp;

    fp = fopen(path, mode);
    if (fp == NULL)
        FATAL(ERROR_FILE_OPEN_FAILED, path, strerror(errno));

    return fp;
}

static inline void
xfseek(FILE *fp, long offset)
{
    if (fseek(fp, offset, SEEK_SET) != 0 && !feof(fp))
        FATAL(ERROR_FILE_SEEK_FAILED, strerror(errno));
}

static inline void
xfgets(FILE *fp, int buf_size, char *buf)
{
    buf[0] = '\0';

    if (fgets(buf, buf_size, fp) == NULL && !feof(fp)) {
        FATAL(ERROR_FILE_READ_FAILED, strerror(errno));
    }
}

static inline int
xfread(FILE *fp, int buf_size, char *buf) 
{
    int n;

    buf[0] = '\0';

    if ((n = fread(buf, 1, buf_size, fp)) == 0 && !feof(fp))
        FATAL(ERROR_FILE_READ_FAILED, strerror(errno));

    return n;
}

static inline int
xstrcpy(char *dest, char *src)
{
    strcpy(dest, src);

    return strlen(dest);
}

static inline int
xstrcat(char *dest, char *src)
{
    strcat(dest, src);

    return strlen(dest);
}

#endif /* no _XUTIL_H */
