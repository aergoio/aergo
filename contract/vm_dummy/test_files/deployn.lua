function constructor()
    testall()
end

function testall()
    deploytest()
    sendtest()
end

function deploytest()
    src = [[
    function default()
      contract.send(system.getSender(), system.getAmount())
      system.setItem('last-amount', system.getAmount())
      return system.getAmount()
    end

    function get_last_amount()
      return system.getItem('last-amount')
    end

    function getargs(...)
      tb = {...}
    end

    abi.payable(default)
    abi.register(getargs, get_last_amount)
    ]]

    addr = contract.deploy(src)
    id = 'deploy_src'; system.setItem(id, addr)
    assert(system.getItem(id) == addr, "deploy_src")

    korean_char_src = [[
    function 함수()
      변수 = 1
      결과 = 변수 + 3
      system.print('결과', 결과)
    end

    abi.register(함수)
    ]]


    korean_char_src222 = [[
    function default()
      contract.send(system.getSender(), system.getAmount())
    end

    function getargs(...)
      tb = {...}
    end

    function x()
      -- empty
    end

    abi.payable(default)
    abi.register(getargs)
    ]]

    korean_addr = contract.deploy(korean_char_src)
    id = 'korean_char_src'; system.setItem(id, korean_addr)
    assert(system.getItem(id) == korean_addr, "korean_char_src")
end

-- also make sure that system.getAmount() is not the value
-- sent on the transaction on the second contract. it must
-- be the amount sent by this contract instead.

function sendtest()
    addr = system.getItem("deploy_src")
    --system.print('ADDRESS', addr, system.getAmount())

    local amount_str = system.getAmount()
    local amount_big = bignum.number(amount_str)

    -- transfer by using contract.send()

    contract.send(addr, amount_big)
    local ret = contract.call(addr, "get_last_amount")
    --system.print('call 1', amount_str, ret)
    assert(ret == amount_str, "amount 1")

    contract.send(addr, amount_str)
    local ret = contract.call(addr, "get_last_amount")
    --system.print('call 2', amount_str, ret)
    assert(ret == amount_str, "amount 2")

    contract.send(addr, bignum.number(0))
    local ret = contract.call(addr, "get_last_amount")
    --system.print('call 3', "0", ret)
    assert(ret == "0", "amount 3")

    contract.send(addr, "0")
    local ret = contract.call(addr, "get_last_amount")
    --system.print('call 4', "0", ret)
    assert(ret == "0", "amount 4")

    -- transfer by calling the 'default' function directly

    local ret = contract.call.value(amount_big)(addr, "default")
    --system.print('call 5', amount_str, ret)
    assert(ret == amount_str, "amount 5")

    local ret = contract.call.value(amount_str)(addr, "default")
    --system.print('call 6', amount_str, ret)
    assert(ret == amount_str, "amount 6")

    local ret = contract.call.value(bignum.number(1))(addr, "default")
    --system.print('call 7', "1", ret)
    assert(ret == "1", "amount 7")

    local ret = contract.call.value(bignum.number(0))(addr, "default")
    --system.print('call 8', "0", ret)
    assert(ret == "0", "amount 8")

    local ret = contract.call(addr, "default")
    --system.print('call 9', "0", ret)
    assert(ret == "0", "amount 9")

end

function default()
    -- do nothing
end

abi.payable(constructor, default, testall)
