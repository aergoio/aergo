function oom_string()
    local s = "hello"
    while 1 do
        s = s .. s
    end
end

function oom_table1()
    local t = {}
    for i = 1, 10000000 do
        t[i] = i
    end
end

function oom_table2()
    local t = {}
    for i = 1, 10000000 do
        local str = "key" .. tostring(i)
        t[str] = str
    end
end

function oom_global()
    local s1 = "hello"
    local s2 = "hello"
    local s3 = "hello"
    local s4 = "hello"
    while 1 do
        s1 = s1 .. s1
        s2 = s2 .. s2
        s3 = s3 .. s3
        s4 = s4 .. s4
    end
end

function pcall_string()
    pcall(oom_string)
end

function pcall_table1()
    pcall(oom_table1)
end

function pcall_table2()
    pcall(oom_table2)
end

function pcall_global()
    pcall(oom_global)
end

function contract_pcall_string()
    contract.pcall(oom_string)
end

function contract_pcall_table1()
    contract.pcall(oom_table1)
end

function contract_pcall_table2()
    contract.pcall(oom_table2)
end

function contract_pcall_global()
    contract.pcall(oom_global)
end

abi.register(oom_string, pcall_string, contract_pcall_string)
abi.register(oom_table1, pcall_table1, contract_pcall_table1)
abi.register(oom_table2, pcall_table2, contract_pcall_table2)
abi.register(oom_global, pcall_global, contract_pcall_global)
