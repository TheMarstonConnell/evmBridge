package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JackalLabs/mulberry/jackal/uploader"
	jWallet "github.com/JackalLabs/mulberry/jackal/wallet"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/ethereum/go-ethereum"
	"github.com/joho/godotenv"

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
	if err != nil {
		log.Fatal(err)
		return
	}

	jackalContract := os.Getenv("JACKAL_CONTRACT")

	log.Printf("Address: %s\n", w.AccAddress())

	q := uploader.NewQueue(w)

	q.Listen()

	// Specify the contract address
	contractAddress := common.HexToAddress(os.Getenv("CONTRACT_ADDRESS"))
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	// Subscribe to the logs
	var sub ethereum.Subscription
	var logs chan types.Log

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {

		client, err := ethclient.Dial(os.Getenv("ETH_RPC"))
		if err != nil {
			log.Printf("Failed to connect to the Ethereum client, retrying in 5 seconds: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		sub, logs, err = subscribeLogs(client, query)
		if err != nil {
			log.Printf("Failed to subscribe, retrying in 5 seconds: %v", err)
			client.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("Ready to listen!")

		// Listening loop
		func() {
			defer func() {
				// Unsubscribe and close client on exit
				if sub != nil {
					sub.Unsubscribe()
				}
				client.Close()
			}()

			for {
				select {
				case <-sigs:
					log.Println("Exiting...")
					return
				case err := <-sub.Err():
					log.Printf("Subscription error, reconnecting: %v", err)
					return // Break out of the loop to retry
				case vLog := <-logs:
					log.Printf("Log received: %s", vLog.Address.Hex())

					// Ensure transaction is confirmed
					receipt, err := waitForReceipt(client, vLog.TxHash)
					if err != nil {
						fmt.Printf("Failed to get transaction receipt: %v\n", err)
						continue
					}

					if receipt != nil {
						// Process logs if receipt is available
						for _, l := range receipt.Logs {
							if l.Address.Hex() == contractAddress.Hex() {
								handleLog(l, w, q, jackalContract)
							}
						}
					} else {
						fmt.Println("No receipt found for the transaction.")
					}
				}
			}
		}()

	}
}

func subscribeLogs(client *ethclient.Client, query ethereum.FilterQuery) (ethereum.Subscription, chan types.Log, error) {
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	return sub, logs, err
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
