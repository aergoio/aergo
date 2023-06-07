function constructor()
end
function send(addr)
    contract.send(addr,1)
    contract.call.value(1)(addr)
end
function send2(addr)
    contract.call.value(1)(addr)
    contract.call.value(3)(addr)
end

abi.register(send, send2, constructor)
abi.payable(constructor)