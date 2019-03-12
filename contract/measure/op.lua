LLINE = "long line aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
LLINE1 = "long line aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

function comp_ops()
    A = 2
    D = 1 if A < D then
    end
    if A <= D then
    end
    if not (A < D) then
    end
    if not (A <= D) then
    end
    if A == D then
    end
    if not (A == D) then
    end
    V = "kslee"
    if V == "kslee" then
    end
    if V ~= "a very long long long line" then
    end
    if LLINE ~= V then
    end
    if LLINE ~= LLINE1 then
    end
    if LLINE == LLINE1 then
    end
    if A == 2 then
    end
    if A ~= 1 then
    end
    B = true
    if B == true then
    end
    if B ~= false then
    end
end

local istc = function (b)
    return b and 1
end

local isfc = function (b)
    return b or 1
end

function unarytest_n_copy_ops()
    -- 16, 17
    NONE = nil
    B = false
    S = "kslee"
    istc(true)
    isfc(true)
    if S and B then
    end
    if NONE or B then
    end
    if NONE or B then
    end
    B = true
    if B or NONE then
    end
end

function unary_ops()
    D = 3
    A = D
    B = not A
    N = -D
    ARR = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 0 }
    L = #ARR
    S = "a very long long long line"
    L = #S
end

function binary_ops()
    A = 0
    B = 2
    A = B + 1
    A = B - 1
    A = B * 1
    A = B / 1
    A = B % 1
    A = B + 1.2
    A = B - 1.2
    A = B * 1.2
    A = B / 1.2
    A = B % 1.2
    A = 1 + B
    A = 1 - B
    A = 1 * B
    A = 1 / B
    A = 1 % B
    A = 1.2 + B
    A = 1.2 - B
    A = 1.2 * B
    A = 1.2 / B
    A = 1.2 % B
    C = 1
    A = B + C
    A = B - C
    A = B * C
    A = B / C
    A = B % C
    A = B ^ C
    C = 1.2
    A = B + C
    A = B - C
    A = B * C
    A = B / C
    A = B % C
    A = B ^ C
    S = "kslee" .. "1981" .. LLINE
end

function constant_ops()
    A, B, C = nil, nil, nil
end

function upvalue_n_func_ops()
    local U = "kslee"
    F = function() U = U .. "1981"; return U end
    F()
    F = function() U = "1981" end
    F()
    local D = 100
    F = function() D = 0 end
    F()
    local B = false
    F = function() B = true end
    F()
end

function f()
    return 4
end

function table_ops()
    -- deprecated(op code not found): TGETR(59), TSETR(64)
    E = {}
    T = { name = "kslee", age = 40, 1, 2, 3 }
    K = "name"
    NAME = T[K]
    AGE = T.age
    ONE = T[1]
    I = 1
    ONE = T[I]
    T[K] = "kslee"
    T.age = 41
    T[4] = I
    return { 1, 2, 3, f() }
end

function varg(...)
    return ...
end

function fixed_varg(a, b, c, ...)
    return a + b + c, ...
end

function call_n_vararg_ops()
    varg("kslee", "1981", 41)
    T = 0
    ARR = { 1, 2, 3, 4, 5, 6, 7, 8, 9, 0 }
    for i, v in ipairs(ARR) do
        T = T + v
    end
    ARR.name = "kslee"
    for k, v in pairs(ARR) do
        if type(k) == 'number' then
            T = T + v
        end
    end
    local x, y, z = fixed_varg(1, 2, 3, varg("kslee", "1981", 41))
    return fixed_varg(1, 2, 3, varg("kslee", "1981", 41))
end

function swap(x, y)
    return y, x
end

function return_ops()
    return swap(1, 2)
end

function range_iter(n, i)
    if i >= n then
        return nil, nil
    end
    return i+1, i+1
end

function range(n)
    return range_iter, n, 0
end

function loop_n_branche_ops()
    -- IFORL(80), IITERL(83)
    local T = 0
    for i = 1, 100 do
        T = T + i    
    end
    for n in range(100) do
        T = T + n
    end
    local i = 0
    while true do
        i = i + 1
      if i == 100 then
          break
      end
    end
end

function function_header_ops()
    -- FUNCF(89), IFUNCV(93), FUNCCW(96)
end

for i = 1, 1000 do
    comp_ops()
    unarytest_n_copy_ops()
    unary_ops()
    binary_ops()
    constant_ops()
    upvalue_n_func_ops()
    table_ops()
    call_n_vararg_ops()
    return_ops()
    loop_n_branche_ops()
    function_header_ops()
end

