#ifndef _VM_H
#define _VM_H

#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>
#include <stdbool.h>
#include <stdint.h>

extern const char *construct_name;

#define ERR_BF_TIMEOUT "contract timeout"

lua_State *vm_newstate(int hardfork_version);
void vm_closestates(lua_State *s[], int count);
int vm_autoload(lua_State *L, char *func_name);
void vm_remove_constructor(lua_State *L);
const char *vm_loadbuff(lua_State *L, const char *code, size_t sz, char *hex_id);
const char *vm_pcall(lua_State *L, int argc, int *nresult);
const char *vm_get_json_ret(lua_State *L, int nresult, bool has_parent, int *err);
const char *vm_copy_result(lua_State *L, lua_State *target, int cnt);
void vm_get_abi_function(lua_State *L, char *fname);
void vm_set_count_hook(lua_State *L, int limit);
void vm_db_release_resource(lua_State *L);
bool vm_is_hardfork(lua_State *L, int version);
void initViewFunction();
void vm_set_timeout_hook(lua_State *L);
void vm_set_timeout_count_hook(lua_State *L, int limit);
int vm_instcount(lua_State *L);
void vm_setinstcount(lua_State *L, int count);
const char *vm_loadcall(lua_State *L);

#endif /* _VM_H */
