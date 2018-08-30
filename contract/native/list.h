/**
 * @file    list.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _LIST_H
#define _LIST_H

#include "common.h"

struct list_node_s {
  	struct list_node_s *next;
  	void *data;
} list_node_t;

typedef struct list_s {
  	list_node_t *head;
} list_t;

#define FOREACH(p, l)                                                          \
	for ((p) = (l)->head; (p); (p) = (p)->next)

#define FOREACH_SAFE(p, n, l)                                                  \
	for ((p) = (l)->head, (n) = (p)->next; (p); (p) = (n), (n) = (n)->next)

static inline list_t *
list_new(void)
{
	list_t *l;

	l = malloc(sizeof(list_t));
	l->head = NULL;

	return l;
}

static inline void
list_destroy(list_t *l)
{
	list_node_t *p, *n;

	FOREACH_SAFE(p, n, l) {
		free(p);
	}
	l->head = NULL;
}

static inline void
list_init(list_t *l)
{
	l->head = NULL;
}

static inline void
list_add(list_t *l, list_node_t *n)
{
	list_node_t *p = l->head;

	while (p != NULL) {
		p = p->next;
	}

	n->next = NULL;
	p->next = n;
}

static inline void
list_del(list_t *l, list_node_t *n)
{
	list_node_t *p = l->head;

	if (p == n) {
		l->head = n->next;
		return;
	}

	while (p->next != n) {
		p = p->next;
	}

	p->next = n->next;
	n->next = NULL;
}

#endif /* no _LIST_H */
