/**
 * @file    common.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _COMMON_H
#define _COMMON_H

#include <stdio.h>
#include <stdlib.h>
#include <stddef.h>
#include <stdarg.h>
#include <stdint.h>
#include <inttypes.h>
#include <float.h>
#include <string.h>
#include <errno.h>
#include <ctype.h>

#define ANSI_NONE           "\x1b[0m"
#define ANSI_RED            "\x1b[31m"
#define ANSI_GREEN          "\x1b[32m"
#define ANSI_YELLOW         "\x1b[33m"
#define ANSI_BLUE           "\x1b[34m"
#define ANSI_PURPLE         "\x1b[35m"
#define ANSI_WHITE          "\x1b[37m"

#define NAME_MAX_LEN        128

#define PATH_MAX_LEN        256
#define PATH_DELIM          '/'

#define FILENAME(f)         strrchr((f), '/') ? strrchr((f), '/') + 1 : (f)
#define __SOURCE__          FILENAME(__FILE__), __LINE__

#define ALIGN(v, n)         (((v) + ((n) - 1)) & ~((n) - 1))
#define ALIGN32(v)          ALIGN(v, 4)
#define ALIGN64(v)          ALIGN(v, 8)

#if !defined(__bool_true_false_are_defined) && !defined(__cplusplus)
typedef unsigned char bool;
#define true                1
#define false               0
#define __bool_true_false_are_defined
#endif

#include "xalloc.h"
#include "assert.h"
#include "flag.h"
#include "src_pos.h"
#include "error.h"

#define CHECK(stmt)                                                                      \
    do {                                                                                 \
        ec_t ec = (stmt);                                                                \
        if (ec != NO_ERROR)                                                              \
            return ec;                                                                   \
    } while (0)

#define RETURN(ec, pos, ...)                                                             \
    do {                                                                                 \
        error_push((ec), LVL_ERROR, (pos), ## __VA_ARGS__);                              \
        return (ec);                                                                     \
    } while (0)

#endif /* ! _COMMON_H */
