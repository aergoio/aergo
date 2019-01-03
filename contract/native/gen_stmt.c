/**
 * @file    gen_stmt.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_id.h"
#include "gen_exp.h"
#include "gen_util.h"

#include "gen_stmt.h"

static BinaryenExpressionRef
stmt_gen_assign(gen_t *gen, ast_stmt_t *stmt)
{
    uint32_t offset = 0;
    ast_exp_t *l_exp = stmt->u_assign.l_exp;
    ast_id_t *id = l_exp->id;
    BinaryenExpressionRef address, value;

    ASSERT(id != NULL);

    if (is_map_type(&id->meta))
        /* TODO: If the type of identifier is map,
         * lvalue and rvalue must be combined into a call expression */
        return NULL;

    value = exp_gen(gen, stmt->u_assign.r_exp);
    if (value == NULL || is_object_type(&id->meta))
        return NULL;

    if (id->is_param) {
        if (is_return_id(id))
            address = BinaryenGetLocal(gen->module, id->idx, BinaryenTypeInt32());
        else
            address = BinaryenGetLocal(gen->module, id->idx, meta_gen(&id->meta));

        return BinaryenStore(gen->module, TYPE_SIZE(l_exp->meta.type), offset, 0,
                             address, value, meta_gen(&l_exp->meta));
    }

    ASSERT1(is_var_id(id), id->kind);

    if (is_local_ref_exp(l_exp))
        return BinaryenSetLocal(gen->module, l_exp->u_lo.idx, value);

    if (is_id_ref_exp(l_exp)) {
        ASSERT(false);
        //ASSERT1(is_global_id(id), id->scope);
        return BinaryenSetGlobal(gen->module, "", value);
    }

    if (is_stack_ref_exp(l_exp)) {
        address = gen_i32(gen, l_exp->u_st.addr);
        offset = l_exp->u_st.offset;
    }
    else {
        gen->is_lval = true;
        address = exp_gen(gen, l_exp);
        offset = 0;
        gen->is_lval = false;
    }

    return BinaryenStore(gen->module, TYPE_SIZE(l_exp->meta.type), offset, 0, address,
                         value, meta_gen(&l_exp->meta));
}

static BinaryenExpressionRef
stmt_gen_ddl(gen_t *gen, ast_stmt_t *stmt)
{
    /* TODO */
    return NULL;
}

BinaryenExpressionRef
stmt_gen(gen_t *gen, ast_stmt_t *stmt)
{
    switch (stmt->kind) {
    case STMT_NULL:
        return BinaryenNop(gen->module);

    case STMT_EXP:
        return exp_gen(gen, stmt->u_exp.exp);

    case STMT_ASSIGN:
        return stmt_gen_assign(gen, stmt);

    case STMT_DDL:
        return stmt_gen_ddl(gen, stmt);

    default:
        ASSERT1(!"invalid statement", stmt->kind);
    }

    return NULL;
}

/* end of gen_stmt.c */