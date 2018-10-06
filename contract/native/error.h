/**
 * @file    error.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ERROR_H
#define _ERROR_H

#include "common.h"

#define ERROR_MAX_DESC_LEN      1024

#define FATAL(ec, ...)          error_exit((ec), LVL_FATAL, ## __VA_ARGS__)

#define ERROR(ec, trc, ...)                                                    \
    error_push((ec), LVL_ERROR, (trc), ## __VA_ARGS__)
#define INFO(ec, trc, ...)                                                     \
    error_push((ec), LVL_INFO, (trc), ## __VA_ARGS__)
#define WARN(ec, trc, ...)                                                     \
    error_push((ec), LVL_WARN, (trc), ## __VA_ARGS__)
#define DEBUG(ec, trc, ...)                                                    \
    error_push((ec), LVL_DEBUG, (trc), ## __VA_ARGS__)

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

    LVL_MAX
} errlvl_t;

typedef struct error_s {
    ec_t code;
    errlvl_t level;
    trace_t trc;
    char desc[ERROR_MAX_DESC_LEN];
} error_t;

char *error_to_string(ec_t ec);
ec_t error_to_code(char *str);

int error_count(void);
ec_t error_first(void);

void error_push(ec_t ec, errlvl_t lvl, trace_t *trc, ...);
error_t *error_pop(void);

void error_clear(void);
void error_dump(void);

void error_exit(ec_t ec, errlvl_t lvl, ...);

static inline int
error_cmp(const void *e1, const void *e2)
{
    return trace_cmp(&((error_t *)e1)->trc, &((error_t *)e2)->trc);
}

#endif /* ! _ERROR_H */
