state.var {
    h = state.map(),
    arr = state.array(10),
    v = state.value()
}

t = {}

function key_table()
    local k = {}
    t[k] = "table"
end

function key_func()
    t[key_table] = "function"
end

function key_statemap(key)
    t[h] = "state.map"
end

function key_statearray(key)
    t[arr] = "state.array"
end

function key_statevalue(key)
    t[v] = "state.value"
end

function key_upval(key)
    local k = {}
    local f = function()
        t[k] = "upval"
    end
    f()
end

function key_nil(key)
    h[nil] = "nil"
end

abi.register(key_table, key_func, key_statemap, key_statearray, key_statevalue, key_upval, key_nil)
