function constructor()
    db.exec("create table if not exists t (col integer primary key)")
    db.exec("insert into t values (1)")
end

function stmt_exec(stmt)
    db.exec(stmt)
end

function stmt_exec_pcall(stmt)
    pcall(db.exec, stmt)
end

function get()
    local rs = db.query("select col from t order by col")
    local t = {}
    while rs:next() do
        local col = rs:get()
        table.insert(t, col)
    end
    return t
end

abi.register(stmt_exec, stmt_exec_pcall, get)
