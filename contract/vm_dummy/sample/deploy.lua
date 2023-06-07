function hello()
	hello = [[
function hello(say)
	return "Hello " .. say 
end

local type_check = {}
function type_check.isValidAddress(address)
    -- check existence of invalid alphabets
    if nil ~= string.match(address, '[^123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]') then
        return false
    end
    -- check lenght is in range
    if 52 ~= string.len(address) then
        return false
    end
    -- TODO add checksum verification?
    return true
end
function type_check.isValidNumber(value)
    if nil ~= string.match(value, '[^0123456789]') then
        return false
    end
    return true
end

-- The a bridge token is a mintable and burnable token controlled by
-- the bridge contract. It represents all tokens locked on the other side of the 
-- bridge with a 1:1 ratio.
-- This contract is depoyed by the merkle bridge when a new type of token 
-- is transfered
state.var {
    Symbol = state.value(),
    Name = state.value(),
    Decimals = state.value(),
    TotalSupply = state.value(),
    Balances = state.map(),
    Nonces = state.map(),
    -- Contract ID is a unique id that cannot be shared by another contract, even one on a sidechain
    -- This is neeeded for replay protection of signed transfer, because users might have the same private key
    -- on different sidechains
    ContractID = state.value(),
    Owner = state.value(),
}

function constructor() 
    Symbol:set("TOKEN")
    Name:set("Standard Token on Aergo")
    Decimals:set(18)
    TotalSupply:set(bignum.number(0))
    Owner:set(system.getSender())
    -- contractID is the hash of system.getContractID (prevent replay between contracts on the same chain) and system.getPrevBlockHash (prevent replay between sidechains).
    -- take the first 16 bytes to save size of signed message
    local id = crypto.sha256(system.getContractID()..system.getPrevBlockHash())
    id = string.sub(id, 3, 32)
    ContractID:set(id)
    return true
end

---------------------------------------
-- Transfer sender's token to target 'to'
-- @type        call
-- @param to    a target address
-- @param value string amount of tokens to send
-- @return      success
---------------------------------------
function transfer(to, value) 
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    assert(type_check.isValidAddress(to), "invalid address format: " .. to)
    local from = system.getSender()
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    assert(bvalue > b0, "invalid value")
    assert(to ~= from, "same sender and receiver")
    assert(Balances[from] and bvalue <= Balances[from], "not enough balance")
    Balances[from] = Balances[from] - bvalue
    Nonces[from] = (Nonces[from] or 0) + 1
    Balances[to] = (Balances[to] or b0) + bvalue
    -- TODO event notification
    return true
end

---------------------------------------
-- Transfer tokens according to signed data from the owner
-- @type  call
-- @param from      sender's address
-- @param to        receiver's address
-- @param value     string amount of token to send in aer
-- @param nonce     nonce of the sender to prevent replay
-- @param fee       string fee given to the tx broadcaster
-- @param deadline  block number before which the tx can be executed
-- @param signature signature proving sender's consent
-- @return          success
---------------------------------------
function signed_transfer(from, to, value, nonce, signature, fee, deadline)
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    assert(type_check.isValidNumber(fee), "invalid fee format (must be string)")
    local bfee = bignum.number(fee)
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    -- check addresses
    assert(type_check.isValidAddress(to), "invalid address format: " .. to)
    assert(type_check.isValidAddress(from), "invalid address format: " .. from)
    assert(to ~= from, "same sender and receiver")
    -- check amounts, fee
    assert(bfee >= b0, "fee must be positive")
    assert(bvalue >= b0, "value must be positive")
    assert(Balances[from] and (bvalue+bfee) <= Balances[from], "not enough balance")
    -- check deadline
    assert(deadline == 0 or system.getBlockheight() < deadline, "deadline has passed")
    -- check nonce
    if Nonces[from] == nil then Nonces[from] = 0 end
    assert(Nonces[from] == nonce, "nonce is invalid or already spent")
    -- construct signed transfer and verifiy signature
    data = crypto.sha256(to..bignum.tostring(bvalue)..tostring(nonce)..bignum.tostring(bfee)..tostring(deadline)..ContractID:get())
    assert(crypto.ecverify(data, signature, from), "signature of signed transfer is invalid")
    -- execute transfer
    Balances[from] = Balances[from] - bvalue - bfee
    Balances[to] = (Balances[to] or b0) + bvalue
    Balances[system.getOrigin()] = (Balances[system.getOrigin()] or b0) + bfee
    Nonces[from] = Nonces[from] + 1
    -- TODO event notification
    return true
end


---------------------------------------
-- mint, burn and signed_burn are specific to the token contract controlled by
-- the merkle bridge contract and representing transfered assets.
---------------------------------------

---------------------------------------
-- Mint tokens to 'to'
-- @type        call
-- @param to    a target address
-- @param value string amount of token to mint
-- @return      success
---------------------------------------
function mint(to, value)
    assert(system.getSender() == Owner:get(), "Only bridge contract can mint")
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    assert(type_check.isValidAddress(to), "invalid address format: " .. to)
    local new_total = TotalSupply:get() + bvalue
    TotalSupply:set(new_total)
    Balances[to] = (Balances[to] or b0) + bvalue;
    -- TODO event notification
    return true
end

---------------------------------------
-- burn the tokens of 'from'
-- @type        call
-- @param from  a target address
-- @param value an amount of token to send
-- @return      success
---------------------------------------
function burn(from, value)
    assert(system.getSender() == Owner:get(), "Only bridge contract can burn")
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    assert(type_check.isValidAddress(from), "invalid address format: " ..from)
    assert(Balances[from] and bvalue <= Balances[from], "Not enough funds to burn")
    new_total = TotalSupply:get() - bvalue
    TotalSupply:set(new_total)
    Balances[from] = Balances[from] - bvalue
    -- TODO event notification
    return true
end

---------------------------------------
-- signed_burn the tokens of 'from' according to signed data from the owner
-- @type            call
-- @param from      a target address
-- @param value     an amount of token to send
-- @param nonce     nonce of the sender to prevent replay
-- @param fee       string fee given to the tx broadcaster
-- @param deadline  block number before which the tx can be executed
-- @param signature signature proving sender's consent
-- @return          success
---------------------------------------
function signed_burn(from, value, nonce, signature, fee, deadline)
    assert(system.getSender() == Owner:get(), "Only bridge contract can burn")
    assert(type_check.isValidNumber(value), "invalid value format (must be string)")
    assert(type_check.isValidNumber(fee), "invalid fee format (must be string)")
    local bfee = bignum.number(fee)
    local bvalue = bignum.number(value)
    local b0 = bignum.number(0)
    -- check addresses
    assert(type_check.isValidAddress(from), "invalid address format: " .. from)
    -- check amounts, fee
    assert(bfee >= b0, "fee must be positive")
    assert(bvalue >= b0, "value must be positive")
    assert(Balances[from] and (bvalue+bfee) <= Balances[from], "not enough balance")
    -- check deadline
    assert(deadline == 0 or system.getBlockheight() < deadline, "deadline has passed")
    -- check nonce
    if Nonces[from] == nil then Nonces[from] = 0 end
    assert(Nonces[from] == nonce, "nonce is invalid or already spent")
    -- construct signed transfer and verifiy signature
    data = crypto.sha256(system.getSender()..bignum.tostring(bvalue)..tostring(nonce)..bignum.tostring(bfee)..tostring(deadline)..ContractID:get())
    assert(crypto.ecverify(data, signature, from), "signature of signed transfer is invalid")
    -- execute burn
    new_total = TotalSupply:get() - bvalue
    TotalSupply:set(new_total)
    Balances[from] = Balances[from] - bvalue - bfee
    Balances[system.getOrigin()] = (Balances[system.getOrigin()] or b0) + bfee
    Nonces[from] = Nonces[from] + 1
    -- TODO event notification
    return true
end


-- register functions to abi
abi.register(transfer, signed_transfer, mint, burn, signed_burn, hello)
	]]
	addr = contract.deploy(hello)
	ret = contract.call(addr, "hello", "world")
	return addr, ret
end

function helloQuery(addr)
	return contract.call(addr, "hello", "world")
end

function testConst()
	src = [[
		function hello(say, key) 
			return "Hello " .. say .. system.getItem(key) 
		end 
		function constructor(key, item) 
			system.setItem(key, item)
			return key, item
		end 
		abi.register(hello) 
		abi.payable(constructor)
	]]
	addr, key, item = contract.deploy.value(100)(src, "key", 2)
	ret = contract.call(addr, "hello", "world", "key")
	return addr, ret
end

function testFail()
	src = [[
		function hello(say, key) 
			return "Hello " .. say .. system.getItem(key) 
		end 
		function constructor()
		end 
		abi.register(hello) 
	]]
	addr = contract.deploy.value(100)(src)
	return addr
end
 
paddr = nil
function deploy()
	src = [[
		function hello(say, key) 
			return "Hello " .. say .. system.getItem(key) 
		end 
		function getcre()
			return system.getCreator()
		end
		function constructor()
		end 
		abi.register(hello, getcre) 
	]]
	paddr = contract.deploy(src)
	system.print("addr :", paddr)
	ret = contract.call(paddr, "hello", "world", "key")
end

function testPcall()
	ret = contract.pcall(deploy)
	return contract.call(paddr, "getcre")
end
function constructor()
end

abi.register(hello, helloQuery, testConst, testFail, testPcall)
abi.payable(constructor)