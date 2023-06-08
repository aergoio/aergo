function infiniteLoop()
    local t = 0
    while true do
        t = t + 1
    end
    return t
end
function infiniteCall()
    infiniteCall()
end
function catch()
    return pcall(infiniteLoop)
end
function contract_catch()
    return contract.pcall(infiniteLoop)
end
abi.register(infiniteLoop, infiniteCall, catch, contract_catch)