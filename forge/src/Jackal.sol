// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import {Strings} from "@openzeppelin/contracts/utils/Strings.sol";

abstract contract Jackal {
    event PostedFile(address sender, string merkle, uint64 size);

    function getPrice() public view virtual returns(int);

    function getStoragePrice(uint64 filesize) public view returns (uint256) {
        uint256 price = uint256(getPrice());

        uint256 storagePrice = 15; // price at 8 decimal places
        uint256 multiplier = 2;
        uint256 months = 200 * 12;

        uint256 fs = filesize;
        if (fs  <= 1024 * 1024) {
            fs = 1024 * 1024; // minimum file size of one MB for pricing
        }

        // Calculate the price in wei
        // 1e8 adjusts for the 8 decimals of USD, 1e18 converts ETH to wei
        uint256 BSM = storagePrice * multiplier * months * fs;
        uint256 p = (BSM * 1e8 * 1e18) / (price * 1099511627776);

        if (p == 0) {
            p = 5000 gwei;
        }

        return p;
    }

    function postFile(string memory merkle, uint64 filesize) public payable{
        require(msg.sender != address(0), "Invalid sender address");

        uint256 pE = getStoragePrice(filesize);

        require(msg.value >= pE, string.concat("Incorrect payment amount, need ", Strings.toString(pE), " wei"));

        emit PostedFile(msg.sender, merkle, filesize);
    }
}
