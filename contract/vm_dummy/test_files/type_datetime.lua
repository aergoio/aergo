state.var {
    cdate = state.value()
}

function constructor()
    cdate:set(906000490)
end

function CreateDate()
    return system.date("%c", cdate:get())
end

function Extract(fmt)
    return system.date(fmt, cdate:get())
end

function Difftime()
    system.print(system.date("%c", cdate:get()))
    s = system.date("*t", cdate:get())
    system.print(s)
    s.hour = 2
    s.min = 0
    s.sec = 0
    system.print(system.date("*t", system.time(s)))
    return system.difftime(cdate:get(), system.time(s))
end

abi.register(CreateDate, Extract, Difftime)
