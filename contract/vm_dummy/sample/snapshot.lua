state.var{
    counts = state.map(),
    data = state.value(),
    array = state.array(10)
}

function inc()
    a = system.getItem("key1")
    if (a == nil) then
        system.setItem("key1", 1)
        return
    end
    system.setItem("key1", a + 1)
    counts["key1"] = a + 1
    data:set(a+1)
    array[1] = a + 1
end
function query(a)
        return system.getItem("key1", a), state.getsnap(counts, "key1", a), state.getsnap(data,a), state.getsnap(array, 1, a)
end
function query2()
        return state.getsnap(array, 1)
end
abi.register(inc, query, query2)
abi.payable(inc)