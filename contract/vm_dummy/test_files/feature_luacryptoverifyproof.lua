function hextobytes(str)
   return (str:gsub('..', function (cc)
   return string.char(tonumber(cc, 16))
   end))
end

function verifyProofRaw(data)
   local k = "a6eef7e35abe7026729641147f7915573c7e97b47efa546f5f6e3230263bcb49"
   local v = "2710"
   local p0 = "f871a0379a71a6fb36a75e085aff02beec9f5934b9648d24e2901da307492219608b3780a006a684f73e33f5c18739fd1339977f6fe328eb5cbe64239244b0cec88744355180808080a023866491ea0336f72e659c2a7daf61285de093b04fa353c48069a807c2ba845f808080808080808080"
   local p1 = "e5a03eb5be412f275a18f6e4d622aee4ff40b21467c926224771b782d4c095d1444b83822710"
   local b = crypto.verifyProof(hextobytes(k), hextobytes(v), crypto.keccak256(hextobytes(p0)), hextobytes(p0), hextobytes(p1))
   return b
end

function verifyProofHex(data)
   local k = "0xa6eef7e35abe7026729641147f7915573c7e97b47efa546f5f6e3230263bcb49"
   local v = "0x2710"
   local p0 = "0xf871a0379a71a6fb36a75e085aff02beec9f5934b9648d24e2901da307492219608b3780a006a684f73e33f5c18739fd1339977f6fe328eb5cbe64239244b0cec88744355180808080a023866491ea0336f72e659c2a7daf61285de093b04fa353c48069a807c2ba845f808080808080808080"
   local p1 = "0xe5a03eb5be412f275a18f6e4d622aee4ff40b21467c926224771b782d4c095d1444b83822710"
   local b = crypto.verifyProof(k, v, crypto.keccak256(p0), p0, p1)
   return b
end

abi.register(verifyProofRaw, verifyProofHex)