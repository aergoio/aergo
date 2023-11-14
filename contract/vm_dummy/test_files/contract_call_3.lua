
function get_call_info(info)

    local call_info = {
      sender = system.getSender(),
      origin = system.getOrigin(),
      ctr_id = system.getContractID()
    }

    if info == nil then info = {} end
    table.insert(info, call_info)

    return info
end

abi.register(get_call_info)
