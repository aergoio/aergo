function testState()
    system.setItem("key1", 999)
    return system.getSender(), system.getTxhash(), system.getContractID(), system.getTimestamp(), system.getBlockheight(), system.getItem("key1")
end

abi.register(testState)

function get_version()
	return system.version()
end

function to_address(pubkey)
  return system.toAddress(pubkey)
end

function to_pubkey(address)
  return system.toPubKey(address)
end

abi.register_view(get_version, to_address, to_pubkey)
