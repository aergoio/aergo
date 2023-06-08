function init()
    db.exec("create table if not exists dt_test (n datetime, b bool)")
    local n = db.exec("insert into dt_test values (?, ?),(date('2004-10-24', '+1 month', '-1 day'), 0)", 10000, 1)
    assert(n == 2, "change count mismatch");
end

function nowNull()
    db.exec("insert into dt_test values (date('now'), 0)")
end

function localtimeNull()
    db.exec("insert into dt_test values (datetime('2018-05-25', ?), 1)", 'localtime')
end

function get()
    local rs = db.query("select n, b from dt_test order by 1, 2")
    local r = {}
    while rs:next() do
        local d, b = rs:get()
        table.insert(r, { date= d, bool= b })
    end
    return r
end
abi.register(init, nowNull, localtimeNull, get)