
function constructor(init)
    system.setItem("count", init)
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

abi.register(inc, get, set)


function get_call_info(address, fname, info)

    local call_info = {
      sender = system.getSender(),
      origin = system.getOrigin(),
      ctr_id = system.getContractID()
    }

    if info == nil then info = {} end
    table.insert(info, call_info)

    return contract.call(address, fname, info)
end

abi.register(get_call_info)
