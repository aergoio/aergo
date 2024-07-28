#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <stdbool.h>
#include "db_msg.h"

void set_error(request *req, char *format, ...) {
	va_list args;
	va_start(args, format);
	vsnprintf(req->error, sizeof(req->error), format, args);
	va_end(args);
  // if out of memory, return a default error message
  if (req->error == NULL) {
    req->error = "failed: out of memory";
  }
}

// serialization

// copy int32 to buffer, stored as little endian, for unaligned access
void __attribute__((inline)) write_int(char *pdest, int value) {
  unsigned char *source = (unsigned char *) &value;
  unsigned char *dest = (unsigned char *) pdest;
  dest[0] = source[0];
  dest[1] = source[1];
  dest[2] = source[2];
  dest[3] = source[3];
}

// read_int32, stored as little endian, for unaligned access
int read_int(char *p) {
  int value;
  unsigned char *source = (unsigned char *) p;
  unsigned char *dest = (unsigned char *) &value;
  dest[0] = source[0];
  dest[1] = source[1];
  dest[2] = source[2];
  dest[3] = source[3];
  return value;
}

// read_int64, stored as little endian, for unaligned access
int64_t read_int64(char *p) {
  int64_t value;
  unsigned char *source = (unsigned char *) p;
  unsigned char *dest = (unsigned char *) &value;
  dest[0] = source[0];
  dest[1] = source[1];
  dest[2] = source[2];
  dest[3] = source[3];
  dest[4] = source[4];
  dest[5] = source[5];
  dest[6] = source[6];
  dest[7] = source[7];
  return value;
}

////////////////////////////////////////////////////////////////////////////////
// add item

// add item with 4 bytes length
void add_item(buffer *buf, char *data, int len) {
  int item_size = 4 + len;
  if (item_size > buf->allocated) {
    // compute new size
    int new_size = buf->allocated;
    if (new_size == 0) {
      new_size = 1024;
    }
    while (new_size < buf->len + item_size) {
      new_size *= 2;
    }
    // reallocate buffer
    buf->allocated = new_size;
    buf->ptr = (char *)realloc(buf->ptr, buf->allocated);
    if (buf->ptr == NULL) {
      // TODO: error handling
    }
  }
  // store the length of the item
  //*(int *)(req->result + req->used_size) = len;
  write_int32(buf->ptr + buf->len, len);
  // copy item to buffer
  memcpy(buf->ptr + buf->len + 4, data, len);
  buf->len += item_size;
}

// now adding an additional byte for type
void add_typed_item(buffer *buf, char type, char *data, int len) {
  int item_size = 4 + 1 + len;
  if (item_size > buf->allocated) {
    // compute new size
    int new_size = buf->allocated;
    if (new_size == 0) {
      new_size = 1024;
    }
    while (new_size < buf->len + item_size) {
      new_size *= 2;
    }
    // reallocate buffer
    buf->allocated = new_size;
    buf->ptr = (char *)realloc(buf->ptr, buf->allocated);
    if (buf->ptr == NULL) {
      // TODO: error handling
    }
  }
  // store the length of the item
  write_int32(buf->ptr + buf->len, len);
  // store the type of the item
  buf->ptr[buf->len + 4] = type;
  // copy item to buffer
  memcpy(buf->ptr + buf->len + 5, data, len);
  buf->len += item_size;
}

// add items with type

void add_string(buffer *buf, char *str) {
  if (str == NULL) str = "";
  add_typed_item(buf, 's', str, strlen(str));
}

void add_int(buffer *buf, int value) {
	add_typed_item(buf, 'i', (char *)&value, 4);
}

void add_int64(buffer *buf, int64_t value) {
	add_typed_item(buf, 'l', (char *)&value, 8);
}

void add_double(buffer *buf, double value) {
	add_typed_item(buf, 'd', (char *)&value, 8);
}

void add_bool(buffer *buf, bool value) {
	add_typed_item(buf, 'b', (char *)&value, 1);
}

void add_bytes(buffer *buf, char *data, int len) {
	add_typed_item(buf, 's', data, len);
}

void add_null(buffer *buf) {
	add_typed_item(buf, 'n', NULL, 0);
}

////////////////////////////////////////////////////////////////////////////////
// read item

// get item at position
char *get_item(bytes *data, int position, int *plen) {
  char *p = data->ptr;
  char *plimit = data->ptr + data->len;
  int len;
  int count = 1;

  if (p == NULL || position < 0) {
    return NULL;
  }

  while (count < position) {
    if (plimit - p < 4) {
      return NULL;
    }
    len = read_int(p);
    p += 4;
    p += len;
    count++;
  }

  if (plimit - p < 4) {
    return NULL;
  }
  len = read_int(p);
  p += 4;
  if (p + len > plimit) {
    return NULL;
  }
  *plen = len;
  return p;
}

// get string at position
char *get_string(bytes *data, int position, int *plen) {
  char *p = get_item(data, position, plen);
  if (p == NULL) {
    return NULL;
  }
  // check type
  if (*p != 's') {
    return NULL;
  }
  // skip type
  p++;
  return p;
}

// get int at position
int get_int(bytes *data, int position) {
  int len;
  char *p = get_item(data, position, &len);
  if (p == NULL || len != 4) {
    return 0;
  }
  // check type
  if (*p != 'i') {
    return 0;
  }
  // skip type
  p++;
  return read_int(p);
}

// get int64 at position
int64_t get_int64(bytes *data, int position) {
  int64_t value;
  int len;
  char *p = get_item(data, position, &len);
  if (p == NULL || len != 8) {
    return 0;
  }
  // check type
  if (*p != 'l') {
    return 0;
  }
  // skip type
  p++;
  memcpy(&value, p, 8);
  return value;
}

// get double at position
double get_double(bytes *data, int position) {
  double value;
  int len;
  char *p = get_item(data, position, &len);
  if (p == NULL || len != 8) {
    return 0;
  }
  // check type
  if (*p != 'd') {
    return 0;
  }
  // skip type
  p++;
  memcpy(&value, p, 8);
  return value;
}

// get bool at position
bool get_bool(bytes *data, int position) {
  int len;
  char *p = get_item(data, position, &len);
  if (p == NULL || len != 1) {
    return false;
  }
  // check type
  if (*p != 'b') {
    return false;
  }
  // skip type
  p++;
  return *p;
}

bool get_bytes(bytes *data, int position, bytes *pbytes) {
  int len;
  char *p = get_item(data, position, &len);
  if (p == NULL) {
    return false;
  }
  pbytes->ptr = p;
  pbytes->len = len;
  return true;
}

////////////////////////////////////////////////////////////////////////////////
// iterate over items

// get the next item
char *get_next_item(bytes *data, char *pdata, int *plen) {
  char *plimit = data->ptr + data->len;
  int len;

  if (pdata == NULL) {
    *plen = read_int(data->ptr);
    return data->ptr + 4;
  }

  if (pdata < data->ptr + 4 || pdata > plimit) {
    return NULL;
  }

  // skip this data
  pdata += *plen;
  // check if there is more data
  if (plimit - pdata < 4) {
    *plen = 0;
    return NULL;
  }
  // get the length of the next item
  len = read_int(pdata);
  // skip the length
  pdata += 4;
  // check if there is more data
  if (pdata + len > plimit) {
    *plen = 0;
    return NULL;
  }
  // return the length of the next item
  *plen = len;
  // return the pointer to the next item
  return pdata;
}

char get_type(char *ptr, int len) {
  char type = *ptr;
  switch (type) {
  case 'i':
    if (len != 4) {
      return 0;
    }
    break;
  case 'l':
    if (len != 8) {
      return 0;
    }
    break;
  case 'd':
    if (len != 8) {
      return 0;
    }
    break;
  case 'b':
    if (len != 1) {
      return 0;
    }
  }
  return type;
}

double read_double(char *p) {
  double value;
  memcpy(&value, p, 8);
  return value;
}

bool read_bool(char *p) {
  return *p;
}

////////////////////////////////////////////////////////////////////////////////

void free_buffer(buffer *buf) {
  if (buf->ptr != NULL) {
    free(buf->ptr);
    buf->ptr = NULL;
  }
}

void free_response(response *resp) {
  free_buffer(resp->result);
  if (resp->error != NULL) {
    free(resp->error);
    resp->error = NULL;
  }
}
