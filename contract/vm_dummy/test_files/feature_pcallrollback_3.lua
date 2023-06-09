function pass(addr)
    contract.send(addr, 1)
end

function add(addr, a, b)
    system.setItem("arg", a)
    contract.pcall(pass, addr)
    return a + b
end

function set(addr)
    contract.send(addr, 1)
    system.setItem("arg", 2)
    status, ret = contract.pcall(add, addr, 1, 2)
end

function set2(addr)
    contract.send(addr, 1)
    system.setItem("arg", 2)
    status, ret = contract.pcall(add, addar, 1)
end

function get()
    return system.getItem("arg")
end

function getBalance()
    return contract.balance()
end

abi.register(set, set2, get, getBalance)
abi.payable(set, set2)
