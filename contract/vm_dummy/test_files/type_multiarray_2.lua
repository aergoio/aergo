state.var {
  -- global map
  alpha = state.map(2),
  beta = state.map(2),
}

function constructor()

  local d = alpha["dict"]
  d["a"] = "A"
  alpha["dict"]["b"] = "B"

  assert(alpha["dict"]["a"]=="A" and alpha["dict"]["b"]=="B")

  -- with local variable
  local d2 = beta["dict"]
  d2["a"] = "A"
  d2["value"] = "v0"
  beta["dict"]["b"] = "B"
  beta["dict"]["value"] = "v1"
  assert(beta["dict"]["a"]=="A" and beta["dict"]["b"]=="B" and beta["dict"]["value"]=="v1")
end

function abc()
  local d = alpha["dict"]
  d["c"] = "C"
  alpha["dict"]["d"] = "D"

  local d = beta["dict"]
  d["a"] = "A"
  d["value"] = "v2"
  beta["dict"]["b"] = "B"
  beta["dict"]["value"] = "v3"
  return alpha["dict"]["c"], alpha["dict"]["d"], beta["dict"]["a"], beta["dict"]["b"], beta["dict"]["value"]
end

function query()
  return alpha["dict"]["a"], alpha["dict"]["b"], alpha["dict"]["c"], alpha["dict"]["d"],
  beta["dict"]["a"], beta["dict"]["b"], beta["dict"]["value"]
end

abi.register(abc, query)