// SPDX-License-Identifier: BUSL
pragma solidity ^0.8.0;

abstract contract Jackal {
    event PostedFile(address sender, string merkle, uint64 size);

    function postFile(string memory merkle, uint64 filesize) public payable {
        require(msg.sender != address(0), "Invalid sender address");

        uint64 fs = filesize;
        if (fs  <= 1024 * 1024) {
            fs = 1024 * 1024;
        }

        require(msg.value == 5000000 * fs, "Incorrect payment amount");

        emit PostedFile(msg.sender, merkle, filesize);
    }
}
