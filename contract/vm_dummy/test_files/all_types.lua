-- Minimal contract for tests/test-aergocli.sh (value/array/map + events).
state.var {
  name = state.value(),
  list = state.array(16),
  list_len = state.value(),
  values = state.map(),
}

function value_set(v)
  name:set(v)
  contract.event("value_set", v)
end

function array_append(v)
  local n = (list_len:get() or 0) + 1
  list[n] = v
  list_len:set(n)
  contract.event("array_append", v)
end

function map_set(k, val)
  values[tostring(k)] = val
  contract.event("map_set", k, val)
end

function value_get()
  return name:get()
end

function array_get(i)
  return list[i]
end

function map_get(k)
  return values[tostring(k)]
end

abi.register(value_set, array_append, map_set)
abi.register_view(value_get, array_get, map_get)
