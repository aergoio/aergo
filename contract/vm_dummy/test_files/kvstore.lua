state.var{
    counts = state.map(),
    name = state.value()
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

function setname(n)
    name:set(n)
end

function getname()
    return name:get()
end

abi.register(inc,get,set,setname,getname)