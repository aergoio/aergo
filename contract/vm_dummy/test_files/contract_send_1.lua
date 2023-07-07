function constructor()
end

function send(addr)
    contract.send(addr, 1)
    contract.call.value(1)(addr)
end

abi.register(send, constructor)
abi.payable(constructor)
