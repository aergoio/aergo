function constructor(init)
    system.setItem("count", init)
end

function init()
    db.exec([[create table if not exists r (
    id integer primary key
    , n integer check(n >= 10)
    , nonull text not null
    , only integer unique)
    ]])
    db.exec("insert into r values (1, 11, 'text', 1)")
end

function pkins1()
    db.exec("insert into r values (3, 12, 'text', 2)")
    db.exec("insert into r values (1, 12, 'text', 2)")
end

function pkins2()
    db.exec("insert into r values (4, 12, 'text', 2)")
end

function pkget()
    local rs = db.query("select count(*) from r")
    if rs:next() then
        local n = rs:get()
        --rs:next()
        return n
    else
        return "error in count()"
    end
end

function inc()
    count = system.getItem("count")
    system.setItem("count", count + 1)
    return count
end

function get()
    return system.getItem("count")
end

function getOrigin()
    return system.getOrigin()
end

function set(val)
    system.setItem("count", val)
end

abi.register(inc, get, set, init, pkins1, pkins2, pkget, getOrigin)
abi.payable(constructor, inc)
