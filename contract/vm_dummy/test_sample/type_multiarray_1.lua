state.var{
    mcounts = state.map(2),
    array = state.array(10, 11),
    tcounts = state.map(3)
}

function inc()
    a = system.getItem("key1")
    if (a == nil) then
        system.setItem("key1", 1)
        return
    end
    system.setItem("key1", a + 1)
    mcounts[system.getSender()]["key1"] = a + 1
    array[1][10] = "k"
    array[10][5] = "l"
    tcounts[0][0][0] = 2
end
function query(a)
    return system.getItem("key1"), mcounts[a]["key1"], tcounts[0][0][0], tcounts[1][2][3], array:length(), array[1]:length()
end
function del()
    tcounts[0][0]:delete(0)
    tcounts[1][2]:delete(3)
end
function iter(a)
    local rv = {}
    for i, x in array:ipairs() do
        for j, y in x:ipairs() do
            if y ~= nil then
                rv[i..","..j] =  y
            end
        end
    end
    return rv
end

function seterror()
    rv, err = pcall(function () mcounts[1]["k2y1"] = 4 end)
    assert(rv == false and string.find(err, "string expected, got number"))
    rv, err = pcall(function () mcounts["middle"] = 4 end)
    assert(rv == false and string.find(err, "not permitted to set intermediate dimension of map"))
    rv, err = pcall(function () array[1] = 4 end)
    assert(rv == false and string.find(err, "not permitted to set intermediate dimension of array"))
    rv, err = pcall(function () tcounts[0]:delete(0) end)
    assert(rv == false and string.find(err, "not permitted to set intermediate dimension of map"))
    rv, err = pcall(function () tcounts[0][1]:delete() end)
    assert(rv == false and string.find(err, "invalid key type: 'no value', state.map: 'tcounts'"))
    rv, err = pcall(function () array[0]:append(2) end)
    assert(rv == false and string.find(err, "the fixed array cannot use 'append' method"))
    rv, err = pcall(function () state.var {k = state.map(6)} end)
    assert(rv == false and string.find(err, "dimension over max limit"), err)
    rv, err = pcall(function () state.var {k = state.array(1,2,3,4,5,6)} end)
    assert(rv == false and string.find(err, "dimension over max limit"), err)
end

abi.register(inc, query, iter, seterror, del)
abi.payable(inc)