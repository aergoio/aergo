function createAndInsert()
    db.exec("create table if not exists dual(dummy char(1))")
    db.exec("insert into dual values ('X')")
    local insertYZ = db.prepare("insert into dual values (?),(?)")
    insertYZ:exec("Y", "Z")
end

function insertRollbackData()
    db.exec("insert into dual values ('A'),('B'),('C')")
end

function query()
    local rt = {}
    local stmt = db.prepare("select ?+1, round(?, 1), dummy || ? as col3 from dual order by col3")
    local rs = stmt:query(1, 3.14, " Hello Blockchain")
    while rs:next() do
        local col1, col2, col3 = rs:get()
        table.insert(rt, col1)
        table.insert(rt, col2)
        table.insert(rt, col3)
    end
    return rt
end

function count()
    local rs = db.query("select count(*) from dual")
    if rs:next() then
        local n = rs:get()
        --rs:next()
        return n
    else
        return "error in count()"
    end
end

function all()
    local rt = {}
    local rs = db.query("select dummy from dual order by 1")
    while rs:next() do
        local col = rs:get()
        table.insert(rt, col)
    end
    return rt
end

abi.register(createAndInsert, insertRollbackData, query, count, all)
