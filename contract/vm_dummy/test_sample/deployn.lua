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
    end

    function getargs(...)
    tb = {...}
    end

    abi.payable(default)
    abi.register(getargs)
  ]]
  
  addr = contract.deploy(src)
  id = 'deploy_src'; system.setItem(id, addr)
  system.print(id, system.getItem(id))

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
  end

  abi.payable(default)
  abi.register(getargs)
  ]]

  korean_addr = contract.deploy(korean_char_src)
  id = 'korean_char_src'; system.setItem(id, korean_addr)
  system.print(id, system.getItem(id))
end

function sendtest()
  addr = system.getItem("deploy_src")
  system.print('ADDRESS', addr, system.getAmount())

  id = 's01'; system.setItem(id,{pcall(function() contract.send(addr, system.getAmount()) end)})
  system.print(id, system.getItem(id))
end

function default()
  -- do nothing
end
  
abi.payable(constructor, default)
abi.register(testall)