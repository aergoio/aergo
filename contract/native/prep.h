/**
 * @file    prep.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PREP_H
#define _PREP_H

#include "common.h"

#include "flag.h"
#include "ast.h"

#ifndef _IOBUF_T
#define _IOBUF_T
typedef struct iobuf_s iobuf_t;
#endif /* ! _IOBUF_T */

typedef struct prep_s {
    flag_t flag;

    char *path;
    char *work_dir;

    int offset;
    iobuf_t *src;

    src_pos_t pos;

    ast_t *ast;
} prep_t;

void prep(iobuf_t *src, flag_t flag, ast_t *ast);

#endif /* ! _PREP_H */
