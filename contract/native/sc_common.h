/**
 * @file    sc_common.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SC_COMMON_H
#define _SC_COMMON_H

#include <stdio.h>
#include <stdlib.h>
#include <stdarg.h>
#include <string.h>

#define SC_EXECUTABLE       "ascc"

#define SC_VERSION_MAJOR    0
#define SC_VERSION_MINOR    1
#define SC_VERSION_PATCH    0

#define SC_PATH_MAX         256
#define SC_PATH_DELIM       '/'

#define sc_exit(ec)                                                            \
    do {                                                                       \
        fflush(stdout);                                                        \
        exit(ec);                                                              \
    } while (0)

#endif /* no _SC_COMMON_H */
