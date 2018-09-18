/**
 * @file    list.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _LIST_H
#define _LIST_H

#include "common.h"

#define list_empty(L)           ((L)->head == NULL)

#define list_foreach(E, T, L)                                                  \
    for ((E) = (T *)((L)->head); (E) != NULL; (E) = (T *)((E)->link.next))

#define list_foreach_safe(E, N, T, L)                                          \
    for ((E) = (T *)((L)->head), (N) = (T *)((E)->link.next); (E) != NULL;     \
         (E) = (N), (N) = (T *)((N)->link.next))

#define list_foreach_reverse(E, T, L)                                          \
    for ((E) = (T *)((L)->tail); (E) != NULL; (E) = (T *)((E)->link.prev))

#define list_foreach_reverse_safe(E, N, T, L)                                  \
    for ((E) = (T *)((L)->tail), (N) = (T *)((E)->link.prev); (E) != NULL;     \
         (E) = (N), (N) = (T *)((N)->link.prev))

#define list_link_init(L)                                                      \
    do {                                                                       \
        (L)->next = NULL;                                                      \
        (L)->prev = NULL;                                                      \
    } while (0)

#define list_add_var(L, E)      list_add((L), ast_var_t, (E))
#define list_add_exp(L, E)      list_add((L), ast_exp_t, (E))
#define list_add_stmt(L, E)     list_add((L), ast_stmt_t, (E))

#define list_add(L, T, E)                                                      \
    do {                                                                       \
        T *p = (T *)((L)->head);                                               \
        if (p == NULL) {                                                       \
            (L)->head = (E);                                                   \
            (L)->tail = (E);                                                   \
            break;                                                             \
        }                                                                      \
        while (p->link.next != NULL) {                                         \
            p = (T *)(p->link.next);                                           \
        }                                                                      \
        (E)->link.prev = p;                                                    \
        p->link.next = (E);                                                    \
        (L)->tail = (E);                                                       \
    } while (0)

#define list_del(L, T, E)                                                      \
    do {                                                                       \
        T *p = (T *)((L)->head);                                               \
        if (p == (E)) {                                                        \
            (L)->head = NULL;                                                  \
            (L)->tail = NULL;                                                  \
            break;                                                             \
        }                                                                      \
        while (p->link.next != (E)) {                                          \
            p = (T *)(p->link.next);                                           \
        }                                                                      \
        if ((L)->tail == (E))                                                  \
            (L)->tail = p;                                                     \
        p->link.next = (E)->link.next;                                         \
        ((T *)((E)->link.next))->prev = p;                                     \
    } while (0)

#define list_clear(L, T)                                                       \
    do {                                                                       \
        T *entry, *next;                                                       \
        list_foreach_safe(entry, next, T, L) {                                 \
            xfree(entry);                                                      \
        }                                                                      \
        (L)->head = NULL;                                                      \
        (L)->tail = NULL;                                                      \
    } while (0)

#define list_destroy(L, T)                                                     \
    do {                                                                       \
        list_clear(L, T);                                                      \
        xfree(L);                                                              \
    } while (0)

typedef struct list_link_s {
  	void *next;
  	void *prev;
} list_link_t;

typedef struct list_s {
  	void *head;
  	void *tail;
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

#endif /* ! _LIST_H */
