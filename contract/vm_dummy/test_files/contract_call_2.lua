
function constructor(addr)
    system.setItem("count", 99)
    system.setItem("addr", addr)
end

function inc()
    count = system.getItem("count")
    system.setItem("count", count + 1)
    return count
end

function get()
    return system.getItem("count")
end

function set(val)
    system.setItem("count", val)
end

function cinc(amount)
    return contract.call.value(amount)(system.getItem("addr"), "inc")
end

function dinc()
    return contract.delegatecall(system.getItem("addr"), "inc")
end

function cget()
    return contract.call(system.getItem("addr"), "get")
end

function dget()
    return contract.delegatecall(system.getItem("addr"), "get")
end

function cset(val)
    contract.call(system.getItem("addr"), "set", val)
end

function dset(val)
    contract.delegatecall(system.getItem("addr"), "set", val)
end

abi.register(inc, cinc, dinc, get, cget, dget, set, cset, dset)


function get_call_info(info)
		info = get_call_info2(info)
    return contract.delegatecall(system.getItem("addr"), "get_call_info", system.getContractID(), "get_call_info2", info)
end

function get_call_info2(info)

    local call_info = {
      sender = system.getSender(),
      origin = system.getOrigin(),
      ctr_id = system.getContractID()
    }

    if info == nil then info = {} end
    table.insert(info, call_info)

    return info
end

abi.register(get_call_info, get_call_info2)
