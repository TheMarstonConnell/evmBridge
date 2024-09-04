pragma solidity ^0.8.26;

contract JackalBridge {
    event PostedFile(address sender, string merkle, uint64 size);



    function postFile(string memory merkle, uint64 filesize) public {
        require(msg.sender != address(0), "Invalid sender address");


        emit PostedFile(msg.sender, merkle, filesize);
    }


}