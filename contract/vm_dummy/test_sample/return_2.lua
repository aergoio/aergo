function foo()
	return {1,2,3}
end
function foo2(bar)
	return bar
end
abi.register(foo,foo2)