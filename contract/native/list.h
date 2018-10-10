/**
 * @file    list.h
 * @copyright defined in aergo/LICENSE.txt
 */

#ifndef _LIST_H
#define _LIST_H

#include "common.h"

#define is_empty_list(list)             ((list)->head == NULL)
#define list_size(list)                 (list)->size
#define list_item(node, type)           (type *)((node)->item)

#define list_foreach(node, list)                                               \
    for ((node) = (list)->head; (node) != NULL; (node) = (node)->next)

#define list_foreach_safe(node, save, list)                                    \
    for ((node) = (list)->head, (save) = (node)->next;                         \
         (node) != NULL; (node) = (save), (save) = (save)->next)

#define list_foreach_reverse(node, list)                                       \
    for ((node) = (list)->tail; (node) != NULL; (node) = (node)->prev)

#define list_foreach_reverse_safe(node, next, type, list)                      \
    for ((node) = (list)->tail, (save) = (node)->prev;                         \
         (node) != NULL; (node) = (save), (save) = (save)->prev)

#define list_node_init(node)                                                   \
    do {                                                                       \
        (node)->next = NULL;                                                   \
        (node)->prev = NULL;                                                   \
        (node)->item = NULL;                                                   \
    } while (0)

#define list_init(list)                                                        \
    do {                                                                       \
        (list)->head = NULL;                                                   \
        (list)->tail = NULL;                                                   \
        (list)->size = 0;                                                      \
    } while (0)

typedef struct list_node_s {
  	struct list_node_s *next;
  	struct list_node_s *prev;
    void *item;
} list_node_t;

typedef struct list_s {
    int size;
  	list_node_t *head;
  	list_node_t *tail;
} list_t;

static inline list_t *
list_new(void)
{
	list_t *list = xmalloc(sizeof(list_t));

    list_init(list);

	return list;
}

static inline list_node_t *
list_node_new(void *item)
{
    list_node_t *node = xmalloc(sizeof(list_node_t));

    list_node_init(node);
    node->item = item;

    return node;
}

static inline void
list_add_first(list_t *list, void *item)
{
    list_node_t *node;

    if (item == NULL)
        return;

    node = list_node_new(item);

    if (list->head == NULL) {
        ASSERT(list->tail == NULL);
        list->head = node;
        list->tail = node;
        return;
    }

    node->next = list->head;
    list->head->prev = node;
    list->head = node;
    list->size++;
}

static inline void
list_add_last(list_t *list, void *item)
{
    list_node_t *node;

    if (item == NULL)
        return;

    node = list_node_new(item);

    if (list->tail == NULL) {
        ASSERT(list->head == NULL);
        list->head = node;
        list->tail = node;
        return;
    }

    node->prev = list->tail;
    list->tail->next = node;
    list->tail = node;
    list->size++;
}

static inline list_node_t *
list_find(list_t *list, void *item)
{
    list_node_t *node;

    if (item == NULL)
        return NULL;

    list_foreach(node, list) {
        if (node->item == item)
            return node;
    }

    return NULL;
}

static inline void
list_del(list_t *list, list_node_t *node)
{
    ASSERT(node != NULL);

    if (list->head == node && list->tail == node) {
        list->head = NULL;
        list->tail = NULL;
    }
    else if (list->head == node) {
        list->head = node->next;
        node->next->prev = NULL;
    }
    else if (list->tail == node) {
        list->tail = node->prev;
        node->prev->next = NULL;
    }
    else {
        node->prev->next = node->next;
        node->next->prev = node->prev;
    }

    node->next = NULL;
    node->prev = NULL;
    list->size--;
}

static inline void
list_clear(list_t *list)
{
    list_node_t *node, *next;

    list_foreach_safe(node, next, list) {
        list_del(list, node);
        xfree(node);
    }

    ASSERT(list_size(list) == 0);
}

static inline void
list_join(list_t *dest, list_t *src)
{
    ASSERT(src != NULL);
    ASSERT(dest != NULL);

    if (dest->head == NULL) {
        ASSERT(dest->tail == NULL);
        dest->head = src->head;
        dest->tail = src->tail;
    }
    else {
        dest->tail->next = src->head;
        src->head->prev = dest->tail;
        dest->tail = src->tail;
    }

    src->head = NULL;
    src->tail = NULL;
}

#endif /* ! _LIST_H */
