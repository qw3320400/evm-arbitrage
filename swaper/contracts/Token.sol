// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract Token is ERC20 {
    constructor(string memory symbol, uint256 amount) ERC20(symbol, symbol)  {
        _mint(msg.sender, amount);
    }
}