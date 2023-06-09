state.var {
    Person = state.value()
}

function constructor()
    Person:set({ name = "kslee", age = 38, address = "blahblah..." })
end

function InvalidUpdateAge(age)
    Person:get().age = age
end

function ValidUpdateAge(age)
    local p = Person:get()
    p.age = age
    Person:set(p)
end

function GetPerson()
    return Person:get()
end

abi.register(InvalidUpdateAge, ValidUpdateAge, GetPerson)
