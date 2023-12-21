// SPDX-License-Identifier: MIT
pragma solidity ^0.8.10;

contract HelloWorld {
    string public greet;

    // Constructor
    constructor() {
        greet = "Hello, World!";
    }

    function hello() public view returns (string memory) {
        return greet;
    }
}
