
function loop(n)
  local a = 0
  for i = 1,n do
    a = a + i
  end
  return a
end

abi.register(loop)
