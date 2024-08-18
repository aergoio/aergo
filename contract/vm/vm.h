#ifndef _VM_H
#define _VM_H

#include <lualib.h>
#include <lauxlib.h>
#include <luajit.h>
#include <stdbool.h>
#include <stdint.h>

extern const char *construct_name;

#define ERR_BF_TIMEOUT "contract timeout"

void initViewFunction();
lua_State *vm_newstate(int hardfork_version);
void vm_push_abi_function(lua_State *L, char *fname);
int  vm_push_global_function(lua_State *L, char *fname);
void vm_remove_constructor(lua_State *L);
const char *vm_load_code(lua_State *L, const char *code, size_t sz, char *hex_id);
const char *vm_pre_run(lua_State *L);
const char *vm_call(lua_State *L, int argc, int *nresult);
const char *vm_get_json_ret(lua_State *L, int nresult, bool has_parent, int *err);
void vm_set_count_hook(lua_State *L, int limit);
void vm_db_release_resource(lua_State *L);
bool vm_is_hardfork(lua_State *L, int version);
void vm_set_timeout_hook(lua_State *L);
void vm_set_timeout_count_hook(lua_State *L, int limit);
int  vm_instcount(lua_State *L);
void vm_setinstcount(lua_State *L, int count);

#endif /* _VM_H */
