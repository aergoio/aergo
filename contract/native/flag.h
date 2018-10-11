/**
 * @file    flag.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _FLAG_H
#define _FLAG_H

#include "common.h"

#define flag_set(x, y)      ((x) |= (y))
#define flag_on(x, y)       (((x) & (y)) == (y))
#define flag_off(x, y)      (((x) & (y)) != (y))

#endif /* ! _FLAG_H */
