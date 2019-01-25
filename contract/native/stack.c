/**
 * @file    stack.c
 * @copyright defined in aergo/LICENSE.txt
 */

#include "common.h"

#include "stack.h"

void
stack_push(stack_t *stack, void *item)
{
    stack_node_t *n = xmalloc(sizeof(stack_node_t));

    ASSERT(stack != NULL);

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

vector_t *
stack_to_vector(stack_t *stack, int (*cmp_fn)(const void *, const void *))
{
    vector_t *vector = vector_new();
    stack_node_t *n;

    stack_foreach(n, stack) {
        vector_sadd(vector, n->item, cmp_fn);
    }

    return vector;
}

/* end of stack.c */
