-- state dbs
state.var {
    _t = state.value(), -- 의미없는 값
}

function init()
    _t:set(0)
end

function justLoop(count)
    local t = 0
    while t < count do
        t = t + 1
        _t:set(t)
    end
    return t
end

function t()
    return _t:get()
end

function catch(count)
    return pcall(justLoop, count)
end

function contract_catch(count)
    return contract.pcall(justLoop, count)
end

abi.register(init, justLoop, catch, contract_catch)
abi.register_view(t)
