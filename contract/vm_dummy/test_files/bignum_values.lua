
function parse_bignum(value)
  local val = bignum.number(value)
  return bignum.tostring(val)
end

abi.register(parse_bignum)
