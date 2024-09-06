// SPDX-License-Identifier: BUSL
pragma solidity ^0.8.26;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {Jackal} from "./Jackal.sol";

contract JackalBridge is Ownable, Jackal {

    address[] public relays;

    constructor(address[] memory _relays) Ownable(msg.sender){
        require (_relays.length > 0, "must provide relays");

        relays = _relays;
    }

    // Modifier to restrict access to owner or relays
    modifier onlyOwnerOrRelay() {
        require(msg.sender == owner() || isRelay(msg.sender), "not owner or relay");
        _;
    }

    function isRelay(address _relay) internal view returns (bool) {
        for (uint i = 0; i < relays.length; i++) {
            if (relays[i] == _relay) {
                return true;
            }
        }
        return false;
    }

    // Function to add a relay, only callable by the owner
    function addRelay(address _relay) public onlyOwner {
        relays.push(_relay);
    }

    // Function to remove a relay, only callable by the owner
    function removeRelay(address _relay) public onlyOwner {
        for (uint i = 0; i < relays.length; i++) {
            if (relays[i] == _relay) {
                relays[i] = relays[relays.length - 1];
                relays.pop();
                break;
            }
        }
    }

    // Function to split the balance: 50% to owner, 50% to relays
    function distributeETH() public onlyOwnerOrRelay {
        uint balance = address(this).balance;
        require(balance > 0, "No ETH to distribute");

        // Calculate 50% for the owner
        uint ownerShare = balance / 2;
        payable(owner()).transfer(ownerShare);

        // If there are relays, split the remaining 50% among them
        if (relays.length > 0) {
            uint relayShare = balance - ownerShare; // Remaining 50%
            uint perRelay = relayShare / relays.length;

            for (uint i = 0; i < relays.length; i++) {
                payable(relays[i]).transfer(perRelay);
            }
        }
    }

    function distributeBalance() public onlyOwnerOrRelay {
        uint256 balance = address(this).balance;
        require(balance > 2, "not enough wei to distribute");

        uint256 ownerShare = balance / 2;
        payable(owner()).transfer(ownerShare);

        // If there are relays, split the remaining 50% among them
        if (relays.length > 0) {
            uint256 relayShare = balance - ownerShare; // Remaining 50%
            uint256 perRelay = relayShare / relays.length;

            for (uint i = 0; i < relays.length; i++) {
                payable(relays[i]).transfer(perRelay);
            }
        }
    }



}