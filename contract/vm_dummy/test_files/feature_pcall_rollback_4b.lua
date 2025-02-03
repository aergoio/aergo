state.var {
    resolver = state.value(),
    name = state.value(),
    values = state.map()
}

function constructor(resolver_address, contract_name, use_db)
    -- initialize state variables
    resolver:set(resolver_address)
    name:set(contract_name)
    -- initialize db
    if use_db then
        db.exec("create table config (value integer primary key) without rowid")
        db.exec("insert into config values (0)")
        db.exec([[create table products (
            id integer primary key,
            name text not null,
            price real)
        ]])
        db.exec("insert into products (name,price) values ('first', 1234.56)")
    end
end

function resolve(name)
    return contract.call(resolver:get(), "resolve", name)
end

--[[
  ['set','x',111],
  ['pcall','A']
],[
  ['set','x',222],
  ['pcall','A'],
  ['fail']
],[
  ['set','x',333]
]]

function test(script)
    -- get the commands for this function to execute
    local commands = table.remove(script, 1)

    -- execute each command
    for i, cmd in ipairs(commands) do
        local action = cmd[1]
        if action == "set" then
            contract.event(name:get() .. ".set", cmd[2], cmd[3])
            values[cmd[2]] = cmd[3]
        elseif action == "pcall" then
            local to_call = cmd[2]
            local amount = cmd[3]
            if to_call == name:get() then
                pcall(test, script)
            elseif amount ~= nil then
                contract.event(name:get() .. " send " .. to_call, amount)
                local address = resolve(to_call)
                success, ret = pcall(function()
                    return contract.call.value(amount)(address, "test", script)
                end)
            else
                local address = resolve(to_call)
                success, ret = pcall(contract.call, address, "test", script)
            end
            --contract.event("result", success, ret)
        elseif action == "send" then
            local to = cmd[2]
            contract.event(name:get() .. " send " .. to, amount)
            contract.send(resolve(to), cmd[3])
        elseif action == "deploy" then
            local code_or_address = resolve_deploy(cmd[2])
            pcall(contract.deploy, code_or_address, unpack(cmd,3))
        elseif action == "deploy.send" then
            contract.event(name:get() .. ".deploy.send", amount)
            local code_or_address = resolve_deploy(cmd[3])
            pcall(function()
                contract.deploy.value(cmd[2])(code_or_address, unpack(cmd,4))
            end)
        elseif action == "db.set" then
            db.exec("update config set value = " .. cmd[2])
        elseif action == "db.insert" then
            db.exec("insert into products (name,price) values ('another',1234.56)")
        elseif action == "fail" then
            assert(1 == 0, "failed")
        end
    end

end

function set(key, value)
    values[key] = value
end

function get(key)
    return values[key]
end

function db_reset()
    db.exec("update config set value = 0")
    db.exec("delete from products where id > 1")
end

function db_get()
    local rs = db.query("select value from config")
    if rs:next() then
        return rs:get()
    else
        return "error"
    end
end

function db_count()
    local rs = db.query("select count(*) from products")
    if rs:next() then
        return rs:get()
    else
        return "error"
    end
end

function default()
    -- only receive
end

function send_back(to)
    contract.send(resolve(to), contract.balance())
end

abi.payable(constructor, test, default)
abi.register(set, send_back, db_reset)
abi.register_view(get, db_get, db_count)
