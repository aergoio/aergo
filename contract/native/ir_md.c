/**
 * @file    ir_md.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ir_abi.h"

#include "ir_md.h"
 
ir_md_t *
md_new(char *name)
{
    ir_md_t *md = xcalloc(sizeof(ir_md_t));

    md->name = name;

    vector_init(&md->abis);
    vector_init(&md->fns);

    sgmt_init(&md->sgmt);

    return md;
}

void
md_add_imp(ir_md_t *md, ir_abi_t *abi)
{
    int i;

    vector_foreach(&md->abis, i) {
        ir_abi_t *old_abi = vector_get_abi(&md->abis, i);

        if (old_abi->param_cnt == abi->param_cnt &&
            memcmp(old_abi->params, abi->params, sizeof(BinaryenType) * old_abi->param_cnt) == 0 &&
            old_abi->result == abi->result &&
            strcmp(old_abi->module, abi->module) == 0 &&
            strcmp(old_abi->name, abi->name) == 0)
            return;
    }

    vector_add_last(&md->abis, abi);
}

void
md_add_fn(ir_md_t *md, ir_fn_t *fn)
{
    vector_add_last(&md->fns, fn);
}

/* end of ir_md.c */
