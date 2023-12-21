// SPDX-License-Identifier: MIT
pragma solidity ^0.8.10;

contract Transfer {
    function transferAergo(
        address payable _recipient,
        uint256 amount
    ) external payable {
        require(amount > 0, "Invalid amount");
        require(msg.value >= amount, "Insufficient Ether sent");

        // 수령자에게 이더를 전송합니다.
        _recipient.transfer(amount);
    }
}
