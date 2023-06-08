state.var{
    counts = state.map()
}
function setCount(key, value)
    counts[key] = value
end
function getCount(key)
    return counts[key]
end
function delCount(key)
    counts:delete(key)
end
abi.register(setCount, getCount, delCount)