function test(addr)
	bal = contract.balance()
	contract.send(addr, bal / 2)
	return contract.balance()
end

function sendS(addr)
	contract.send(addr, "1 gaer 99999")
	return contract.balance()
end

function testBignum()
	bg = bignum.number("999999999999999999999999999999")
	system.setItem("big", bg)
	bi = system.getItem("big")
	return tostring(bi)
end

function argBignum(a)
	b = a + 1
	return tostring(b)
end

function calladdBignum(addr, a)
	return tostring(contract.call(addr, "add", a, 2) + 3)
end

function checkBignum()
	a = 1
	b = bignum.number(1)
	
	return bignum.isbignum(a), bignum.isbignum(b), bignum.isbignum("2333")
end
function calcBignum()
	bg1 = bignum.number("999999999999999999999999999999")
	bg2 = bignum.number("999999999999999999999999999999")
	bg3 = bg1 + bg2
	bg4 = bg1 * 2
	bg5 = 2 * bg1
	n1 = 999999999999999
	system.print(n1)
	bg6 = bignum.number(n1)
	assert (bg3 == bg4 and bg4 == bg5)
	bg5 = bg1 - bg3 
	assert (bignum.isneg(bg5) and bg5 == bignum.neg(bg1))
	system.print(bg3, bg5, bg6)
	bg6 = bignum.number(1)
	assert (bg6 > bg5)
	a = bignum.number(2)
	b = bignum.number(8)
	pow = a ^ b
	system.print(pow, a, b)
	assert(pow == bignum.number(256) and a == bignum.number(2) and b == bignum.number(8))
	assert(bignum.compare(bg6, 1) == 0)
	system.print((bg6 == 1), bignum.isbignum(pow))
	div1 = bignum.number(3)/2
	assert(bignum.compare(div1, 1) == 0)
	div = bg6 / 0
end

function negativeBignum()
	bg1 = bignum.number("-2")
	bg2 = bignum.sqrt(bg1)
end

function byteBignum()
	 state.var {
        value = state.value()
    }
	value = bignum.tobyte(bignum.number("177"))
	return bignum.frombyte(value)
end

function constructor()
end

abi.register(test, sendS, testBignum, argBignum, calladdBignum, checkBignum, calcBignum, negativeBignum, byteBignum)
abi.payable(constructor)
