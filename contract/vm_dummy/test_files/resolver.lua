state.var {
    xx = state.map()
}

function set(name, address)
    xx[name] = address
end

function resolve(name)
    return xx[name]
end

abi.register(set)
abi.register_view(resolve)
