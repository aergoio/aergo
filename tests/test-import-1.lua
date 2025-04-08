import "test-import-2.lua"
import "test-name-service.lua"

function test()
  return "test"
end

abi.register(test)
