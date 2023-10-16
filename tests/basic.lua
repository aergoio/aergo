state.var {
  values = state.map(),
}

function set_value(key, value)
  assert(system.getSender() == system.getCreator(), "permission denied")
  values[tostring(key)] = value
end

function get_value(key)
  return values[tostring(key)]
end

abi.register(set_value)
abi.register_view(get_value)
