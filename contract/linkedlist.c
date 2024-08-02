// correct use:
// llist_add(&first, item);
// llist_prepend(&first, item);
// llist_remove(&first, item);
// llist_count(first);
// llist_get(first, pos);

#include <stddef.h>
#include "linkedlist.h"

typedef struct llitem llitem;
struct llitem {
    llitem *next;
};

void llist_add(void *pfirst, void *pto_add) {
  llitem **first, *to_add, *item;

  first = (llitem **) pfirst;
  to_add = (llitem *) pto_add;

  item = *first;
  if (item == 0) {
    *first = to_add;
  } else {
    while (item->next != 0) {
      item = item->next;
    }
    item->next = to_add;
  }

}

void llist_prepend(void *pfirst, void *pto_add) {
  llitem **first, *to_add, *item;

  first = (llitem **) pfirst;
  to_add = (llitem *) pto_add;

  item = *first;
  *first = to_add;
  to_add->next = item;

}

/* safer version: other threads can be iterating the list while item is removed */
/* caller should not release the memory immediately */
void llist_safe_remove(void *pfirst, void *pto_del) {
  llitem **first, *to_del, *item;

  first = (llitem **) pfirst;
  to_del = (llitem *) pto_del;

  item = *first;
  if (to_del == item) {
    *first = to_del->next;
  } else {
    while (item != NULL) {
      if (item->next == to_del) {
        item->next = to_del->next;
        break;
      }
      item = item->next;
    }
  }

}

void llist_remove(void *pfirst, void *pto_del) {
  llitem *to_del = (llitem *) pto_del;
  llist_safe_remove(pfirst, pto_del);
  to_del->next = NULL;  /* unsafe for concurrent threads without mutex */
}

int llist_count(void *list) {
  llitem *item = (llitem *) list;
  int count = 0;

  while (item) {
    count++;
    item = item->next;
  }

  return count;
}

void* llist_get(void *list, int pos) {
  llitem *item = (llitem *) list;
  int count = 0;

  while (item) {
    if (count==pos) return item;
    count++;
    item = item->next;
  }

  return NULL;
}
