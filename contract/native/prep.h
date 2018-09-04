/**
 * @file    preprocess.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _PREPROCESS_H
#define _PREPROCESS_H

#include "common.h"

#include "compile.h"

int preprocess(char *infile, char *outfile, opt_t opt);

void mark_fpos(char *path, int line, strbuf_t *sb);

#endif /*_PREPROCESS_H */
