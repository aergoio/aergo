#ifndef LINKEDLIST_H
#define LINKEDLIST_H

void llist_add(void *pfirst, void *pto_add);
void llist_prepend(void *pfirst, void *pto_add);
void llist_remove(void *pfirst, void *pto_del);
int llist_count(void *list);
void* llist_get(void *list, int pos);

#endif
