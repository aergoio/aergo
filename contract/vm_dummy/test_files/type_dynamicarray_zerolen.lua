state.var {
    fixedArray = state.array(0)
}

function Length()
	return fixedArray:length()
end

abi.register(Length)