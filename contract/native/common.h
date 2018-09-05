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
#include <ctype.h>

#include "xalloc.h"
#include "option.h"
#include "error.h"

#define EXECUTABLE          "aergoscc"

#define PATH_MAX_LEN        256
#define PATH_DELIM          '/'

#ifndef bool
#define bool                unsigned char
#endif

#ifndef true
#define true                1
#define false               0
#endif

#endif /*_COMMON_H */
