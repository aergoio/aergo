package contract

import (
	luacUtil "github.com/aergoio/aergo/v2/cmd/aergoluac/util"
)

var multicall_compiled luacUtil.LuaCode

const multicall_code = `
vars = {}

skip = {store=true,let=true,set=true,insert=true,assert=true,send=true}

action = {

  -- contract call
  call = function (...) return contract.call(...) end,
  ["call + send"] = function (amount,...) return contract.call.value(amount)(...) end,
  ["try call"] = function (...) return try_call(contract.call,...) end,
  ["try call + send"] = function (amount,...) return try_call(contract.call.value(amount),...) end,

  -- aergo balance and transfer
  ["get balance"] = function (address) return bignum.number(contract.balance(address)) end,
  send = function (address,amount) return contract.send(address, amount) end,

  -- variables
  let = function (x,y,z) if z then y = convert_bignum(y,z) end vars[x] = y end,
  ["store result as"] = function (n) vars[n] = vars['last result'] end,

  -- tables
  get = function (o,k) return o[k] end,
  set = function (o,k,v) o[k] = v end,
  insert = function (...) table.insert(...) end,   -- inserts at the end if no pos informed
  remove = function (...) return table.remove(...) end,   -- returns the removed item
  ["get size"] = function (x) return #x end,
  ["get keys"] = function (obj)
      local list = {}
      for key,_ in pairs(obj) do
        list[#list + 1] = key
      end
      table.sort(list)  -- for a deterministic output
      return list
    end,

  -- math
  add = function (x,y) return x+y end,
  ["subtract"] = function (x,y) return x-y end,
  ["multiply"] = function (x,y) return x*y end,
  ["divide"] = function (x,y) return x/y end,
  ["remainder"] = function (x,y) return x%y end,
  ["exponentiate"] = function (x,y) return x^y end,
  ["square root"] = function (x)
    if bignum.isbignum(x) then
      return bignum.sqrt(x)
    else
      return x ^ 0.5
    end
  end,

  -- strings
  ["combine"] = function (...) return table.concat({...}) end,
  format = function (...) return string.format(...) end, -- for concat: ['format','%s%s','%val1%','%val2%']
  ["extract"] = function (...) return string.sub(...) end,
  find = function (...) return string.match(...) end,
  replace = function (...) return string.gsub(...) end,

  -- conversions
  ["to big number"] = function (x) if x==nil then x=vars['last result'] end; return bignum.number(x) end,
  ["to number"] = function (x) if x==nil then x=vars['last result'] end; return tonumber(x) end,
  ["to string"] = function (x) if x==nil then x=vars['last result'] end; return tostring(x) end,
  ["to json"]   = function (x) if x==nil then x=vars['last result'] end; return json.encode(x) end,
  ["from json"] = function (x) if x==nil then x=vars['last result'] end; return json.decode(x) end,

  -- assertion
  assert = function (...) assert(eval(...),"assertion failed: " .. json.encode({...})) end,

}

function try_call(fn, ...)
  local success, result = pcall(fn, ...)
  vars['call succeeded'] = success
  return result
end

function process_arg(arg)
  if type(arg) == 'string' then
    if #arg >= 3 and string.sub(arg, 1, 1) == '%' and string.sub(arg, -1, -1) == '%' then
      local varname = string.sub(arg, 2, -2)
      arg = get_arg_value(varname, arg)
    end
  elseif type(arg) == 'table' then
    for k,v in pairs(arg) do arg[k] = process_arg(v) end
  end
  return arg
end

function get_arg_value(varname, default)
  local value = default
  if varname == "my account address" then
    value = system.getContractID()
  elseif varname == "my aergo balance" then
    value = action["get balance"]()
  elseif vars[varname] ~= nil then
    value = vars[varname]
  end
  return value
end

function execute(calls)

  local if_on = true
  local if_done = false

  local for_cmdpos
  local for_var, for_var2, for_type
  local for_obj, for_list, for_pos
  local for_last, for_increment
  local skip_for = false

  local cmdpos = 1
  while cmdpos <= #calls do
    local call = calls[cmdpos]
    local args = {}  -- use a copy of the list because of loops

    -- copy values and process variables
    for i,arg in ipairs(call) do
      if i == 1 then
        args[i] = arg
      else
        args[i] = process_arg(arg)
      end
    end

    -- process the command
    local cmd = table.remove(args, 1)
    local fn = action[cmd]
    if fn and if_on then
      local result = fn(unpack(args))
      if not skip[cmd] then
        vars['last result'] = result
      end

    -- if elif else end
    elseif cmd == "if" then
      if_on = eval(unpack(args))
      if_done = if_on
    elseif cmd == "else if" then
      if if_on then
        if_on = false
      elseif not if_done then
        if_on = eval(unpack(args))
        if_done = if_on
      end
    elseif cmd == "else" then
      if_on = (not if_on) and (not if_done)
    elseif cmd == "end" then
      if_on = true

    -- for foreach break loop
    elseif cmd == "for each" and if_on then
      -- "for each", "item", "in", "list"
      if args[2] == "in" then
        for_var2 = "__"
        for_obj = {}
        for_list = args[3]
      -- "for each", "key", "value", "in", "object"
      elseif args[3] == "in" then
        for_var2 = args[2]
        for_obj = args[4]
        for_list = action["get keys"](for_obj)
        vars[for_var2] = for_obj[for_list[1]]
      else
        assert(false, "for each: invalid syntax")
      end
      for_cmdpos = cmdpos
      for_type = "each"
      for_var = args[1]
      for_pos = 1
      vars[for_var] = for_list[1]
      skip_for = (for_list[1] == nil)  -- if the list is empty or it is a dictionary
    elseif cmd == "for" and if_on then
      for_cmdpos = cmdpos
      for_type = "number"
      for_var = args[1]
      for_last = args[3]
      for_increment = args[4] or 1
      vars[for_var] = args[2]
      skip_for = ((for_increment > 0 and vars[for_var] > for_last) or (for_increment < 0 and vars[for_var] < for_last))
    elseif cmd == "break" and if_on then
      if table.remove(args, 1) == "if" then
        skip_for = eval(unpack(args))
      else
        skip_for = true
      end
    elseif cmd == "loop" and if_on then
      if for_type == "each" then
        for_pos = for_pos + 1
        if for_pos <= #for_list then
          vars[for_var] = for_list[for_pos]
          vars[for_var2] = for_obj[for_list[for_pos]]
          cmdpos = for_cmdpos
        end
      elseif for_type == "number" then
        vars[for_var] = vars[for_var] + for_increment
        if (for_increment > 0 and vars[for_var] > for_last) or (for_increment < 0 and vars[for_var] < for_last) then
          -- quit loop (continue to the next command)
        else
          cmdpos = for_cmdpos
        end
      else
        cmdpos = 0
      end

    -- return
    elseif cmd == "return" and if_on then
      return unpack(args)  -- or the array itself
    elseif if_on then
      assert(false, "command not found: " .. cmd)
    end

    if skip_for then
      repeat
        cmdpos = cmdpos + 1
        call = calls[cmdpos]
      until call == nil or call[1] == "loop"
      skip_for = false
    end

    cmdpos = cmdpos + 1
  end

end

function eval(...)
  local args = {...}
  if #args == 1 and type(args[1]) == 'boolean' then
    return args[1]
  end
  local v1 = args[1]
  local op = args[2]
  local v2 = args[3]
  local neg = false
  local matches = false
  -- if v1 is a bignum and v2 is a string, convert v2 to a bignum
  if bignum.isbignum(v1) and type(v2) == 'string' then
    v2 = to_bignum(v2)
  end
  if string.sub(op,1,1) == "!" then
    neg = true
    op = string.sub(op, 2)
  end
  if v1 == nil and op ~= "=" then
    -- does not match
  elseif op == "=" then
    matches = v1 == v2
  elseif op == ">" then
    matches = v1 > v2
  elseif op == ">=" then
    matches = v1 >= v2
  elseif op == "<" then
    matches = v1 < v2
  elseif op == "<=" then
    matches = v1 <= v2
  elseif op == "match" then
    matches = string.match(v1, v2) ~= nil
  else
    assert(false, "operator not known: " .. op)
  end
  if neg then matches = not matches end
  if #args > 3 then
    op = args[4]
    local matches2 = eval(unpack(args, 5, #args))
    if op == "and" then
      return (matches and matches2)
    elseif op == "or" then
      return (matches or matches2)
    else
      assert(false, "operator not known: " .. op)
    end
  end
  return matches
end

-- convert a number to a big number with the correct number of decimals
function convert_bignum(x, token)
  -- convert the number to a string
  if type(x) ~= 'string' then
    x = tostring(x)
  end

  -- check if the string contains invalid characters
  assert(string.match(x, '[^0-9.]') == nil, "the amount contains invalid character")

  -- count the number of dots
  local _, count = string.gsub(x, "%.", "")
  assert(count <= 1, "the amount is invalid")

  -- get the number of decimals from the token
  local num_decimals
  if token:lower() == 'aergo' then
    num_decimals = 18
  else
    num_decimals = contract.call(token, "decimals")
  end
  assert(num_decimals >= 0 and num_decimals <= 18, "token with invalid decimals")

  -- split the amount into integer and fractional parts
  local p1, p2
  if count == 1 then
    p1, p2 = string.match('0' .. x .. '0', '(%d+)%.(%d+)')
  else
    p1 = x
    p2 = ""
  end
  -- compute the number of zeros to add or remove
  local to_add = num_decimals - #p2
  if to_add > 0 then
    p2 = p2 .. string.rep('0', to_add)
  elseif to_add < 0 then
    p2 = string.sub(p2, 1, num_decimals)
  end
  -- join the two parts
  x = p1 .. p2
  -- remove leading zeros
  x = string.gsub(x, '0*', '', 1)
  -- if the result is empty, return 0
  if #x == 0 then x = '0' end
  -- convert to big number
  return bignum.number(x)
end

function to_bignum(x)
  -- if there is a space, separate the amount and the symbol
  local amount, symbol = string.match(x, '([^%s]+)%s+(.+)')
  if amount and symbol then
    return convert_bignum(amount, symbol)
  else
    return bignum.number(x)
  end
end

abi.register(execute)
`
