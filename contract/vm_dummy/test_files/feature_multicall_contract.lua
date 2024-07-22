state.var {
  dict = state.map()
}

function set_value(key, value)
  dict[key] = value
end

function get_value(key)
  return dict[key]
end

abi.register(set_value)
abi.register_view(get_value)

function call(...)
  return contract.call(...)
end

function delegate_call(...)
  return contract.delegatecall(...)
end

function multicall(script)
  return contract.delegatecall("multicall", script)
end

function multicall_and_check(script)
  local result1, result2 = contract.delegatecall("multicall", script)
  assert(contract.balance() == "875000000000000000")
  assert(contract.balance("AmhXhR3Eguhu5qjVoqcg7aCFMpw1GGZJfqDDqfy6RsTP7MrpWeJ9") == "125000000000000000")
  return result1, result2
end

abi.register(call, delegate_call, multicall, multicall_and_check)

function recv_aergo()
  -- does nothing
end

function default()
  -- does nothing
end

function send_to(address, amount)
  contract.send(address, amount)
end

function get_balance()
  return contract.balance()
end

abi.payable(recv_aergo, default)
abi.register(send_to)
abi.register_view(get_balance)
