function a()
    src = [[
    while true do
    end
    function b()
    end
    abi.register(b)
    ]]
    contract.deploy(src)
end

abi.register(a)
