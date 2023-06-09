function test_view(a)
    contract.event("ev1", 1, "local", 2, "form")
    contract.event("ev1", 3, "local", 4, "form")
end

function k(a)
    return a
end

function tx_in_view_function()
    k2()
end

function k2()
    test_view()
end

function k3()
    ret = contract.pcall(test_view)
    assert(ret == false)
    contract.event("ev2", 4, "global")
end

function tx_after_view_function()
    assert(k(1) == 1)
    contract.event("ev1", 1, "local", 2, "form")
end

function sqltest()
    db.exec([[create table if not exists book (
    page number,
    contents text
    )]])
end

abi.register(test_view, tx_after_view_function, k2, k3)
abi.register_view(test_view, k, tx_in_view_function, sqltest)
