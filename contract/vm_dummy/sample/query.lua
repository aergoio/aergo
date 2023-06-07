function inc()
    a = system.getItem("key1")
    if (a == nil) then
        system.setItem("key1", 1)
        return
    end
    system.setItem("key1", a + 1)
    end
    function query(a)
        return system.getItem(a)
    end
abi.register(query)
abi.payable(inc)