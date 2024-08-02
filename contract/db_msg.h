#ifndef DB_MSG_H
#define DB_MSG_H

typedef struct {
	char *ptr;
  int  len;
	int  allocated;
} buffer;

typedef struct {
	char *ptr;
	int  len;
} bytes;

typedef struct {
	int service;
	buffer result;
	char *error;
} request;

typedef struct {
	bytes result;
	char *error;
} rresponse;


void set_error(request *req, const char *format, ...);
void write_int(char *pdest, int value);
int read_int(char *p);
int64_t read_int64(char *p);
double read_double(char *p);

void add_item(buffer *buf, const char *data, int len);
void add_typed_item(buffer *buf, char type, const char *data, int len);
void add_string(buffer *buf, const char *str);
void add_string_ex(buffer *buf, const char *str, int len);
void add_int(buffer *buf, int value);
void add_int64(buffer *buf, int64_t value);
void add_double(buffer *buf, double value);
void add_bool(buffer *buf, bool value);
void add_bytes(buffer *buf, const char *data, int len);
void add_null(buffer *buf);

char *get_item(bytes *data, int position, int *plen);
int get_count(bytes *data);
char *get_string(bytes *data, int position);
char *get_string_ex(bytes *data, int position, int *plen);
int get_int(bytes *data, int position);
int64_t get_int64(bytes *data, int position);
double get_double(bytes *data, int position);
bool get_bool(bytes *data, int position);
bool get_bytes(bytes *data, int position, bytes *pbytes);
char *get_next_item(bytes *data, char *pdata, int *plen);
char get_type(char *ptr, int len);
double read_double(char *p);
bool read_bool(char *p);

void free_buffer(buffer *buf);
void free_response(rresponse *resp);


#endif // DB_MSG_H
