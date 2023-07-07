state.var {
  call_info = state.map(),
  total_calls = state.value()
}

function call_me(contracts, call_depth, max_depth)
  local last_call = (total_calls:get() or 0) + 1
  total_calls:set(last_call)
  call_info[tostring(last_call)] = call_depth
  if call_depth == max_depth then
    return call_depth
  else
    local next_contract = contracts[call_depth % #contracts + 1]
    return contract.call(next_contract, "call_me", contracts, call_depth + 1, max_depth)
  end
end

function get_total_calls()
  return total_calls:get()
end

function get_call_info(key)
  return call_info[tostring(key)]
end

abi.register(call_me)
abi.register_view(get_total_calls, get_call_info)
