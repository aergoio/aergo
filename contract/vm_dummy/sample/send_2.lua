function default()
    system.print("default called")
end
abi.register(default)
abi.payable(default)