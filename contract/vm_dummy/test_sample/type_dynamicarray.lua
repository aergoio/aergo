state.var {
    dArr = state.array()
}

function Append(val)
	dArr:append(val)
end

function Get(idx)
	return dArr[idx]
end

function Set(idx, val)
	dArr[idx] = val
end

function Length()
	return dArr:length()
end

abi.register(Append, Get, Set, Length)