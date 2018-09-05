/**
 * @file    error.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ERROR_H
#define _ERROR_H

#include "common.h"

#define ERROR_MAX_DESC_LEN      1024

#define ANSI_NONE               "\x1b[0m"
#define ANSI_RED                "\x1b[31m"
#define ANSI_GREEN              "\x1b[32m"
#define ANSI_YELLOW             "\x1b[33m"
#define ANSI_BLUE               "\x1b[34m"
#define ANSI_PURPLE             "\x1b[35m"
#define ANSI_WHITE              "\x1b[37m"

#define EXIT(ec)                                                               \
    do {                                                                       \
        fflush(stdout);                                                        \
        exit(ec);                                                              \
    } while (0)

#define CHECK(cond)                                                            \
    do {                                                                       \
        if (!(cond))                                                           \
            WARN(WARN_INTERNAL, __FILE__, __LINE__,                            \
                 "check failed with condition: " #cond);                       \
    } while (0)

#define ASSERT(cond)                                                           \
    do {                                                                       \
        if (!(cond))                                                           \
            FATAL(ERROR_INTERNAL, __FILE__, __LINE__,                          \
                  "assertion failed with condition: " #cond);                  \
    } while (0)

#define ASSERT2(cond, err)                                                     \
    do {                                                                       \
        if (!(cond))                                                           \
            FATAL(ERROR_INTERNAL, __FILE__, __LINE__,                          \
                  "assertion failed with error: " #err);                       \
    } while (0)

#define FATAL(ec, ...)                                                         \
    do {                                                                       \
        error_push((ec), LVL_FATAL, ## __VA_ARGS__);                           \
        error_dump();                                                          \
        EXIT(EXIT_FAILURE);                                                    \
    } while (0)

#define ERROR(ec, ...)                                                         \
    error_push((ec), LVL_ERROR, ## __VA_ARGS__)

#define INFO(ec, ...)                                                          \
    error_push((ec), LVL_INFO, ## __VA_ARGS__)

#define WARN(ec, ...)                                                          \
    error_push((ec), LVL_WARN, ## __VA_ARGS__)

#define DEBUG(ec, ...)                                                         \
    error_push((ec), LVL_DEBUG, ## __VA_ARGS__)

#define error_empty()        (error_count() == 0)

typedef enum ec_e {
    NO_ERROR = 0,
#undef error
#define error(code, msg)    code,
#include "error.list"

    ERROR_MAX
} ec_t;

typedef enum lvl_e {
    LVL_FATAL = 0,
    LVL_ERROR,
    LVL_INFO,
    LVL_WARN,
    LVL_DEBUG,

    LEVEL_MAX
} lvl_t;

typedef struct error_s {
    ec_t code;
    lvl_t level;
    char desc[ERROR_MAX_DESC_LEN];
} error_t;

int error_count(void);
ec_t error_last(void);
char *error_text(ec_t ec);

void error_push(ec_t ec, lvl_t lvl, ...);
error_t *error_pop(void);

void error_clear(void);

void error_dump(void);

#endif /*_ERROR_H */
