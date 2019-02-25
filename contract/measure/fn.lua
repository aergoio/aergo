loop_cnt = 1000
arr = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 0 }
tbl = { name= "kslee", year= 1981, age= 41 }
long_str = [[Looks for the first match of pattern in the string s. If it finds a match, then find returns the indices of s where this occurrence starts and ends; otherwise, it returns nil. A third, optional numerical argument init specifies where to start the search; its default value is 1 and can be negative. A value of true as a fourth, optional argument plain turns off the pattern matching facilities, so the function does a plain "find substring" operation, with no characters in pattern being considered "magic". Note that if plain is given, then init must be given as well.
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
    m1k(tostring, 1)
    m1k(tostring, true)
    m1k(tostring, nil)
    m1k(tostring, 3.14)
    m1k(tostring, x)
    print("type")
    m1k(type, '112134')
    m1k(type, 112134)
    m1k(type, {})
    m1k(type, type)
    print("unpack")
    m1k(unpack, {1, 2, 3, 4, 5}, 2, 4)
end

function string_fns()
    print("string.byte")
    m1k(string.byte, "hello string", 3, 7)
    print("string.char")
    m1k(string.char, 72, 101, 108, 108, 111)
    m1k(string.char, 72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100)
    print("string.dump")
    m1k(string.dump, m1k)
    m1k(string.dump, basic_fns)
    print("string.find")
    m1k(string.find, long_str, "numerical")
    m1k(string.find, long_str, "pattern")
    m1k(string.find, long_str, "pattern", 200)
    m1k(string.find, "hello world hellow", "hello")
    m1k(string.find, "hello world hellow", "hello", 3)
    print("string.format")
    m1k(string.format, "string.format %d, %.9f, %e, %g: %s", 1, 1.999999999, 10e9, 3.14, "end of string")
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
    print("string.lower")
    m1k(string.lower, s)
    print("string.match")
    m1k(string.match, s, "(L%w+)")
    print("string.rep")
    m1k(string.rep, s, 2)
    m1k(string.rep, s, 4)
    m1k(string.rep, s, 8)
    m1k(string.rep, s, 16)
    print("string.reverse")
    m1k(string.reverse, s)
    m1k(string.reverse, s2)
    m1k(string.reverse, s2)
    print("string.sub")
    m1k(string.sub, s, 10, 13)
    m1k(string.sub, s, 10, -3)
    print("string.upper")
    m1k(string.upper, s)
    m1k(string.upper, s2)
end

function table_fns()
    local a = copy_arr(arr)
    print("table.concat")
    m1k(table.concat, a, ',')
    m1k(table.concat, a, ',', 3, 7)
    print("table.insert")
    m1k(table.insert, a, 11)
    m1k(table.insert, a, 5, -5)
    print("table.maxn")
    m1k(table.maxn, a)
    print("table.remove")
    m1k(table.remove, a)
    m1k(table.remove, a, 5)
    print("table.sort")
    m1k(table.sort, a)
    m1k(table.sort, a, function(x,y) return x > y end)
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
    md(math.ceil, d)
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
    m1k(bit.lshift, 1, 8)
    print("bit.rshift")
    m1k(bit.rshift, 256, 8)
    print("bit.ashift")
    m1k(bit.arshift, 0x87654321, 12)
    print("bit.rol")
    m1k(bit.rol, 0x12345678, 12)
    print("bit.ror")
    m1k(bit.ror, 0x12345678, 12)
    print("bit.bswap")
    m1k(bit.bswap, 0x12345678)
end

function run_test()
    basic_fns()
    string_fns()
    table_fns()
    math_fns()
    bit_fns()
end

abi.register(run_test)
