function infinite_loop()
    local t = 0
    while true do
        t = t + 1
    end
    return t
end

function infinite_call()
    infinite_call()
end

function catch_loop()
    return pcall(infinite_loop)
end

function catch_call()
    return pcall(infinite_call)
end

function contract_catch_loop()
    return contract.pcall(infinite_loop)
end

function contract_catch_call()
    return contract.pcall(infinite_call)
end

abi.register(infinite_loop, catch_loop, contract_catch_loop)
abi.register(infinite_call, catch_call, contract_catch_call)
