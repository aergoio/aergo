loop_cnt = 1000
arr = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 0 }
tbl = { name= "kslee", year= 1981, age= 41 }
long_str = [[Looks for the first match of pattern in the string s. If it finds a match, then find returns the indices of s where this occurrence starts and ends; otherwise, it returns nil. A third, optional numerical argument init specifies where to start the search; its default value is 1 and can be negative. A value of true as a fourth, optional argument plain turns off the pattern matching facilities, so the function does a plain "find substring" operation, with no characters in pattern being considered "magic". Note that if plain is given, then init must be given as well.
]]
long_str1 = [[Looks for the first match of pattern in the string s. If it finds a match, then find returns the indices of s where this occurrence starts and ends; otherwise, it returns nil. A third, optional numerical argument init specifies where to start the search; its default value is 1 and can be negative. A value of true as a fourth, optional argument plain turns off the pattern matching facilities, so the function does a plain "find substring" operation, with no characters in pattern being considered "magic". Note that if plain is given, then init must be given as werr.
]]

local print = function(...)
    system.print(...)
end

function copy_arr(arr)
    local a = {}
    for i, v in ipairs(arr) do
        a[i] = v
    end
    return a
end

function measure(fn, ...)
    local start = nsec()
    fn(...)
    return nsec() - start
end

function m1k(fn, ...)
    local r1k = function(fn)
        return function(...)
            for i = 1, loop_cnt do
                fn(...)
            end
        end
    end
    t = measure(r1k(fn), ...)
    print(string.format('%.9f', t / loop_cnt))
end

function print_chain(fn)
    return function(...)
        print(fn(...))
    end
end

function m1kp(fn, ...)
    m1k(print_chain(fn), ...)
end

function basic_fns()
    print("assert")
    m1k(assert, true, 'true')
    m1k(assert, 1 == 1, 'true')
    m1k(assert, 1 ~= 2, 'true')
    m1k(assert, long_str == long_str, 'true')
    m1k(assert, long_str ~= long_str1, 'true')
    print("getfenv")
    m1k(getfenv, basic_fns)
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
    print("getmetable")
    m1k(getmetatable, x)
    print("ipairs")
    m1k(ipairs, arr)
    print("next")
    m1k(next, tbl)
    m1k(next, tbl, "year")
    m1k(next, tbl, "name")
    print("pairs")
    m1k(pairs, tbl)
    print("rawequal")
    m1k(rawequal, 1, 1)
    m1k(rawequal, 1, 2)
    m1k(rawequal, 1.4, 2.1)
    m1k(rawequal, 1.4, 2)
    m1k(rawequal, "hello", "world")
    m1k(rawequal, arr, tbl)
    m1k(rawequal, arr, arr)
    m1k(rawequal, tbl, tbl)
    print("rawget")
    local a = arr
    m1k(rawget, a, 1)
    m1k(rawget, a, 10)
    m1k(rawget, tbl, "age")
    print("rawset")
    local a = copy_arr(arr)
    m1k(rawset, a, 1, 0)
    m1k(rawset, a, 10, 1)
    m1k(rawset, a, 11, -1)
    m1k(rawset, tbl, "addr", "aergo")
    m1k(rawset, tbl, "age", 42)
    print("select")
    m1k(select, '#', 'a', 'b', 'c', 'd')
    m1k(select, '#', arr)
    m1k(select, '2', 'a', 'b', 'c', 'd')
    m1k(select, '2', arr)
    m1k(select, '6', arr)
    m1k(select, '9', arr)
    print("setfenv")
    fenv = getfenv(basic_fns)
    m1k(setfenv, basic_fns, fenv)
    print("setmetatable")
    m1k(setmetatable, x, mt)
    print("tonumber")
    m1k(tonumber, 0x10, 16)
    m1k(tonumber, '112134', 16)
    m1k(tonumber, '112134')
    m1k(tonumber, 112134)
    print("tostring")
    m1k(tostring, 'i am a string')
    m1k(tostring, 1)
    m1k(tostring, 112134)
    m1k(tostring, true)
    m1k(tostring, nil)
    m1k(tostring, 3.14)
    m1k(tostring, 3.14159267)
    m1k(tostring, x)
    print("type")
    m1k(type, '112134')
    m1k(type, 112134)
    m1k(type, {})
    m1k(type, type)
    print("unpack")
    m1k(unpack, {1, 2, 3, 4, 5}, 2, 4)
    m1k(unpack, arr, 2, 4)
    m1k(unpack, {1, 2, 3, 4, 5})
    m1k(unpack, arr)
    local a = {}
    for i = 1, 100 do
        a[i] = i * i
    end
    m1k(unpack, a, 2, 4)
    m1k(unpack, a, 10, 40)
    m1k(unpack, a)
end

function string_fns()
    print("string.byte")
    m1k(string.byte, "hello string", 3, 7)
    m1k(string.byte, long_str, 3, 7)
    m1k(string.byte, long_str, 1, #long_str)
    print("string.char")
    m1k(string.char, 72, 101, 108, 108, 111)
    m1k(string.char, 72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100)
    m1k(string.char, string.byte(long_str, 1, #long_str))
    print("string.dump")
    m1k(string.dump, m1k)
    m1k(string.dump, basic_fns)
    print("string.find")
    m1k(string.find, long_str, "nume.....")
    m1k(string.find, long_str, "we..")
    m1k(string.find, long_str, "wi..")
    m1k(string.find, long_str, "pattern")
    m1k(string.find, long_str, "pattern", 200)
    m1k(string.find, "hello world hellow", "hello")
    m1k(string.find, "hello world hellow", "hello", 3)
    m1k(string.find, long_str, "head")
    print("string.format")
    m1k(string.format, "string.format %d, %.9f, %e, %g: %s", 1, 1.999999999, 10e9, 3.14, "end of string")
    m1k(string.format, "string.format %d, %.9f, %e", 1, 1.999999999, 10e9)
    print("stirng.gmatch")
    s = "hello world from Lua"
    m1k(string.gmatch, s, "%a+")
    m1k(function() 
        for w in string.gmatch(s, "%a+") do
        end
    end)
    print("string.gsub")
    m1k(string.gsub, s, "(%w+)", "%1 %1")
    s2 = s .. ' ' .. s
    m1k(string.gsub, s2, "(%w+)", "%1 %1")
    m1k(string.gsub, s, "(%w+)%s*(%w+)", "%2 %1")
    print("string.len")
    m1k(string.len, s)
    m1k(string.len, s2)
    m1k(string.len, long_str)
    print("string.lower")
    m1k(string.lower, s)
    m1k(string.lower, long_str)
    print("string.match")
    m1k(string.match, s, "(L%w+)")
    print("string.rep")
    m1k(string.rep, s, 2)
    m1k(string.rep, s, 4)
    m1k(string.rep, s, 8)
    m1k(string.rep, s, 16)
    m1k(string.rep, long_str, 16)
    print("string.reverse")
    m1k(string.reverse, s)
    m1k(string.reverse, s2)
    m1k(string.reverse, long_str)
    print("string.sub")
    m1k(string.sub, s, 10, 13)
    m1k(string.sub, s, 10, -3)
    m1k(string.sub, long_str, 10, 13)
    m1k(string.sub, long_str, 10, -3)
    print("string.upper")
    m1k(string.upper, s)
    m1k(string.upper, s2)
    m1k(string.upper, long_str)
end

function table_fns1()
    local a = copy_arr(arr)
    local a100 = {}
    for i = 1, 100 do
        a100[i] = i * i
    end
    print("table.concat")
    m1k(table.concat, a, ',')
    m1k(table.concat, a, ',', 3, 7)
    m1k(table.concat, a100, ',')
    m1k(table.concat, a100, ',', 3, 7)
    m1k(table.concat, a100, ',', 3, 32) 
    local as10 = {}
    for i = 1, 10 do
        as10[i] = "hello"
    end
    local as100 = {}
    for i = 1, 100 do
        as100[i] = "hello"
    end
    m1k(table.concat, as10, ',')
    m1k(table.concat, as10, ',', 3, 7)
    m1k(table.concat, as100, ',')
    m1k(table.concat, as100, ',', 3, 7)
    m1k(table.concat, as100, ',', 3, 32) 
    for i = 1, 10 do
        as10[i] = "h"
    end
    for i = 1, 100 do
        as100[i] = "h"
    end
    m1k(table.concat, as10, ',')
    m1k(table.concat, as10, ',', 3, 7)
    m1k(table.concat, as100, ',')
    m1k(table.concat, as100, ',', 3, 7)
    m1k(table.concat, as100, ',', 3, 32) 
end

function table_fns2()
    local a = copy_arr(arr)
    local a100 = {}
    for i = 1, 100 do
        a100[i] = i * i
    end
    print("table.insert")
    m1k(table.insert, a, 11)
    for i = 1, 1000 do
        table.remove(a)
    end
    m1k(table.insert, a, 5, 11)
    for i = 1, 1000 do
        table.remove(a, 5)
    end
    --m1k(table.insert, a, 5, -5)
    --m1k(table.insert, a, 1, -5)
    m1k(table.insert, a100, 11)
    for i = 1, 1000 do
        table.remove(a100)
    end
    m1k(table.insert, a100, 5, -5)
    for i = 1, 1000 do
        table.remove(a100, 5)
    end
    --m1k(table.insert, a100, 1, -5)
end

function table_fns3()
    local a = copy_arr(arr)
    local a100 = {}
    for i = 1, 100 do
        a100[i] = i * i
    end
    print("table.maxn")
    m1k(table.maxn, a)
    m1k(table.maxn, a100)
    print("table.remove")
    for i = 1, 1000 do
        table.insert(a, i)
    end
    m1k(table.remove, a)
    for i = 1, 1000 do
        table.insert(a, 5, i)
    end
    m1k(table.remove, a, 5)
    for i = 1, 1000 do
        table.insert(a100, i)
    end
    m1k(table.remove, a100)
    for i = 1, 1000 do
        table.insert(a100, 5, i)
    end
    m1k(table.remove, a100, 5)
end

function table_fns4()
    local a = copy_arr(arr)
    local a100 = {}
    for i = 1, 100 do
        a100[i] = i * i
    end
    print("table.sort")
    m1k(table.sort, a)
    m1k(table.sort, a, function(x,y) return x > y end)
    m1k(table.sort, a100)
    m1k(table.sort, a100, function(x,y) return x > y end)
end

function math_fns()
    d = {}
    for i = 1, loop_cnt do
        d[i] = -500 + i
    end
    for i = 1, loop_cnt, 10 do
        d[i] = d[i] + 0.5
    end
    for i = 1, loop_cnt, 13 do
        d[i] = d[i] + 0.3
    end
    for i = 1, loop_cnt, 17 do
        d[i] = d[i] + 0.7
    end
    f = {}
    for i = 1, loop_cnt do
        f[i] = -1 + i * 0.002
    end
    local md = function (fn, d, ...)
        local x = function (fn, d)
            return function(...)
                for i, v in ipairs(d) do
                    fn(v, ...)
                end
            end
        end
        t = measure(x(fn, d), ...)
        print(string.format('%.9f', t / #d))
    end
    print("math.abs")
    md(math.abs, d)
    print("math.acos")
    md(math.acos, f)
    print("math.asin")
    md(math.asin, f)
    print("math.atan")
    md(math.atan, f)
    print("math.atan2")
    md(math.atan2, f, 2)
    print("math.ceil")
    md(math.ceil, f)
    print("math.cos")
    md(math.cos, d)
    md(math.cos, f)
    print("math.cosh")
    md(math.cosh, d)
    md(math.cosh, f)
    print("math.deg")
    md(math.deg, f)
    print("math.exp")
    md(math.exp, d)
    md(math.exp, f)
    print("math.floor")
    md(math.floor, d)
    md(math.floor, f)
    print("math.fmod")
    md(math.fmod, f, 1.4)
    print("math.frexp")
    md(math.frexp, d)
    md(math.frexp, f)
    print("math.log")
    local filter = function(l, cond)
        r = {}
        for i, v in ipairs(l) do
            if cond(v) then
                table.insert(r, v) 
            end
        end
        return r
    end
    ud = filter(d, function(v) return v >= 0 end)
    uf = filter(f, function(v) return v >= 0 end)
    md(math.log, ud)
    md(math.log, uf)
    print("math.log10")
    md(math.log10, ud)
    md(math.log10, uf)
    print("math.max")
    m1k(math.max, unpack(ud))
    m1k(math.max, unpack(d))
    m1k(math.max, unpack(uf))
    m1k(math.max, unpack(f))
    print("math.min")
    m1k(math.min, unpack(ud))
    m1k(math.min, unpack(d))
    m1k(math.min, unpack(uf))
    m1k(math.min, unpack(f))
    print("math.modf")
    md(math.modf, d)
    md(math.modf, f)
    print("math.pow")
    md(math.pow, d, 2)
    md(math.pow, d, 4)
    md(math.pow, d, 8)
    md(math.pow, f, 2)
    md(math.pow, f, 4)
    md(math.pow, f, 8)
    md(math.pow, ud, 8.4)
    md(math.pow, uf, 8.4)
    print("math.sin")
    md(math.sin, d)
    md(math.sin, f)
    print("math.sinh")
    md(math.sinh, d)
    md(math.sinh, f)
    print("math.sqrt")
    md(math.sqrt, ud)
    md(math.sqrt, uf)
    print("math.tan")
    md(math.tan, d)
    md(math.tan, f)
    print("math.tanh")
    md(math.tanh, f)
end

function bit_fns()
    local printx = function (fn)
        return function(...)
            print("0x"..bit.tohex(fn(...)))
        end
    end
    print("bit.tobit")
    m1k(bit.tobit, 0xffffffff)
    m1k(bit.tobit, 0xffffffff + 1)
    m1k(bit.tobit, 2^40 + 1234)
    print("bit.tohex")
    m1k(bit.tohex, 1)
    m1k(bit.tohex, -1)
    m1k(bit.tohex, -1, -8)
    m1k(bit.tohex, 0x87654321, 4)
    print("bit.bnot")
    m1k(bit.bnot, 0)
    m1k(bit.bnot, 0x12345678)
    print("bit.bor")
    m1k(bit.bor, 1)
    m1k(bit.bor, 1, 2)
    m1k(bit.bor, 1, 2, 4)
    m1k(bit.bor, 1, 2, 4, 8)
    print("bit.band")
    m1k(bit.band, 0x12345678, 0xff)
    m1k(bit.band, 0x12345678, 0xff, 0x3f)
    m1k(bit.band, 0x12345678, 0xff, 0x3f, 0x1f)
    print("bit.xor")
    m1k(bit.bxor, 0xa5a5f0f0, 0xaa55ff00)
    m1k(bit.bxor, 0xa5a5f0f0, 0xaa55ff00, 0x18000000)
    m1k(bit.bxor, 0xa5a5f0f0, 0xaa55ff00, 0x18000000, 0x00000033)
    print("bit.lshift")
    m1k(bit.lshift, 1, 0)
    m1k(bit.lshift, 1, 8)
    m1k(bit.lshift, 1, 40)
    print("bit.rshift")
    m1k(bit.rshift, 256, 0)
    m1k(bit.rshift, 256, 8)
    m1k(bit.rshift, 256, 40)
    print("bit.ashift")
    m1k(bit.arshift, 0x87654321, 0)
    m1k(bit.arshift, 0x87654321, 12)
    m1k(bit.arshift, 0x87654321, 40)
    print("bit.rol")
    m1k(bit.rol, 0x12345678, 0)
    m1k(bit.rol, 0x12345678, 12)
    m1k(bit.rol, 0x12345678, 40)
    print("bit.ror")
    m1k(bit.ror, 0x12345678, 0)
    m1k(bit.ror, 0x12345678, 12)
    m1k(bit.ror, 0x12345678, 40)
    print("bit.bswap")
    m1k(bit.bswap, 0x12345678)
end

abi.register(basic_fns,string_fns,table_fns1,table_fns2,table_fns3,table_fns4,math_fns,bit_fns)
