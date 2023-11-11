state.var {
    v = state.value()
}

function assert2(condition, message)
    if not condition then
        error(message or "Assertion failed!")
    end
end

function override_functions()

    -- override the assert function
    assert = function(condition, message)
        v:set("overridden")
        return "overridden"
    end

    -- override system.getSender and system.getOrigin
    system.getSender = function() return "overridden" end
    system.getOrigin = function() return "overridden" end

    -- override contract.balance
    contract.balance = function() return "123" end

    -- override the __add metamethod on bignum module
    getmetatable(bignum.number(0)).__add = function(x,y) return x-y end

end

function test_vm_isolation(address)

    override_functions()

    -- check overriden assert()
    v:set("")
    local success, ret = pcall(test_assert)
    assert2(success, "override failed 1")
    assert2(v:get() == "overridden", "override failed 2")

    -- test it on another call to this contract - it should fail
    v:set("")
    local success, ret = pcall(contract.call, system.getContractID(), "test_assert")
    assert2(success == false, "override worked on another instance 1")
    assert2(v:get() ~= "overridden", "override worked on another instance 2")

    -- test it on another contract - it should fail
    v:set("")
    local success, ret = pcall(contract.call, address, "test_assert")
    assert2(success == false, "override worked on another contract 1")
    assert2(ret ~= "overridden", "override worked on another contract 2")


    -- check overriden functions
    assert2(test_sender() == "overridden", "system.getSender() override failed")
    assert2(test_origin() == "overridden", "system.getOrigin() override failed")
    assert2(test_balance() == "123", "contract.balance() override failed")
    assert2(test_bignum() == bignum.number(3), "metamethod override failed")

    -- test them on another call to this instance
    ret = contract.call(system.getContractID(), "test_sender")
    assert2(ret ~= "overridden", "override worked on another instance 3")
    ret = contract.call(system.getContractID(), "test_origin")
    assert2(ret ~= "overridden", "override worked on another instance 4")
    ret = contract.call(system.getContractID(), "test_balance")
    assert2(ret ~= "123", "override worked on another instance 5")
    ret = contract.call(system.getContractID(), "test_bignum")
    assert2(ret ~= "7", "override worked on another instance 6")

    -- test them on another contract
    ret = contract.call(address, "test_sender")
    assert2(ret ~= "overridden", "override worked on another contract 3")
    ret = contract.call(address, "test_origin")
    assert2(ret ~= "overridden", "override worked on another contract 4")
    ret = contract.call(address, "test_balance")
    assert2(ret ~= "123", "override worked on another contract 5")
    ret = contract.call(address, "test_bignum")
    assert2(ret ~= "7", "override worked on another contract 6")

end

function test_assert()
    assert(1 == 0, "original assert")
end

function test_sender()
    return system.getSender()
end

function test_origin()
    return system.getOrigin()
end

function test_balance()
    return contract.balance()
end

function test_bignum()
    return bignum.number(5) + bignum.number(2)
end

abi.register(test_vm_isolation, test_assert, test_sender, test_origin, test_balance, test_bignum)
