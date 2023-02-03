package contract

import (
	"fmt"
	"math"
	"testing"

	"github.com/aergoio/aergo/config"
	"github.com/aergoio/aergo/types"
	"github.com/stretchr/testify/require"
)

// TODO: move to hardfork_test.go
func TestHardForkVersion(t *testing.T) {
	for _, test := range []struct {
		blockNumV2Fork  uint64
		blockNumV3Fork  uint64
		blockNumCurrent uint64
		expectVersion   int32
	}{
		{0, 0, 0, 3}, // version 3

		{0, 100, 0, 2},   // version 2
		{0, 100, 99, 2},  // version 2
		{0, 100, 100, 3}, // version 3

		{100, 200, 0, 0},   // version 0
		{100, 200, 99, 0},  // version 0
		{100, 200, 100, 2}, // version 2
		{100, 200, 199, 2}, // version 2
		{100, 200, 200, 3}, // version 3
	} {
		// set hardfork config
		HardforkConfig = &config.HardforkConfig{
			V2: test.blockNumV2Fork,
			V3: test.blockNumV3Fork,
		}
		// check fork version
		version := HardforkConfig.Version(test.blockNumCurrent)
		require.Equal(t, test.expectVersion, version, "failed to check fork version")
	}
}

// common test function for hardfork
func LoadDummyChainForked(forkVersion int, fn func(d *DummyChain) error) error {
	bc, err := LoadDummyChain(OnPubNet)
	if err != nil {
		return err
	}
	defer bc.Release()

	// set fork version
	switch forkVersion {
	case 0, 1:
		HardforkConfig = &config.HardforkConfig{
			V2: math.MaxUint64,
			V3: math.MaxUint64,
		}
	case 2:
		HardforkConfig = &config.HardforkConfig{
			V2: 0,
			V3: math.MaxUint64,
		}
	case 3:
		HardforkConfig = &config.HardforkConfig{
			V2: 0,
			V3: 0,
		}
	default:
		return fmt.Errorf("invalid fork version %d", forkVersion)
	}
	return fn(bc)
}

func TestMaxCallDepth(t *testing.T) {
	for _, test := range []struct {
		forkVersion        int
		expectMaxCallDepth int32
		isMaxCallExceed    bool
	}{
		// version 3
		{3, maxCallDepth, false},
		{3, maxCallDepth + 1, true},

		// version 2
		{2, maxCallDepthOld, false},
		{2, maxCallDepthOld + 1, true},

		// version 0
		{0, maxCallDepthOld, false},
		{0, maxCallDepthOld + 1, true},
	} {
		err := LoadDummyChainForked(test.forkVersion, func(bc *DummyChain) error {
			return bc.ConnectBlock(
				NewLuaTxAccount("kch", 1e18),
				NewLuaTxDef("kch", "make_call", 0, `
function make_call(remaining)
	remaining = remaining - 1
	if remaining >= 0 then
		contract.call(system.getContractID(), "make_call", remaining)
	end
end
abi.register(make_call)
	`),
				NewLuaTxCall("kch", "make_call", 0, fmt.Sprintf(`{"Name":"make_call", "Args":[%d]}`, test.expectMaxCallDepth)),
			)
		})
		require.EqualValues(t, test.isMaxCallExceed, err != nil, "failed to check max call depth")
	}
}

// TODO : listing failed case
func TestDiscardEventFailedTransaction(t *testing.T) {
	var contractSucceed = `
function make_call()
end
abi.register(make_call)
`
	var contractFailed = `
function make_call()
	contract.call(system.getContractID(), "make_call")
end
abi.register(make_call)
`

	for _, test := range []struct {
		forkVersion  int
		contract     string
		isEventFound bool
	}{
		// version 3
		{3, contractSucceed, true},
		{3, contractFailed, false},

		// version 2
		{2, contractSucceed, true},
		{2, contractFailed, true},

		// version 0
		{0, contractSucceed, true},
		{0, contractFailed, true},
	} {
		err := LoadDummyChainForked(test.forkVersion, func(bc *DummyChain) error {
			return bc.ConnectBlock(
				NewLuaTxAccount("kch", 1e18),
				NewLuaTxDef("kch", "make_call", 0, test.contract),
				NewLuaTxCall("kch", "make_call", 0, `{"Name":"make_call", "Args":[]}`),
			)
		})
		require.NoError(t, err, "failed to execute")
	}

}

var newTokenFactory = `
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

  if (x and t == 'address') then -- a string of alphanumeric char. except for '0, I, O, l'

    assert(type(x) == 'string', "ARC1: address must be string type")

    -- check address length
    assert(52 == #x, string.format("ARC1: invalid address length: %s (%s)", x, #x))

    -- check character
    local invalidChar = string.match(x, '[^123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]')
    assert(nil == invalidChar, string.format("ARC1: invalid address format: %s contains invalid char %s", x, invalidChar or 'nil'))

  elseif (x and t == 'ubig') then   -- a positive big number

    -- check unsigned bignum
    assert(bignum.isbignum(x), string.format("ARC1: invalid type: %s != %s", type(x), t))
    assert(x >= bignum.number(0), string.format("ARC1: %s must be positive number", bignum.tostring(x)))

  elseif (x and t == 'uint') then   -- a positive number

    assert(type(x) == 'number', string.format("ARC1: invalid type: %s != number", type(x)))
    assert(math.floor(x) == x, "ARC1: the number must be an integer")
    assert(x >= 0, "ARC1: the number must be 0 or positive")

  else
    -- check default lua types
    assert(type(x) == t, string.format("ARC1: invalid type: %s != %s", type(x), t or 'nil'))

  end
end

function _check_bignum(x)
  if type(x) == 'string' then
    assert(string.match(x, '[^0-9.]') == nil, "ARC1: amount contains invalid character")
    local _, count = string.gsub(x, "%.", "")
    assert(count <= 1, "ARC1: the amount is invalid")
    if count == 1 then
      local num_decimals = _decimals:get()
      local p1, p2 = string.match('0' .. x .. '0', '(%d+)%.(%d+)')
      local to_add = num_decimals - #p2
      if to_add > 0 then
        p2 = p2 .. string.rep('0', to_add)
      elseif to_add < 0 then
        p2 = string.sub(p2, 1, num_decimals)
      end
      x = p1 .. p2
      x = string.gsub(x, '0*', '', 1)  -- remove leading zeros
      if #x == 0 then x = '0' end
    end
    x = bignum.number(x)
  end
  _typecheck(x, 'ubig')
  return x
end

-- call this at constructor
-- @type internal
-- @param name (string) name of this token
-- @param symbol (string) symbol of this token
-- @param decimals (number) decimals of this token
-- @param owner (optional:address) the owner of this contract

local function _init(name, symbol, decimals, owner)

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

-- Get a token name
-- @type    query
-- @return  (string) name of this token

function name()
  return _name:get()
end


-- Get a token symbol
-- @type    query
-- @return  (string) symbol of this token

function symbol()
  return _symbol:get()
end


-- Get a token decimals
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


-- return total supply.
-- @type    query
-- @return  (ubig) total supply of this token

function totalSupply()
  return _totalSupply:get()
end


abi.register(set_metadata)
abi.register_view(name, symbol, decimals, get_metadata, totalSupply, balanceOf)

-- Hook "tokensReceived" function on the recipient after a 'transfer'
-- @type internal
-- @param   from   (address) sender's address
-- @param   to     (address) recipient's address
-- @param   amount (ubig) amount of token to send
-- @param   ...    additional data, is sent unaltered in call to 'tokensReceived' on 'to'
-- @return  value returned from 'tokensReceived' callback, or nil

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

  assert(_balances[from] and _balances[from] >= amount, "ARC1: not enough balance")

  _balances[from] = _balances[from] - amount
  _balances[to] = (_balances[to] or bignum.number(0)) + amount

  return _callTokensReceived(from, to, amount, ...)
end

-- Mint new tokens to an account
-- @type    internal
-- @param   to      (address) recipient's address
-- @param   amount  (ubig) amount of tokens to mint
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

  assert(_balances[from] and _balances[from] >= amount, "ARC1: not enough balance")

  _totalSupply:set(_totalSupply:get() - amount)
  _balances[from] = _balances[from] - amount
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


function set_contract_owner(address)
  assert(system.getSender() == _contract_owner:get(), "ARC1: permission denied")
  _typecheck(address, "address")
  _contract_owner:set(address)
end

-- returns a JSON string containing the list of ARC1 extensions
-- that were included on the contract
function arc1_extensions()
  local list = {}
  for name,_ in pairs(extensions) do
    table.insert(list, name)
  end
  return list
end


abi.register(transfer, set_contract_owner)
abi.register_view(arc1_extensions)
]]

arc1_burnable = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- Burnable
------------------------------------------------------------------------------

extensions["burnable"] = true

-- Burn tokens (from TX sender)
-- @type    call
-- @param   amount  (ubig) amount of token to burn
-- @event   burn(account, amount)

function burn(amount)
  amount = _check_bignum(amount)

  local sender = system.getSender()

  _burn(sender, amount)

  contract.event("burn", sender, bignum.tostring(amount))
end

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

-- set Max Supply
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

-- Renounce the Minter Role of TX sender
-- @type    call
-- @event   removeMinter(TX sender)

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

-- return Max Supply
-- @type    query
-- @return  amount   (ubig) amount of tokens to mint

function maxSupply()
  return _max_supply:get() or bignum.number(0)
end

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
  return (_paused:get() == true)
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


abi.register(pause, unpause, removePauser, renouncePauser, addPauser)
abi.register_view(paused, isPauser)
]]

arc1_blacklist = [[
------------------------------------------------------------------------------
-- Aergo Standard Token Interface (Proposal) - 20211028
-- Blacklist
------------------------------------------------------------------------------

extensions["blacklist"] = true

-- Add accounts to blacklist.
-- @type    call
-- @param   account_list    (list of address)
-- @event   addToBlacklist(account_list)

function addToBlacklist(account_list)
  assert(system.getSender() == _contract_owner:get(), "ARC1: only owner can blacklist anothers")

  for i = 1, #account_list do
    _typecheck(account_list[i], 'address')
    _blacklist[account_list[i] ] = true
  end

  contract.event("addToBlacklist", account_list)
end


-- removes accounts from blacklist
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


-- Retrun true when an account is on blacklist
-- @type    query
-- @param   account   (address)

function isOnBlacklist(account)
  _typecheck(account, 'address')

  return _blacklist[account] == true
end


abi.register(addToBlacklist,removeFromBlacklist)
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

-- Allow an account to use all TX sender's tokens
-- @type    call
-- @param   operator  (address) operator's address
-- @param   approved  (boolean) true/false
-- @event   setApprovalForAll(TX sender, operator, approved)

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

-- Transfer tokens from an account to another, Tx sender have to be approved to spend from the account
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


abi.register(setApprovalForAll, transferFrom)
abi.register_view(isApprovedForAll)


-- Burn tokens from an account, the operator needs to be approved to spend from the account
-- @type    call
-- @param   from    (address) sender's address
-- @param   amount  (ubig)    amount of tokens to send
-- @event   burn(from, amount, operator)

if extensions["burnable"] == true then

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

-- Approve an account to spend the specified amount of Tx sender's tokens
-- @type    call
-- @param   operator (address) operator's address
-- @param   amount   (ubig)    amount of allowed tokens
-- @event   approve(Tx sender, operator, amount)

function approve(operator, amount)
  _typecheck(operator, 'address')
  amount = _check_bignum(amount)

  local owner = system.getSender()

  assert(owner ~= operator, "ARC1: cannot approve self as operator")

  _allowance[owner .. "/" .. operator] = amount

  contract.event("approve", owner, operator, bignum.tostring(amount))
end


-- Increase the amount of tokens that Tx sender allowed to an account
-- @type    call
-- @param   operator (address) operator's address
-- @param   amount   (ubig)    amount of increased tokens
-- @event   increaseAllowance(Tx sender, operator, amount)

function increaseAllowance(operator, amount)
  _typecheck(operator, 'address')
  amount = _check_bignum(amount)

  local owner = system.getSender()
  local pair = owner .. "/" .. operator

  assert(_allowance[pair], "ARC1: not approved")

  _allowance[pair] = _allowance[pair] + amount

  contract.event("increaseAllowance", owner, operator, bignum.tostring(amount))
end


-- Decrease the amount of tokens that Tx sender allowed to an account
-- @type    call
-- @param   operator (address) operator's address
-- @param   amount   (ubig)    amount of decreased tokens
-- @event   decreaseAllowance(Tx sender, operator, amount)

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


abi.register(approve, increaseAllowance, decreaseAllowance, limitedTransferFrom)
abi.register_view(allowance)


-- Burn tokens from an account using the allowance mechanism
-- @type    call
-- @param   from    (address) sender's address
-- @param   amount  (ubig)    amount of tokens to burn
-- @event   burn(from, amount, operator)

if extensions["burnable"] == true then

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
`

var airdropFactory = `
-------------------------------------------------------------------
-- AIRDROP CONTRACT - DIFFERENT AMOUNT FOR EACH
-------------------------------------------------------------------

airdrop_diff_amount = [[

state.var {
  recipients = state.map()
}

creator = "%creator_address%"
token_address = "%token_address%"

local function _typecheck(x, t)
  if (x and t == 'address') then -- a string of alphanumeric char. except for '0, I, O, l'
    assert(type(x) == 'string', "AirDrop: address must be string type")
    assert(#x == 52, string.format("AirDrop: invalid address length: %s (%s)", x, #x))
    local invalidChar = string.match(x, '[^123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]')
    assert(nil == invalidChar, string.format("AirDrop: invalid address format: %s contains invalid char %s", x, invalidChar or 'nil'))
  elseif (x and t == 'str_ubig') then   -- a positive big number in string format
    assert(type(x) == 'string', "AirDrop: amount must be string")
    assert(string.match(x, '[^0-9]') == nil, "AirDrop: amount contains invalid character")
    x = bignum.number(x)
    assert(x >= bignum.number(0), string.format("AirDrop: %s must be positive number", bignum.tostring(x)))
  else
    assert(type(x) == t, string.format("AirDrop: invalid type: %s != %s", type(x), t or 'nil'))
  end
end

function add_recipients(list)

  assert(system.getSender() == creator, "permission denied")

  for address,amount in pairs(list) do
    _typecheck(address, "address")
    _typecheck(amount, "str_ubig")
    recipients[address] = amount
  end

end

function token()
  return token_address  
end

function has_tokens(address)
  _typecheck(address, "address")
  return recipients[address]
end

function withdraw()

  local address = system.getSender()
  local amount = recipients[address]
  if amount ~= nil then
    amount = bignum.number(amount)
  end
  assert(amount ~= nil and amount > bignum.number(0), "no amount to withdraw")

  -- update state before external call to avoid re-entrancy attack
  recipients[address] = nil

  -- transfer tokens to the requester
  contract.call(token_address, "transfer", address, amount)

end

function tokensReceived(operator, from, amount, ...)
  -- used to receive tokens from the contract
end

abi.register(add_recipients, withdraw, tokensReceived)
abi.register_view(token, has_tokens)
]]

-------------------------------------------------------------------
-- AIRDROP CONTRACT - SAME AMOUNT FOR ALL
-------------------------------------------------------------------

airdrop_same_amount = [[

state.var {
  recipients = state.map()
}

creator = "%creator_address%"
token_address = "%token_address%"
withdraw_amount = "%withdraw_amount%"

local function _typecheck(x, t)
  if (x and t == 'address') then -- a string of alphanumeric char. except for '0, I, O, l'
    assert(type(x) == 'string', "AirDrop: address must be string type")
    assert(#x == 52, string.format("AirDrop: invalid address length: %s (%s)", x, #x))
    local invalidChar = string.match(x, '[^123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]')
    assert(nil == invalidChar, string.format("AirDrop: invalid address format: %s contains invalid char %s", x, invalidChar or 'nil'))
  elseif (x and t == 'str_ubig') then   -- a positive big number in string format
    assert(type(x) == 'string', "AirDrop: amount must be string")
    assert(string.match(x, '[^0-9]') == nil, "AirDrop: amount contains invalid character")
    x = bignum.number(x)
    assert(x >= bignum.number(0), string.format("AirDrop: %s must be positive number", bignum.tostring(x)))
  else
    assert(type(x) == t, string.format("AirDrop: invalid type: %s != %s", type(x), t or 'nil'))
  end
end

function add_recipients(list)

  assert(system.getSender() == creator, "permission denied")

  for _,address in ipairs(list) do
    _typecheck(address, "address")
    recipients[address] = true
  end

end

function token()
  return token_address  
end

function has_tokens(address)
  _typecheck(address, "address")
  if recipients[address] == true then
    return withdraw_amount
  else
    return nil
  end
end

function withdraw()

  local address = system.getSender()
  assert(recipients[address] == true, "no amount to withdraw")

  -- update state before external call to avoid re-entrancy attack
  recipients[address] = nil

  -- transfer tokens to the requester
  contract.call(token_address, "transfer", address, withdraw_amount)

end

function tokensReceived(operator, from, amount, ...)
  -- used to receive tokens from the contract
end

abi.register(add_recipients, withdraw, tokensReceived)
abi.register_view(token, has_tokens)
]]

-------------------------------------------------------------------
-- FACTORY
-------------------------------------------------------------------

state.var {
  _num_airdrops = state.value(),
  _airdrops = state.map()
}

local function _typecheck(x, t)
  if (x and t == 'address') then -- a string of alphanumeric char. except for '0, I, O, l'
    assert(type(x) == 'string', "AirDrop: address must be string type")
    assert(#x == 52, string.format("AirDrop: invalid address length: %s (%s)", x, #x))
    local invalidChar = string.match(x, '[^123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]')
    assert(nil == invalidChar, string.format("AirDrop: invalid address format: %s contains invalid char %s", x, invalidChar or 'nil'))
  elseif (x and t == 'str_ubig') then   -- a positive big number in string format
    assert(type(x) == 'string', "AirDrop: amount must be string")
    assert(string.match(x, '[^0-9]') == nil, "AirDrop: amount contains invalid character")
    x = bignum.number(x)
    assert(x >= bignum.number(0), string.format("AirDrop: %s must be positive number", bignum.tostring(x)))
  else
    assert(type(x) == t, string.format("AirDrop: invalid type: %s != %s", type(x), t or 'nil'))
  end
end

function new_airdrop(owner, token, airdrop_type, amount)
  _typecheck(token, "address")

  if owner == nil or owner == '' then
    owner = system.getSender()
  else
    _typecheck(owner, "address")
  end

  -- check if it is an ARC1 token - discard the result
  local result = contract.call(token, "arc1_extensions")

  local contract_code

  if airdrop_type == "same" then
    _typecheck(amount, "str_ubig")
    contract_code = airdrop_same_amount
    contract_code = string.gsub(contract_code, "%%withdraw_amount%%", amount)
  elseif airdrop_type == "diff" then
    contract_code = airdrop_diff_amount
  else
    assert(false, "invalid airdrop type: '" .. airdrop_type .. "'")
  end

  contract_code = string.gsub(contract_code, "%%creator_address%%", owner)
  contract_code = string.gsub(contract_code, "%%token_address%%", token)

  local address = contract.deploy(contract_code)

  local num_airdrops = (_num_airdrops:get() or 0) + 1
  _num_airdrops:set(num_airdrops)
  _airdrops[tostring(num_airdrops)] = address

  contract.event("new_airdrop", address)

  return address
end

function list_airdrops(first, count)

  if first == nil or first == 0 then
    first = 1
  end
  if count == nil or count == 0 then
    count = 100
  end
  local last = first + count - 1

  local list = {}

  for n = first,last do
    local address = _airdrops[tostring(n)]
    if address == nil then break end
    list[#list + 1] = address
  end

  return list
end

function has_tokens(account, first, count)

  if first == nil or first == 0 then
    first = 1
  end
  if count == nil or count == 0 then
    count = 100
  end
  local last = first + count - 1

  local num_airdrops = _num_airdrops:get() or 0
  if first > num_airdrops then
    return nil
  end

  local list = {}

  for n = first,last do
    local address = _airdrops[tostring(n)]
    if address == nil then break end
    local amount = contract.call(address, "has_tokens", account)
    if amount ~= nil then
      local token = contract.call(address, "token")
      local name = contract.call(token, "name")
      local symbol = contract.call(token, "symbol")
      local decimals = contract.call(token, "decimals")
      -- list[#list + 1] = {address, amount, name, symbol, decimals}
      local item = {
        address = address,
        amount = amount,
        name = name,
        symbol = symbol,
        decimals = decimals
      }
      list[#list + 1] = item
    end
  end

  return list
end

abi.register(new_airdrop)
abi.register_view(list_airdrops, has_tokens)
`

func TestDifferentFee(t *testing.T) {
	// version 2
	err := LoadDummyChainForked(2, func(bc *DummyChain) error {
		// deploy token_factory
		txDeployTokenFactory := NewLuaTxDef("kch", "token_factory", 0, newTokenFactory)
		err := bc.ConnectBlock(
			NewLuaTxAccount("kch", 1e18),
			txDeployTokenFactory,
		)
		if err != nil {
			return err
		}

		// get contract address for deploy token_factory
		receiptDeployTokenFactory := bc.GetReceipt(txDeployTokenFactory.Hash())
		contractAddressContract := types.EncodeAddress(receiptDeployTokenFactory.ContractAddress)
		fmt.Println(contractAddressContract)

		// call new_token
		txTokenFactory := NewLuaTxCall("kch", contractAddressContract, 0, `{"Name":"new_token", "Args":[
			"Confetti","CFT",18,"500",{
				"mintable": true,
				"burnable": true,
				"blacklist": true,
				"pausable": true,
				"max_supply": "1000"
			},null
		  ]}`)
		err = bc.ConnectBlock(txTokenFactory)
		if err != nil {
			return err
		}

		// get token address
		receiptCallTokenFactory := bc.GetReceipt(txTokenFactory.Hash())
		var tokenName string
		for _, event := range receiptCallTokenFactory.GetEvents() {
			tokenName = event.JsonArgs[2 : len(event.JsonArgs)-2]
		}
		fmt.Println("token name :", tokenName)

		// deploy airdrop_factory
		TxAirdropFactory := NewLuaTxDef("kch", "airdrop_factory", 0, airdropFactory)
		err = bc.ConnectBlock(
			TxAirdropFactory,
		)
		if err != nil {
			return err
		}

		// get contract address for deploy airdrop_factory
		receiptAirdropFactory := bc.GetReceipt(TxAirdropFactory.Hash())
		contractAddressAirdropFactory := types.EncodeAddress(receiptAirdropFactory.ContractAddress)

		traceState = true

		// call airdrop and check fee
		TxCallNewAirdrop := NewLuaTxCall("kch", contractAddressAirdropFactory, 0, fmt.Sprintf(`{"Name":"new_airdrop", "Args":[
			null, "%s", "same", "1250000000000000000"
		]}`, tokenName))
		err = bc.ConnectBlock(TxCallNewAirdrop)
		if err != nil {
			return err
		}

		// get receipt about new_airdrop tx call
		receiptCallAirdrop := bc.GetReceipt(TxCallNewAirdrop.Hash())
		fmt.Println("gas used :", receiptCallAirdrop.GasUsed)
		return nil
	})

	require.NoError(t, err, "failed to execute")
}
