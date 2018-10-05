/**
 * @file    stack.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "stack.h"

void
stack_push(stack_t *stack, void *item)
{
    stack_node_t *n;

    ASSERT(stack != NULL);

    n = xmalloc(sizeof(stack_node_t));

    n->item = item;
    n->next = NULL;

    if (stack->head == NULL) {
        ASSERT(stack->tail == NULL);
        stack->head = n;
    }
    else {
        ASSERT(stack->tail != NULL);
        stack->tail->next = n;
    }

    stack->size++;
    stack->tail = n;
}

void *
stack_pop(stack_t *stack)
{
    stack_node_t *n;
    void *item;

    ASSERT(stack != NULL);

    if (stack->tail == NULL)
        return NULL;

    if (stack->head == stack->tail) {
        n = stack->head;
        stack->head = NULL;
        stack->tail = NULL;
    }
    else {
        stack_node_t *prev = stack->head;
        while (prev->next != stack->tail) {
            prev = prev->next;
        }
        n = stack->tail;
        stack->tail = prev;
        prev->next = NULL;
    }

    item = n->item;
    free(n);

    stack->size--;
    ASSERT(stack->size >= 0);

    return item;
}

array_t *
stack_to_array(stack_t *stack, int (*cmp_fn)(const void *, const void *))
{
    array_t *array = array_new();
    stack_node_t *n;

    stack_foreach(n, stack) {
        array_sadd(array, n->item, cmp_fn);
    }

    return array;
}

/* end of stack.c */
