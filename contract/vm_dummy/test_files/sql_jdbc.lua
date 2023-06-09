function init()
    db.exec("create table if not exists total(a int, b int, c text)")
    db.exec("insert into total(a,c) values (1,2)")
    db.exec("insert into total values (2,2,3)")
    db.exec("insert into total values (3,2,3)")
    db.exec("insert into total values (4,2,3)")
    db.exec("insert into total values (5,2,3)")
    db.exec("insert into total values (6,2,3)")
    db.exec("insert into total values (7,2,3)")
end

function exec(sql, ...)
    local stmt = db.prepare(sql)
    stmt:exec(...)
end

function query(sql, ...)
    local stmt = db.prepare(sql)
    local rs = stmt:query(...)
    local r = {}
    local colcnt = rs:colcnt()
    local colmetas
    while rs:next() do
        if colmetas == nil then
            colmetas = stmt:column_info()
        end

        local k = {rs:get()}
        for i = 1, colcnt do
            if k[i] == nil then
                k[i] = {}
            end
        end
        table.insert(r, k)
    end
    --  if (#r == 0) then
    --      return {"colcnt":0, "rowcnt":0}
    --  end

    return {snap=db.getsnap(), colcnt=colcnt, rowcnt=#r, data=r, colmetas=colmetas}
end

function queryS(snap, sql, ...)
    db.open_with_snapshot(snap)

    local stmt = db.prepare(sql)
    local rs = stmt:query(...)
    local r = {}
    local colcnt = rs:colcnt()
    local colmetas
    while rs:next() do
        if colmetas == nil then
            colmetas = stmt:column_info()
        end

        local k = {rs:get()}
        for i = 1, colcnt do
            if k[i] == nil then
                k[i] = {}
            end
        end
        table.insert(r, k)
    end
    --  if (#r == 0) then
    --      return {"colcnt":0, "rowcnt":0}
    --  end

    return {snap=db.getsnap(), colcnt=colcnt, rowcnt=#r, data=r, colmetas=colmetas}
end
function getmeta(sql)
    local stmt = db.prepare(sql)

    return stmt:column_info(), stmt:bind_param_cnt()
end
abi.register(init, exec, query, getmeta, queryS)