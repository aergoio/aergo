function random(...)
	return system.random(...)
end

function get_numbers(count)
	local list = {}

	for i = 1, count do
		local num = system.random(1, 100)
		table.insert(list, num)
	end

	return list
end

abi.register(random, get_numbers)
