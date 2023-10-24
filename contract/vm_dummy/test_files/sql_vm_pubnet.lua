function createAndInsert()
    db.exec("create table if not exists dual(dummy char(1))")
    db.exec("insert into dual values ('X')")
    local insertYZ = db.prepare("insert into dual values (?),(?)")
    insertYZ:exec("Y", "Z")
end

abi.register(createAndInsert)
