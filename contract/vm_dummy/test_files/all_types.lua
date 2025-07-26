state.var {
  name = state.value(),
  list = state.array(),
  values = state.map()
}

-- value type

function value_set(new_name)
  name:set(new_name)
  contract.event("value_set", new_name)
end

function value_get()
  return name:get()
end

-- map type

function map_set(key, val)
  values[key] = val
  contract.event("map_set", key, val)
end

function map_get(key)
  return values[key]
end

-- array type

function array_append(val)
  list:append(val)
  contract.event("array_append", val)
end

function array_set(idx, val)
  list[idx] = val
  contract.event("array_set", idx, val)
end

function array_get(idx)
  return list[idx]
end

function array_length()
  return list:length()
end

-- write functions
abi.register(value_set, map_set, array_append, array_set)
-- read-only functions
abi.register_view(value_get, map_get, array_get, array_length)
