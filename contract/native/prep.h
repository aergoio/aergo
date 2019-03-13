/**
 * @file    prep.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PREP_H
#define _PREP_H

#include "common.h"

#include "flag.h"
#include "ast.h"
#include "iobuf.h"

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
