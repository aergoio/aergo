state.var{
    whitelist = state.map(),
}

function reg(k)
    if (k == nil) then
        whitelist[system.getSender()] = true
    else
        whitelist[k] = true
    end
end

function query(a)
    if (system.isFeeDelegation() == true) then
        whitelist[system.getSender()] = false
    end
    return 1,2,3,4,5
end
function check_delegation(fname,k)
    if (fname == "query") then
        return whitelist[system.getSender()]
    end
    return false
end
function default()
end
abi.register(reg, query)
abi.payable(default)
abi.fee_delegation(query)