function test_gov()
    contract.stake("10000 aergo")
    contract.vote("16Uiu2HAm2gtByd6DQu95jXURJXnS59Dyb9zTe16rDrcwKQaxma4p")
end

function error_case()
    contract.stake("10000 aergo")
    assert(false)
end

function test_pcall()
    return contract.pcall(error_case)
end

abi.register(test_gov, test_pcall, error_case)
abi.payable(test_gov, test_pcall)
