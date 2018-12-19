/**
 * @file    gen_bb.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _GEN_BB_H
#define _GEN_BB_H

#include "common.h"

#include "ir_bb.h"
#include "gen.h"

void bb_gen(gen_t *gen, ir_bb_t *bb); 
void br_gen(gen_t *gen, ir_bb_t *bb); 

#endif /* no _GEN_BB_H */
