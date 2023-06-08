state.var {
    c = state.map(),
}

function constructor()
    c[fromhex('00')] = "kk"
    c[fromhex('61')] = "kk"
    system.setItem(fromhex('00'), "kk")
end

function fromhex(str)
    return (str:gsub('..', function (cc)
        return string.char(tonumber(cc, 16))
    end))
end
function get()
	return c[fromhex('00')], system.getItem(fromhex('00')), system.getItem(fromhex('0000'))
end
function getcre()
	return system.getCreator()
end
abi.register(get, getcre)