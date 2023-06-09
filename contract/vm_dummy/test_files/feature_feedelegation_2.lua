state.var {
    whitelist = state.map(),
}

function reg(k)
    if (k == nil) then
        whitelist[system.getSender()] = true
    else
        whitelist[k] = true
    end
end

function query()
    whitelist[system.getSender()] = false
    return 1, 2, 3, 4, 5
end

function default()
end

abi.register(reg, query)
abi.payable(default)
abi.fee_delegation(query)
