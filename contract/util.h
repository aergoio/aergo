#include "lua.h"

typedef struct sbuff {
	char *buf;
	int idx;
	int buf_len;
} sbuff_t;

typedef struct blockchain_ctx bc_ctx_t;

void lua_util_sbuf_init(sbuff_t *sbuf, int len);
char *lua_util_get_json_from_ret (lua_State *L, int nresult, sbuff_t *sbuf);
char *lua_util_get_json (lua_State *L, int idx);
char *lua_util_get_db_key(const bc_ctx_t *bc_ctx, const char *key);
int lua_util_json_to_lua (lua_State *L, char *json);
