-- Define global variables.
state.var {
  Name = state.value()
}

-- Initialize a name in this contract.
function constructor()
  -- a constructor is called at a deployment, only one times
  -- set initial name
  Name:set("world")
end

-- Update a name.
-- @call
-- @param name          string: new name.
function set_name(name)
  Name:set(name)
end
  
-- Say hello.
-- @query
-- @return              string: 'hello ' + name
function hello()
  return "hello " .. Name:get()
end
  
 -- register functions to expose
abi.register(set_name, hello)