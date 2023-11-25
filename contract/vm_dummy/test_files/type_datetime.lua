state.var {
    cdate = state.value()
}

function constructor()
    cdate:set(905465118)
end

function SetTimestamp(value)
    cdate:set(value)
end

function CreateDate(format, timestamp)
    return system.date(format, timestamp)
end

function Extract(format)
    return system.date(format, cdate:get())
end

function Difftime()
    -- test convertion to table
    s = system.date("*t", cdate:get())
    -- modification of table
    s.hour = 2
    s.min = 0
    s.sec = 0
    -- system.difftime() and system.time()
    diff = system.difftime(cdate:get(), system.time(s))
    -- conversion of diff to hours
    return diff, system.date("%T",diff)
end

abi.register(CreateDate, SetTimestamp, Extract, Difftime)
