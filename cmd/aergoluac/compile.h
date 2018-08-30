#ifndef _COMPILE_H
#define _COMPILE_H

typedef struct lua_State lua_State;

lua_State *vm_newstate();
void vm_close(lua_State *L);
const char *vm_compile(lua_State *L, const char *code, const char *byte, const char *abi);
const char *vm_stringdump(lua_State *L, const char *code);

#endif /* _COMPILE_H */