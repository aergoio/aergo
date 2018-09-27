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
            assert_exit(#cond, __SOURCE__, 0);                                 \
    } while (0)
#define ASSERT1(cond, p1)                                                      \
    do {                                                                       \
        if (!(cond))                                                           \
            assert_exit(#cond, __SOURCE__, 1, #p1, sizeof(p1), p1);            \
    } while (0)
#define ASSERT2(cond, p1, p2)                                                  \
    do {                                                                       \
        if (!(cond))                                                           \
            assert_exit(#cond, __SOURCE__, 2, #p1, sizeof(p1), p1,             \
                        #p2, sizeof(p2), p2);                                  \
    } while (0)

#define FATAL(ec, ...)          error_exit((ec), LVL_FATAL, ## __VA_ARGS__)

#define ERROR(ec, ...)                                                         \
    error_push((ec), LVL_ERROR, NULL, ## __VA_ARGS__)
#define INFO(ec, ...)                                                          \
    error_push((ec), LVL_INFO, NULL, ## __VA_ARGS__)
#define WARN(ec, ...)                                                          \
    error_push((ec), LVL_WARN, NULL, ## __VA_ARGS__)
#define DEBUG(ec, ...)                                                         \
    error_push((ec), LVL_DEBUG, NULL, ## __VA_ARGS__)
#define TRACE(ec, pos, ...)                                                    \
    error_push((ec), LVL_TRACE, (pos), FILENAME((pos)->path),                  \
               (pos)->first_line, ## __VA_ARGS__)

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

void error_exit(ec_t ec, errlvl_t lvl, ...);

void assert_exit(char *cond, char *file, int line, int argc, ...);

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
