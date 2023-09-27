state.var {
  arc1_factory = state.value(),
  arc2_factory = state.value(),
}

function constructor(f1, f2)
  arc1_factory:set(f1)
  arc2_factory:set(f2)
end

function deploy_tokens(n1, n2)

  for n = 1,n1 do
    address = contract.call(arc1_factory:get(), "new_token", "Test", "TST", 18, 1000, {burnable=true, mintable=true, pausable=true, blacklist=true, all_approval=true, limited_approval=true})
    assert(contract.call(address, "symbol") == "TST", "deployed contract is not working")
  end

  for n = 1,n2 do
    address = contract.call(arc2_factory:get(), "new_arc2_nft", "Test NFT 1", "NFT1", null, {burnable=true, mintable=true, metadata=true, pausable=true, blacklist=true, approval=true, searchable=true, non_transferable=true, recallable=true})
    assert(contract.call(address, "symbol") == "NFT1", "deployed contract is not working")
  end

end

function tokensReceived(operator, from, value, ...)
  -- do nothing
end

function nonFungibleReceived(operator, from, tokenId, ...)
  -- do nothing
end

abi.register(deploy_tokens, tokensReceived, nonFungibleReceived)
