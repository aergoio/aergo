function constructor(addr)
    system.setItem("count", 99)
    system.setItem("addr", addr)
end
function add(amount)
    first = contract.call.value(amount)(system.getItem("addr"), "inc")
    status, res = pcall(contract.call.value(1000000), system.getItem("addr"), "inc")
    if status == false then
        return first
    end
    return res
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
function send(addr, amount)
    contract.send(addr, amount)
    status, res = pcall(contract.call.value(1000000000)(system.getItem("addr"), "inc"))
    return status
end
function sql()
    contract.call(system.getItem("addr"), "init")
    pcall(contract.call, system.getItem("addr"), "pkins1")
    contract.call(system.getItem("addr"), "pkins2")
    return status
end

function sqlget()
    return contract.call(system.getItem("addr"), "pkget")
end

function getOrigin()
    return contract.call(system.getItem("addr"), "getOrigin")
end
abi.register(add, dadd, get, dget, send, sql, sqlget, getOrigin)
abi.payable(constructor,add)