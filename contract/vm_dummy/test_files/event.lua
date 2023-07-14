function test_ev()
    contract.event("ev1", 1, "local", 2, "form")
    contract.event("ev1", 3, "local", 4, "form")
end

abi.register(test_ev)
