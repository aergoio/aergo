function r()
    local t = {}
    t["name"] = "user1"
    t["self"] = t
    return t
end

abi.register(r)
