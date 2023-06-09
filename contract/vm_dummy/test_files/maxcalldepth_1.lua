state.var {
    call_info = state.map(),
    last_call = state.value(),
    total_calls = state.value()
}

function call_me(call_depth, max_depth)
    call_info[tostring(call_depth)] = call_depth
    last_call:set(call_depth)
    total_calls:set((total_calls:get() or 0) + 1)
    if call_depth == max_depth then
        return call_depth
    else
        return contract.call(system.getContractID(), "call_me", call_depth + 1, max_depth)
    end
end

function get_total_calls()
    return last_call:get(), total_calls:get()
end

function get_call_info(key)
    return call_info[key]
end

function check_state()
    assert(last_call:get() == 64, "last_call")
    assert(total_calls:get() == 64, "total_calls")
    for i = 1, 64 do
        assert(call_info[tostring(i)] == i, "call_info[" .. tostring(i) .. "] = " .. tostring(call_info[tostring(i)]))
    end
    return true
end

abi.register(call_me)
abi.register_view(get_total_calls, get_call_info, check_state)
