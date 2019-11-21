loop_cnt = 1000
arr = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 0 }
tbl = { name= "kslee", year= 1981, age= 41 }
long_str = [[Looks for the first match of pattern in the string s. If it finds a match, then find returns the indices of s where this occurrence starts and ends; otherwise, it returns nil. A third, optional numerical argument init specifies where to start the search; its default value is 1 and can be negative. A value of true as a fourth, optional argument plain turns off the pattern matching facilities, so the function does a plain "find substring" operation, with no characters in pattern being considered "magic". Note that if plain is given, then init must be given as well.
]]
long_str1 = [[Looks for the first match of pattern in the string s. If it finds a match, then find returns the indices of s where this occurrence starts and ends; otherwise, it returns nil. A third, optional numerical argument init specifies where to start the search; its default value is 1 and can be negative. A value of true as a fourth, optional argument plain turns off the pattern matching facilities, so the function does a plain "find substring" operation, with no characters in pattern being considered "magic". Note that if plain is given, then init must be given as werr.
]]

function copy_arr(arr)
    local a = {}
    for i, v in ipairs(arr) do
        a[i] = v
    end
    return a
end

function m1k(fn, ...)
    for i = 1, loop_cnt do
        fn(...)
    end
end


function basic_fns()
    m1k(assert, true, 'true')
    m1k(assert, 1 == 1, 'true')
    m1k(assert, 1 ~= 2, 'true')
    m1k(assert, long_str == long_str, 'true')
    m1k(assert, long_str ~= long_str1, 'true')
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
    m1k(getmetatable, x)
    m1k(ipairs, arr)
    m1k(next, tbl)
    m1k(next, tbl, "year")
    m1k(next, tbl, "name")
    m1k(pairs, tbl)
    m1k(rawequal, 1, 1)
    m1k(rawequal, 1, 2)
    m1k(rawequal, 1.4, 2.1)
    m1k(rawequal, 1.4, 2)
    m1k(rawequal, "hello", "world")
    m1k(rawequal, arr, tbl)
    m1k(rawequal, arr, arr)
    m1k(rawequal, tbl, tbl)
    local a = arr
    m1k(rawget, a, 1)
    m1k(rawget, a, 10)
    m1k(rawget, tbl, "age")
    local a = copy_arr(arr)
    m1k(rawset, a, 1, 0)
    m1k(rawset, a, 10, 1)
    m1k(rawset, a, 11, -1)
    m1k(rawset, tbl, "addr", "aergo")
    m1k(rawset, tbl, "age", 42)
    m1k(select, '#', 'a', 'b', 'c', 'd')
    m1k(select, '#', arr)
    m1k(select, '2', 'a', 'b', 'c', 'd')
    m1k(select, '2', arr)
    m1k(select, '6', arr)
    m1k(select, '9', arr)
    fenv = getfenv(basic_fns)
    m1k(setfenv, basic_fns, fenv)
    m1k(setmetatable, x, mt)
    m1k(tonumber, 0x10, 16)
    m1k(tonumber, '112134', 16)
    m1k(tonumber, '112134')
    m1k(tonumber, 112134)
    m1k(tostring, 'i am a string')
    m1k(tostring, 1)
    m1k(tostring, 112134)
    m1k(tostring, true)
    m1k(tostring, nil)
    m1k(tostring, 3.14)
    m1k(tostring, 3.14159267)
    m1k(tostring, x)
    m1k(type, '112134')
    m1k(type, 112134)
    m1k(type, {})
    m1k(type, type)
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
    m1k(string.byte, "hello string", 3, 7)
    m1k(string.byte, long_str, 3, 7)
    m1k(string.byte, long_str, 1, #long_str)
    m1k(string.char, 72, 101, 108, 108, 111)
    m1k(string.char, 72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100)
    m1k(string.char, string.byte(long_str, 1, #long_str))
    m1k(string.dump, m1k)
    m1k(string.dump, basic_fns)
    m1k(string.find, long_str, "nume.....")
    m1k(string.find, long_str, "we..")
    m1k(string.find, long_str, "wi..")
    m1k(string.find, long_str, "pattern")
    m1k(string.find, long_str, "pattern", 200)
    m1k(string.find, "hello world hellow", "hello")
    m1k(string.find, "hello world hellow", "hello", 3)
    m1k(string.find, long_str, "head")
    m1k(string.format, "string.format %d, %.9f, %e, %g: %s", 1, 1.999999999, 10e9, 3.14, "end of string")
    m1k(string.format, "string.format %d, %.9f, %e", 1, 1.999999999, 10e9)
    s = "hello world from Lua"
    m1k(string.gmatch, s, "%a+")
    m1k(function() 
        for w in string.gmatch(s, "%a+") do
        end
    end)
    m1k(string.gsub, s, "(%w+)", "%1 %1")
    s2 = s .. ' ' .. s
    m1k(string.gsub, s2, "(%w+)", "%1 %1")
    m1k(string.gsub, s, "(%w+)%s*(%w+)", "%2 %1")
    m1k(string.len, s)
    m1k(string.len, s2)
    m1k(string.len, long_str)
    m1k(string.lower, s)
    m1k(string.lower, long_str)
    m1k(string.match, s, "(L%w+)")
    m1k(string.rep, s, 2)
    m1k(string.rep, s, 4)
    m1k(string.rep, s, 8)
    m1k(string.rep, s, 16)
    m1k(string.rep, long_str, 16)
    m1k(string.reverse, s)
    m1k(string.reverse, s2)
    m1k(string.reverse, long_str)
    m1k(string.sub, s, 10, 13)
    m1k(string.sub, s, 10, -3)
    m1k(string.sub, long_str, 10, 13)
    m1k(string.sub, long_str, 10, -3)
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
    m1k(table.maxn, a)
    m1k(table.maxn, a100)
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
        for i, v in ipairs(d) do
            fn(v, ...)
        end
    end
    md(math.abs, d)
    md(math.ceil, d)
    md(math.ceil, f)
    md(math.floor, d)
    md(math.floor, f)
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
    m1k(math.max, unpack(ud))
    m1k(math.max, unpack(d))
    m1k(math.max, unpack(uf))
    m1k(math.max, unpack(f))
    m1k(math.min, unpack(ud))
    m1k(math.min, unpack(d))
    m1k(math.min, unpack(uf))
    m1k(math.min, unpack(f))
    md(math.pow, d, 2)
    md(math.pow, d, 4)
    md(math.pow, d, 8)
    md(math.pow, f, 2)
    md(math.pow, f, 4)
    md(math.pow, f, 8)
    md(math.pow, ud, 8.4)
    md(math.pow, uf, 8.4)
end

function bit_fns()
    m1k(bit.tobit, 0xffffffff)
    m1k(bit.tobit, 0xffffffff + 1)
    m1k(bit.tobit, 2^40 + 1234)
    m1k(bit.tohex, 1)
    m1k(bit.tohex, -1)
    m1k(bit.tohex, -1, -8)
    m1k(bit.tohex, 0x87654321, 4)
    m1k(bit.bnot, 0)
    m1k(bit.bnot, 0x12345678)
    m1k(bit.bor, 1)
    m1k(bit.bor, 1, 2)
    m1k(bit.bor, 1, 2, 4)
    m1k(bit.bor, 1, 2, 4, 8)
    m1k(bit.band, 0x12345678, 0xff)
    m1k(bit.band, 0x12345678, 0xff, 0x3f)
    m1k(bit.band, 0x12345678, 0xff, 0x3f, 0x1f)
    m1k(bit.bxor, 0xa5a5f0f0, 0xaa55ff00)
    m1k(bit.bxor, 0xa5a5f0f0, 0xaa55ff00, 0x18000000)
    m1k(bit.bxor, 0xa5a5f0f0, 0xaa55ff00, 0x18000000, 0x00000033)
    m1k(bit.lshift, 1, 0)
    m1k(bit.lshift, 1, 8)
    m1k(bit.lshift, 1, 40)
    m1k(bit.rshift, 256, 0)
    m1k(bit.rshift, 256, 8)
    m1k(bit.rshift, 256, 40)
    m1k(bit.arshift, 0x87654321, 0)
    m1k(bit.arshift, 0x87654321, 12)
    m1k(bit.arshift, 0x87654321, 40)
    m1k(bit.rol, 0x12345678, 0)
    m1k(bit.rol, 0x12345678, 12)
    m1k(bit.rol, 0x12345678, 40)
    m1k(bit.ror, 0x12345678, 0)
    m1k(bit.ror, 0x12345678, 12)
    m1k(bit.ror, 0x12345678, 40)
    m1k(bit.bswap, 0x12345678)
end

function main()
    basic_fns()
    string_fns()
    table_fns1()
    table_fns2()
    table_fns3()
    table_fns4()
    math_fns()
    bit_fns()
end

abi.register(main)
