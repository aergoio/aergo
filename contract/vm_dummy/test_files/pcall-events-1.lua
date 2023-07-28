state.var {
  parent = state.value()
}

function constructor(address)
  parent:set(address)
end

function error_handler(err_msg)
  return "oh no! " .. err_msg
end

function test_pcall()
  contract.event("before pcall")
  local s, r = pcall(do_work, parent:get(), "pcall")
  contract.event("after pcall")
	assert(s == false, "call not failed")
	return r
end

function test_xpcall()
  contract.event("before xpcall")
  local s, r = xpcall(do_work, error_handler, parent:get(), "xpcall")
  contract.event("after xpcall")
	assert(s == false, "call not failed")
	return r
end

function test_contract_pcall()
  contract.event("before contract.pcall")
  local s, r = contract.pcall(do_work, parent:get(), "contract.pcall")
  contract.event("after contract.pcall")
	assert(s == false, "call not failed")
	return r
end

function do_work(contract_address, caller)
  contract.event("inside " .. caller .. " before")
  local r = contract.call(contract_address, "call_me")
  contract.event("inside " .. caller .. " after")
	return r
end

abi.register(test_pcall, test_xpcall, test_contract_pcall)
