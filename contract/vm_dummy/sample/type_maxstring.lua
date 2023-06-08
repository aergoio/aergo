function oom()
    local s = "hello"

    while 1 do
        s = s .. s
    end
end

function p()
   pcall(oom)
end

function cp()
   contract.pcall(oom)
end
abi.register(oom, p, cp)