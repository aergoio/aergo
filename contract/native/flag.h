/**
 * @file    flag.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _FLAG_H
#define _FLAG_H

#include "common.h"

#include "enum.h"

#define flag_set(x, y)              BIT_SET((x).val, (y))
#define flag_unset(x, y)            BIT_UNSET((x).val, (y))

#define is_flag_on(x, y)            IS_BIT_ON((x).val, (y))
#define is_flag_off(x, y)           IS_BIT_OFF((x).val, (y))

typedef struct flag_s {
    flag_val_t val;

    char *path;

    uint8_t opt_lvl;
    uint32_t stack_size;
} flag_t;

#endif /* ! _FLAG_H */
