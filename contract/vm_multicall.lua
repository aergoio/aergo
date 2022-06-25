-------------------------------------------------------------------
-- COMPOSABLE EXECUTION
-------------------------------------------------------------------

vars = {}

skip = {store=true,let=true,set=true,insert=true,assert=true,send=true}

action = {

  -- contract call
  call = function (...) return contract.call(...) end,
  ["call-send"] = function (amount,...) return contract.call.value(amount)(...) end,
  ["pcall"] = function (...) return {pcall(contract.call,...)} end,
  ["pcall-send"] = function (amount,...) return {pcall(contract.call.value(amount),...)} end,

  -- aergo balance and transfer
  balance = function (address) return contract.balance(address) end,
  send = function (address,amount) return contract.send(address, amount) end,

  -- variables
  let = function (x,y,z) if z then y = convert_bignum(y,z) end vars[x] = y end,
  store = function (n) vars[n] = vars['last_result'] end,

  -- tables
  get = function (o,k) return o[k] end,
  set = function (o,k,v) o[k] = v end,
  insert = function (...) table.insert(...) end,   -- inserts at the end if no pos informed
  remove = function (...) table.remove(...) end,   -- returns the removed item

  -- math
  add = function (x,y) return x+y end,
  sub = function (x,y) return x-y end,
  mul = function (x,y) return x*y end,
  div = function (x,y) return x/y end,
  pow = function (x,y) return x^y end,
  mod = function (x,y) return x%y end,
  sqrt = function (x) return bignum.sqrt(x) end,  -- use pow(0.5) for numbers

  -- strings
  format = function (...) return string.format(...) end, -- for concat: ['format','%s%s','%val1%','%val2%']
  substr = function (...) return string.sub(...) end,
  find = function (...) return string.match(...) end,
  replace = function (...) return string.gsub(...) end,

  -- conversions
  tobignum = function (x) return bignum.number(x) end,
  tonumber = function (x) return tonumber(x) end,
  tostring = function (x) return tostring(x) end,     -- bignum to string
  tojson   = function (x) return json.encode(x) end,
  fromjson = function (x) return json.decode(x) end,  -- create tables

  -- assertion
  assert = function (...) assert(eval(...),"assertion failed: " .. json.encode({...})) end,

}

function process_arg(arg)
  if type(arg) == 'string' then
    if #arg >= 3 and string.sub(arg, 1, 1) == '%' and string.sub(arg, -1, -1) == '%' then
      local varname = string.sub(arg, 2, -2)
      if vars[varname] ~= nil then
        arg = vars[varname]
      end
    end
  elseif type(arg) == 'table' then
    for k,v in pairs(arg) do arg[k] = process_arg(v) end
  end
  return arg
end

function execute(calls)

  local if_on = true
  local if_done = false

  local for_cmdpos
  local for_var, for_type
  local for_list, for_pos
  local for_last, for_increment

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
        vars['last_result'] = result
      end

    -- if elif else end
    elseif cmd == "if" then
      if_on = eval(unpack(args))
      if_done = if_on
    elseif cmd == "elif" then
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

    -- for foreach loop
    elseif cmd == "foreach" and if_on then
      for_cmdpos = cmdpos
      for_type = "each"
      for_var = args[1]
      for_list = args[2]
      for_pos = 1
      vars[for_var] = for_list[for_pos]
    elseif cmd == "for" and if_on then
      for_cmdpos = cmdpos
      for_type = "number"
      for_var = args[1]
      for_last = args[3]
      for_increment = args[4] or 1
      vars[for_var] = args[2]
    elseif cmd == "loop" and if_on then
      if for_type == "each" then
        for_pos = for_pos + 1
        if for_pos <= #for_list then
          vars[for_var] = for_list[for_pos]
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

    cmdpos = cmdpos + 1
  end

end

function eval(...)
  local args = {...}
  local v1 = args[1]
  local op = args[2]
  local v2 = args[3]
  local neg = false
  local matches = false
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

function convert_bignum(x, token)
  if type(x) ~= 'string' then
    x = tostring(x)
  end
  assert(string.match(x, '[^0-9.]') == nil, "the amount contains invalid character")
  local _, count = string.gsub(x, "%.", "")
  assert(count <= 1, "the amount is invalid")
  if count == 1 then
    local num_decimals
    if token:lower() == 'aergo' then
      num_decimals = 18
    else
      num_decimals = contract.call(token, "decimals")
    end
    assert(num_decimals >= 0 and num_decimals <= 18, "token with invalid decimals")
    local p1, p2 = string.match('0' .. x .. '0', '(%d+)%.(%d+)')
    local to_add = num_decimals - #p2
    if to_add > 0 then
      p2 = p2 .. string.rep('0', to_add)
    elseif to_add < 0 then
      p2 = string.sub(p2, 1, num_decimals)
    end
    x = p1 .. p2
    x = string.gsub(x, '0*', '', 1)  -- remove leading zeros
    if #x == 0 then x = '0' end
  end
  return bignum.number(x)
end

abi.register(execute)
