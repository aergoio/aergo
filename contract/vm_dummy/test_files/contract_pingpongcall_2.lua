function constructor(addr)
    system.setItem("key", "empty")
    system.setItem("addr", addr)
end

function called()
    system.setItem("key", "called")
    contract.call(system.getItem("addr"), "callback")
end

function get()
    return system.getItem("key")
end

abi.register(called, get)
