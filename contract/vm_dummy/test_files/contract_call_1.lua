function constructor(init)
    system.setItem("count", init)
end

function inc()
    count = system.getItem("count")
    system.setItem("count", count + 1)
    return count
end

function get()
    return system.getItem("count")
end

function set(val)
    system.setItem("count", val)
end

abi.register(inc, get, set)
