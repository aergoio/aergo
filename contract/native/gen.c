/**
 * @file    gen.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_blk.h"
#include "gen_id.h"

#include "gen.h"

static void
gen_init(gen_t *gen, ast_t *ast, flag_t flag)
{
    gen->flag = flag;
    gen->root = ast->root;
    gen->module = BinaryenModuleCreate();
}

void
gen(ast_t *ast, flag_t flag)
{
    int i;
    gen_t gen;

    if (ast == NULL)
        return;

    gen_init(&gen, ast, flag);

    for (i = 0; i < array_size(&gen.root->ids); i++) {
        id_gen(&gen, array_item(&gen.root->ids, i, ast_id_t));
    }
}

/* end of gen.c */
