pragma solidity ^0.4.0;

contract Contract {

    uint32 public perm;

    function Contract(uint32 input){
        perm = input;
    }

    function storageTest(address add) returns (uint32, address){
        perm = perm+1;
        return (perm, add);
    }
}