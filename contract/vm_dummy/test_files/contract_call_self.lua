-- simple case

function call_myself()
    return contract.call(system.getContractID(), "test")
end

function test()
    return 123
end

abi.register(call_myself, test)

-- recursive call with state variable

state.var {
    v = state.value()
}

function call_me_again(n, max)
    n = n + 1
    v:set(n)
    if n < max then
        contract.call(system.getContractID(), "call_me_again", n, max)
    end
    if n == 1 then
        return v:get()
    end
end

abi.register(call_me_again)
