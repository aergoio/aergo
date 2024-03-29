function get(a)
    return crypto.sha256(a)
end

function checkEther()
    return crypto.ecverify("0xce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008",
        "0x90f27b8b488db00b00606796d2987f6a5f59ae62ea05effe84fef5b8b0e549984a691139ad57a3f0b906637673aa2f63d1f55cb1a69199d4009eea23ceaddc9301",
        "0xbcf9061f21320aa7e824b00d0152398b2d7a6e44")
end

function checkAergo()
    return crypto.ecverify("11e96f2b58622a0ce815b81f94da04ae7a17ba17602feb1fd5afa4b9f2467960",
        "304402202e6d5664a87c2e29856bf8ff8b47caf44169a2a4a135edd459640be5b1b6ef8102200d8ea1f6f9ecdb7b520cdb3cc6816d773df47a1820d43adb4b74fb879fb27402",
        "AmPbWrQbtQrCaJqLWdMtfk2KiN83m2HFpBbQQSTxqqchVv58o82i")
end

function keccak256(s)
    return crypto.keccak256(s)
end

abi.register(get, checkEther, checkAergo, keccak256)
