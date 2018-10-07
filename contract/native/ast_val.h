/**
 * @file    ast_val.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _AST_VAL_H
#define _AST_VAL_H

#include "common.h"

#define is_null_val(val)            ((val)->kind == VAL_NULL)
#define is_bool_val(val)            ((val)->kind == VAL_BOOL)
#define is_int_val(val)             ((val)->kind == VAL_INT)
#define is_float_val(val)           ((val)->kind == VAL_FP)
#define is_string_val(val)          ((val)->kind == VAL_STR)

#ifndef _AST_VAL_T
#define _AST_VAL_T
typedef struct ast_val_s ast_val_t;
#endif /* ! _AST_VAL_T */

typedef enum val_kind_e {
    VAL_NULL        = 0,
    VAL_BOOL,
    VAL_INT,
    VAL_FP,
    VAL_STR,
    VAL_MAX
} val_kind_t;

struct ast_val_s {
    val_kind_t kind;

    union {
        bool bv;
        int64_t iv;
        double dv;
        char *sv;
    };
};

static inline void
ast_val_init(ast_val_t *val)
{
    ASSERT(val != NULL);
    memset(val, 0x00, sizeof(ast_val_t));
}

static inline void
val_set_null(ast_val_t *val)
{
    val->kind = VAL_NULL;
}

static inline void 
val_set_bool(ast_val_t *val, bool bv)
{
    val->kind = VAL_BOOL;
    val->bv = bv;
}

static inline void 
val_set_int(ast_val_t *val, char *str)
{
    val->kind = VAL_INT;
    sscanf(str, "%"SCNd64, &val->iv);
}

static inline void 
val_set_hexa(ast_val_t *val, char *str)
{
    val->kind = VAL_INT;
    sscanf(str, "%"SCNx64, &val->iv);
}

static inline void 
val_set_fp(ast_val_t *val, char *str)
{
    val->kind = VAL_FP;
    sscanf(str, "%lf", &val->dv);
}

static inline void 
val_set_str(ast_val_t *val, char *str)
{
    val->kind = VAL_STR;
    val->sv = str;
}

#endif /* ! _AST_VAL_H */
