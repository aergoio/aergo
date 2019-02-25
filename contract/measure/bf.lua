loop_cnt = 1000
arr = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 0 }
tbl = { name= "kslee", year= 1981, age= 41 }

local print = function(...)
    system.print(...)
end

local copy_arr = function(arr)
    local a = {}
    for i, v in ipairs(arr) do
        a[i] = v
    end
    return a
end

function measure(fn, ...)
    local start = nsec()
    fn(...)
    return nsec() - start
end

function m1k(fn, ...)
    local r1k = function(fn)
        return function(...)
            for i = 1, loop_cnt do
                fn(...)
            end
        end
    end
    t = measure(r1k(fn), ...)
    print(string.format('%.9f', t / loop_cnt))
end

function print_chain(fn)
    return function(...)
        print(fn(...))
    end
end

function m1kp(fn, ...)
    m1k(print_chain(fn), ...)
end

function system_fns()
    print("sysem.getSender")
    m1k(system.getSender)
    print("system.getBlockheight")
    m1k(system.getBlockheight)
    print("system.getTxhash")
    m1k(system.getTxhash)
    print("system.getTimestamp")
    m1k(system.getTimestamp)
    print("system.getContractID")
    m1k(system.getContractID)
    print("system.setItem")
    m1k(system.setItem, "key100", "100ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
    m1k(system.setItem, "key200", "200ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
    m1k(system.setItem, "key400", "400ddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
    print("system.getItem")
    m1k(system.getItem, "key100")
    m1k(system.getItem, "key200")
    m1k(system.getItem, "key400")
    print("system.getAmount")
    m1k(system.getAmount)
    print("system.getCreator")
    m1k(system.getCreator)
    print("system.Origin")
    m1k(system.getOrigin)
    print("system.getPrevBlockHash")
    m1k(system.getPrevBlockHash)
end

--TODO
function contract_fns()
end

local s100 = "100aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
local s200 =  "200aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
local s400 = "400aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

function json_fns()
    print("json.encode")
    m1k(json.encode, 1)
    m1k(json.encode, 100)
    m1k(json.encode, 3.14)
    m1k(json.encode, s100)
    m1k(json.encode, s200)
    m1k(json.encode, s400)
    m1k(json.encode, arr)
    local long_arr = {}
    for i = 1, 100 do
        long_arr[i] = i
    end
    local long_arr100 = json.encode(long_arr)
    m1k(json.encode, long_arr100)
    for i = #long_arr + 1, loop_cnt do
        long_arr[i] = i
    end
    m1k(json.encode, long_arr)
    m1k(json.encode, tbl)
    print("json.decode")
    m1k(json.decode, json.encode(1))
    m1k(json.decode, json.encode(100))
    m1k(json.decode, json.encode(3.14))
    m1k(json.decode, json.encode(s100))
    m1k(json.decode, json.encode(s200))
    m1k(json.decode, json.encode(s300))
    m1k(json.decode, json.encode(arr))
    m1k(json.decode, long_arr100)
    m1k(json.decode, json.encode(long_arr))
    m1k(json.decode, json.encode(tbl))
end

function crypto_fns()
    print("crypto.sha256")
	m1k(crypto.sha256, "0x616200e490aa")
	m1k(crypto.sha256, s100)
	m1k(crypto.sha256, s200)
	m1k(crypto.sha256, s400)
    print("crypto.ecverify")
	m1k(crypto.ecverify, "11e96f2b58622a0ce815b81f94da04ae7a17ba17602feb1fd5afa4b9f2467960", "304402202e6d5664a87c2e29856bf8ff8b47caf44169a2a4a135edd459640be5b1b6ef8102200d8ea1f6f9ecdb7b520cdb3cc6816d773df47a1820d43adb4b74fb879fb27402", "AmPbWrQbtQrCaJqLWdMtfk2KiN83m2HFpBbQQSTxqqchVv58o82i")
end

function bignum_fns()
    print("bignum.number")
    local nines = "999999999999999999999999999999"
    local nines1 = "999999999"
    local nines2 = 999999999
    local one = 1
    local b = 3.14e19
    local pi = 3.14
    local pi1 = 3.141592653589
    m1k(bignum.number, nines)
    m1k(bignum.number, nines1)
    m1k(bignum.number, nines2)
    m1k(bignum.number, one)
    m1k(bignum.number, b)
    m1k(bignum.number, pi)
    m1k(bignum.number, pi1)
    local nines_b = bignum.number(nines)
    local nines1_b = bignum.number(nines1)
    local nines2_b = bignum.number(nines2)
    local one_b = bignum.number(one)
    local b_b = bignum.number(b)
    local pi_b = bignum.number(pi)
    local pi1_b = bignum.number(pi1)
    print("bignum.isneg")
    m1k(bignum.isneg, nines_b)
    m1k(bignum.isneg, nines1_b)
    m1k(bignum.isneg, nines2_b)
    m1k(bignum.isneg, one_b)
    m1k(bignum.isneg, b_b)
    m1k(bignum.isneg, pi_b)
    m1k(bignum.isneg, pi1_b)
    print("bignum.iszero")
    m1k(bignum.iszero, nines_b)
    m1k(bignum.iszero, nines1_b)
    m1k(bignum.iszero, nines2_b)
    m1k(bignum.iszero, one_b)
    m1k(bignum.iszero, b_b)
    m1k(bignum.iszero, pi_b)
    m1k(bignum.iszero, pi1_b)
    print("bignum.tonumber")
    m1k(bignum.tonumber, nines_b)
    m1k(bignum.tonumber, nines1_b)
    m1k(bignum.tonumber, nines2_b)
    m1k(bignum.tonumber, one_b)
    m1k(bignum.tonumber, b_b)
    m1k(bignum.tonumber, pi_b)
    m1k(bignum.tonumber, pi1_b)
    print("bignum.tostring")
    m1k(bignum.tostring, nines_b)
    m1k(bignum.tostring, nines1_b)
    m1k(bignum.tostring, nines2_b)
    m1k(bignum.tostring, one_b)
    m1k(bignum.tostring, b_b)
    m1k(bignum.tostring, pi_b)
    m1k(bignum.tostring, pi1_b)
    print("bignum.neg")
    m1k(bignum.neg, nines_b)
    m1k(bignum.neg, nines1_b)
    m1k(bignum.neg, nines2_b)
    m1k(bignum.neg, one_b)
    m1k(bignum.neg, b_b)
    m1k(bignum.neg, pi_b)
    m1k(bignum.neg, pi1_b)
    print("bignum.sqrt")
    m1k(bignum.sqrt, nines_b)
    m1k(bignum.sqrt, nines1_b)
    m1k(bignum.sqrt, nines2_b)
    m1k(bignum.sqrt, one_b)
    m1k(bignum.sqrt, b_b)
    m1k(bignum.sqrt, pi_b)
    m1k(bignum.sqrt, pi1_b)
    print("bignum.compare")
    m1k(bignum.compare, one_b, one_b)
    m1k(bignum.compare, nines_b, one_b)
    m1k(bignum.compare, nines_b, nines_b)
    m1k(bignum.compare, pi_b, pi_b)
    m1k(bignum.compare, pi1_b, pi1_b)
    m1k(bignum.compare, pi1_b, one_b)
    m1k(bignum.compare, pi_b, one_b)
    print("bignum.add")
    m1k(bignum.add, one_b, one_b)
    m1k(bignum.add, nines_b, one_b)
    m1k(bignum.add, nines_b, nines_b)
    m1k(bignum.add, pi_b, pi_b)
    m1k(bignum.add, pi_b, pi1_b)
    m1k(bignum.add, pi_b, one_b)
    m1k(bignum.add, pi_b, nines_b)
    m1k(bignum.add, pi1_b, one_b)
    print("bignum.sub")
    m1k(bignum.sub, one_b, one_b)
    m1k(bignum.sub, bignum.number(-1), one_b)
    m1k(bignum.sub, nines_b, one_b)
    m1k(bignum.sub, nines_b, nines_b)
    m1k(bignum.sub, nines_b, b_b)
    m1k(bignum.sub, pi_b, nines_b)
    m1k(bignum.sub, pi1_b, nines_b)
    print("bignum.mul")
    m1k(bignum.mul, one_b, one_b)
    m1k(bignum.mul, bignum.number(-1), one_b)
    m1k(bignum.mul, nines_b, one_b)
    m1k(bignum.mul, nines_b, nines1_b)
    m1k(bignum.mul, nines_b, nines_b)
    m1k(bignum.mul, pi_b, one_b)
    m1k(bignum.mul, pi_b, nines_b)
    m1k(bignum.mul, pi_b, pi_b)
    m1k(bignum.mul, pi1_b, pi1_b)
    print("bignum.div")
    m1k(bignum.div, one_b, one_b)
    m1k(bignum.div, bignum.number(-1), one_b)
    m1k(bignum.div, nines_b, one_b)
    m1k(bignum.div, nines_b, nines_b)
    m1k(bignum.div, nines_b, bignum.number(333))
    m1k(bignum.div, pi1_b, bignum.number(1))
    m1k(bignum.div, pi_b, pi1_b)
    print("bignum.mod")
    m1k(bignum.mod, one_b, one_b)
    m1k(bignum.mod, bignum.number(-1), one_b)
    m1k(bignum.mod, nines_b, one_b)
    m1k(bignum.mod, nines_b, nines1_b)
    m1k(bignum.mod, nines_b, bignum.number(333))
    m1k(bignum.mod, bignum.number(333), nines_b)
    m1k(bignum.mod, nines1_b, bignum.number(333))
    m1k(bignum.mod, bignum.number(333), nines1_b)
    m1k(bignum.mod, pi1_b, pi_b)
    m1k(bignum.mod, pi1_b, one_b)
    m1k(bignum.mod, nines_b, nines1_b)
    print("bignum.pow")
    m1k(bignum.pow, one_b, one_b)
    m1k(bignum.pow, bignum.number(-1), one_b)
    m1k(bignum.pow, nines_b, one_b)
    m1k(bignum.pow, nines_b, bignum.number(3))
    m1k(bignum.pow, nines_b, bignum.number(9))
    m1k(bignum.pow, nines1_b, bignum.number(3))
    m1k(bignum.pow, nines1_b, bignum.number(9))
    m1k(bignum.pow, pi1_b, one_b)
    m1k(bignum.pow, pi1_b, pi_b)
    m1k(bignum.pow, pi1_b, pi1_b)
    print("bignum.divmod")
    m1k(bignum.divmod, one_b, one_b)
    m1k(bignum.divmod, bignum.number(-1), one_b)
    m1k(bignum.divmod, nines_b, one_b)
    m1k(bignum.divmod, nines_b, nines1_b)
    m1k(bignum.divmod, bignum.number(3), nines_b)
    m1k(bignum.divmod, bignum.number(9), nines_b)
    m1k(bignum.divmod, bignum.number(3), nines1_b)
    m1k(bignum.divmod, bignum.number(9), nines1_b)
    m1k(bignum.divmod, pi1_b, pi_b)
    m1k(bignum.divmod, pi1_b, pi1_b)
    m1k(bignum.divmod, pi1_b, one_b)
    print("bignum.powmod")
    m1k(bignum.powmod, one_b, one_b, bignum.number(3))
    m1k(bignum.powmod, bignum.number(-1), one_b, bignum.number(3))
    m1k(bignum.powmod, nines_b, one_b, bignum.number(3))
    m1k(bignum.powmod, nines_b, bignum.number(3), bignum.number(4))
    m1k(bignum.powmod, nines_b, bignum.number(9), bignum.number(4))
    m1k(bignum.powmod, nines1_b, bignum.number(3), bignum.number(4))
    m1k(bignum.powmod, nines1_b, bignum.number(9), bignum.number(4))
    m1k(bignum.powmod, nines_b, bignum.number(9), bignum.number(4))
    m1k(bignum.powmod, pi_b, bignum.number(9), bignum.number(4))
    m1k(bignum.powmod, pi1_b, bignum.number(9), bignum.number(4))
end

function run_test()
    system_fns()
    json_fns()
    crypto_fns()
    bignum_fns()
end

abi.register(run_test)

