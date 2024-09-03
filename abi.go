package main

import (
	_ "embed"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
	"strings"
)

var ABI string = `[
    {
      "type": "function",
      "name": "dispatchEvent",
      "inputs": [
        {
          "name": "value",
          "type": "uint256",
          "internalType": "uint256"
        }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "event",
      "name": "Dispatch",
      "inputs": [
        {
          "name": "sender",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "value",
          "type": "uint256",
          "indexed": false,
          "internalType": "uint256"
        },
        {
          "name": "message",
          "type": "string",
          "indexed": false,
          "internalType": "string"
        }
      ],
      "anonymous": false
    }
  ]`

func handleLog(vLog types.Log) {
	contractAbi, err := abi.JSON(strings.NewReader(ABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	event := struct {
		Sender  string
		Value   *big.Int
		Message string
	}{}

	err = contractAbi.UnpackIntoInterface(&event, "Dispatch", vLog.Data)
	if err != nil {
		log.Fatalf("Failed to unpack log data: %v", err)
	}

	log.Printf("Event details: %+v", event)
}
