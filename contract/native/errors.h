/**
 * @file    errors.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ERRORS_H
#define _ERRORS_H

#include "common.h"

#include "errors.list"

#define ERROR_MAX_DESC_LEN      1024

#define ANSI_RED                "\x1b[31m"
#define ANSI_GREEN              "\x1b[32m"
#define ANSI_YELLOW             "\x1b[33m"
#define ANSI_BLUE               "\x1b[34m"
#define ANSI_PURPLE             "\x1b[35m"
#define ANSI_WHITE              "\x1b[37m"
#define ANSI_DEFAULT            "\x1b[0m"

#define EXIT(ec)                                                               \
    do {                                                                       \
        fflush(stdout);                                                        \
        exit(ec);                                                              \
    } while (0)

#define ASSERT(cond)                                                           \
    do {                                                                       \
        if (!(cond))                                                           \
            FATAL(EXECUTABLE, ERROR_INTERNAL,                                  \
                  "assertion failed with condition: "#cond);                   \
    } while (0)

#define INFO(fmt, ...)                                                         \
    xperror(EXECUTABLE, ANSI_WHITE"info", (fmt), ## __VA_ARGS__)

#define WARN(fmt, ...)                                                         \
    xperror(EXECUTABLE, ANSI_YELLOW"warning", (fmt), ## __VA_ARGS__)

#define ERROR(loc, fmt, ...)                                                   \
    xperror((loc), ANSI_RED"error", (fmt), ## __VA_ARGS__)

#define FATAL(fmt, ...)                                                        \
    do {                                                                       \
        xperror(EXECUTABLE, ANSI_RED"fatal", (fmt), ## __VA_ARGS__);           \
        exit(EXIT_FAILURE);                                                    \
    } while (0)

static inline void
xperror(char *loc, char *lvl, char *fmt, ...) 
{
    va_list vargs;
    char errdesc[ERROR_MAX_DESC_LEN];

    va_start(vargs, fmt);
    vsnprintf(errdesc, sizeof(errdesc), fmt, vargs);
    va_end(vargs);

    fprintf(stderr, ANSI_WHITE"%s: %s: "ANSI_DEFAULT"%s\n", loc, lvl, errdesc);
}

#endif /* no _ERRORS_H */
