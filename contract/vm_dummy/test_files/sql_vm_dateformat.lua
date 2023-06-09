function init()
    db.exec("drop table if exists dateformat_test")
    db.exec([[create table if not exists dateformat_test
    (
    col1 date ,
    col2 datetime ,
    col3 text
    )]])
    db.exec(
        "insert into dateformat_test values (date('2004-10-24 11:11:11'), datetime('2004-10-24 11:11:11'),strftime('%Y%m%d%H%M%S','2004-10-24 11:11:11'))")
    db.exec(
        "insert into dateformat_test values (date(1527504338,'unixepoch'), datetime(1527504338,'unixepoch'), strftime('%Y%m%d%H%M%S',1527504338,'unixepoch') )")
end

function get()
    local rt = {}
    local rs = db.query([[select col1, col2, col3
    from dateformat_test ]])
    while rs:next() do
        local col1, col2, col3 = rs:get()
        table.insert(rt, { col1, col2, col3 })
    end
    return rt
end

abi.register(init, get)
