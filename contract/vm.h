#ifndef _VM_H
#define _VM_H

#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>
#include "sqlite3-binding.h"

extern const char *construct_name;

lua_State *vm_newstate();
int vm_isnil(lua_State *L, int idx);
void vm_getfield(lua_State *L, const char *name);
void vm_get_constructor(lua_State *L);
void vm_remove_constructor(lua_State *L);
const char *vm_loadbuff(lua_State *L, const char *code, size_t sz, char *hex_id, int *service);
const char *vm_pcall(lua_State *L, int argc, int* nresult);
const char *vm_get_json_ret(lua_State *L, int nresult);
const char *vm_tostring(lua_State *L, int idx);
const char *vm_copy_result(lua_State *L, lua_State *target, int cnt);
sqlite3 *vm_get_db(lua_State *L);
void vm_get_abi_function(lua_State *L, char *fname);
int vm_is_payable_function(lua_State *L, char *fname);
char *vm_resolve_function(lua_State *L, char *fname, int *viewflag, int *payflag);
void vm_set_count_hook(lua_State *L, int limit);

#endif /* _VM_H */
