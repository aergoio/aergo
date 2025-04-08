
function resolve(name)
  return name_service.resolve(name)
end

abi.register(resolve)

-- from http://lua-users.org/wiki/LuaTypeChecking
--
-- Type check function that decorates functions.
-- supported type: string, number, function, boolean, nil, userdata, address, bignum
-- Example:
--   sum = typecheck('number', 'number', '->', 'number')(
--     function(x, y) return x + y end
--   )

function typecheck(...)
  local types = {...}
  
  local function check(x, f)
    if (x and f == 'address') then
      assert(type(x) == 'string', "address must be string type") 
      -- check address length
      assert(52 == #x, string.format("invalid address length: %s (%s)", x, #x))
      -- check character
      local invalidChar = string.match(x, '[^123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]')
      assert(nil == invalidChar,
        string.format("invalid address format: %s contains invalid char %s", x, invalidChar or 'nil'))
    elseif (x and f == 'bignum') then
      -- check bignum
      assert(bignum.isbignum(x), string.format("invalid format: %s != %s", type(x), f))
    else
      -- check default lua types
      assert(type(x) == f, string.format("invalid format: %s != %s", type(x), f or 'nil'))
    end
  end
  
  return function(f)
    local function returncheck(i, ...)
      -- Check types of return values.
      if types[i] == "->" then i = i + 1 end
      local j = i
      while types[i] ~= nil do
        check(select(i - j + 1, ...), types[i])
        i = i + 1
      end
      return ...
    end
    return function(...)
      -- Check types of input parameters.
      local i = 1
      while types[i] ~= nil and types[i] ~= "->" do
        check(select(i, ...), types[i])
        i = i + 1
      end
      return returncheck(i, f(...))  -- call function
    end
  end
end


function test2()
  return "test2"
end

abi.register(test2)


function test()
  return "test"
end

abi.register(test)
