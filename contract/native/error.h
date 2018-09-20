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
            WARN(ERROR_INTERNAL, __SOURCE__,                                   \
                 "optional requirements failed with condition '"#cond"'");     \
    } while (0)

#define ASSERT(cond)                                                           \
    do {                                                                       \
        if (!(cond))                                                           \
            FATAL(ERROR_INTERNAL, __SOURCE__,                                  \
                  "internal error with condition '"#cond"'");                  \
    } while (0)

#define ASSERT2(cond, msg)                                                     \
    do {                                                                       \
        if (!(cond))                                                           \
            FATAL(ERROR_INTERNAL, __SOURCE__,                                  \
                  "internal error with '"#msg"'");                             \
    } while (0)

#define FATAL(ec, ...)                                                         \
    do {                                                                       \
        error_push((ec), LVL_FATAL, NULL, ## __VA_ARGS__);                     \
        error_dump();                                                          \
        exit(EXIT_FAILURE);                                                    \
    } while (0)

#define ERROR(ec, ...)                                                         \
    error_push((ec), LVL_ERROR, NULL, ## __VA_ARGS__)
#define INFO(ec, ...)                                                          \
    error_push((ec), LVL_INFO, NULL, ## __VA_ARGS__)
#define WARN(ec, ...)                                                          \
    error_push((ec), LVL_WARN, NULL, ## __VA_ARGS__)
#define DEBUG(ec, ...)                                                         \
    error_push((ec), LVL_DEBUG, NULL, ## __VA_ARGS__)
#define TRACE(ec, pos, ...)                                                    \
    error_push((ec), LVL_TRACE, (pos), ## __VA_ARGS__)

#define error_empty()           (error_count() == 0)

typedef enum ec_e {
    NO_ERROR = 0,
#undef error
#define error(code, msg)    code,
#include "error.list"

    ERROR_MAX
} ec_t;

typedef enum errlvl_e {
    LVL_FATAL = 0,
    LVL_ERROR,
    LVL_INFO,
    LVL_WARN,
    LVL_DEBUG,
    LVL_TRACE,

    LEVEL_MAX
} errlvl_t;

typedef struct errpos_s {
    char *path;
    int first_line;
    int first_col;
    int first_offset;
    int last_line;
    int last_col;
    int last_offset;
} errpos_t;

typedef struct error_s {
    ec_t code;
    errlvl_t level;
    errpos_t pos;
    char desc[ERROR_MAX_DESC_LEN];
} error_t;

char *error_to_string(ec_t ec);
ec_t error_to_code(char *str);

int error_count(void);

ec_t error_first(void);
ec_t error_last(void);

void error_push(ec_t ec, errlvl_t lvl, errpos_t *pos, ...);
error_t *error_pop(void);

void error_clear(void);
void error_dump(void);

static inline void
errpos_init(errpos_t *pos, char *path)
{
    pos->path = path;
    pos->first_line = 1;
    pos->first_col = 1;
    pos->first_offset = 0;
    pos->last_line = 1;
    pos->last_col = 1;
    pos->last_offset = 0;
}

#endif /* ! _ERROR_H */
