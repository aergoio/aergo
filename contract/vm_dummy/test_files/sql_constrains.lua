function init()
    db.exec([[create table if not exists r (
    id integer primary key
    , n integer check(n >= 10)
    , nonull text not null
    , only integer unique)
    ]])
    db.exec("insert into r values (1, 11, 'text', 1)")
    db.exec("create table if not exists s (rid integer references r(id))")
end

function pkFail()
    db.exec("insert into r values (1, 12, 'text', 2)")
end

function checkFail()
    db.exec("insert into r values (2, 9, 'text', 3)")
end

function fkFail()
    db.exec("insert into s values (2)")
end

function notNullFail()
    db.exec("insert into r values (2, 13, null, 2)")
end

function uniqueFail()
    db.exec("insert into r values (2, 13, 'text', 1)")
end

abi.register(init, pkFail, checkFail, fkFail, notNullFail, uniqueFail)