state.var{
	Var1 = state.value(),
}
function GetVar1()
	return Var1:get()
end
function Work()
	state.var{
		Var1 = state.value(),
	}
end
abi.register(GetVar1, Work)