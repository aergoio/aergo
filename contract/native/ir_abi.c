/**
 * @file    ir_abi.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "gen_util.h"

#include "ir_abi.h"

ir_abi_t *
abi_new(ast_id_t *id)
{
    int i, j = 0;
    ast_id_t *ret_id = id->u_fn.ret_id;
    ir_abi_t *abi = xcalloc(sizeof(ir_abi_t));

    ASSERT(id != NULL);
    ASSERT(id->up != NULL);
    ASSERT1(is_fn_id(id), id->kind);

    abi->module = id->up->name;
    abi->name = id->u_fn.alias != NULL ? id->u_fn.alias : id->name;

    abi->param_cnt = vector_size(id->u_fn.param_ids);
    abi->params = xmalloc(sizeof(BinaryenType) * abi->param_cnt);

    vector_foreach(id->u_fn.param_ids, i) {
        ast_id_t *param_id = vector_get_id(id->u_fn.param_ids, i);

        abi->params[j] = meta_gen(&param_id->meta);
        param_id->idx = j++;
    }

    if (ret_id != NULL)
        abi->result = meta_gen(&ret_id->meta);
    else
        abi->result = BinaryenTypeNone();

    // TODO multiple return values
#if 0
    if (is_ctor_id(id)) {
        abi->params = xmalloc(sizeof(BinaryenType) * abi->param_cnt);

        vector_foreach(id->u_fn.param_ids, i) {
            ast_id_t *param_id = vector_get_id(id->u_fn.param_ids, i);

            abi->params[j] = meta_gen(&param_id->meta);
            param_id->idx = j++;
        }

        abi->result = meta_gen(&ret_id->meta);
    }
    else {
        if (ret_id != NULL) {
            if (is_tuple_id(ret_id))
                abi->param_cnt += vector_size(ret_id->u_tup.elem_ids);
            else
                abi->param_cnt++;
        }

        abi->params = xmalloc(sizeof(BinaryenType) * abi->param_cnt);

        vector_foreach(id->u_fn.param_ids, i) {
            ast_id_t *param_id = vector_get_id(id->u_fn.param_ids, i);

            abi->params[j] = meta_gen(&param_id->meta);
            param_id->idx = j++;
        }

        /* The return value is always passed as an address */
        if (ret_id != NULL) {
            if (is_tuple_id(ret_id)) {
                vector_foreach(ret_id->u_tup.elem_ids, i) {
                    ast_id_t *elem_id = vector_get_id(ret_id->u_tup.elem_ids, i);

                    abi->params[j] = BinaryenTypeInt32();
                    elem_id->idx = j++;
                }
            }
            else {
                abi->params[j] = BinaryenTypeInt32();
                ret_id->idx = j;
            }
        }

        abi->result = BinaryenTypeNone();
    }
#endif

    abi->spec = NULL;

    return abi;
}

/* end of ir_abi.c */
