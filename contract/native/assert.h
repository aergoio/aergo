/**
 * @file    assert.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _ASSERT_H
#define _ASSERT_H

#include "common.h"

#define ASSERT(cond)                                                                     \
    do {                                                                                 \
        if (!(cond))                                                                     \
            assert_exit(#cond, __SOURCE__, 0);                                           \
    } while (0)

#define ASSERT1(cond, p1)                                                                \
    do {                                                                                 \
        if (!(cond))                                                                     \
            assert_exit(#cond, __SOURCE__, 1, #p1, sizeof(p1), p1);                      \
    } while (0)

#define ASSERT2(cond, p1, p2)                                                            \
    do {                                                                                 \
        if (!(cond))                                                                     \
            assert_exit(#cond, __SOURCE__, 2, #p1, sizeof(p1), p1, #p2, sizeof(p2), p2); \
    } while (0)

void assert_exit(char *cond, char *file, int line, int argc, ...);

#endif /* no _ASSERT_H */
