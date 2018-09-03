/**
 * @file    version.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _VERSION_H
#define _VERSION_H

#include "common.h"

#define MAJOR_VER       0
#define MINOR_VER       1

#define version_check(major, minor)                                            \
    ((major) == MAJOR_VER && (minor) == MINOR_VER)

#endif /*_VERSION_H */
