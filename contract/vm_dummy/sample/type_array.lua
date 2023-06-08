state.var{
    counts = state.array(10)
}

function inc(key)
    if counts[key] == nil then
        counts[key] = 0
    end
    counts[key] = counts[key] + 1
end

function get(key)
    return counts[key]
end

function set(key,val)
    counts[key] = val
end

function len()
    return counts:length()
end

function iter()
    local rv = {}
    for i, v in counts:ipairs() do 
        if v == nil then
            rv[i] = "nil"
        else
            rv[i] = v
        end
    end
    return rv
end

abi.register(inc,get,set,len,iter)