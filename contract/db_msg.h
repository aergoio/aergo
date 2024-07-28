
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
} response;
