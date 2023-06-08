function testState()
	system.setItem("key1", 999)
	return system.getSender(), system.getTxhash(),system.getContractID(), system.getTimestamp(), system.getBlockheight(), system.getItem("key1")
end 
abi.register(testState)