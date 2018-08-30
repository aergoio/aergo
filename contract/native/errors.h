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
            FATAL(ERROR_INTERNAL, "assertion failed with condition: "#cond);   \
    } while (0)

#define INFO(ec, ...)                                                          \
    xperror(ANSI_WHITE"info", (ec), ## __VA_ARGS__)

#define WARN(ec, ...)                                                          \
    xperror(ANSI_YELLOW"warning", (ec), ## __VA_ARGS__)

#define ERROR(ec, ...)                                                         \
    xperror(ANSI_RED"error", (ec), ## __VA_ARGS__)

#define FATAL(ec, ...)                                                         \
    do {                                                                       \
        xperror(ANSI_RED"fatal", (ec), ## __VA_ARGS__);                        \
        exit(EXIT_FAILURE);                                                    \
    } while (0)

typedef enum ec_e {
    ERROR_NONE = 0,
#undef error
#define error(code, msg)    code,
#include "errors.list"

    ERROR_MAX
} ec_t;

extern char *errmsgs_[ERROR_MAX];

static inline void
xperror(char *lvl, ec_t ec, ...)
{
    va_list vargs;
    char errdesc[ERROR_MAX_DESC_LEN];

    va_start(vargs, ec);
    vsnprintf(errdesc, sizeof(errdesc), errmsgs_[ec], vargs);
    va_end(vargs);

    fprintf(stderr, "%s: "ANSI_DEFAULT"%s\n", lvl, errdesc);
}

#endif /* no _ERRORS_H */
