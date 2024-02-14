
function test1()
  local gas_before = bignum.number(contract.gasLeft())
  local a = 0
  for i = 1,10 do
    a = a + i
  end
  local gas_after = bignum.number(contract.gasLeft())
  local used_gas = gas_before - gas_after
  return gas_before, gas_after, used_gas
end

function test2(...)
  local gas_before = bignum.number(contract.gasLeft())
  contract.call(...)
  local gas_after = bignum.number(contract.gasLeft())
  local used_gas = gas_before - gas_after
  return gas_before, gas_after, used_gas
end

function test3(address)
  local g1 = contract.gasLeft()
  local g2, g3 = contract.call(address, "test")
  local g4 = contract.gasLeft()
  return g1, g2, g3, g4
end

function callback()
  return contract.gasLeft()
end

abi.register(test1, test2, test3, callback)

--- this is used in another contract

function test()
  local g1 = contract.gasLeft()
  local g2 = contract.call(system.getSender(), "callback")
  return g1, g2
end

abi.register(test)
