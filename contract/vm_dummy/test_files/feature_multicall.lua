state.var {
  name = state.value(),
  last = state.value(),
  dict = state.map()
}

function get_dict()
  return {one = 1, two = 2, three = 3}
end

function get_list()
  return {'first', 'second', 'third', 123, 12.5, true}
end

function get_table()
  return { name = "Test", second = 'Te"st', number = 123, bool = true, array = {11,22,33} }
end

function works()
  return 123
end

function fails()
  assert(false, "this call should fail")
end

function hello(name)
  return 'hello ' .. name
end

function set_name(val)
  name:set(val)
  assert(type(val)=='string', "must be string")
end

function get_name()
  return name:get()
end

function set(key, value)
  dict[key] = value
end

function inc(key)
  dict[key] = (dict[key] or 0) + 1
  contract.event("new_value", dict[key])
end

function add(value)
  local key = (last:get() or 0) + 1
  dict[tostring(key)] = value
  last:set(key)
end

function get(key)
  return dict[key]
end

function sort(list)
  table.sort(list)
  return list
end

abi.register(add, set, inc, set_name)
abi.register_view(get_dict, get_list, get_table, works, fails, get, get_name, sort, hello)

function call(...)
  return contract.call(...)
end

function is_contract(address)
  return system.isContract(address)
end

function sender()
  return system.getSender()
end

function origin()
  return system.getOrigin()
end

abi.register(call)
abi.register_view(is_contract, sender, origin)

function recv_aergo()
  -- does nothing
end

abi.payable(recv_aergo)
