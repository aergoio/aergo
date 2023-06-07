paddr = nil
function deploy()
src = [[
    function hello(say, key)
return "Hello " .. say .. key
end
function getcre()
return system.getCreator()
end
function constructor()
end
abi.register(hello, getcre)
]]
paddr = contract.deploy(src)
system.print("addr :", paddr)
ret = contract.call(paddr, "hello", "world", "key")
end
function testPcall()
    ret = contract.pcall(deploy)
    return contract.call(paddr, "getcre")
end

abi.register(testPcall)