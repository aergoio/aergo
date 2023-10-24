arc2_core = [[
------------------------------------------------------------------------------
-- Aergo Standard NFT Interface (Proposal) - 20210425
------------------------------------------------------------------------------

extensions = {}

address0 = '1111111111111111111111111111111111111111111111111111'

-- A internal type check function
-- @type internal
-- @param x variable to check
-- @param t (string) expected type
local function _typecheck(x, t)
  if (x and t == 'address') then
    assert(type(x) == 'string', "address must be string type")
    -- check address length
    assert(52 == #x, string.format("invalid address length: %s (%s)", x, #x))
    -- check address checksum
    if x ~= address0 then
      local success = pcall(system.isContract, x)
      assert(success, "invalid address: " .. x)
    end
  elseif (x and t == 'str128') then
    assert(type(x) == 'string', "str128 must be string type")
    -- check address length
    assert(#x <= 128, string.format("too long str128 length: %s", #x))
  elseif (x and t == 'uint') then
    -- check unsigned integer
    assert(type(x) == 'number', string.format("invalid type: %s != number", type(x)))
    assert(math.floor(x) == x, "the number must be an integer")
    assert(x >= 0, "the number must be 0 or positive")
  else
    -- check default lua types
    assert(type(x) == t, string.format("invalid type: %s != %s", type(x), t or 'nil'))
  end
end


state.var {
  _contract_owner = state.value(),  -- string

  _name = state.value(),            -- string
  _symbol = state.value(),          -- string

  _num_burned = state.value(),      -- integer
  _last_index = state.value(),      -- integer
  _ids = state.map(),               -- integer -> str128
  _tokens = state.map(),            -- str128 -> { index: integer, owner: address, approved: address }
  _user_tokens = state.map(),       -- address -> array of integers (index to tokenId)

  -- Pausable
  _paused = state.value(),          -- boolean

  -- Blacklist
  _blacklist = state.map()          -- address -> boolean
}


-- call this at constructor
local function _init(name, symbol, owner)
  _typecheck(name, 'string')
  _typecheck(symbol, 'string')

  if owner == nil or owner == '' then
    owner = system.getCreator()
  elseif owner == 'none' then
    owner = nil
  else
    _typecheck(owner, "address")
  end
  _contract_owner:set(owner)

  _name:set(name)
  _symbol:set(symbol)

  _last_index:set(0)
  _num_burned:set(0)

  _paused:set(false)
end

local function _callReceiverCallback(from, to, tokenId, ...)
  if to ~= address0 and system.isContract(to) then
    return contract.call(to, "nonFungibleReceived", system.getSender(), from, tokenId, ...)
  else
    return nil
  end
end

local function _exists(tokenId)
  return _tokens[tokenId] ~= nil
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

-- Count of all NFTs
-- @type    query
-- @return  (integer) the number of non-fungible tokens on this contract
function totalSupply()
  return _last_index:get() - _num_burned:get()
end

-- Count of all NFTs assigned to an owner
-- @type    query
-- @param   owner  (address) a target address
-- @return  (integer) the number of non-fungible tokens of owner
function balanceOf(owner)
  local list = _user_tokens[owner] or {}
  return #list
end

-- Find the owner of an NFT
-- @type    query
-- @param   tokenId  (str128) the non-fungible token id
-- @return  (address) the address of the owner, or nil if the token does not exist
function ownerOf(tokenId)
  local token = _tokens[tokenId]
  if token == nil then
    return nil
  else
    return token["owner"]
  end
end


local function add_to_owner(index, owner)
  local list = _user_tokens[owner] or {}
  table.insert(list, index)
  _user_tokens[owner] = list
end

local function remove_from_owner(index, owner)
  local list = _user_tokens[owner] or {}
  for position,value in ipairs(list) do
    if value == index then
      table.remove(list, position)
      break
    end
  end
  if #list > 0 then
    _user_tokens[owner] = list
  else
    _user_tokens:delete(owner)
  end
end


local function _mint(to, tokenId, metadata, ...)
  _typecheck(to, 'address')
  _typecheck(tokenId, 'str128')

  assert(not _paused:get(), "ARC2: paused contract")
  assert(not _blacklist[to], "ARC2: recipient is on blacklist")

  assert(not _exists(tokenId), "ARC2: mint - already minted token")
  assert(metadata==nil or type(metadata)=="table", "ARC2: invalid metadata")

  local index = _last_index:get() + 1
  _last_index:set(index)
  _ids[tostring(index)] = tokenId

  local token = {
    index = index,
    owner = to
  }
  if metadata ~= nil then
    assert(extensions["metadata"], "ARC2: this token has no support for metadata")
    for key,value in pairs(metadata) do
      assert(not is_reserved_metadata(key), "ARC2: reserved metadata")
      token[key] = value
    end
  end
  _tokens[tokenId] = token

  add_to_owner(index, to)

  contract.event("mint", to, tokenId)

  return _callReceiverCallback(nil, to, tokenId, ...)
end


local function _burn(tokenId)
  _typecheck(tokenId, 'str128')

  local token = _tokens[tokenId]
  assert(token ~= nil, "ARC2: burn: token not found")
  local index = token["index"]
  local owner = token["owner"]

  assert(not _paused:get(), "ARC2: paused contract")
  assert(not _blacklist[owner], "ARC2: owner is on blacklist")

  _ids:delete(tostring(index))

  _tokens:delete(tokenId)

  remove_from_owner(index, owner)

  _num_burned:set(_num_burned:get() + 1)

  contract.event("burn", owner, tokenId)
end


local function _transfer(from, to, tokenId, ...)
  assert(not _paused:get(), "ARC2: paused contract")
  assert(not _blacklist[from], "ARC2: sender is on blacklist")
  assert(not _blacklist[to], "ARC2: recipient is on blacklist")

  local token = _tokens[tokenId]
  token["owner"] = to
  token["approved"] = nil   -- clear approval
  _tokens[tokenId] = token

  local index = token["index"]
  remove_from_owner(index, from)
  add_to_owner(index, to)

  return _callReceiverCallback(from, to, tokenId, ...)
end


-- Transfer a token
-- @type    call
-- @param   to      (address) the receiver address
-- @param   tokenId (str128) the NFT token to send
-- @param   ...     (Optional) additional data, is sent unaltered in a call to 'nonFungibleReceived' on 'to'
-- @return  value returned from the 'nonFungibleReceived' callback, or nil
-- @event   transfer(from, to, tokenId)
function transfer(to, tokenId, ...)
  _typecheck(to, 'address')
  _typecheck(tokenId, 'str128')

  local token = _tokens[tokenId]
  assert(token ~= nil, "ARC2: transfer - nonexisting token")

  assert(extensions["non_transferable"] == nil and
              token["non_transferable"] == nil, "ARC2: this token is non-transferable")

  local sender = system.getSender()
  local owner = token["owner"]
  assert(sender == owner, "ARC2: transfer of token that is not own")

  contract.event("transfer", sender, to, tokenId)

  return _transfer(sender, to, tokenId, ...)
end

-- Transfer a non-fungible token of 'from' to 'to'
-- @type    call
-- @param   from    (address) the owner address
-- @param   to      (address) the receiver address
-- @param   tokenId (str128) the non-fungible token to send
-- @param   ...     (Optional) additional data, is sent unaltered in a call to 'nonFungibleReceived' on 'to'
-- @return  value returned from the 'nonFungibleReceived' callback, or nil
-- @event   transfer(from, to, tokenId, operator)
function transferFrom(from, to, tokenId, ...)
  _typecheck(from, 'address')
  _typecheck(to, 'address')
  _typecheck(tokenId, 'str128')

  local token = _tokens[tokenId]
  assert(token ~= nil, "ARC2: transferFrom - nonexisting token")

  local owner = token["owner"]
  assert(from == owner, "ARC2: transferFrom - token is not from account")

  local operator = system.getSender()

  -- if recallable, the creator/issuer can transfer the token
  if extensions["mintable"] then
    operator_can_recall = isMinter(operator)
  else
    operator_can_recall = (operator == _contract_owner:get())
  end
  local is_recall = (extensions["recallable"] or token["recallable"]) and operator_can_recall

  if not is_recall then
    assert(extensions["approval"], "ARC2: approval extension not included")
    -- check allowance
    assert(operator == token["approved"] or isApprovedForAll(owner, operator),
           "ARC2: transferFrom - caller is not approved")
    -- check if it is a non-transferable token
    assert(extensions["non_transferable"] == nil and
                token["non_transferable"] == nil, "ARC2: this token is non-transferable")
  end

  contract.event("transfer", from, to, tokenId, operator)

  return _transfer(from, to, tokenId, ...)
end


-- Token Enumeration Functions --


-- Retrieves the next token in the contract
-- @type    query
-- @param   prev_index (integer) the index of the previous returned token. use `0` in the first call
-- @return  (index, tokenId) the index of the next token and its token id, or `nil,nil` if no more tokens
function nextToken(prev_index)
  _typecheck(prev_index, 'uint')

  local index = prev_index
  local last_index = _last_index:get()
  local tokenId

  while tokenId == nil and index < last_index do
    index = index + 1
    tokenId = _ids[tostring(index)]
  end

  if tokenId == nil then
    index = nil
  end

  return index, tokenId
end

-- Retrieves the token from the given user at the given position
-- @type    query
-- @param   user      (address) ..
-- @param   position  (integer) the position of the token in the incremental sequence
-- @return  tokenId   (str128) the token id, or `nil` if no more tokens on this account
function tokenFromUser(user, position)
  _typecheck(user, 'address')
  _typecheck(position, 'uint')

  local list = _user_tokens[user] or {}
  local index = list[position]
  local tokenId = _ids[tostring(index)]
  return tokenId
end


function set_contract_owner(address)
  assert(system.getSender() == _contract_owner:get(), "ARC2: permission denied")
  _typecheck(address, "address")
  _contract_owner:set(address)
end


-- Returns a JSON string containing the list of ARC2 extensions
-- that were included on the contract.
-- @type    query
function arc2_extensions()
  local list = {}
  for name,_ in pairs(extensions) do
    table.insert(list, name)
  end
  return list
end


abi.register(transfer, transferFrom, set_contract_owner)
abi.register_view(name, symbol, balanceOf, ownerOf, totalSupply, nextToken, tokenFromUser, arc2_extensions)
]]

arc2_burnable = [[
------------------------------------------------------------------------------
-- Aergo Standard NFT Interface (Proposal) - 20210425
------------------------------------------------------------------------------

extensions["burnable"] = true

-- Burn a non-fungible token
-- @type    call
-- @param   tokenId  (str128) the identifier of the token to be burned
-- @event   burn(owner, tokenId)
function burn(tokenId)
  _typecheck(tokenId, 'str128')

  local owner = ownerOf(tokenId)
  assert(owner ~= nil, "ARC2: burn - nonexisting token")
  assert(system.getSender() == owner, "ARC2: cannot burn a token that is not own")

  _burn(tokenId)
end

abi.register(burn)
]]

arc2_mintable = [[
------------------------------------------------------------------------------
-- Aergo Standard NFT Interface (Proposal) - 20210425
------------------------------------------------------------------------------

extensions["mintable"] = true

state.var {
  _minter = state.map(),       -- address -> boolean
  _max_supply = state.value()  -- integer
}

-- set Max Supply
-- @type    internal
-- @param   amount   (integer) amount of mintable tokens

local function _setMaxSupply(amount)
  _typecheck(amount, 'uint')
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

-- Add an account to the minters group
-- @type    call
-- @param   account  (address)
-- @event   addMinter(account)

function addMinter(account)
  _typecheck(account, 'address')

  local creator = _contract_owner:get()
  assert(system.getSender() == creator, "ARC2: only the contract owner can add a minter")
  assert(account ~= creator, "ARC2: the contract owner is always a minter")

  _minter[account] = true

  contract.event("addMinter", account)
end

-- Remove an account from the minters group
-- @type    call
-- @param   account  (address)
-- @event   removeMinter(account)

function removeMinter(account)
  _typecheck(account, 'address')

  local creator = _contract_owner:get()
  assert(system.getSender() == creator, "ARC2: only the contract owner can remove a minter")
  assert(account ~= creator, "ARC2: the contract owner is always a minter")
  assert(isMinter(account), "ARC2: not a minter")

  _minter:delete(account)

  contract.event("removeMinter", account)
end

-- Renounce the minter role
-- @type    call
-- @event   removeMinter(account)

function renounceMinter()
  local sender = system.getSender()
  assert(sender ~= _contract_owner:get(), "ARC2: contract owner can't renounce minter role")
  assert(isMinter(sender), "ARC2: only minter can renounce minter role")

  _minter:delete(sender)

  contract.event("removeMinter", sender)
end

-- Mint a new non-fungible token
-- @type    call
-- @param   to       (address) recipient's address
-- @param   tokenId  (str128) the non-fungible token id
-- @param   metadata (table) lua table containing key-value pairs
-- @param   ...      additional data, is sent unaltered in a call to 'nonFungibleReceived' on 'to'
-- @return  value returned from the 'nonFungibleReceived' callback, or nil
-- @event   mint(to, tokenId)

function mint(to, tokenId, metadata, ...)
  assert(isMinter(system.getSender()), "ARC2: only minter can mint")
  local max_supply = _max_supply:get()
  assert(not max_supply or (totalSupply() + 1) <= max_supply, "ARC2: TotalSupply is over MaxSupply")

  return _mint(to, tokenId, metadata, ...)
end

-- Retrieve the Max Supply
-- @type    query
-- @return  amount   (integer) the maximum amount of tokens that can be active on the contract

function maxSupply()
  return _max_supply:get() or 0
end

abi.register(mint, addMinter, removeMinter, renounceMinter)
abi.register_view(isMinter, maxSupply)
]]

arc2_metadata = [[
------------------------------------------------------------------------------
-- Aergo Standard NFT Interface (Proposal) - 20210425
------------------------------------------------------------------------------

extensions["metadata"] = true

state.var {
  _immutable_metadata = state.value(),
  _incremental_metadata = state.value(),
  _contract_metadata = state.value()
}

reserved_metadata = { "index", "owner", "approved" }

function is_reserved_metadata(name)
  for _,reserved in ipairs(reserved_metadata) do
    if name == reserved then
      return true
    end
  end
  return false
end

local function check_metadata_update(name, prev_value, new_value)

  assert(not is_reserved_metadata(name), "ARC2: reserved metadata")

  local immutable = _immutable_metadata:get() or {}
  local incremental = _incremental_metadata:get() or {}

  for _,value in ipairs(immutable) do
    if value == name then
      assert(false, "ARC2: immutable metadata")
    end
  end
  for _,value in ipairs(incremental) do
    if value == name then
      assert(new_value ~= nil and type(new_value) == type(prev_value) and
             new_value >= prev_value, "ARC2: incremental metadata")
      break
    end
  end

end

--- Exported Functions ---------------------------------------------------------

-- Store non-fungible token metadata
-- @type    call
-- @param   tokenId  (str128) the non-fungible token id, or nil for contract metadata
-- @param   metadata (table)  lua table containing key-value pairs
function set_metadata(tokenId, metadata)

  if extensions["mintable"] then
    assert(isMinter(system.getSender()), "ARC2: permission denied")
  else
    assert(system.getSender() == _contract_owner:get(), "ARC2: permission denied")
  end
  assert(not _paused:get(), "ARC2: paused contract")

  if tokenId == nil then
    local contract_metadata = _contract_metadata:get() or {}
    for key,value in pairs(metadata) do
      contract_metadata[key] = value
    end
    _contract_metadata:set(contract_metadata)
    return
  end

  _typecheck(tokenId, 'str128')

  local token = _tokens[tokenId]
  assert(token ~= nil, "ARC2: nonexisting token")
  for key,value in pairs(metadata) do
    check_metadata_update(key, token[key], value)
    assert(key ~= "non_transferable" and key ~= "recallable", "ARC2: permission denied")
    token[key] = value
  end
  _tokens[tokenId] = token
end

-- Remove non-fungible token metadata
-- @type    call
-- @param   tokenId  (str128) the non-fungible token id
-- @param   list     (table)  lua table containing list of keys to remove
function remove_metadata(tokenId, list)

  if extensions["mintable"] then
    assert(isMinter(system.getSender()), "ARC2: permission denied")
  else
    assert(system.getSender() == _contract_owner:get(), "ARC2: permission denied")
  end
  assert(not _paused:get(), "ARC2: paused contract")

  _typecheck(tokenId, 'str128')

  local token = _tokens[tokenId]
  assert(token ~= nil, "ARC2: nonexisting token")
  for _,key in ipairs(list) do
    check_metadata_update(key, token[key], nil)
    token[key] = nil
  end
  _tokens[tokenId] = token
end

-- Retrieve non-fungible token metadata
-- @type    query
-- @param   tokenId  (str128) the non-fungible token id, or nil for contract metadata
-- @param   key      (string) the metadata key
-- @return  (string) if key is nil, return all metadata from token or contract,
--                   otherwise return the value linked to the given key
function get_metadata(tokenId, key)

  if tokenId == nil then
    local contract_metadata = _contract_metadata:get() or {}
    if key ~= nil then
      return contract_metadata[key]
    end
    return contract_metadata
  end

  _typecheck(tokenId, 'str128')

  local token = _tokens[tokenId]
  assert(token ~= nil, "ARC2: nonexisting token")

  -- token["index"] = nil
  -- token["owner"] = nil
  -- token["approved"] = nil

  if key == nil then
    return token
  end

  return token[key]
end

-- Mark a specific metadata key as immutable. This means that once this metadata
-- is set on a token, it can no longer be modified. And once this property is set
-- on a metadata, it cannot be removed. It gives the guarantee to the owner that
-- the creator/issuer will not modify or remove this specific metadata.
-- @type    call
-- @param   key  (string) the metadata key
function make_metadata_immutable(key)
  _typecheck(key, 'string')
  assert(#key > 0, "ARC2: invalid key")
  assert(not is_reserved_metadata(key), "ARC2: reserved metadata")
  assert(system.getSender() == _contract_owner:get(), "ARC2: permission denied")
  assert(not _paused:get(), "ARC2: paused contract")

  local immutable = _immutable_metadata:get() or {}

  for _,value in ipairs(immutable) do
    if value == key then return end
  end

  table.insert(immutable, key)
  _immutable_metadata:set(immutable)
end

-- Mark a specific metadata key as incremental. This means that once this metadata
-- is set on a token, it can only be incremented. Useful for expiration time.
-- Once this property is set on a metadata, it cannot be removed. It gives the
-- guarantee to the owner that the creator/issuer can only increment this value.
-- @type    call
-- @param   key  (string) the metadata key
function make_metadata_incremental(key)
  _typecheck(key, 'string')
  assert(#key > 0, "ARC2: invalid key")
  assert(not is_reserved_metadata(key), "ARC2: reserved metadata")
  assert(system.getSender() == _contract_owner:get(), "ARC2: permission denied")
  assert(not _paused:get(), "ARC2: paused contract")

  local incremental = _incremental_metadata:get() or {}

  for _,value in ipairs(incremental) do
    if value == key then return end
  end

  table.insert(incremental, key)
  _incremental_metadata:set(incremental)
end

-- Retrieve the list of immutable and incremental metadata
-- @type    query
-- @return  (string) a JSON object with each metadata as key and the property
--                   as value. Example: {"expiration": "incremental", "index": "immutable"}
function get_metadata_info()
  local immutable = _immutable_metadata:get() or {}
  local incremental = _incremental_metadata:get() or {}

  local list = {}

  list["index"] = "immutable"

  for _,value in ipairs(immutable) do
    list[value] = "immutable"
  end
  for _,value in ipairs(incremental) do
    list[value] = "incremental"
  end

  return list
end


abi.register_view(get_metadata, get_metadata_info)
abi.register(set_metadata, remove_metadata, make_metadata_immutable, make_metadata_incremental)
]]

arc2_pausable = [[
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

  assert(system.getSender() == _contract_owner:get(), "ARC2: only contract owner can grant pauser role")

  _pauser[account] = true

  contract.event("addPauser", account)
end

-- Removes the pauser role from an account
-- @type    call
-- @param   account  (address)
-- @event   removePauser(account)

function removePauser(account)
  _typecheck(account, 'address')

  assert(system.getSender() == _contract_owner:get(), "ARC2: only owner can remove pauser role")
  assert(_pauser[account] == true, "ARC2: account does not have pauser role")

  _pauser[account] = nil

  contract.event("removePauser", account)
end

-- Renounce the granted pauser role
-- @type    call
-- @event   removePauser(account)

function renouncePauser()
  local sender = system.getSender()
  assert(sender ~= _contract_owner:get(), "ARC2: owner can't renounce pauser role")
  assert(_pauser[sender] == true, "ARC2: account does not have pauser role")

  _pauser[sender] = nil

  contract.event("removePauser", sender)
end

-- Indicate if the contract is paused
-- @type    query
-- @return  (bool) true/false

function paused()
  return (_paused:get() == true)
end

-- Put the contract in a paused state
-- @type    call
-- @event   pause(caller)

function pause()
  local sender = system.getSender()
  assert(not _paused:get(), "ARC2: contract is paused")
  assert(isPauser(sender), "ARC2: only pauser can pause")

  _paused:set(true)

  contract.event("pause", sender)
end

-- Return the contract to the normal state
-- @type    call
-- @event   unpause(caller)

function unpause()
  local sender = system.getSender()
  assert(_paused:get(), "ARC2: contract is unpaused")
  assert(isPauser(sender), "ARC2: only pauser can unpause")

  _paused:set(false)

  contract.event("unpause", sender)
end


abi.register(pause, unpause, removePauser, renouncePauser, addPauser)
abi.register_view(paused, isPauser)
]]

arc2_blacklist = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- Blacklist
------------------------------------------------------------------------------

extensions["blacklist"] = true

-- Add accounts to the blacklist
-- @type    call
-- @param   account_list (list of address)
-- @event   addToBlacklist(account_list)

function addToBlacklist(account_list)
  assert(system.getSender() == _contract_owner:get(), "ARC2: only owner can blacklist accounts")

  for i = 1, #account_list do
    _typecheck(account_list[i], 'address')
    _blacklist[account_list[i] ] = true
  end

  contract.event("addToBlacklist", account_list)
end

-- Remove accounts from the blacklist
-- @type    call
-- @param   account_list  (list of address)
-- @event   removeFromBlacklist(account_list)

function removeFromBlacklist(account_list)
  assert(system.getSender() == _contract_owner:get(), "ARC2: only owner can blacklist accounts")

  for i = 1, #account_list do
    _typecheck(account_list[i], 'address')
    _blacklist[account_list[i] ] = nil
  end

  contract.event("removeFromBlacklist", account_list)
end

-- Indicate if an account is on the blacklist
-- @type    query
-- @param   account   (address)
-- @return  (bool) true/false

function isOnBlacklist(account)
  _typecheck(account, 'address')

  return _blacklist[account] == true
end


abi.register(addToBlacklist, removeFromBlacklist)
abi.register_view(isOnBlacklist)
]]

arc2_approval = [[
------------------------------------------------------------------------------
-- Aergo Standard NFT Interface (Proposal) - 20210425
------------------------------------------------------------------------------

extensions["approval"] = true

state.var {
  _operatorApprovals = state.map()  -- address/address -> bool
}

-- Approve an account to operate on the given non-fungible token.
-- Use `nil` on the operator to remove the approval
-- @type    call
-- @param   operator    (address) the new approved NFT controller
-- @param   tokenId     (str128) the NFT token to be controlled
-- @event   approve(owner, operator, tokenId)
function approve(operator, tokenId)
  _typecheck(tokenId, 'str128')
  if operator ~= nil then
    _typecheck(operator, 'address')
  end

  local token = _tokens[tokenId]
  assert(token ~= nil, "ARC2: approve - nonexisting token")
  local owner = token["owner"]

  assert(not _paused:get(), "ARC2: paused contract")
  assert(not _blacklist[owner], "ARC2: owner is on blacklist")
  if operator ~= nil then
    assert(not _blacklist[operator], "ARC2: operator is on blacklist")
  end

  assert(owner ~= operator, "ARC2: approve - to current owner")
  local sender = system.getSender()
  assert(sender == owner or isApprovedForAll(owner, sender),
    "ARC2: approve - caller is not owner nor approved for all")

  token["approved"] = operator
  _tokens[tokenId] = token

  contract.event("approve", owner, operator, tokenId)
end

-- Get the approved operator address for a given non-fungible token
-- @type    query
-- @param   tokenId  (str128) the NFT token to find the approved operator
-- @return  (address) the approved operator address for this NFT, or nil if there is none
function getApproved(tokenId)
  _typecheck(tokenId, 'str128')
  local token = _tokens[tokenId]
  assert(token ~= nil, "ARC2: getApproved - nonexisting token")
  return token["approved"]
end

-- Allow an operator to control all the sender's tokens
-- @type    call
-- @param   operator  (address) the operator address
-- @param   approved  (boolean) true if the operator is approved, false to revoke approval
-- @event   approvalForAll(owner, operator, approved)
function setApprovalForAll(operator, approved)
  _typecheck(operator, 'address')
  _typecheck(approved, 'boolean')

  local owner = system.getSender()

  assert(not _paused:get(), "ARC2: paused contract")
  assert(not _blacklist[owner], "ARC2: owner is on blacklist")
  if approved then
    assert(not _blacklist[operator], "ARC2: operator is on blacklist")
  end

  assert(operator ~= owner, "ARC2: setApprovalForAll - to caller")

  _operatorApprovals[owner .. '/' .. operator] = approved

  contract.event("approvalForAll", owner, operator, approved)
end

-- Check if the given operator is approved to control the owner's tokens
-- @type    query
-- @param   owner       (address) owner address
-- @param   operator    (address) operator address
-- @return  (bool) true/false
function isApprovedForAll(owner, operator)
  return _operatorApprovals[owner .. '/' .. operator] or false
end


abi.register(approve, setApprovalForAll)
abi.register_view(getApproved, isApprovedForAll)
]]

arc2_searchable = [[
------------------------------------------------------------------------------
-- Aergo Standard NFT Interface (Proposal) - 20210425
------------------------------------------------------------------------------

extensions["searchable"] = true

local function escape(str)
  return str:gsub("[%(%)%.%%%+%-%*%?%[%]%^%$]", function(c) return "%" .. c end)
end

local function token_matches(tokenId, query)

  if tokenId == nil then
    return false
  end

  local token = _tokens[tokenId]

  local pattern = query["pattern"]
  local metadata = query["metadata"]

  if pattern then
    if not tokenId:match(pattern) then
      return false
    end
  end

  if metadata then
    for key,find in pairs(metadata) do
      local op = find["op"]
      local value = find["value"]
      local neg = false
      local matches = false
      if string.sub(op,1,1) == "!" then
        neg = true
        op = string.sub(op, 2)
      end
      local token_value = token[key]
      if token_value == nil and op ~= "=" then
        -- does not match
      elseif op == "=" then
        matches = token_value == value
      elseif op == ">" then
        matches = token_value > value
      elseif op == ">=" then
        matches = token_value >= value
      elseif op == "<" then
        matches = token_value < value
      elseif op == "<=" then
        matches = token_value <= value
      elseif op == "between" then
        matches = (token_value >= value and token_value <= find["value2"])
      elseif op == "match" then
        matches = string.match(token_value, value) ~= nil
      else
        assert(false, "operator not known: " .. op)
      end
      if neg then matches = not matches end
      if not matches then return false end
    end
  end

  return true
end

-- retrieve the first token found that mathes the query
-- the query is a table that can contain these fields:
--   owner    - the owner of the token (address)
--   contains - check if the tokenId contains this string
--   pattern  - check if the tokenId matches this Lua regex pattern
-- the prev_index must be 0 in the first call
-- for the next calls, just inform the returned index from the previous call
-- return value: (2 values) index, tokenId
-- if no token is found with the given query, it returns (nil, nil)
function findToken(query, prev_index)
  _typecheck(query, 'table')
  _typecheck(prev_index, 'uint')

  local contains = query["contains"]
  if contains then
    query["pattern"] = escape(contains)
  end

  local index, tokenId
  local owner = query["owner"]

  if owner then
    -- iterate over the tokens from this user
    local list = _user_tokens[owner] or {}
    local check_tokens = (prev_index == 0)

    for position,index2 in ipairs(list) do
      if check_tokens then
        tokenId = _ids[tostring(index2)]
        if token_matches(tokenId, query) then
          index = index2
          break
        else
          tokenId = nil
        end
      elseif index2 == prev_index then
        check_tokens = true
      end
    end

  else
    -- iterate over all the tokens
    local last_index = _last_index:get()
    index = prev_index

    while tokenId == nil and index < last_index do
      index = index + 1
      tokenId = _ids[tostring(index)]
      if not token_matches(tokenId, query) then
        tokenId = nil
      end
    end
  end

  if tokenId == nil then
    index = nil
  end

  return index, tokenId
end


abi.register_view(findToken)
]]

arc2_non_transferable = [[
------------------------------------------------------------------------------
-- Aergo Standard NFT Interface (Proposal) - 20210425
------------------------------------------------------------------------------

extensions["non_transferable"] = true
]]

arc2_recallable = [[
------------------------------------------------------------------------------
-- Aergo Standard NFT Interface (Proposal) - 20210425
------------------------------------------------------------------------------

extensions["recallable"] = true
]]

arc2_constructor = [[
function constructor(name, symbol, initial_supply, max_supply, owner)
  _init(name, symbol, owner)
  if initial_supply then
    for _,token in ipairs(initial_supply) do
      _mint(owner, token[1], token[2])
    end
  end
  if max_supply then
    _setMaxSupply(max_supply)
  end
end
]]

function new_arc2_nft(name, symbol, initial_supply, options, owner)

  if options == nil or options == '' then
    options = {}
  end

  if owner == nil or owner == '' then
    owner = system.getSender()
  end

  local contract_code = arc2_core

  if options["burnable"] then
    contract_code = contract_code .. arc2_burnable
  end
  if options["mintable"] then
    contract_code = contract_code .. arc2_mintable
  end
  if options["metadata"] then
    contract_code = contract_code .. arc2_metadata
  end
  if options["pausable"] then
    contract_code = contract_code .. arc2_pausable
  end
  if options["blacklist"] then
    contract_code = contract_code .. arc2_blacklist
  end
  if options["approval"] then
    contract_code = contract_code .. arc2_approval
  end
  if options["searchable"] then
    contract_code = contract_code .. arc2_searchable
  end
  if options["non_transferable"] then
    contract_code = contract_code .. arc2_non_transferable
  end
  if options["recallable"] then
    contract_code = contract_code .. arc2_recallable
  end

  contract_code = contract_code .. arc2_constructor

  if initial_supply then
    for index,value in ipairs(initial_supply) do
      if type(value) == "string" then
        initial_supply[index] = {value}
      end
    end
  end

  local max_supply = options["max_supply"]
  if max_supply then
    assert(options["mintable"], "max_supply is only available with the mintable extension")
    max_supply = tonumber(max_supply)
    if initial_supply then
      assert(max_supply >= #initial_supply, "the max supply must be bigger than the initial supply count")
    end
  end

  local address = contract.deploy(contract_code, name, symbol, initial_supply, max_supply, owner)

  contract.event("new_arc2_token", address)

  return address
end

abi.register(new_arc2_nft)
