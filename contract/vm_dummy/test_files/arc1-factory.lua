arc1_core = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- Core
------------------------------------------------------------------------------

extensions = {}

---- State Data for Token
state.var {
  _contract_owner = state.value(),

  _balances = state.map(),        -- address -> unsigned_bignum
  _totalSupply = state.value(),   -- unsigned_bignum

  _name = state.value(),          -- string
  _symbol = state.value(),        -- string
  _decimals = state.value(),      -- string

  -- Token Metadata
  _metakeys = state.map(),        -- number -> string
  _metadata = state.map(),        -- string -> string

  -- Pausable
  _paused = state.value(),        -- boolean

  -- Blacklist
  _blacklist = state.map()        -- address -> boolean
}

address0 = '1111111111111111111111111111111111111111111111111111' -- null address

-- Type check
-- @type internal
-- @param x variable to check
-- @param t (string) expected type
local function _typecheck(x, t)
  if (x and t == 'address') then    -- a string containing an address
    assert(type(x) == 'string', "ARC1: address must be a string")
    -- check address length
    assert(#x == 52, string.format("ARC1: invalid address length (%s): %s", #x, x))
    -- check address checksum
    if x ~= address0 then
      local success = pcall(system.isContract, x)
      assert(success, "ARC1: invalid address: " .. x)
    end
  elseif (x and t == 'ubig') then   -- a positive big integer
    -- check unsigned bignum
    assert(bignum.isbignum(x), string.format("ARC1: invalid type: %s != %s", type(x), t))
    assert(x >= bignum.number(0), string.format("ARC1: %s must be positive number", bignum.tostring(x)))
  elseif (x and t == 'uint') then   -- a positive lua integer
    assert(type(x) == 'number', string.format("ARC1: invalid type: %s != number", type(x)))
    assert(math.floor(x) == x, "ARC1: the number must be an integer")
    assert(x >= 0, "ARC1: the number must be 0 or positive")
  else
    -- check default lua types
    assert(type(x) == t, string.format("ARC1: invalid type: %s != %s", type(x), t or 'nil'))
  end
end

-- check or convert the input to a bignum
function _check_bignum(x)
  -- if the input is a string, convert it to bignum
  if type(x) == 'string' then
    -- check for valid characters: 0-9 and .
    assert(string.match(x, '[^0-9.]') == nil, "ARC1: amount contains invalid character")
    -- count the number of dots
    local _, count = string.gsub(x, "%.", "")
    assert(count <= 1, "ARC1: the amount is invalid")
    -- if the number has a dot
    if count == 1 then
      local num_decimals = _decimals:get()
      -- separate the integer part and the decimal part
      local p1, p2 = string.match('0' .. x .. '0', '(%d+)%.(%d+)')
      -- calculate the number of digits to add
      local to_add = num_decimals - #p2
      if to_add > 0 then
        -- add trailing zeros
        p2 = p2 .. string.rep('0', to_add)
      elseif to_add < 0 then
        -- do not remove trailing digits
        --p2 = string.sub(p2, 1, num_decimals)
        assert(false, "ARC1: too many decimal digits")
      end
      -- join the integer part and the decimal part
      x = p1 .. p2
      -- remove leading zeros
      x = string.gsub(x, '0*', '', 1)
      -- if the result is an empty string, set it to '0'
      if #x == 0 then x = '0' end
    end
    -- convert the string to bignum
    x = bignum.number(x)
  end
  -- check if it is a valid unsigned bignum
  _typecheck(x, 'ubig')
  return x
end

-- initialize the token contract
-- this function should be called by the constructor
-- this function can be called only once
-- @type internal
-- @param name (string) name of this token
-- @param symbol (string) symbol of this token
-- @param decimals (number) decimals of this token
-- @param owner (optional:address) the owner of this contract
local function _init(name, symbol, decimals, owner)

  -- check if the contract is already initialized
  assert(_name:get() == nil, "ARC1: the contract is already initialized")

  if owner == nil or owner == '' then
    owner = system.getCreator()
  elseif owner == 'none' then
    owner = nil
  else
    _typecheck(owner, "address")
  end
  _contract_owner:set(owner)

  _typecheck(name, 'string')
  _typecheck(symbol, 'string')
  _typecheck(decimals, 'uint')

  assert(decimals >= 0 and decimals <= 18, "decimals must be between 0 and 18")

  _name:set(name)
  _symbol:set(symbol)
  _decimals:set(decimals)

  _totalSupply:set(bignum.number(0))
  _paused:set(false)

end

-- Get the token name
-- @type    query
-- @return  (string) name of this token
function name()
  return _name:get()
end

-- Get the token symbol
-- @type    query
-- @return  (string) symbol of this token
function symbol()
  return _symbol:get()
end

-- Get the token decimals
-- @type    query
-- @return  (number) decimals of this token
function decimals()
  return _decimals:get()
end

-- Store token metadata
-- @type    call
-- @param   metadata (table)  lua table containing key-value pairs
function set_metadata(metadata)

  assert(system.getSender() == _contract_owner:get(), "ARC1: permission denied")

  for key,value in pairs(metadata) do
    for i=1,1000,1 do
      local skey = _metakeys[tostring(i)]
      if skey == nil then
        _metakeys[tostring(i)] = key
        break
      end
      if skey == key then
        break
      end
    end
    _metadata[key] = value
  end

end

-- Get token metadata
-- @type    query
-- @return  (string) if key is nil, return all metadata from token,
--                   otherwise return the value linked to the key
function get_metadata(key)

  if key ~= nil then
    return _metadata[key]
  end

  local items = {}
  for i=1,1000,1 do
    key = _metakeys[tostring(i)]
    if key == nil then break end
    local value = _metadata[key]
    items[key] = value
  end
  return items

end

-- Get the balance of an account
-- @type    query
-- @param   owner  (address)
-- @return  (ubig) balance of owner
function balanceOf(owner)
  if owner == nil then
    owner = system.getSender()
  else
    _typecheck(owner, 'address')
  end

  return _balances[owner] or bignum.number(0)
end

-- Return the total supply of this token
-- @type    query
-- @return  (ubig) total supply of this token
function totalSupply()
  return _totalSupply:get()
end

-- register exported functions
abi.register(set_metadata)
abi.register_view(name, symbol, decimals, get_metadata, totalSupply, balanceOf)

-- Call the tokensReceived() function on the recipient contract after a transfer or mint
-- @type internal
-- @param   from   (address) sender's address
-- @param   to     (address) recipient's address
-- @param   amount (ubig) the amount of token that was sent
-- @param   ...    additional data which is sent unaltered in the call
-- @return  value returned from the 'tokensReceived' callback, or nil
local function _callTokensReceived(from, to, amount, ...)
  if to ~= address0 and system.isContract(to) then
    return contract.call(to, "tokensReceived", system.getSender(), from, amount, ...)
  else
    return nil
  end
end

-- Transfer tokens from an account to another
-- @type    internal
-- @param   from    (address) sender's address
-- @param   to      (address) recipient's address
-- @param   amount  (ubig)    amount of token to send
-- @param   ...     additional data, is sent unaltered in call to 'tokensReceived' on 'to'
-- @return  value returned from 'tokensReceived' callback, or nil
local function _transfer(from, to, amount, ...)
  assert(not _paused:get(), "ARC1: paused contract")
  assert(not _blacklist[from], "ARC1: sender is on blacklist")
  assert(not _blacklist[to], "ARC1: recipient is on blacklist")

  -- block transfers of `0` amount
  assert(amount > bignum.number(0), "ARC1: invalid amount")

  local balance = _balances[from] or bignum.number(0)
  assert(balance >= amount, "ARC1: not enough balance")

  _balances[from] = balance - amount
  _balances[to] = (_balances[to] or bignum.number(0)) + amount

  return _callTokensReceived(from, to, amount, ...)
end

-- Mint new tokens to an account
-- @type    internal
-- @param   to      (address) recipient's address
-- @param   amount  (ubig) amount of tokens to mint
-- @param   ...     additional data, is sent unaltered in call to 'tokensReceived' on 'to'
-- @return  value returned from 'tokensReceived' callback, or nil
local function _mint(to, amount, ...)
  assert(not _paused:get(), "ARC1: paused contract")
  assert(not _blacklist[to], "ARC1: recipient is on blacklist")

  _totalSupply:set((_totalSupply:get() or bignum.number(0)) + amount)
  _balances[to] = (_balances[to] or bignum.number(0)) + amount

  return _callTokensReceived(system.getSender(), to, amount, ...)
end

-- Burn tokens from an account
-- @type    internal
-- @param   from   (address)
-- @param   amount  (ubig) amount of tokens to burn
local function _burn(from, amount)
  assert(not _paused:get(), "ARC1: paused contract")
  assert(not _blacklist[from], "ARC1: sender is on blacklist")

  assert(amount > bignum.number(0), "ARC1: invalid amount")

  local balance = _balances[from] or bignum.number(0)
  assert(balance >= amount, "ARC1: not enough balance")

  _balances[from] = balance - amount
  _totalSupply:set(_totalSupply:get() - amount)
end

-- Transfer tokens to an account (from caller)
-- @type    call
-- @param   to      (address) recipient's address
-- @param   amount  (ubig) amount of tokens to send
-- @param   ...     additional data, is sent unaltered in call to 'tokensReceived' on 'to'
-- @return  value returned from 'tokensReceived' callback, or nil
-- @event   transfer(from, to, amount)
function transfer(to, amount, ...)
  _typecheck(to, 'address')
  amount = _check_bignum(amount)
  local from = system.getSender()

  contract.event("transfer", from, to, bignum.tostring(amount))

  return _transfer(from, to, amount, ...)
end

-- Define a new contract owner
function set_contract_owner(address)
  assert(system.getSender() == _contract_owner:get(), "ARC1: permission denied")
  _typecheck(address, "address")
  _contract_owner:set(address)
end

-- Returns a JSON string containing the list of ARC1 extensions
-- that were included on the contract
function arc1_extensions()
  local list = {}
  for name,_ in pairs(extensions) do
    table.insert(list, name)
  end
  return list
end

-- register exported functions
abi.register(transfer, set_contract_owner)
abi.register_view(arc1_extensions)
]]

arc1_burnable = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- Burnable
------------------------------------------------------------------------------

extensions["burnable"] = true

-- Burn tokens from the caller account
-- @type    call
-- @param   amount  (ubig) amount of tokens to burn
-- @event   burn(account, amount)
function burn(amount)
  amount = _check_bignum(amount)

  local sender = system.getSender()

  _burn(sender, amount)

  contract.event("burn", sender, bignum.tostring(amount))
end

-- register the exported functions
abi.register(burn)
]]

arc1_mintable = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- Mintable
------------------------------------------------------------------------------

extensions["mintable"] = true

state.var {
  _minter = state.map(),       -- address -> boolean
  _max_supply = state.value()  -- unsigned_bignum
}

-- Set the Maximum Supply of tokens
-- @type    internal
-- @param   amount   (ubig) amount of mintable tokens
local function _setMaxSupply(amount)
  _typecheck(amount, 'ubig')
  _max_supply:set(amount)
end

-- Indicate if an account is a minter
-- @type    query
-- @param   account  (address)
-- @return  (bool) true/false
function isMinter(account)
  _typecheck(account, 'address')
  return (account == _contract_owner:get()) or (_minter[account] == true)
end

-- Add an account to minters
-- @type    call
-- @param   account  (address)
-- @event   addMinter(account)
function addMinter(account)
  _typecheck(account, 'address')

  assert(system.getSender() == _contract_owner:get(), "ARC1: only the contract owner can add a minter")

  _minter[account] = true

  contract.event("addMinter", account)
end

-- Remove an account from minters
-- @type    call
-- @param   account  (address)
-- @event   removeMinter(account)
function removeMinter(account)
  _typecheck(account, 'address')

  local contract_owner = _contract_owner:get()
  assert(system.getSender() == contract_owner, "ARC1: only the contract owner can remove a minter")
  assert(account ~= contract_owner, "ARC1: the contract owner is always a minter")
  assert(isMinter(account), "ARC1: not a minter")

  _minter:delete(account)

  contract.event("removeMinter", account)
end

-- Renounce the Minter Role
-- @type    call
-- @event   removeMinter(account)
function renounceMinter()
  local sender = system.getSender()
  assert(sender ~= _contract_owner:get(), "ARC1: contract owner can't renounce minter role")
  assert(isMinter(sender), "ARC1: only minter can renounce minter role")

  _minter:delete(sender)

  contract.event("removeMinter", sender)
end

-- Mint new tokens at an account
-- @type    call
-- @param   account  (address) recipient's address
-- @param   amount   (ubig) amount of tokens to mint
-- @param   ...      additional data, is sent unaltered in call to 'tokensReceived' on 'to'
-- @return  value returned from 'tokensReceived' callback, or nil
-- @event   mint(account, amount) 
function mint(account, amount, ...)
  _typecheck(account, 'address')
  amount = _check_bignum(amount)

  assert(isMinter(system.getSender()), "ARC1: only minter can mint")
  assert(not _max_supply:get() or (_totalSupply:get()+amount) <= _max_supply:get(), "ARC1: totalSupply is over MaxSupply")

  contract.event("mint", account, bignum.tostring(amount))

  return _mint(account, amount, ...)
end

-- Return the Max Supply
-- @type    query
-- @return  amount   (ubig) amount of tokens to mint
function maxSupply()
  return _max_supply:get() or bignum.number(0)
end

-- register the exported functions
abi.register(mint, addMinter, removeMinter, renounceMinter)
abi.register_view(isMinter, maxSupply)
]]

arc1_pausable = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- Pausable
------------------------------------------------------------------------------

extensions["pausable"] = true

state.var {
  _pauser = state.map()    -- address -> boolean
}

-- Indicate whether an account has the pauser role
-- @type    query
-- @param   account  (address)
-- @return  (bool) true/false
function isPauser(account)
  _typecheck(account, 'address')
  return (account == _contract_owner:get()) or (_pauser[account] == true)
end

-- Grant the pauser role to an account
-- @type    call
-- @param   account  (address)
-- @event   addPauser(account)
function addPauser(account)
  _typecheck(account, 'address')

  assert(system.getSender() == _contract_owner:get(), "ARC1: only contract owner can grant pauser role")

  _pauser[account] = true

  contract.event("addPauser", account)
end

-- Removes the pauser role from an account
-- @type    call
-- @param   account  (address)
-- @event   removePauser(account)
function removePauser(account)
  _typecheck(account, 'address')

  assert(system.getSender() == _contract_owner:get(), "ARC1: only owner can remove pauser role")
  assert(_pauser[account] == true, "ARC1: account does not have pauser role")

  _pauser[account] = nil

  contract.event("removePauser", account)
end

-- Renounce the granted pauser role
-- @type    call
-- @event   removePauser(account)
function renouncePauser()
  local sender = system.getSender()
  assert(sender ~= _contract_owner:get(), "ARC1: owner can't renounce pauser role")
  assert(_pauser[sender] == true, "ARC1: account does not have pauser role")

  _pauser[sender] = nil

  contract.event("removePauser", sender)
end

-- Indicate if the contract is paused
-- @type    query
-- @return  (bool) true/false
function paused()
  return _paused:get()
end

-- Put the contract in a paused state
-- @type    call
-- @event   pause(caller)
function pause()
  local sender = system.getSender()
  assert(not _paused:get(), "ARC1: contract is paused")
  assert(isPauser(sender), "ARC1: only pauser can pause")

  _paused:set(true)

  contract.event("pause", sender)
end

-- Return the contract to the normal state
-- @type    call
-- @event   unpause(caller)
function unpause()
  local sender = system.getSender()
  assert(_paused:get(), "ARC1: contract is unpaused")
  assert(isPauser(sender), "ARC1: only pauser can unpause")

  _paused:set(false)

  contract.event("unpause", sender)
end

-- register the exported functions
abi.register(pause, unpause, removePauser, renouncePauser, addPauser)
abi.register_view(paused, isPauser)
]]

arc1_blacklist = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- Blacklist
------------------------------------------------------------------------------

extensions["blacklist"] = true

-- Add accounts to the blacklist
-- @type    call
-- @param   account_list    (list of address)
-- @event   addToBlacklist(account_list)
function addToBlacklist(account_list)
  assert(system.getSender() == _contract_owner:get(), "ARC1: only owner can blacklist accounts")

  for i = 1, #account_list do
    _typecheck(account_list[i], 'address')
    _blacklist[account_list[i] ] = true
  end

  contract.event("addToBlacklist", account_list)
end

-- Removes accounts from the blacklist
-- @type    call
-- @param   account_list    (list of address)
-- @event   removeFromBlacklist(account_list)
function removeFromBlacklist(account_list)
  assert(system.getSender() == _contract_owner:get(), "ARC1: only owner can blacklist anothers")

  for i = 1, #account_list do
    _typecheck(account_list[i], 'address')
    _blacklist[account_list[i] ] = nil
  end

  contract.event("removeFromBlacklist", account_list)
end

-- Return true when an account is on the blacklist
-- @type    query
-- @param   account   (address)
function isOnBlacklist(account)
  _typecheck(account, 'address')
  return _blacklist[account] == true
end

-- register the exported functions
abi.register(addToBlacklist, removeFromBlacklist)
abi.register_view(isOnBlacklist)
]]

arc1_all_approval = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- All approval
------------------------------------------------------------------------------

extensions["all_approval"] = true

state.var {
  _operators = state.map(),   -- address/address -> boolean
}

-- Indicate the allowance from an account to another
-- @type    query
-- @param   owner       (address) owner's address
-- @param   operator    (address) allowed address
-- @return  (bool) true/false
function isApprovedForAll(owner, operator)
  return _operators[owner .. "/" .. operator] == true
end

-- Allow an account to use all caller's tokens
-- @type    call
-- @param   operator  (address) operator's address
-- @param   approved  (boolean) true/false
-- @event   setApprovalForAll(caller, operator, approved)
function setApprovalForAll(operator, approved)
  _typecheck(operator, 'address')
  _typecheck(approved, 'boolean')

  local sender = system.getSender()

  assert(sender ~= operator, "ARC1: cannot approve self as operator")

  if approved then
    _operators[sender .. "/" .. operator] = true
  else
    _operators[sender .. "/" .. operator] = nil
  end

  contract.event("setApprovalForAll", sender, operator, approved)
end

-- Transfer tokens from an account to another, the caller must be an approved operator
-- @type    call
-- @param   from    (address) sender's address
-- @param   to      (address) recipient's address
-- @param   amount  (ubig)    amount of tokens to send
-- @param   ...     additional data, is sent unaltered in call to 'tokensReceived' on 'to'
-- @return  value returned from 'tokensReceived' callback, or nil
-- @event   transfer(from, to, amount, operator)
function transferFrom(from, to, amount, ...)
  _typecheck(from, 'address')
  _typecheck(to, 'address')
  amount = _check_bignum(amount)

  local operator = system.getSender()

  assert(operator ~= from, "ARC1: use the transfer function")
  assert(isApprovedForAll(from, operator), "ARC1: caller is not approved by holder")

  contract.event("transfer", from, to, bignum.tostring(amount), operator)

  return _transfer(from, to, amount, ...)
end

-- register the exported functions
abi.register(setApprovalForAll, transferFrom)
abi.register_view(isApprovedForAll)


if extensions["burnable"] == true then

-- Burn tokens from an account, the caller must be an approved operator
-- @type    call
-- @param   from    (address) sender's address
-- @param   amount  (ubig)    amount of tokens to send
-- @event   burn(from, amount, operator)
function burnFrom(from, amount)
  _typecheck(from, 'address')
  amount = _check_bignum(amount)

  assert(extensions["burnable"], "ARC1: burnable extension not available")

  local operator = system.getSender()

  assert(operator ~= from, "ARC1: use the burn function")
  assert(isApprovedForAll(from, operator), "ARC1: caller not approved by holder")

  contract.event("burn", from, bignum.tostring(amount), operator)

  _burn(from, amount)
end

abi.register(burnFrom)

end
]]

arc1_limited_approval = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- Allowed approval
------------------------------------------------------------------------------

extensions["limited_approval"] = true

state.var {
  _allowance = state.map(),   -- address/address -> unsigned_bignum
}

-- Approve an account to spend the specified amount of caller's tokens
-- @type    call
-- @param   operator (address) operator's address
-- @param   amount   (ubig)    amount of allowed tokens
-- @event   approve(caller, operator, amount)
function approve(operator, amount)
  _typecheck(operator, 'address')
  amount = _check_bignum(amount)

  local owner = system.getSender()

  assert(owner ~= operator, "ARC1: cannot approve self as operator")

  _allowance[owner .. "/" .. operator] = amount

  contract.event("approve", owner, operator, bignum.tostring(amount))
end

-- Increase the amount of tokens allowed to an operator
-- @type    call
-- @param   operator (address) operator's address
-- @param   amount   (ubig)    amount to increase
-- @event   increaseAllowance(caller, operator, amount)
function increaseAllowance(operator, amount)
  _typecheck(operator, 'address')
  amount = _check_bignum(amount)

  local owner = system.getSender()
  local pair = owner .. "/" .. operator

  assert(_allowance[pair], "ARC1: not approved")

  _allowance[pair] = _allowance[pair] + amount

  contract.event("increaseAllowance", owner, operator, bignum.tostring(amount))
end

-- Decrease the amount of tokens allowed to an operator
-- @type    call
-- @param   operator (address) operator's address
-- @param   amount   (ubig)    amount of allowance to decrease
-- @event   decreaseAllowance(caller, operator, amount)
function decreaseAllowance(operator, amount)
  _typecheck(operator, 'address')
  amount = _check_bignum(amount)

  local owner = system.getSender()
  local pair = owner .. "/" .. operator

  assert(_allowance[pair], "ARC1: not approved")

  if _allowance[pair] < amount then
    _allowance[pair] = 0
  else
    _allowance[pair] = _allowance[pair] - amount
  end

  contract.event("decreaseAllowance", owner, operator, bignum.tostring(amount))
end

-- Get amount of remaining tokens that an account allowed to another
-- @type    query
-- @param   owner    (address) owner's address
-- @param   operator (address) operator's address
-- @return  (number) amount of remaining tokens
function allowance(owner, operator)
  _typecheck(owner, 'address')
  _typecheck(operator, 'address')

  return _allowance[owner .."/".. operator] or bignum.number(0)
end

-- Transfer tokens from an account to another using the allowance mechanism
-- @type    call
-- @param   from   (address) sender's address
-- @param   to     (address) recipient's address
-- @param   amount (ubig)    amount of tokens to send
-- @param   ...    additional data, is sent unaltered in call to 'tokensReceived' on 'to'
-- @return  value returned from 'tokensReceived' callback, or nil
-- @event   transfer(from, to, amount, operator)
function limitedTransferFrom(from, to, amount, ...)
  _typecheck(from, 'address')
  _typecheck(to, 'address')
  amount = _check_bignum(amount)

  local operator = system.getSender()

  assert(operator ~= from, "ARC1: use the transfer function")

  local pair = from .. "/" .. operator

  assert(_allowance[pair], "ARC1: not approved")
  assert(_allowance[pair] >= amount, "ARC1: insufficient allowance")

  _allowance[pair] = _allowance[pair] - amount

  contract.event("transfer", from, to, bignum.tostring(amount), operator)

  return _transfer(from, to, amount, ...)
end

-- register the exported functions
abi.register(approve, increaseAllowance, decreaseAllowance, limitedTransferFrom)
abi.register_view(allowance)


if extensions["burnable"] == true then

-- Burn tokens from an account using the allowance mechanism
-- @type    call
-- @param   from    (address) sender's address
-- @param   amount  (ubig)    amount of tokens to burn
-- @event   burn(from, amount, operator)
function limitedBurnFrom(from, amount)
  _typecheck(from, 'address')
  amount = _check_bignum(amount)

  assert(extensions["burnable"], "ARC1: burnable extension not available")

  local operator = system.getSender()

  assert(operator ~= from, "ARC1: use the burn function")

  local pair = from .. "/" .. operator

  assert(_allowance[pair], "ARC1: caller not approved by holder")
  assert(_allowance[pair] >= amount, "ARC1: insufficient allowance")

  _allowance[pair] = _allowance[pair] - amount

  contract.event("burn", from, bignum.tostring(amount), operator)

  _burn(from, amount)
end

abi.register(limitedBurnFrom)

end
]]

arc1_constructor = [[
function constructor(name, symbol, decimals, initial_supply, max_supply, owner)
  _init(name, symbol, decimals, owner)
  local decimal_str = "1" .. string.rep("0", decimals)
  if initial_supply > bignum.number(0) then
    _mint(owner, initial_supply * bignum.number(decimal_str))
  end
  if max_supply then
    _setMaxSupply(max_supply * bignum.number(decimal_str))
  end
end
]]

function new_token(name, symbol, decimals, initial_supply, options, owner)

  if options == nil or options == '' then
    options = {}
  end

  if owner == nil or owner == '' then
    owner = system.getSender()
  end

  local contract_code = arc1_core

  if options["burnable"] then
    contract_code = contract_code .. arc1_burnable
  end
  if options["mintable"] then
    contract_code = contract_code .. arc1_mintable
  end
  if options["pausable"] then
    contract_code = contract_code .. arc1_pausable
  end
  if options["blacklist"] then
    contract_code = contract_code .. arc1_blacklist
  end
  if options["all_approval"] then
    contract_code = contract_code .. arc1_all_approval
  end
  if options["limited_approval"] then
    contract_code = contract_code .. arc1_limited_approval
  end

  contract_code = contract_code .. arc1_constructor

  if not bignum.isbignum(initial_supply) then
    initial_supply = bignum.number(initial_supply)
  end
  assert(initial_supply >= bignum.number(0), "invalid initial supply")
  local max_supply = options["max_supply"]
  if max_supply then
    assert(options["mintable"], "max_supply is only available with the mintable extension")
    max_supply = bignum.number(max_supply)
    assert(max_supply >= initial_supply, "invalid max supply")
  end

  local address = contract.deploy(contract_code, name, symbol, decimals, initial_supply, max_supply, owner)

  contract.event("new_arc1_token", address)

  return address
end

abi.register(new_token)
