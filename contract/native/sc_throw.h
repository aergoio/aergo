/**
 * @file    sc_error.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SC_ERROR_H
#define _SC_ERROR_H

#include "sc_common.h"

#define ERROR_MAX_DESC_LEN      1024

#define ANSI_RED                "\x1b[31m"
#define ANSI_GREEN              "\x1b[32m"
#define ANSI_YELLOW             "\x1b[33m"
#define ANSI_WHITE              "\x1b[37m"
#define ANSI_DEFAULT            "\x1b[0m"

#define ERROR_INTERNAL          "internal error: %s"
#define ERROR_INVALID_OPTION    "invalid command option: %s"
#define ERROR_FILE_NOT_FOUND    "file not found: %s"
#define ERROR_FILE_READ_FAILED  "%s failed: %s"
#define ERROR_PARSE_FAILED      "syntax error occurred"

#define sc_assert(cond)                                                        \
    do {                                                                       \
        if (!(cond))                                                           \
            sc_fatal(ERROR_INTERNAL,                                           \
                     "assertion failed with condition '"#cond"'");             \
    } while (0)

static inline void
sc_perror(char *loc, char *lvl, char *fmt, va_list errargs) 
{
    char errdesc[ERROR_MAX_DESC_LEN];

    vsnprintf(errdesc, sizeof(errdesc), fmt, errargs);

    fprintf(stderr, ANSI_WHITE"%s: "ANSI_RED"%s: "ANSI_DEFAULT"%s\n", loc, lvl, 
            errdesc);
}

static inline void
sc_warn(char *fmt, ...)
{
    va_list errargs;

    va_start(errargs, fmt);
    sc_perror(SC_EXECUTABLE, ANSI_YELLOW"warning", fmt, errargs);
    va_end(errargs);
}

static inline void
sc_error(char *loc, char *fmt, ...)
{
    va_list errargs;

    va_start(errargs, fmt);
    sc_perror(loc, ANSI_RED"error", fmt, errargs);
    va_end(errargs);
}

static inline void
sc_fatal(char *fmt, ...)
{
    va_list errargs;

    va_start(errargs, fmt);
    sc_perror(SC_EXECUTABLE, ANSI_RED"fatal", fmt, errargs);
    va_end(errargs);

    sc_exit(EXIT_FAILURE);
}

#endif /* no _SC_ERROR_H */
