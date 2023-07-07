function copy(arr)
    assert(type(arr) == "table", "table expected")
    local rv = {}
    for i, v in ipairs(arr) do
        table.insert(rv, i, v)
    end
    return rv
end

function two_arr(arr1, arr2)
    assert(type(arr1) == "table", "table expected")
    assert(type(arr2) == "table", "table expected")
    local rv = {}
    table.insert(rv, 1, #arr1)
    table.insert(rv, 2, #arr2)
    return rv
end

function mixed_args(arr1, map1, n)
    assert(type(arr1) == "table", "table expected")
    assert(type(map1) == "table", "table expected")
    local rv = {}
    table.insert(rv, 1, arr1)
    table.insert(rv, 2, map1)
    table.insert(rv, 3, n)
    return rv
end

abi.register(copy, two_arr, mixed_args)
