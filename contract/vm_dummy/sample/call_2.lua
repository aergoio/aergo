function constructor(addr)
    system.setItem("count", 99)
    system.setItem("addr", addr)
end
function add(amount)
    return contract.call.value(amount)(system.getItem("addr"), "inc")
end
function dadd()
    return contract.delegatecall(system.getItem("addr"), "inc")
end
function get()
    addr = system.getItem("addr")
    a = contract.call(addr, "get")
    return a
end
function dget()
    addr = system.getItem("addr")
    a = contract.delegatecall(addr, "get")
    return a
end
function set(val)
    contract.call(system.getItem("addr"), "set", val)
end
function dset(val)
    contract.delegatecall(system.getItem("addr"), "set", val)
end
abi.register(add,dadd, get, dget, set, dset)
