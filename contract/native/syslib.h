/**
 * @file    syslib.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SYSLIB_H
#define _SYSLIB_H

#include "common.h"

#include "trans.h"
#include "gen.h"

#define SYSLIB_MODULE               "system"
#define SYSLIB_MAX_ARGS             16

#define SYS_FN(kind)                (&sys_fntab_[(kind)])

#define syslib_call_1(gen, kind, arg)                                                              \
    syslib_call((gen), (kind), 1, (arg))

#define syslib_call_2(gen, kind, arg1, arg2)                                                       \
    syslib_call((gen), (kind), 2, (arg1), (arg2))

#define syslib_make_malloc(size, align, pos)                                                       \
    syslib_make_alloc(FN_MALLOC, (size), (align), (pos))

#define syslib_make_alloca(size, align, pos)                                                       \
    syslib_make_alloc(FN_ALLOCA, (size), (align), (pos))

#ifndef _IR_ABI_T
#define _IR_ABI_T
typedef struct ir_abi_s ir_abi_t;
#endif /* ! _IR_ABI_T */

#ifndef _AST_EXP_T
#define _AST_EXP_T
typedef struct ast_exp_s ast_exp_t;
#endif /* ! _AST_EXP_T */

typedef struct sys_fn_s {
    char *name;
    char *qname;

    int param_cnt;
    type_t params[SYSLIB_MAX_ARGS];

    type_t result;
} sys_fn_t;

extern sys_fn_t sys_fntab_[FN_MAX];

void syslib_load(ast_t *ast);

ir_abi_t *syslib_abi(sys_fn_t *sys_fn);

BinaryenExpressionRef syslib_call(gen_t *gen, fn_kind_t kind, int argc, ...);

void syslib_gen(gen_t *gen, fn_kind_t kind);

ast_exp_t *syslib_make_alloc(fn_kind_t kind, uint32_t size, uint32_t align, src_pos_t *pos);
ast_exp_t *syslib_make_memcpy(ast_exp_t *dest_exp, ast_exp_t *src_exp, uint32_t size,
                              src_pos_t *pos);
ast_exp_t *syslib_make_memset(ast_exp_t *addr_exp, uint32_t val, uint32_t size, src_pos_t *pos);
ast_exp_t *syslib_make_strlen(ast_exp_t *addr_exp, src_pos_t *pos);
ast_exp_t *syslib_make_char_set(ast_exp_t *addr_exp, ast_exp_t *idx_exp, ast_exp_t *val_exp,
                                src_pos_t *pos);

#endif /* ! _SYSLIB_H */
