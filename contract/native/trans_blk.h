/**
 * @file    trans_blk.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _TRANS_BLK_H
#define _TRANS_BLK_H

#include "common.h"

#include "ast_blk.h"
#include "trans.h"

void blk_trans(trans_t *trans, ast_blk_t *blk);

#endif /* ! _TRANS_BLK_H */
