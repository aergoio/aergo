function testState()
    system.setItem("key1", 999)
    return system.getSender(), system.getTxhash(), system.getContractID(), system.getTimestamp(), system.getBlockheight(),
        system.getItem("key1")
end

abi.register(testState)


function get_version()
	return system.version()
end

abi.register_view(get_version)
