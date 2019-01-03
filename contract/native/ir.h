/**
 * @file    ir.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _IR_H
#define _IR_H

#include "common.h"

#include "array.h"

#ifndef _AST_ID_T
#define _AST_ID_T
typedef struct ast_id_s ast_id_t;
#endif /* ! _AST_ID_T */

#ifndef _IR_FN_T
#define _IR_FN_T
typedef struct ir_fn_s ir_fn_t;
#endif /* ! _IR_FN_T */

#ifndef _IR_SGMT_T
#define _IR_SGMT_T
typedef struct ir_sgmt_s ir_sgmt_t;
#endif /* ! _IR_SGMT_T */

typedef struct ir_s {
    array_t fns;

    ir_sgmt_t *sgmt;
} ir_t;

ir_t *ir_new(void);

void ir_add_global(ir_t *ir, ast_id_t *id);
void ir_add_fn(ir_t *ir, ir_fn_t *fn);

#endif /* no _IR_H */
