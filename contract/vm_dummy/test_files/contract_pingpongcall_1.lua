function constructor()
    system.setItem("key", "empty")
end

function start(addr)
    system.setItem("key", "start")
    contract.call(addr, "called")
    return system.getItem("key")
end

function callback()
    system.setItem("key", "callback")
end

function get()
    return system.getItem("key")
end

abi.register(start, callback, get)
