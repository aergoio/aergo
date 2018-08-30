/**
 * @file    version.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _VERSION_H
#define _VERSION_H

#include "common.h"

#define MAJOR_VER       0
#define MINOR_VER       1

static inline boolean
version_check(int major, int minor)
{
    return major == MAJOR_VER && minor == MINOR_VER;
}

#endif /* no _VERSION_H */
