package main

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("ws://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	// Specify the contract address
	contractAddress := common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3")
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	// Subscribe to the logs
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to the logs: %v", err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("Subscription error: %v", err)
		case vLog := <-logs:
			log.Printf("Received log: %v", vLog) // Handle the log
			handleLog(vLog)
		}
	}
}
