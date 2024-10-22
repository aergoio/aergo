function is_table_equal(t1, t2)
    local ty1 = type(t1)
    local ty2 = type(t2)
    if ty1 ~= ty2 then return false end
    -- non-table types can be directly compared
    if ty1 ~= 'table' and ty2 ~= 'table' then return t1 == t2 end
    -- if table, compare each key-value pair
    for k1, v1 in pairs(t1) do
        local v2 = t2[k1]
        if v2 == nil or not is_table_equal(v1, v2) then return false end
    end
    for k2, v2 in pairs(t2) do
        local v1 = t1[k2]
        if v1 == nil or not is_table_equal(v1, v2) then return false end
    end
    return true
end

function r()
    local t = {}
    t[10000] = "1234"
    system.setItem("k", t)
    k = system.getItem("k")
    if is_table_equal(t, k) then
        return 1
    end
    return 0
end

abi.register(r)
