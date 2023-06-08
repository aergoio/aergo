function constructor()
    db.exec("create table if not exists aergojdbc001 (name text, yyyymmdd text)")
    db.exec("insert into aergojdbc001 values ('홍길동', '20191007')")
    db.exec("insert into aergojdbc001 values ('홍길동', '20191007')")
    db.exec("insert into aergojdbc001 values ('홍길동', '20191007')")
end

function inserts()
    db.exec("insert into aergojdbc001 select * from aergojdbc001")
end

abi.register(inserts)