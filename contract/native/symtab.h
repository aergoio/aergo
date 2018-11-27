/**
 * @file    symtab.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _SYMTAB_H
#define _SYMTAB_H

#include "common.h"

#include "array.h"
#include "meta.h"
#include "value.h"

#define SYMTAB_BUCKET_SIZE          32

/*
typedef struct sym_rec_s {
    array_t fld_syms;
} sym_rec_t;

typedef struct sym_enum_s {
    array_t fld_syms;
} sym_enum_t;

typedef struct sym_func_s {
    array_t param_syms;
    array_t ret_syms;
} sym_func_t;

typedef struct sym_entry_s {
    sym_kind_t kind;
    char *name;

    union {
        sym_enum_t u_enum;
        sym_rec_t u_st;
        sym_func_t u_func;
    };

    bool is_used;
    meta_t *meta;
    value_t *val;
} sym_entry_t;
*/

typedef struct sym_rec_s {
    array_t *fld_syms;
} sym_rec_t;

typedef struct sym_func_s {
    array_t *param_syms;
    array_t *ret_syms;
} sym_func_t;

typedef struct symbol_s {
    sym_kind_t kind;
    modifier_t mod;

    char *name;

    union {
        sym_rec_t u_rec;
        sym_func_t u_func;
    };

    bool is_used;       /* whether it is referenced */
    bool is_checked;    /* whether it is checked */

    meta_t meta;        /* type metadata */
    value_t *val;       /* constant or aggregated value */

    int idx;            /* index of variable */
} symbol_t;

typedef struct symtab_s {
    array_t buckets[SYMTAB_BUCKET_SIZE];
} symtab_t;

symbol_t *symtab_lookup(symtab_t *symtab, char *name);

static inline symtab_t *
symtab_new(void)
{
    int i;
    symtab_t *symtab = xmalloc(sizeof(symtab_t));

    for (i = 0; i < SYMTAB_BUCKET_SIZE; i++) {
        array_init(&symtab->buckets[i]);
    }

    return symtab;
}

#endif /* no _SYMTAB_H */
