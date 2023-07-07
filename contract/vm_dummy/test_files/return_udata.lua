function test_die()
    return contract.call(system.getContractID(), "return_object")
end

function return_object()
    return db.query("select 1")
end

abi.register(test_die, return_object)
