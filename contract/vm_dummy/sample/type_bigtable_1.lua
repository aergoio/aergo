function constructor()
    db.exec("create table if not exists table1 (cid integer PRIMARY KEY, rgtime datetime)")
    db.exec("insert into table1 (rgtime) values (datetime('2018-10-30 16:00:00'))")
end

function inserts(n)
    for i = 1, n do
        db.exec("insert into table1 (rgtime) select rgtime from table1")
    end
end

abi.register(inserts)