/**
 * @file    stack.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "stack.h"

void
stack_push(stack_t *stack, void *item)
{
    stack_node_t *node = xmalloc(sizeof(stack_node_t));

    node->item = item;
    node->next = stack->top;
    
    stack->size++;
    stack->top = node;
}

void *
stack_pop(stack_t *stack)
{
    stack_node_t *node = stack->top;
    void *item;

    if (node == NULL)
        return NULL;

    stack->top = node->next;
    item = node->item;

    xfree(node);

    return item;
}

/* end of stack.c */
