package main

import (
	"context"
	"evmBridge/jackal/uploader"
	jWallet "evmBridge/jackal/wallet"
	"fmt"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/ethereum/go-ethereum"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	w, err := jWallet.CreateWallet(os.Getenv("SEED"), "m/44'/118'/0'/0/0", walletTypes.ChainConfig{
		Bech32Prefix:  "jkl",
		RPCAddr:       os.Getenv("RPC"),
		GRPCAddr:      os.Getenv("GRPC"),
		GasPrice:      "0.02ujkl",
		GasAdjustment: 1.5,
	})

	log.Printf("Address: %s\n", w.AccAddress())

	q := uploader.NewQueue(w)

	q.Listen()

	client, err := ethclient.Dial(os.Getenv("ETH_RPC"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	defer client.Close()

	// Specify the contract address
	contractAddress := common.HexToAddress(os.Getenv("CONTRACT_ADDRESS"))
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	// Subscribe to the logs
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("Failed to subscribe to the logs: %v", err)
	}

	log.Println("Ready to listen!")
	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("Subscription error: %v", err)
		case vLog := <-logs:
			log.Printf("Log received: \n")
			log.Printf("Address: %s\n", vLog.Address.Hex())

			// Ensure transaction is confirmed
			receipt, err := waitForReceipt(client, vLog.TxHash)
			if err != nil {
				fmt.Printf("Failed to get transaction receipt: %v\n", err)
				continue
			}

			if receipt != nil {
				// Process logs if receipt is available
				for _, log := range receipt.Logs {
					if log.Address.Hex() == contractAddress.Hex() {
						handleLog(log, w, q)
					}
				}
			} else {
				fmt.Println("No receipt found for the transaction.")
			}

		}
	}
}

// waitForReceipt polls for the transaction receipt until it's available
func waitForReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil && receipt != nil {

			log.Println(receipt.Status)
			return receipt, nil
		}

		// Wait before retrying
		time.Sleep(1 * time.Second)
	}
}
