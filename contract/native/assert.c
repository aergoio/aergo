/**
 * @file    assert.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "assert.h"
 
void
assert_exit(char *cond, char *file, int line, int argc, ...)
{
    int i;
    va_list vargs;
    char errdesc[DESC_MAX_LEN];

    snprintf(errdesc, sizeof(errdesc), "%s:%d: internal error with condition '%s'", 
             file, line, cond);

    fprintf(stderr, ANSI_RED"fatal"ANSI_RED": "ANSI_NONE"%s\n", errdesc);

    va_start(vargs, argc);

    for (i = 0; i < argc; i++) {
        int size;
        char *name;
        char c;
        uint32_t i32;
        uint64_t i64;

        name = va_arg(vargs, char *);
        size = va_arg(vargs, int);

        fprintf(stderr, "    %s = ", name);

        switch (size) {
        case 1:
            c = (char)va_arg(vargs, int);
            fprintf(stderr, "'%c' = 0x%x\n", c, c);
            break;
        case 2:
        case 4:
            i32 = va_arg(vargs, uint32_t);
            fprintf(stderr, "%d = %u = 0x%x\n", i32, i32, i32);
            break;
        case 8:
            i64 = va_arg(vargs, uint64_t);
            fprintf(stderr, "%"PRId64" = %"PRIu64" = 0x%"PRIx64"\n", 
                    (int64_t)i64, i64, i64);
            break;
        }
    }

    va_end(vargs);

    exit(EXIT_FAILURE);
}

/* end of assert.c */
