function string.tohex(str)
    return (str:gsub('.', function (c)
    return string.format('%02X', string.byte(c))
    end))
end

function query()
    assert (utf8.char(256) == json.decode('"\\u0100"'), "test1")
    a = utf8.char(256,128)
    b = utf8.char(256,10000,45)
    assert(string.len(a) == 4 and utf8.len(a) == 2, "test2")

    for p,c in utf8.codes(a) do
        if p == 1 then
            assert(c == 256, "test11")
        else
            assert(c == 128, "test12")
        end
    end
    assert(utf8.offset(b,1)==1, "test3")
    assert(utf8.offset(b,2)==3, "test4")
    assert(utf8.offset(b,3)==6, "test5")

    assert(utf8.codepoint(b,1)==256, "test6")

    k1, k2, k3 = utf8.codepoint(b,1,3)
    assert(k1 == 256 and k2 == 10000 and k3 == nil, "test7" .. k1 .. k2)

    k1, k2, k3 = utf8.codepoint(b,1,6)
    assert(k1 == 256 and k2 == 10000 and k3 == 45, "test7" .. k1 .. k2 .. k3)
end

function query2()
    a = bignum.number(1000000000000)
    b = bignum.number(0)
    return (bignum.tobyte(a)):tohex(), (bignum.tobyte(b)):tohex()
end

function query3()
    a = bignum.number(-1)
    return (bignum.tobyte(a)):tohex()
end
abi.register(query, query2, query3)
