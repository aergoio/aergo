/**
 * @file    stack.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _STACK_H
#define _STACK_H

#include "common.h"

#define stack_size(stack)       ((stack)->size)
#define stack_empty(stack)      ((stack)->size == 0)
#define stack_head(stack)       ((stack)->head)
#define stack_tail(stack)       ((stack)->tail)
#define stack_top(stack)        stack_tail(stack)

#define stack_foreach(node, stack)                                             \
    for ((node) = (stack)->head; (node) != NULL; (node) = (node)->next)

typedef struct stack_node_s {
    struct stack_node_s *next;
    void *item;
} stack_node_t;

typedef struct stack_s {
    int size;
    stack_node_t *head;
    stack_node_t *tail;
} stack_t; 

void stack_push(stack_t *stack, void *item);
void *stack_pop(stack_t *stack);

static inline void
stack_init(stack_t *stack)
{
    ASSERT(stack != NULL);

    stack->size = 0;
    stack->head = NULL;
    stack->tail = NULL;
}

#endif /* ! _STACK_H */
