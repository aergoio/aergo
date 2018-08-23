/**
 * @file    sc_util.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SC_UTIL_H
#define _SC_UTIL_H

#include "sc_common.h"

#include "sc_error.h"

#define sc_fclose       fclose

static inline FILE *
sc_fopen(char *path, char *mode)
{
    FILE *fp;

    fp = fopen(path, mode);
    if (fp == NULL)
        sc_fatal(ERROR_FILE_NOT_FOUND, path);

    return fp;
}

static inline void
sc_fseek(FILE *fp, long offset)
{
    if (fseek(fp, offset, SEEK_SET) != 0)
        sc_fatal(ERROR_FILE_READ_FAILED, ferror(fp));
}

static inline void
sc_fgets(FILE *fp, int buf_size, char *buf)
{
    if (fgets(buf, buf_size, fp) == NULL)
        sc_fatal(ERROR_FILE_READ_FAILED, ferror(fp));
}

#endif /* no _SC_UTIL_H */
