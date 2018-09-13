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
#include <string.h>
#include <errno.h>
#include <ctype.h>

#include "xalloc.h"
#include "option.h"
#include "error.h"

#define PATH_MAX_LEN        256
#define PATH_DELIM          '/'

#define FILENAME(f)         strrchr((f), '/') ? strrchr((f), '/') + 1 : (f)
#define __SOURCE__          FILENAME(__FILE__), __LINE__

#if !defined(__bool_true_false_are_defined) && !defined(__cplusplus)
typedef unsigned char bool;
#define true                1
#define false               0
#define __bool_true_false_are_defined
#endif

#endif /*_COMMON_H */
