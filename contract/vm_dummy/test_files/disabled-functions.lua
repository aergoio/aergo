
function check_disabled_functions()

  -- check the disabled modules
  assert(os == nil, "os is available")
  assert(io == nil, "io is available")
  assert(debug == nil, "debug is available")
  assert(jit == nil, "jit is available")
  assert(ffi == nil, "ffi is available")
  assert(coroutine == nil, "coroutine is available")
  assert(package == nil, "package is available")

  -- check the disabled functions
  assert(collectgarbage == nil, "collectgarbage is available")
  assert(gcinfo == nil, "gcinfo is available")
  assert(module == nil, "module is available")
  assert(require == nil, "require is available")
  assert(dofile == nil, "dofile is available")
  assert(load == nil, "load is available")
  assert(loadlib == nil, "loadlib is available")
  assert(loadfile == nil, "loadfile is available")
  assert(loadstring == nil, "loadstring is available")
  assert(print == nil, "print is available")
  assert(getmetatable == nil, "getmetatable is available")
  assert(setmetatable == nil, "setmetatable is available")
  assert(string.dump == nil, "string.dump is available")

  local success, result = pcall(function() newproxy() end)
  assert(success == false and result:match(".* 'newproxy' not supported"), "newproxy is available")
  local success, result = pcall(function() setfenv() end)
  assert(success == false and result:match(".* 'setfenv' not supported"), "setfenv is available")
  local success, result = pcall(function() getfenv() end)
  assert(success == false and result:match(".* 'getfenv' not supported"), "getfenv is available")

  -- make sure the tostring does not return a pointer
  local tab = {}
  assert(not tostring(type):match("0x[%x]+"), "tostring returns a pointer for function")
  assert(not tostring(system):match("0x[%x]+"), "tostring returns a pointer for internal table")
  assert(not tostring(tab):match("0x[%x]+"), "tostring returns a pointer for table")

end

abi.register(check_disabled_functions)
