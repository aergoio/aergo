function init()
    db.exec("create table if not exists auto_test (a integer primary key autoincrement, b text)")
    local n = db.exec("insert into auto_test(b) values (?),(?)", 10000, 1)
    assert(n == 2, "change count mismatch");
end

function get()
    db.exec("insert into auto_test(b) values ('ss')")
    assert(db.last_insert_rowid() == 3, "id is not valid")
end

abi.register(init, get)
