/**
 * @file    list.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _LIST_H
#define _LIST_H

#include "common.h"

#define list_empty(l)           ((l)->head == NULL)

#define list_foreach(p, l)                                                     \
	for ((p) = (l)->head; (p); (p) = (p)->next)

#define list_foreach_safe(p, n, l)                                             \
	for ((p) = (l)->head, (n) = (p)->next; (p); (p) = (n), (n) = (n)->next)

#define list_foreach_reverse(p, l)                                             \
	for ((p) = (l)->tail; (p); (p) = (p)->prev)

#define list_foreach_reverse_safe(p, n, l)                                     \
	for ((p) = (l)->tail, (n) = (p)->prev; (p); (p) = (n), (n) = (n)->prev)

typedef struct list_node_s {
  	struct list_node_s *next;
  	struct list_node_s *prev;
  	void *item;
} list_node_t;

typedef struct list_s {
  	list_node_t *head;
  	list_node_t *tail;
} list_t;

static inline void
list_init(list_t *l)
{
	l->head = NULL;
	l->tail = NULL;
}

static inline list_t *
list_new(void)
{
	list_t *l = xmalloc(sizeof(list_t));

    list_init(l);

	return l;
}

static inline void
list_node_init(list_node_t *n)
{
    n->next = NULL;
    n->prev = NULL;
    n->item = NULL;
}

static inline list_node_t *
list_node_new(void)
{
    list_node_t *n = xmalloc(sizeof(list_node_t));

    list_node_init(n);

    return n;
}

static inline void
list_clear(list_t *l)
{
	list_node_t *p, *n;

	list_foreach_safe(p, n, l) {
		xfree(p);
	}

	l->head = NULL;
}

static inline void
list_destroy(list_t *l)
{
    list_clear(l);
    xfree(l);
}

static inline void
list_add(list_t *l, list_node_t *n)
{
	list_node_t *p = l->head;

    if (p == NULL) {
        l->head = n;
        l->tail = n;
        return;
    }

	while (p->next != NULL) {
		p = p->next;
	}

    n->prev = p;
	p->next = n;
    l->tail = n;
}

static inline void
list_del(list_t *l, list_node_t *n)
{
	list_node_t *p = l->head;

	if (p == n) {
		l->head = NULL;
		l->tail = NULL;
		return;
	}

	while (p->next != n) {
		p = p->next;
	}

    if (l->tail == n)
        l->tail = p;

	p->next = n->next;
    n->next->prev = p;
}

#endif /*_LIST_H */
