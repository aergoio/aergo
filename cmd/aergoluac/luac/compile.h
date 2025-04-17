#ifndef _COMPILE_H
#define _COMPILE_H

typedef struct lua_State lua_State;

lua_State *luac_vm_newstate();
void luac_vm_close(lua_State *L);
const char *vm_compile(lua_State *L, const char *code, const char *byte, const char *abi);
const char *vm_loadfile(lua_State *L, const char *filename);
const char *vm_loadstring(lua_State *L, const char *source);
const char *vm_stringdump(lua_State *L);

#endif /* _COMPILE_H */