state.var {
  Say = state.value()
}

function hello(say)
  return "Hello " .. say
end

abi.register(hello)