#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>
#include <stdarg.h>
#include "db_msg.h"

void set_error(request *req, const char *format, ...) {

  if (req == NULL) {
    return;  // avoid null pointer dereference
  }

  va_list args;
  va_start(args, format);

  // determine the required buffer size
  va_list args_copy;
  va_copy(args_copy, args);
  int size = vsnprintf(NULL, 0, format, args_copy) + 1;  // +1 for null terminator
  va_end(args_copy);

  if (size <= 0) {
    va_end(args);
    return;  // error in formatting
  }

  // allocate memory for the new error message
  char *new_error = malloc(size);
  if (new_error == NULL) {
    va_end(args);
    return;  // memory allocation failed
  }

  // format the error message
  vsnprintf(new_error, size, format, args);
  va_end(args);

  // free the old error message if it exists
  if (req->error != NULL) {
    free(req->error);
  }
  // set the new error message
  req->error = new_error;
}

// serialization

// copy int32 to buffer, stored as little endian, for unaligned access
void write_int(char *pdest, int value) {
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

double read_double(char *p) {
  double value;
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
void add_item(buffer *buf, const char *data, int len) {
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
  write_int(buf->ptr + buf->len, len);
  // copy item to buffer
  memcpy(buf->ptr + buf->len + 4, data, len);
  buf->len += item_size;
}

// now adding an additional byte for type
void add_typed_item(buffer *buf, char type, const char *data, int len) {
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
  write_int(buf->ptr + buf->len, len + 1);
  // store the type of the item
  buf->ptr[buf->len + 4] = type;
  // copy item to buffer
  memcpy(buf->ptr + buf->len + 5, data, len);
  buf->len += item_size;
}

// add items with type

void add_string(buffer *buf, const char *str) {
  if (str == NULL) str = "";
  add_typed_item(buf, 's', str, strlen(str) + 1);
}

void add_string_ex(buffer *buf, const char *str, int len) {
  if (str == NULL) str = "";
  add_typed_item(buf, 's', str, len + 1);
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

void add_bytes(buffer *buf, const char *data, int len) {
	add_typed_item(buf, 'y', data, len);
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

  if (p == NULL || position <= 0) {
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
  if (plen != NULL) *plen = len;
  return p;
}

int get_count(bytes *data) {
  int count = 0;
  int len;
  char *p = data->ptr;
  while (p < data->ptr + data->len) {
    len = read_int(p);
    p += 4;
    p += len;
    count++;
  }
  return count;
}

// get string at position
char *get_string(bytes *data, int position) {
  int len;
  char *p = get_item(data, position, &len);
  if (p == NULL || len < 2 || *p != 's') {
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
  if (p == NULL || len != 1+4) {
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
  if (p == NULL || len != 1+8) {
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
  if (p == NULL || len != 1+8) {
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
  if (p == NULL || len != 1+1) {
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
  if (p == NULL || len < 1 || *p != 'y') {
    return false;
  }
  p++; len--;
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
    if (len != 1+4) {
      return 0;
    }
    break;
  case 'l':
    if (len != 1+8) {
      return 0;
    }
    break;
  case 'd':
    if (len != 1+8) {
      return 0;
    }
    break;
  case 'b':
    if (len != 1+1) {
      return 0;
    }
  }
  return type;
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

void free_response(rresponse *resp) {
  if(resp->result.ptr != NULL) {
    free(resp->result.ptr);
    resp->result.ptr = NULL;
  }
  if (resp->error != NULL) {
    free(resp->error);
    resp->error = NULL;
  }
}
