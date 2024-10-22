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
    end

    -- override system.getSender and system.getOrigin
    system.getSender = function() return "overridden" end
    system.getOrigin = function() return "overridden" end

    -- override contract.balance
    contract.balance = function() return "123" end

    if getmetatable ~= nil then
      -- override the __add metamethod on bignum module
      getmetatable(bignum.number(0)).__add = function(x,y) return x-y end
    end

end

function check_local_overridden_functions()

    v:set("")
    local success, ret = pcall(test_assert)
    assert2(success, "assert override failed 1")
    assert2(v:get() == "overridden", "assert override failed 2")

    assert2(test_sender() == "overridden", "system.getSender() override failed")
    assert2(test_origin() == "overridden", "system.getOrigin() override failed")
    assert2(test_balance() == "123", "contract.balance() override failed")
    if getmetatable ~= nil then
      assert2(test_bignum() == bignum.number(3), "metamethod override failed")
    end

end

function test_vm_isolation_forward(address)

    override_functions()

    check_local_overridden_functions()


    -- test assert on another call to this contract - it should fail
    v:set("")
    local success, ret = pcall(contract.call, system.getContractID(), "test_assert")
    assert2(success == false, "override worked on another instance 1")
    assert2(v:get() ~= "overridden", "override worked on another instance 2")

    -- test assert on another contract - it should fail
    local success, ret = pcall(contract.call, address, "test_assert")
    assert2(success == false, "override worked on another contract 1")
    ret = contract.call(address, "get")
    assert2(ret ~= "overridden", "override worked on another contract 2")

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

function get()
    return v:get()
end

abi.register(test_vm_isolation_forward, test_assert, test_sender, test_origin, test_balance, test_bignum)
abi.register_view(get)

--[[
The above is used when contract A overrides the functions and test on called contract B.
Below is the opposite: contract A calls B, B overrides the functions and return, then
A checks if the functions were overridden on its contract
]]

function test_vm_isolation_reverse(address)

    -- contract A calls B (it can be the same contract)
    local ret = contract.call(address, "override_and_return")
    assert2(ret == "overridden", "override did not work on contract B")

    -- check if the functions on this contract were overridden
    v:set("")
    local success, ret = pcall(test_assert)
    assert2(success == false, "assert reverse override worked 1")
    assert2(v:get() ~= "overridden", "assert reverse override worked 2")

    assert2(test_sender() ~= "overridden", "system.getSender() reverse override worked")
    assert2(test_origin() ~= "overridden", "system.getOrigin() reverse override worked")
    assert2(test_balance() ~= "123", "contract.balance() reverse override worked")
    assert2(test_bignum() ~= "7", "bignum.number() reverse override worked")

end

function override_and_return()
    override_functions()
    check_local_overridden_functions()
    return v:get()
end

abi.register(test_vm_isolation_reverse, override_and_return)
