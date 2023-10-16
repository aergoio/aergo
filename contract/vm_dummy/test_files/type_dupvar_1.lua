state.var {
    Var1 = state.value(),
}
function GetVar1()
    return Var1:get()
end

state.var {
    Var1 = state.value(),
}
abi.register(GetVar1)
