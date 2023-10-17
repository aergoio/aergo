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
state.var {
    _version = state.value(),
    _implementation = state.value(),
    _payableList = state.map(),
    _payableArray = state.value()
}
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
-- @event UpgradedPayableList()
-- @event Upgraded(version, implementation)
function upgradeTo(version, implementation, payableList)
    _onlyProxyOwner()
    _upgradePayableList(payableList)
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
function _deletePayableList(oldList)
    if nil ~= _payableArray:get() then
        for i, payable in ipairs(oldList) do
            _payableList:delete(payable)
        end
    end
end
function _setPayableList(newList)
    _payableArray:set(newList)
    if nil ~= newList then
        for i, payable in ipairs(newList) do
            _payableList[payable] = true
        end
    end
end
function _upgradePayableList(newList)
    assert(nil == newList or 'table' == type(newList), "invalid payable list")
    local oldList = _payableArray:get()
    _deletePayableList(oldList)
    _setPayableList(newList)
    -- temporary commit
    -- contract.event("UpgradedPayableList", oldList, newList)
    contract.event("UpgradedPayableList")
end
function upgradePayableList(newList)
    _onlyProxyOwner()
    _upgradePayableList(newList)
end
function payableList()
    return _payableArray:get()
end
function _isPayable(name)
    return _payableList[name]
end
function refund(addr, amount)
    _onlyProxyOwner()
    _typecheck(addr, 'address')
    contract.send(addr, amount)
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
    if not bignum.iszero(system.getAmount()) then
        assert(_isPayable(callName), string.format("[%s] is not payable.", callName))
    end
    return contract.delegatecall(_implementation:get(), callName, ...)
end
-- Used to clearly show that it is a query-only
function query(callName, ...)
    assert(nil ~= _implementation:get(), "implementation is nil")
    assert(nil ~= callName, "callName is nil")
    return contract.delegatecall(_implementation:get(), callName, ...)
end

function check_delegation(fname, ...)
    return contract.delegatecall(_implementation:get(), "checkDelegation", ...)
end
abi.register(upgradeTo, upgradePayableList, refund)
abi.register_view(proxyOwner, version, implementation, payableList, query)
abi.payable(default,invoke)
abi.fee_delegation(invoke)
