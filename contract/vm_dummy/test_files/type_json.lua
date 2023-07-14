state.var {
    table = state.value()
}

function set(val)
    table:set(json.decode(val))
end

function get()
    return table:get()
end

function getenc()
    return json.encode(table:get())
end

function getlen()
    a = table:get()
    return a[1], string.len(a[1])
end

function getAmount()
    return system.getAmount()
end

abi.register(set, get, getenc, getlen)
abi.payable(getAmount)
