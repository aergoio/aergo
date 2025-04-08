import "test-name-service.lua"
import "https://raw.githubusercontent.com/aergoio/aergo-contract-ex/refs/heads/master/contracts/typecheck/typecheck.lua"

function test2()
  return "test2"
end

abi.register(test2)
