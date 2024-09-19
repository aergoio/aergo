state.var {
    registry = state.map()
}

function set(name, address)
    registry[name] = address
end

function resolve(name)
    return registry[name]
end

abi.register(set)
abi.register_view(resolve)
