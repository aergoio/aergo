state.var {
  next_contract = state.value(),
  call_info = state.map(),
  total_calls = state.value()
}

function set_next_contract(next_contract_id)
  next_contract:set(next_contract_id)
end

function call_me(call_depth, max_depth)
  local last_call = (total_calls:get() or 0) + 1
  total_calls:set(last_call)
  call_info[tostring(last_call)] = call_depth
  if call_depth == max_depth then
    return call_depth
  else
    return contract.call(next_contract:get(), "call_me", call_depth + 1, max_depth)
  end
end

function get_total_calls()
  return total_calls:get()
end

function get_call_info(key)
  return call_info[tostring(key)]
end

abi.register(set_next_contract, call_me)
abi.register_view(get_total_calls, get_call_info)
