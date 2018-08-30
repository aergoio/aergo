/**
 * @file    common.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _COMMON_H
#define _COMMON_H

#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include <string.h>
#include <errno.h>

#define EXECUTABLE          "aergoscc"

#define PATH_MAX_LEN        256
#define PATH_DELIM          '/'

#define RC_OK               0
#define RC_ERROR            (-1)

#define boolean             unsigned char
#ifndef true
#define true                1
#define false               0
#endif

#endif /* no _COMMON_H */
