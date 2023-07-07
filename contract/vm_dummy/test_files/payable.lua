state.var {
    Data = state.value()
}
function save(data)
    Data:set(data)
end

function load()
    return Data:get()
end

abi.register(load)
abi.payable(save)
