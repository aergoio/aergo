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
            WARN(INTERNAL_ERROR, __SOURCE__, "check failed, " #cond);          \
    } while (0)

#define ASSERT(cond)                                                           \
    do {                                                                       \
        if (!(cond))                                                           \
            FATAL(INTERNAL_ERROR, __SOURCE__, "assertion failed, " #cond);     \
    } while (0)

#define ASSERT2(cond, msg)                                                     \
    do {                                                                       \
        if (!(cond))                                                           \
            FATAL(INTERNAL_ERROR, __SOURCE__, "assertion failed, " #msg);      \
    } while (0)

#define FATAL(ec, ...)                                                         \
    do {                                                                       \
        error_dump((ec), LVL_FATAL, ## __VA_ARGS__);                           \
        exit(EXIT_FAILURE);                                                    \
    } while (0)

#define ERROR(ec, ...)          error_dump((ec), LVL_ERROR, ## __VA_ARGS__)
#define INFO(ec, ...)           error_dump((ec), LVL_INFO, ## __VA_ARGS__)
#define WARN(ec, ...)           error_dump((ec), LVL_WARN, ## __VA_ARGS__)
#define DEBUG(ec, ...)          error_dump((ec), LVL_DEBUG, ## __VA_ARGS__)

//#define error_empty()           (error_count() == 0)

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

char *error_text(ec_t ec);

/*
int error_count(void);

ec_t error_first(void);
ec_t error_last(void);

void error_push(ec_t ec, lvl_t lvl, ...);
error_t *error_pop(void);

void error_clear(void);
void error_dump(void);
*/

ec_t error_top(void);
void error_clear(void);
void error_dump(ec_t ec, lvl_t lvl, ...);

#endif /*_ERROR_H */
