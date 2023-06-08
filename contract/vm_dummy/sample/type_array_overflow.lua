state.var{
    counts = state.array(1000000000000)
}

function get()
    return "hello"
end

abi.register(get)