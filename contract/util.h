#include "lua.h"

typedef struct sbuff {
	char *buf;
	int idx;
	int buf_len;
} sbuff_t;

void lua_util_sbuf_init(sbuff_t *sbuf, int len);
char *lua_util_get_json_from_ret (lua_State *L, int nresult, sbuff_t *sbuf);
