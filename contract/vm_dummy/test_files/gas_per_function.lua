state.var {
  values = state.map(),
  _value1 = state.value(),
  _value2 = state.value(),
}

arr = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 0 }

tbl = { name= "user2", year= 1981, age= 41,
 name1= "user2", year1= 1981, age1= 41,
 name2= "user2", year2= 1981, age2= 41,
 xxx= true}

local s1 = "hello world from Lua"
local s2 = "hello world from Lua Smart Contracts in Aergo"

local s100 = "100aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
local s200 = "200aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
local s400 = "400aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

long_str = [[Looks for the first match of pattern in the string s. If it finds a match, then find returns the indices of s where this occurrence starts and ends; otherwise, it returns nil. A third, optional numerical argument init specifies where to start the search; its default value is 1 and can be negative. A value of true as a fourth, optional argument plain turns off the pattern matching facilities, so the function does a plain "find substring" operation, with no characters in pattern being considered "magic". Note that if plain is given, then init must be given as well.
]]
long_str1 = [[Looks for the first match of pattern in the string s. If it finds a match, then find returns the indices of s where this occurrence starts and ends; otherwise, it returns nil. A third, optional numerical argument init specifies where to start the search; its default value is 1 and can be negative. A value of true as a fourth, optional argument plain turns off the pattern matching facilities, so the function does a plain "find substring" operation, with no characters in pattern being considered "magic". Note that if plain is given, then init must be given as werr.
]]

LLINE1 = "long line aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
LLINE2 = "long line aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

local nines = "999999999999999999999999999999"
local ninesn = "-999999999999999999999999999999"
local nines1 = "999999999"
local nines2 = 999999999
local one = 1
local b = 3.14e19
local pi = 3.14
local pi1 = 3.141592653589


local istc = function (b)
  return b and 1
end

local isfc = function (b)
  return b or 1
end

function f()
  return 4
end

function varg(...)
    return ...
end

function fixed_varg(a, b, c, ...)
    return a + b + c, ...
end

function swap(x, y)
  return y, x
end

function range_iter(n, i)
  if i >= n then
      return nil, nil
  end
  return i+1, i+1
end

function range(n)
  return range_iter, n, 0
end

function copy_arr(arr)
    local a = {}
    for i, v in ipairs(arr) do
        a[i] = v
    end
    return a
end

function error_handler(x)
  --print("an error ocurred", x)
  return "oh no!"
end


funcs = {

  -- Lua operators --

  ["comp_ops"] = function()
    A = 2
    D = 1 if A < D then
    end
    if A <= D then
    end
    if not (A < D) then
    end
    if not (A <= D) then
    end
    if A == D then
    end
    if not (A == D) then
    end
    V = "user2"
    if V == "user2" then
    end
    if V ~= "a very long long long line" then
    end
    if LLINE1 ~= V then
    end
    if LLINE1 ~= LLINE2 then
    end
    if LLINE1 == LLINE2 then
    end
    if A == 2 then
    end
    if A ~= 1 then
    end
    B = true
    if B == true then
    end
    if B ~= false then
    end
    if B == nil then
    end
    if B ~= nil then
    end
  end,

  ["unarytest_n_copy_ops"] = function()
    -- 16, 17
    NONE = nil
    B = false
    S = "user2"
    istc(true)
    isfc(true)
    if S and B then
    end
    if NONE or B then
    end
    if NONE or B then
    end
    B = true
    if B or NONE then
    end
  end,

  ["unary_ops"] = function()
    D = 3
    A = D
    B = not A
    N = -D
    ARR = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 0 }
    L = #ARR
    S = "a very long long long line"
    L = #S
  end,

  ["binary_ops"] = function()
    A = 0
    B = 2
    A = B + 1
    A = B - 1
    A = B * 1
    A = B / 1
    A = B % 1
    A = B + 1.2
    A = B - 1.2
    A = B * 1.2
    A = B / 1.2
    A = B % 1.2
    A = 1 + B
    A = 1 - B
    A = 1 * B
    A = 1 / B
    A = 1 % B
    A = 1.2 + B
    A = 1.2 - B
    A = 1.2 * B
    A = 1.2 / B
    A = 1.2 % B
    C = 1
    A = B + C
    A = B - C
    A = B * C
    A = B / C
    A = B % C
    A = B ^ C
    C = 1.2
    A = B + C
    A = B - C
    A = B * C
    A = B / C
    A = B % C
    A = B ^ C
    S = "user2" .. "1981" .. LLINE1
  end,

  ["constant_ops"] = function()
    A, B, C = nil, nil, nil
  end,

  ["upvalue_n_func_ops"] = function()
    local U = "user2"
    F = function() U = U .. "1981"; return U end
    F()
    F = function() U = "1981" end
    F()
    local D = 100
    F = function() D = 0 end
    F()
    local B = false
    F = function() B = true end
    F()
  end,

  ["table_ops"] = function()
    -- deprecated(op code not found): TGETR(59), TSETR(64)
    E = {}
    T = { name = "user2", age = 40, 1, 2, 3 }
    K = "name"
    NAME = T[K]
    AGE = T.age
    ONE = T[1]
    I = 1
    ONE = T[I]
    T[K] = "user2"
    T.age = 41
    T[4] = I
    return { 1, 2, 3, f() }
  end,

  ["call_n_vararg_ops"] = function()
    varg("user2", "1981", 41)
    T = 0
    ARR = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 0 }
    for i, v in ipairs(ARR) do
        T = T + v
    end
    ARR.name = "user2"
    for k, v in pairs(ARR) do
        if type(k) == 'number' then
            T = T + v
        end
    end
    local x, y, z = fixed_varg(1, 2, 3, varg("user2", "1981", 41))
    return fixed_varg(1, 2, 3, varg("user2", "1981", 41))
  end,

  ["return_ops"] = function()
    return swap(1, 2)
  end,

  ["loop_n_branche_ops"] = function()
    -- IFORL(80), IITERL(83)
    local T = 0
    for i = 1, 100 do
        T = T + i    
    end
    for n in range(100) do
        T = T + n
    end
    local i = 0
    while true do
        i = i + 1
      if i == 100 then
          break
      end
    end
  end,

  ["function_header_ops"] = function()
    -- FUNCF(89), IFUNCV(93), FUNCCW(96)
  end,


  -- Lua functions

  -- assert
  ["assert"] = function()
    assert(true, 'true')
    assert(1 == 1, 'true')
    assert(1 ~= 2, 'true')
    assert(long_str == long_str, 'true')
    assert(long_str ~= long_str1, 'true')
  end,
  -- getfenv
  ["getfenv"] = function()
    getfenv(run_test)
  end,
  -- metatable
  ["metatable"] = function()
    local x = {value = 5} 
    local mt = {
        __add = function (lhs, rhs) 
            return { value = lhs.value + rhs.value }
        end,
        __tostring = function (self)
            return "Hello Aergo"
        end
    }
    setmetatable(x, mt)
    getmetatable(x)
  end,
  -- ipairs
  ["ipairs"] = function()
    ipairs(arr)
  end,
  -- pairs
  ["pairs"] = function()
    pairs(tbl)
  end,
  -- next
  ["next"] = function()
    next(tbl)
    next(tbl, "year")
    next(tbl, "name")
  end,
  -- rawequal
  ["rawequal"] = function()
    rawequal(1, 1)
    rawequal(1, 2)
    rawequal(1.4, 2.1)
    rawequal(1.4, 2)
    rawequal("hello", "world")
    rawequal(arr, tbl)
    rawequal(arr, arr)
    rawequal(tbl, tbl)
  end,
  -- rawget
  ["rawget"] = function()
    local a = arr
    rawget(a, 1)
    rawget(a, 10)
    rawget(tbl, "age")
  end,
  -- rawset
  ["rawset"] = function()
    local a = copy_arr(arr)
    rawset(a, 1, 0)
    rawset(a, 10, 1)
    rawset(a, 11, -1)
    rawset(tbl, "addr", "aergo")
    rawset(tbl, "age", 18)
  end,
  -- select
  ["select"] = function()
    select('#', 'a', 'b', 'c', 'd')
    select('#', arr)
    select('2', 'a', 'b', 'c', 'd')
    select('2', arr)
    select('6', arr)
    select('9', arr)
  end,
  -- setfenv
  ["setfenv"] = function()
    fenv = getfenv(run_test)
    setfenv(run_test, fenv)
  end,
  -- tonumber
  ["tonumber"] = function()
    tonumber(0x10, 16)
    tonumber('112134', 16)
    tonumber('112134')
    tonumber(112134)
  end,
  -- tostring
  ["tostring"] = function()
    tostring('i am a string')
    tostring(1)
    tostring(112134)
    tostring(true)
    tostring(nil)
    tostring(3.14)
    tostring(3.14159267)
    tostring(x)
  end,
  -- type
  ["type"] = function()
    type('112134')
    type(112134)
    type({})
    type(type)
  end,
  -- unpack
  ["unpack"] = function()
    unpack({1, 2, 3, 4, 5}, 2, 4)
    unpack(arr, 2, 4)
    unpack({1, 2, 3, 4, 5})
    unpack(arr)
    local a = {}
    for i = 1, 100 do
        a[i] = i * i
    end
    unpack(a, 2, 4)
    unpack(a, 10, 40)
    unpack(a)
  end,
  -- pcall
  ["pcall"] = function()
    -- successfull calls
    local status, result = pcall(function() return 1 end)
    local status, result = pcall(swap, 1, 2)
    local status, result = pcall(copy_arr, arr)
    -- failed calls
    local status, err = pcall(function() error("error") end)
    local status, err = pcall(swap, 1)
    local status, err = pcall(copy_arr, 1)
  end,
  -- xpcall
  ["xpcall"] = function()
    -- successfull calls
    local status, result = xpcall(function() return 1 end, function() return "failed!" end)
    local status, result = xpcall(swap, error_handler, 1, 2)
    local status, result = xpcall(copy_arr, error_handler, arr)
    -- failed calls
    local status, err = xpcall(function() error("error") end, function() return "failed!" end)
    local status, err = xpcall(swap, error_handler, 1)
    local status, err = xpcall(copy_arr, error_handler, 1)
  end,

  -- string functions --

  -- string.byte
  ["string.byte"] = function()
    string.byte("hello string", 3, 7)
    string.byte(long_str, 3, 7)
    string.byte(long_str, 1, #long_str)
  end,
  -- string.char
  ["string.char"] = function()
    string.char(72, 101, 108, 108, 111)
    string.char(72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100)
    string.char(string.byte(long_str, 1, #long_str))
  end,
  -- string.dump
  ["string.dump"] = function()
    --string.dump(string.dump)
    string.dump(fixed_varg)
    string.dump(copy_arr)
  end,
  -- string.find
  ["string.find"] = function()
    string.find(long_str, "nume.....")
    string.find(long_str, "we..")
    string.find(long_str, "wi..")
    string.find(long_str, "pattern")
    string.find(long_str, "pattern", 200)
    string.find("hello world hellow", "hello")
    string.find("hello world hellow", "hello", 3)
    string.find(long_str, "head")
  end,
  -- string.format
  ["string.format"] = function()
    string.format("string.format %d, %.9f, %e, %g: %s", 1, 1.999999999, 10e9, 3.14, "end of string")
    string.format("string.format %d, %.9f, %e", 1, 1.999999999, 10e9)
  end,
  -- string.gmatch
  ["string.gmatch"] = function()
    s = "hello world from Lua"
    string.gmatch(s, "%a+")
    for w in string.gmatch(s, "%a+") do
    end
  end,
  -- string.gsub
  ["string.gsub"] = function()
    string.gsub(s1, "(%w+)", "%1 %1")
    local ss = s1 .. ' ' .. s1
    string.gsub(ss, "(%w+)", "%1 %1")
    string.gsub(ss, "(%w+)%s*(%w+)", "%2 %1")
  end,
  -- string.len
  ["string.len"] = function()
    string.len(s1)
    string.len(s2)
    string.len(long_str)
  end,
  -- string.lower
  ["string.lower"] = function()
    string.lower(s1)
    string.lower(s2)
    string.lower(long_str)
  end,
  -- string.match
  ["string.match"] = function()
    string.match(s1, "(L%w+)")
  end,
  -- string.rep
  ["string.rep"] = function()
    string.rep(s1, 2)
    string.rep(s1, 4)
    string.rep(s1, 8)
    string.rep(s1, 16)
    string.rep(long_str, 16)
  end,
  -- string.reverse
  ["string.reverse"] = function()
    string.reverse(s1)
    string.reverse(s2)
    string.reverse(long_str)
  end,
  -- string.sub
  ["string.sub"] = function()
    string.sub(s1, 10, 13)
    string.sub(s1, 10, -3)
    string.sub(long_str, 10, 13)
    string.sub(long_str, 10, -3)
  end,
  -- string.upper
  ["string.upper"] = function()
    string.upper(s1)
    string.upper(s2)
    string.upper(long_str)
  end,


  -- table functions --

  ["table.concat"] = function()
    local a = copy_arr(arr)
    local a100 = {}
    for i = 1, 100 do
        a100[i] = i * i
    end
    table.concat(a, ',')
    table.concat(a, ',', 3, 7)
    table.concat(a100, ',')
    table.concat(a100, ',', 3, 7)
    table.concat(a100, ',', 3, 32)
    local as10 = {}
    for i = 1, 10 do
        as10[i] = "hello"
    end
    local as100 = {}
    for i = 1, 100 do
        as100[i] = "hello"
    end
    table.concat(as10, ',')
    table.concat(as10, ',', 3, 7)
    table.concat(as100, ',')
    table.concat(as100, ',', 3, 7)
    table.concat(as100, ',', 3, 32)
    for i = 1, 10 do
        as10[i] = "h"
    end
    for i = 1, 100 do
        as100[i] = "h"
    end
    table.concat(as10, ',')
    table.concat(as10, ',', 3, 7)
    table.concat(as100, ',')
    table.concat(as100, ',', 3, 7)
    table.concat(as100, ',', 3, 32)
  end,
  ["table.insert"] = function()
    local a = copy_arr(arr)
    local a100 = {}
    for i = 1, 100 do
        a100[i] = i * i
    end
    for i = 1, 100 do
        table.insert(a, i)
    end
    table.remove(a)
    for i = 1, 100 do
        table.insert(a, 5, i)
    end
    table.remove(a, 5)
    for i = 1, 100 do
        table.insert(a100, i)
    end
    table.remove(a100)
    for i = 1, 100 do
        table.insert(a100, 5, i)
    end
    table.remove(a100, 5)
  end,
  ["table.remove"] = function()
    local a = copy_arr(arr)
    local a100 = {}
    for i = 1, 100 do
        a100[i] = i * i
    end
    table.insert(a, 11)
    for i = 1, 2 do
        table.remove(a)
    end
    table.insert(a, 5, 11)
    for i = 1, 10 do
        table.remove(a, 5)
    end
    --table.insert(a, 5, -5)
    --table.insert(a, 1, -5)
    table.insert(a100, 11)
    for i = 1, 2 do
        table.remove(a100)
    end
    table.insert(a100, 5, -5)
    for i = 1, 10 do
        table.remove(a100, 5)
    end
    --table.insert(a100, 1, -5)
  end,
  ["table.maxn"] = function()
    local a = copy_arr(arr)
    local a100 = {}
    for i = 1, 100 do
        a100[i] = i * i
    end
    local res1 = table.maxn(a)
    local res2 = table.maxn(a100)
  end,
  ["table.sort"] = function()
    local a = copy_arr(arr)
    local a100 = {}
    for i = 1, 100 do
        a100[i] = i * i
    end
    table.sort(a)
    table.sort(a, function(x,y) return x > y end)
    table.sort(a100)
    table.sort(a100, function(x,y) return x > y end)
  end,

  -- math functions --

  -- math.abs
  ["math.abs"] = function()
    math.abs(-1)
    math.abs(-1.4)
    math.abs(1)
    math.abs(1.4)
    math.abs(-1e9)
    math.abs(1e9)
    math.abs(-1e-9)
    math.abs(1e-9)
  end,
  -- math.ceil
  ["math.ceil"] = function()
    math.ceil(-1)
    math.ceil(-1.4)
    math.ceil(1)
    math.ceil(1.4)
    math.ceil(-1e9)
    math.ceil(1e9)
    math.ceil(-1e-9)
    math.ceil(1e-9)
  end,
  -- math.floor
  ["math.floor"] = function()
    math.floor(-1)
    math.floor(-1.4)
    math.floor(1)
    math.floor(1.4)
    math.floor(-1e9)
    math.floor(1e9)
    math.floor(-1e-9)
    math.floor(1e-9)
  end,
  -- math.max
  ["math.max"] = function()
    math.max(-1)
    math.max(-1.4)
    math.max(1)
    math.max(1.4)
    math.max(-1e9)
    math.max(1e9)
    math.max(-1e-9)
    math.max(1e-9)
    math.max(-1, 1)
    math.max(-1.4, 1.4)
    math.max(1, -1)
    math.max(1.4, -1.4)
    math.max(-1e9, 1e9)
    math.max(1e9, -1e9)
    math.max(-1e-9, 1e-9)
    math.max(1e-9, -1e-9)
    math.max(-1, 1, -1)
    math.max(-1.4, 1.4, -1.4)
    math.max(1, -1, 1)
    math.max(1.4, -1.4, 1.4)
    math.max(-1e9, 1e9, -1e9)
    math.max(1e9, -1e9, 1e9)
    math.max(-1e-9, 1e-9, -1e-9)
    math.max(1e-9, -1e-9, 1e-9)
  end,
  -- math.min
  ["math.min"] = function()
    math.min(-1)
    math.min(-1.4)
    math.min(1)
    math.min(1.4)
    math.min(-1e9)
    math.min(1e9)
    math.min(-1e-9)
    math.min(1e-9)
    math.min(-1, 1)
    math.min(-1.4, 1.4)
    math.min(1, -1)
    math.min(1.4, -1.4)
    math.min(-1e9, 1e9)
    math.min(1e9, -1e9)
    math.min(-1e-9, 1e-9)
    math.min(1e-9, -1e-9)
    math.min(-1, 1, -1)
    math.min(-1.4, 1.4, -1.4)
    math.min(1, -1, 1)
    math.min(1.4, -1.4, 1.4)
    math.min(-1e9, 1e9, -1e9)
    math.min(1e9, -1e9, 1e9)
    math.min(-1e-9, 1e-9, -1e-9)
    math.min(1e-9, -1e-9, 1e-9)
  end,
  -- math.pow
  ["math.pow"] = function()
    math.pow(-1, 2)
    math.pow(-1.4, 2)
    math.pow(1, 2)
    math.pow(1.4, 2)
    math.pow(-1e9, 2)
    math.pow(1e9, 2)
    math.pow(-1e-9, 2)
    math.pow(1e-9, 2)
    math.pow(-1, 4)
    math.pow(-1.4, 4)
    math.pow(1, 4)
    math.pow(1.4, 4)
    math.pow(-1e9, 4)
    math.pow(1e9, 4)
    math.pow(-1e-9, 4)
    math.pow(1e-9, 4)
    math.pow(-1, 8)
    math.pow(-1.4, 8)
    math.pow(1, 8)
    math.pow(1.4, 8)
    math.pow(-1e9, 8)
    math.pow(1e9, 8)
    math.pow(-1e-9, 8)
    math.pow(1e-9, 8)
  end,

  -- bit functions --

  -- bit.tobit
  ["bit.tobit"] = function()
    bit.tobit(0xffffffff)
    bit.tobit(0xffffffff + 1)
    bit.tobit(2^40 + 1234)
  end,
  -- bit.tohex
  ["bit.tohex"] = function()
    bit.tohex(1)
    bit.tohex(-1)
    bit.tohex(-1, -8)
    bit.tohex(0x87654321, 4)
  end,
  -- bit.bnot
  ["bit.bnot"] = function()
    bit.bnot(0)
    bit.bnot(0x12345678)
  end,
  -- bit.bor
  ["bit.bor"] = function()
    bit.bor(1)
    bit.bor(1, 2)
    bit.bor(1, 2, 4)
    bit.bor(1, 2, 4, 8)
  end,
  -- bit.band
  ["bit.band"] = function()
    bit.band(0x12345678, 0xff)
    bit.band(0x12345678, 0xff, 0x3f)
    bit.band(0x12345678, 0xff, 0x3f, 0x1f)
  end,
  -- bit.xor
  ["bit.xor"] = function()
    bit.bxor(0xa5a5f0f0, 0xaa55ff00)
    bit.bxor(0xa5a5f0f0, 0xaa55ff00, 0x18000000)
    bit.bxor(0xa5a5f0f0, 0xaa55ff00, 0x18000000, 0x00000033)
  end,
  -- bit.lshift
  ["bit.lshift"] = function()
    bit.lshift(1, 0)
    bit.lshift(1, 8)
    bit.lshift(1, 40)
  end,
  -- bit.rshift
  ["bit.rshift"] = function()
    bit.rshift(256, 0)
    bit.rshift(256, 8)
    bit.rshift(256, 40)
  end,
  -- bit.ashift
  ["bit.ashift"] = function()
    bit.arshift(0x87654321, 0)
    bit.arshift(0x87654321, 12)
    bit.arshift(0x87654321, 40)
  end,
  -- bit.rol
  ["bit.rol"] = function()
    bit.rol(0x12345678, 0)
    bit.rol(0x12345678, 12)
    bit.rol(0x12345678, 40)
  end,
  -- bit.ror
  ["bit.ror"] = function()
    bit.ror(0x12345678, 0)
    bit.ror(0x12345678, 12)
    bit.ror(0x12345678, 40)
  end,
  -- bit.bswap
  ["bit.bswap"] = function()
    bit.bswap(0x12345678)
  end,


  -- bignum --

  ["bignum.number"] = function()
    local nines_b = bignum.number(nines)
    local ninesn_b = bignum.number(ninesn)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local b_b = bignum.number(b)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
  end,
  ["bignum.isneg"] = function()
    local nines_b = bignum.number(nines)
    local ninesn_b = bignum.number(ninesn)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local b_b = bignum.number(b)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.isneg(nines_b)
    bignum.isneg(ninesn_b)
    bignum.isneg(nines1_b)
    bignum.isneg(nines2_b)
    bignum.isneg(one_b)
    bignum.isneg(b_b)
    bignum.isneg(pi_b)
    bignum.isneg(pi1_b)
  end,
  ["bignum.iszero"] = function()
    local nines_b = bignum.number(nines)
    local ninesn_b = bignum.number(ninesn)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local b_b = bignum.number(b)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.iszero(nines_b)
    bignum.iszero(ninesn_b)
    bignum.iszero(nines1_b)
    bignum.iszero(nines2_b)
    bignum.iszero(one_b)
    bignum.iszero(b_b)
    bignum.iszero(pi_b)
    bignum.iszero(pi1_b)
  end,
  ["bignum.tonumber"] = function()
    local nines_b = bignum.number(nines)
    local ninesn_b = bignum.number(ninesn)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local b_b = bignum.number(b)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.tonumber(nines_b)
    bignum.tonumber(ninesn_b)
    bignum.tonumber(nines1_b)
    bignum.tonumber(nines2_b)
    bignum.tonumber(one_b)
    bignum.tonumber(b_b)
    bignum.tonumber(pi_b)
    bignum.tonumber(pi1_b)
  end,
  ["bignum.tostring"] = function()
    local nines_b = bignum.number(nines)
    local ninesn_b = bignum.number(ninesn)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local b_b = bignum.number(b)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.tostring(nines_b)
    bignum.tostring(ninesn_b)
    bignum.tostring(nines1_b)
    bignum.tostring(nines2_b)
    bignum.tostring(one_b)
    bignum.tostring(b_b)
    bignum.tostring(pi_b)
    bignum.tostring(pi1_b)
  end,
  ["bignum.neg"] = function()
    local nines_b = bignum.number(nines)
    local ninesn_b = bignum.number(ninesn)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local b_b = bignum.number(b)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.neg(nines_b)
    bignum.neg(ninesn_b)
    bignum.neg(nines1_b)
    bignum.neg(nines2_b)
    bignum.neg(one_b)
    bignum.neg(b_b)
    bignum.neg(pi_b)
    bignum.neg(pi1_b)
  end,
  ["bignum.sqrt"] = function()
    local nines_b = bignum.number(nines)
    --local ninesn_b = bignum.number(ninesn)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local b_b = bignum.number(b)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.sqrt(nines_b)
    --bignum.sqrt(ninesn_b)
    bignum.sqrt(nines1_b)
    bignum.sqrt(nines2_b)
    bignum.sqrt(one_b)
    bignum.sqrt(b_b)
    bignum.sqrt(pi_b)
    bignum.sqrt(pi1_b)
  end,
  ["bignum.compare"] = function()
    local nines_b = bignum.number(nines)
    local ninesn_b = bignum.number(ninesn)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local b_b = bignum.number(b)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.compare(one_b, one_b)
    bignum.compare(nines_b, one_b)
    bignum.compare(nines_b, nines_b)
    bignum.compare(pi_b, pi_b)
    bignum.compare(pi1_b, pi1_b)
    bignum.compare(pi1_b, one_b)
    bignum.compare(pi_b, one_b)
  end,
  ["bignum.add"] = function()
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.add(one_b, one_b)
    bignum.add(nines_b, one_b)
    bignum.add(nines_b, nines_b)
    bignum.add(pi_b, pi_b)
    bignum.add(pi_b, pi1_b)
    bignum.add(pi_b, one_b)
    bignum.add(pi_b, nines_b)
    bignum.add(pi1_b, one_b)
  end,
  ["bignum.sub"] = function()
    local one_b = bignum.number(one)
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.sub(one_b, one_b)
    bignum.sub(bignum.number(-1), one_b)
    bignum.sub(nines_b, one_b)
    bignum.sub(nines_b, nines_b)
    bignum.sub(nines_b, pi_b)
    bignum.sub(pi_b, nines_b)
    bignum.sub(pi1_b, nines_b)
  end,
  ["bignum.mul"] = function()
    local one_b = bignum.number(one)
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    bignum.mul(one_b, one_b)
    bignum.mul(bignum.number(-1), one_b)
    bignum.mul(nines_b, one_b)
    bignum.mul(nines_b, nines1_b)
    bignum.mul(nines_b, nines_b)
    bignum.mul(pi_b, one_b)
    bignum.mul(pi_b, nines_b)
    bignum.mul(pi_b, pi_b)
    bignum.mul(pi1_b, pi1_b)
  end,
  ["bignum.div"] = function()
    local one_b = bignum.number(one)
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local pi1_b = bignum.number(pi1)
    local pi_b = bignum.number(pi)
    bignum.div(one_b, one_b)
    bignum.div(bignum.number(-1), one_b)
    bignum.div(nines_b, one_b)
    bignum.div(nines_b, nines_b)
    bignum.div(nines_b, bignum.number(333))
    bignum.div(pi1_b, bignum.number(1))
    bignum.div(pi_b, pi1_b)
  end,
  ["bignum.mod"] = function()
    local one_b = bignum.number(one)
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local pi1_b = bignum.number(pi1)
    local pi_b = bignum.number(pi)
    bignum.mod(one_b, one_b)
    bignum.mod(bignum.number(-1), one_b)
    bignum.mod(nines_b, one_b)
    bignum.mod(nines_b, nines1_b)
    bignum.mod(nines_b, bignum.number(333))
    bignum.mod(bignum.number(333), nines_b)
    bignum.mod(nines1_b, bignum.number(333))
    bignum.mod(bignum.number(333), nines1_b)
    bignum.mod(pi1_b, pi_b)
    bignum.mod(pi1_b, one_b)
  end,
  ["bignum.pow"] = function()
    local one_b = bignum.number(one)
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local pi1_b = bignum.number(pi1)
    local pi_b = bignum.number(pi)
    bignum.pow(one_b, one_b)
    bignum.pow(bignum.number(-1), one_b)
    bignum.pow(nines_b, one_b)
    --bignum.pow(nines_b, bignum.number(3))
    --bignum.pow(nines_b, bignum.number(9))
    bignum.pow(nines1_b, bignum.number(3))
    --bignum.pow(nines1_b, bignum.number(9))
    bignum.pow(pi1_b, one_b)
    bignum.pow(pi1_b, pi_b)
    bignum.pow(pi1_b, pi1_b)
  end,
  ["bignum.divmod"] = function()
    local one_b = bignum.number(one)
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local pi1_b = bignum.number(pi1)
    local pi_b = bignum.number(pi)
    bignum.divmod(one_b, one_b)
    bignum.divmod(bignum.number(-1), one_b)
    bignum.divmod(nines_b, one_b)
    bignum.divmod(nines_b, nines1_b)
    bignum.divmod(bignum.number(3), nines_b)
    bignum.divmod(bignum.number(9), nines_b)
    bignum.divmod(bignum.number(3), nines1_b)
    bignum.divmod(bignum.number(9), nines1_b)
    bignum.divmod(pi1_b, pi_b)
    bignum.divmod(pi1_b, pi1_b)
    bignum.divmod(pi1_b, one_b)
  end,
  ["bignum.powmod"] = function()
    local one_b = bignum.number(one)
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local pi1_b = bignum.number(pi1)
    local pi_b = bignum.number(pi)
    bignum.powmod(one_b, one_b, bignum.number(3))
    bignum.powmod(bignum.number(-1), one_b, bignum.number(3))
    bignum.powmod(nines_b, one_b, bignum.number(3))
    bignum.powmod(nines_b, bignum.number(3), bignum.number(4))
    bignum.powmod(nines_b, bignum.number(9), bignum.number(4))
    bignum.powmod(nines1_b, bignum.number(3), bignum.number(4))
    bignum.powmod(nines1_b, bignum.number(9), bignum.number(4))
    bignum.powmod(pi_b, bignum.number(9), bignum.number(4))
    bignum.powmod(pi1_b, bignum.number(9), bignum.number(4))
  end,
  ["bignum.operators"] = function()
    local one_b = bignum.number(one)
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local pi1_b = bignum.number(pi1)
    local pi_b = bignum.number(pi)
    -- unary
    local nines3_b = nines_b + nines1_b
    local nines4_b = nines_b - nines1_b
    local nines5_b = nines_b * nines1_b
    local nines6_b = nines_b / nines1_b
    local nines7_b = nines_b % nines1_b
    local nines8_b = nines_b ^ one_b
    -- comparison
    local result1 = nines_b < nines1_b
    local result2 = nines_b <= nines1_b
    local result3 = nines_b > nines1_b
    local result4 = nines_b >= nines1_b
    local result5 = nines_b == nines1_b
    local result6 = nines_b ~= nines1_b
  end,

  -- json module --

  ["json"] = function()
    json.decode(json.encode(1))
    json.decode(json.encode(100))
    json.decode(json.encode(10000000000))
    json.decode(json.encode(3.14))
    json.decode(json.encode(300000000.14159))
    json.decode(json.encode(s100))
    json.decode(json.encode(s200))
    json.decode(json.encode(s400))
    json.decode(json.encode(arr))
    json.decode(json.encode(long_arr100))
    json.decode(json.encode(long_arr1000))
    json.decode(json.encode(tbl))
  end,

  -- crypto module --

  ["crypto.sha256"] = function()
    crypto.sha256("0x616200e490aa")
    crypto.sha256(s100)
    crypto.sha256(s200)
    crypto.sha256(s400)
  end,
  ["crypto.ecverify"] = function()
    crypto.ecverify("11e96f2b58622a0ce815b81f94da04ae7a17ba17602feb1fd5afa4b9f2467960", "304402202e6d5664a87c2e29856bf8ff8b47caf44169a2a4a135edd459640be5b1b6ef8102200d8ea1f6f9ecdb7b520cdb3cc6816d773df47a1820d43adb4b74fb879fb27402", "AmPbWrQbtQrCaJqLWdMtfk2KiN83m2HFpBbQQSTxqqchVv58o82i")
  end,

  -- state module --

  ["state.set"] = function()
    values["key1"] = "value1"
    values["key2"] = "value2"
    _value1:set("hello")
    _value2:set("world")
  end,
  ["state.get"] = function()
    v1 = values["key1"]
    v2 = values["key2"]
    v3 = _value1:get()
    v4 = _value2:get()
    return v1 .. " " .. v2 .. " " .. v3 .. " " .. v4
  end,
  ["state.delete"] = function()
    values["key1"] = nil
    values:delete("key2")
    _value1:set(nil)
    _value2:set(nil)
  end,

  -- system module --

  ["system.getSender"] = function(...)
    return system.getSender()
  end,
  ["system.getBlockheight"] = function(...)
    return system.getBlockheight()
  end,
  ["system.getTxhash"] = function(...)
    return system.getTxhash()
  end,
  ["system.getTimestamp"] = function(...)
    return system.getTimestamp()
  end,
  ["system.getContractID"] = function(...)
    return system.getContractID()
  end,
  ["system.setItem"] = function(...)
    system.setItem("value", "this is a string")
    system.setItem("another_value", bignum.number(nines))
  end,
  ["system.getItem"] = function(...)
    return system.getItem("value") .. " | " .. bignum.tostring(system.getItem("another_value"))
  end,
  ["system.getAmount"] = function(...)
    return system.getAmount()
  end,
  ["system.getCreator"] = function(...)
    return system.getCreator()
  end,
  ["system.getOrigin"] = function(...)
    return system.getOrigin()
  end,
  ["system.getPrevBlockHash"] = function(...)
    return system.getPrevBlockHash()
  end,

  -- contract module --

  -- send
  ["contract.send"] = function()
    contract.send("AmPbWrQbtQrCaJqLWdMtfk2KiN83m2HFpBbQQSTxqqchVv58o82i", 10000000000)
    contract.send("Amh4S9pZgoJpxdCoMGg6SXEpAstTaTQNfQdZFsE26NpkqPwmaWod", "1 aergo 10 gaer")
    contract.send("Amg3EzRrAhyWmWUMHai5ZDZufbTXrorGutbwuHv6khQc8Cs2KWgA", bignum.number("999999999999999"))
  end,
  -- balance
  ["contract.balance"] = function()
    local b1 = contract.balance("AmPbWrQbtQrCaJqLWdMtfk2KiN83m2HFpBbQQSTxqqchVv58o82i")
    local b2 = contract.balance("Amh4S9pZgoJpxdCoMGg6SXEpAstTaTQNfQdZFsE26NpkqPwmaWod")
    local b3 = contract.balance("Amg3EzRrAhyWmWUMHai5ZDZufbTXrorGutbwuHv6khQc8Cs2KWgA")
  end,
  -- deploy
  ["contract.deploy"] = function()
    src = [[
      function hello(say)
          return "Hello " .. say .. " " .. system.getItem("name")
      end
      function emit(name, ...)
        contract.event(name, ...)
      end
      function constructor(creator)
          system.setItem("name", creator)
      end
      abi.register(hello, emit)
      abi.payable(constructor)
    ]]
    address1 = contract.deploy(src, "first")
    address2 = contract.deploy.value("1 aergo")(src, "second")
    address3 = contract.deploy.value(bignum.number("999999999999999"))(src, "third")
    system.setItem("address1", address1)
    system.setItem("address2", address2)
    system.setItem("address3", address3)
  end,
  -- call
  ["contract.call"] = function()
    address1 = system.getItem("address1")
    address2 = system.getItem("address2")
    address3 = system.getItem("address3")
    contract.call(address1, "hello", "world")
    contract.call(address2, "hello", "world")
    contract.call(address3, "hello", "world")
  end,
  -- pcall
  ["contract.pcall"] = function()
    address1 = system.getItem("address1")
    address2 = system.getItem("address2")
    address3 = system.getItem("address3")
    contract.pcall(contract.call, address1, "hello", "world")
    contract.pcall(contract.call, address2, "hello", "world")
    contract.pcall(contract.call, address3, "hello", "world")
  end,
  -- delegatecall
  ["contract.delegatecall"] = function()
    src = [[
      function update(key, value)
        system.setItem(key, value)
      end
      abi.register(update)
    ]]
    address = contract.deploy(src)
    contract.delegatecall(address, "update", "key1", "modified")
    assert(system.getItem("key1") == "modified")
  end,
  -- event
  ["contract.event"] = function()
    address1 = system.getItem("address1")
    address2 = system.getItem("address2")
    address3 = system.getItem("address3")
    contract.call(address1, "emit", "event1", "arg1", "arg2")
    contract.call(address2, "emit", "event2", "arg1", "arg2")
    contract.call(address3, "emit", "event3", "arg1", "arg2")
  end,

}

function run_test(function_name, ...)

  -- run the requested function
  local func = funcs[function_name]
  if func then
    return func(...)
  else
    error("unknown function: " .. function_name)
  end

end

function default()
  -- do nothing, only receive native aergo tokens
end

abi.register(run_test)
abi.payable(default)
