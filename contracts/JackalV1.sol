pragma solidity ^0.8.26;

import "@openzeppelin/contracts/access/Ownable.sol";

contract JackalBridge {
    event PostedFile(address sender, string merkle, uint64 size);

    address[] public relays;
    address public owner;

    constructor(address[] memory _relays) {
        require (_relays.length > 0, "must provide relays");

        relays = _relays;
    }

    // Function to add a relay, only callable by the owner
    function addRelay(address _relay) public onlyOwner {
        relays.push(_relay);
    }

    // Function to remove a relay, only callable by the owner
    function removeRelay(address _relay) public onlyOwner {
        for (uint i = 0; i < relays.length; i++) {
            if (relays[i] == _relay) {
                relays[i] = relays[relays.length - 1]; // Move last element to current index
                relays.pop(); // Remove the last element
                break;
            }
        }
    }

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