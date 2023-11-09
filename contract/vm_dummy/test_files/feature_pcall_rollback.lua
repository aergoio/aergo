state.var {
    resolver = state.value(),
    name = state.value(),
    values = state.map()
}

function constructor(resolver_address, contract_name)
    resolver:set(resolver_address)
    name:set(contract_name)
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
            if to_call == name:get() then
                contract.pcall(test, script)
            else
                contract.pcall(contract.call, resolve(to_call), "test", script)
            end
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

abi.register(test, set)
abi.register_view(get)
