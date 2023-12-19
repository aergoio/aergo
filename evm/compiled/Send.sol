// SPDX-License-Identifier: MIT
pragma solidity ^0.8.10;

contract EtherTransfer {
    address payable public recipient;

    event EtherTransferred(
        address indexed from,
        address indexed to,
        uint256 amount
    );

    function transferEther(
        address payable _recipient,
        uint256 amount
    ) external payable {
        require(amount > 0, "Invalid amount");

        // 수령자에게 이더를 전송합니다.
        _recipient.transfer(amount);

        // 이벤트를 발생시켜 전송 내역을 기록합니다.
        emit EtherTransferred(msg.sender, _recipient, msg.value);
    }
}
