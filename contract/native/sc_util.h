/**
 * @file    sc_util.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SC_UTIL_H
#define _SC_UTIL_H

#include "sc_common.h"

#include "sc_throw.h"

#define sc_fclose               fclose

static inline FILE *
sc_fopen(char *path, char *mode)
{
    FILE *fp;

    fp = fopen(path, mode);
    if (fp == NULL)
        sc_fatal(ERROR_FILE_OPEN_FAILED, path, strerror(errno));

    return fp;
}

static inline void
sc_fseek(FILE *fp, long offset)
{
    if (fseek(fp, offset, SEEK_SET) != 0 && !feof(fp))
        sc_fatal(ERROR_FILE_SEEK_FAILED, strerror(errno));
}

static inline void
sc_fgets(FILE *fp, int buf_size, char *buf)
{
    buf[0] = '\0';

    if (fgets(buf, buf_size, fp) == NULL && !feof(fp)) {
        sc_fatal(ERROR_FILE_READ_FAILED, strerror(errno));
    }
}

static inline int
sc_fread(FILE *fp, int buf_size, char *buf) 
{
    int n;

    buf[0] = '\0';

    if ((n = fread(buf, 1, buf_size, fp)) == 0 && !feof(fp))
        sc_fatal(ERROR_FILE_READ_FAILED, strerror(errno));

    return n;
}

static inline int
sc_strcpy(char *dest, char *src)
{
    strcpy(dest, src);

    return strlen(dest);
}

static inline int
sc_strcat(char *dest, char *src)
{
    strcat(dest, src);

    return strlen(dest);
}

#endif /* no _SC_UTIL_H */
