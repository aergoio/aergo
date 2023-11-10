
function call_myself()
    return contract.call(system.getContractID(), "test")
end

function test()
    return 123
end

abi.register(call_myself, test)
