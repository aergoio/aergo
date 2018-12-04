#include "lua.h"
#include "vm.h"
#include <stdbool.h>

typedef struct sbuff {
	char *buf;
	int idx;
	int buf_len;
} sbuff_t;

void lua_util_sbuf_init(sbuff_t *sbuf, int len);
char *lua_util_get_json (lua_State *L, int idx, bool json_form);
char *lua_util_get_db_key(const char *key);
int lua_util_json_to_lua (lua_State *L, char *json, bool check);
char *lua_util_get_json_from_stack (lua_State *L, int start, int end, bool json_form);
int luaopen_json(lua_State *L);
