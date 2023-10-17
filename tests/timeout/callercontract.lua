state.var {
    _version = state.value(),
    _implementation = state.value(),
}

-- A internal type check function
-- @type internal
-- @param x variable to check
-- @param t (string) expected type
local function _typecheck(x, t)
    if (x and t == 'address') then
      assert(type(x) == 'string', string.format("%s must be string type", t))
      -- check address length
      assert(52 == #x, string.format("invalid address length: %s (%s)", x, #x))
      -- check character
      local invalidChar = string.match(x, '[^123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]')
      assert(nil == invalidChar, string.format("invalid address format: %s contains invalid char %s", x, invalidChar or 'nil'))
    else
      -- check default lua types
      assert(type(x) == t, string.format("invalid type: %s != %s", type(x), t or 'nil'))
    end
end

function _onlyProxyOwner()
    assert(system.getCreator() == system.getSender(), string.format("only proxy owner. Owner: %s | sender: %s", system.getCreator(), system.getSender()))
end

-- Tells the address of the proxy owner
-- @type query
-- @return (address) The address of the proxy owner
function proxyOwner()
    return system.getCreator()
end
-- Allows the upgradeability owner to upgrade the current version of the proxy.
-- @type call
-- @param version (string) Representing the version name of the new implementation to be set.
-- @param implementation (contractAddress) Representing the address of the new implementation to be set.
-- @event Upgraded(version, implementation)
function upgradeTo(version, implementation)
    _onlyProxyOwner()
    _upgradeTo(version, implementation)
    contract.delegatecall(_implementation:get(), "init")
end
function _upgradeTo(version, implementation)
    _typecheck(version, 'string')
    _typecheck(implementation, 'address')
    assert(system.isContract(implementation), string.format("[upgradeTo] invalid address format: %s", implementation))
    assert(_implementation:get() ~= implementation, string.format("[upgradeTo] same implementation address %s", implementation))
    _version:set(version)
    _implementation:set(implementation)
    contract.event("Upgraded", version, implementation)
end

-- Tells the version name of the current implementation
-- @type query
-- @return (string) Representing the name of the current version
function version()
    return _version:get()
end
-- Tells the address of the implementation where every call will be delegated.
-- @type query
-- @return (address) of the implementation to which it will be delegated
function implementation()
    return _implementation:get()
end
function default()
end
function invoke(callName, ...)
    assert(nil ~= _implementation:get(), "implementation is nil")
    if nil == callName then
        callName = "default"
    end
    return contract.delegatecall(_implementation:get(), callName, ...)
end

-- Used to clearly show that it is a query-only
function query(callName, ...)
    assert(nil ~= _implementation:get(), "implementation is nil")
    assert(nil ~= callName, "callName is nil")
    return contract.delegatecall(_implementation:get(), callName, ...)
end

function directInvoke(callName, ...)
    assert(nil ~= _implementation:get(), "implementation is nil")
    if nil == callName then
        callName = "default"
    end
    return contract.call(_implementation:get(), callName, ...)
end

-- Used to clearly show that it is a query-only
function directQuery(callName, ...)
    assert(nil ~= _implementation:get(), "implementation is nil")
    assert(nil ~= callName, "callName is nil")
    return contract.call(_implementation:get(), callName, ...)
end

function check_delegation(fname, ...)
    return contract.delegatecall(_implementation:get(), "checkDelegation", ...)
end

abi.register(upgradeTo, invoke, directInvoke)
abi.register_view(proxyOwner, version, implementation, query, directQuery)
abi.fee_delegation(invoke, directInvoke)
