state.var {
  Map = state.map(),
}
function map(a)
  return Map[a]
end
function constructor()
end
function pcall3(to)
  contract.send(to, "1 aergo")
end
function pcall2(addr, to)
  status = pcall(contract.call, addr, "pcall3", to)
  system.print(status)
  assert(false)
end
function pcall1(addr, to)
  status = pcall(contract.call, addr, "pcall2", addr, to)
  system.print(status)
  Map[addr] = 2
  status = pcall(contract.call, addr, "pcall3", to)
  system.print(status)
  status = pcall(contract.call, addr, "pcall2", addr, to)
  system.print(status)
end
function default()
end
abi.register(map, pcall1, pcall2, pcall3, default)
abi.payable(pcall1, default, constructor)