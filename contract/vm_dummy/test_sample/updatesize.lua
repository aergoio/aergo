function infiniteLoop()
    for i = 1, 100000000000000 do
        system.setItem("key_"..i, "value_"..i)
    end
end
abi.register(infiniteLoop)