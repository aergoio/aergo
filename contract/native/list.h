/**
 * @file    list.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _LIST_H
#define _LIST_H

#include "common.h"

#define list_empty(list)                ((list)->head == NULL)

#define list_foreach(entry, type, list)                                        \
    for ((entry) = (type *)((list)->head); (entry) != NULL;                    \
         (entry) = (type *)((entry)->link.next))

#define list_foreach_safe(entry, N, type, list)                                \
    for ((entry) = (type *)((list)->head), (N) = (type *)((entry)->link.next); \
         (entry) != NULL; (entry) = (N), (N) = (type *)((N)->link.next))

#define list_foreach_reverse(entry, type, list)                                \
    for ((entry) = (type *)((list)->tail); (entry) != NULL;                    \
         (entry) = (type *)((entry)->link.prev))

#define list_foreach_reverse_safe(entry, N, type, list)                        \
    for ((entry) = (type *)((list)->tail), (N) = (type *)((entry)->link.prev); \
         (entry) != NULL; (entry) = (N), (N) = (type *)((N)->link.prev))

#define list_link_init(list)                                                   \
    do {                                                                       \
        (list)->next = NULL;                                                   \
        (list)->prev = NULL;                                                   \
    } while (0)

#define list_add_var(list, entry)       list_add((list), ast_var_t, (entry))
#define list_add_exp(list, entry)       list_add((list), ast_exp_t, (entry))
#define list_add_stmt(list, entry)      list_add((list), ast_stmt_t, (entry))
#define list_add_struct(list, entry)    list_add((list), ast_struct_t, (entry))

#define list_add(list, type, entry)                                            \
    do {                                                                       \
        type *p = (type *)((list)->head);                                      \
        if (p == NULL) {                                                       \
            (list)->head = (entry);                                            \
            (list)->tail = (entry);                                            \
            break;                                                             \
        }                                                                      \
        while (p->link.next != NULL) {                                         \
            p = (type *)(p->link.next);                                        \
        }                                                                      \
        (entry)->link.prev = p;                                                \
        p->link.next = (entry);                                                \
        (list)->tail = (entry);                                                \
    } while (0)

#define list_del(list, type, entry)                                            \
    do {                                                                       \
        type *p = (type *)((list)->head);                                      \
        if (p == (entry)) {                                                    \
            (list)->head = NULL;                                               \
            (list)->tail = NULL;                                               \
            break;                                                             \
        }                                                                      \
        while (p->link.next != (entry)) {                                      \
            p = (type *)(p->link.next);                                        \
        }                                                                      \
        if ((list)->tail == (entry))                                           \
            (list)->tail = p;                                                  \
        p->link.next = (entry)->link.next;                                     \
        ((type *)((entry)->link.next))->prev = p;                              \
    } while (0)

#define list_clear(list, type)                                                 \
    do {                                                                       \
        type *entry, *next;                                                    \
        list_foreach_safe(entry, next, type, list) {                           \
            xfree(entry);                                                      \
        }                                                                      \
        (list)->head = NULL;                                                   \
        (list)->tail = NULL;                                                   \
    } while (0)

#define list_destroy(list, type)                                               \
    do {                                                                       \
        list_clear(list, type);                                                \
        xfree(list);                                                           \
    } while (0)

#define list_join(dest, type, src)                                             \
    do {                                                                       \
        if ((dest)->head == NULL) {                                            \
            ASSERT((dest)->tail == NULL);                                      \
            (dest)->head = (src)->head;                                        \
            (dest)->tail = (src)->tail;                                        \
        }                                                                      \
        else {                                                                 \
            ((type *)((dest)->tail))->link.next = (src)->head;                 \
            ((type *)((src)->head))->link.prev = (dest)->tail;                 \
            (dest)->tail = (src)->tail;                                        \
        }                                                                      \
        (src)->head = NULL;                                                    \
        (src)->tail = NULL;                                                    \
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
