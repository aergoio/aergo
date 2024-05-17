state.var {
  parent = state.value()
}

function constructor(address)
  parent:set(address)
end

function call_me()
  contract.event("contract-2 before call")
  -- call contract-3
  local result = contract.call(parent:get(), "call_me")
  contract.event("contract-2 after call")
  -- raises an error:
  assert(1 == 0)
  contract.event("contract-2 returning")
end

abi.register(call_me)
