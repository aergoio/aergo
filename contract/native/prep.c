/**
 * @file    prep.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "ast_imp.h"
#include "parse.h"
#include "check.h"
#include "stack.h"

#include "prep.h"

static bool
push_imp(stack_t *refs, ast_imp_t *imp)
{
    stack_node_t *node;

    stack_foreach(node, refs) {
        if (strcmp(node->item, imp->path) == 0) {
            ERROR(ERROR_INFINITE_IMPORT, &imp->pos, FILENAME(imp->path));
            return false;
        }
    }

    stack_push(refs, imp->path);

    return true;
}

static void
pop_imp(stack_t *refs)
{
    stack_pop(refs);
}

static void
attach(ast_id_t *id, vector_t *ids)
{
    int i;

    ASSERT(id->is_checked);

    vector_foreach(ids, i) {
        if (strcmp(vector_get_id(ids, i)->name, id->name) == 0)
            return;
    }

    id->is_imported = true;

    vector_add_last(ids, id);
}

static void
subst(ast_t *ast, flag_t flag, stack_t *refs, vector_t *ids)
{
    int i, j;

    vector_foreach(&ast->imps, i) {
        ast_imp_t *imp = vector_get_imp(&ast->imps, i);

        if (push_imp(refs, imp)) {
            ast_t *imp_ast = NULL;

            parse(imp->path, flag, &imp_ast);

            if (imp_ast != NULL) {
                ast_blk_t *root = imp_ast->root;

                ASSERT(is_empty_vector(&root->stmts));

                subst(imp_ast, flag, refs, ids);
                check(imp_ast, flag);

                vector_foreach(&root->ids, j) {
                    attach(vector_get_id(&root->ids, j), ids);
                }
            }

            pop_imp(refs);
        }
    }
}

void
prep(ast_t *ast, flag_t flag, char *path)
{
    stack_t refs;
    vector_t ids;
    ast_blk_t *root = ast->root;

    ASSERT(is_empty_vector(&root->stmts));

    stack_init(&refs);
    vector_init(&ids);

    stack_push(&refs, path);

    subst(ast, flag, &refs, &ids);

    vector_join_first(&root->ids, &ids);

    stack_pop(&refs);
}

/* end of prep.c */
