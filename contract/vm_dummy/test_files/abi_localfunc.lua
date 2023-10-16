function dummy()
    system.print("dummy")
end

local function helper()
    system.print("helper")
end
abi.register(helper)
