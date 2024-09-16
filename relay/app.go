package relay

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/rs/zerolog/log"

	"github.com/JackalLabs/mulberry/config"
	"github.com/JackalLabs/mulberry/jackal/uploader"
	jWallet "github.com/JackalLabs/mulberry/jackal/wallet"
	"github.com/cosmos/go-bip39"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/viper"
)

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Level(zerolog.DebugLevel)
}

func (a *App) ListenToNetwork(network config.NetworkConfig, wg *sync.WaitGroup) {
	log.Printf("Now listening to %s...", network.Name)

	jackalContract := a.cfg.JackalConfig.Contract

	// Specify the contract address
	contractAddress := common.HexToAddress(network.Contract)
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	// Subscribe to the logs
	var sub ethereum.Subscription
	var logs chan types.Log

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	stopped := false
	for !stopped {

		client, err := ethclient.Dial(network.RPC)
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

		log.Print("Ready to listen!")

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
					log.Print("Exiting...")
					stopped = true
					return
				case err := <-sub.Err():
					log.Printf("Subscription error, reconnecting: %v", err)
					return // Break out of the loop to retry
				case vLog := <-logs:
					log.Printf("Log received: %s", vLog.Address.Hex())

					go func() { // run async
						// Ensure transaction is confirmed
						err := waitForReceipt(client, vLog.TxHash, network.Finality, func(receipt *types.Receipt) {
							for _, l := range receipt.Logs {
								if l.Address.Hex() == contractAddress.Hex() {
									handleLog(l, a.w, a.q, network.ChainID, jackalContract)
								}
							}
						})
						if err != nil {
							log.Printf("error getting reciept for tx %s: %v", vLog.TxHash.String(), err)
						}
					}()

				}
			}
		}()

	}
	wg.Done()
}

func (a *App) Start() error {
	a.q.Listen()

	var wg sync.WaitGroup

	for _, networkConfig := range a.cfg.NetworksConfig {
		wg.Add(1)
		go a.ListenToNetwork(networkConfig, &wg)
	}

	wg.Wait()

	return nil
}

func MakeApp(homePath string) (*App, error) {
	_, err := os.Stat(homePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(homePath, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("cannot make the home directory at %s | %w", homePath, err)
			}
		} else {
			return nil, fmt.Errorf("something is wrong with the home directory | %w", err)
		}
	}

	configPath := path.Join(homePath, "config.yaml")
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	_, err = os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			defaultConfig, err := config.DefaultConfig().Export()
			if err != nil {
				return nil, fmt.Errorf("cannot export the default config | %w", err)
			}
			err = os.WriteFile(configPath, defaultConfig, os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("cannot write the default config | %w", err)
			}
		} else {
			return nil, fmt.Errorf("something is wrong with the config file in the home directory | %w", err)
		}
	}

	err = viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot read the config at %s | %w", configPath, err)
	}

	var cfg config.Config

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal the config | %w", err)
	}

	initLogger()

	seedFile := cfg.JackalConfig.SeedFile
	seedPath := path.Join(homePath, seedFile)

	_, err = os.Stat(seedPath)
	if err != nil {
		if os.IsNotExist(err) {
			entropy, err := bip39.NewEntropy(256)
			if err != nil {
				return nil, fmt.Errorf("cannot generate entropy | %w", err)
			}

			mnemonic, err := bip39.NewMnemonic(entropy)
			if err != nil {
				return nil, fmt.Errorf("cannot generate seed phrase | %w", err)
			}

			err = os.WriteFile(seedPath, []byte(mnemonic), os.ModePerm)
			if err != nil {
				return nil, fmt.Errorf("cannot write seed file | %w", err)
			}

			fmt.Printf("You have just generated a new seed phrase for this relay at %s", seedPath)
		} else {
			return nil, fmt.Errorf("something is wrong with the seed file | %w", err)
		}
	}

	seedData, err := os.ReadFile(seedPath)
	if err != nil {
		return nil, fmt.Errorf("cannot find seed file | %w", err)
	}

	seed := string(seedData)

	seed = strings.TrimSpace(seed)

	w, err := jWallet.CreateWallet(seed, "m/44'/118'/0'/0/0", walletTypes.ChainConfig{
		Bech32Prefix:  "jkl",
		RPCAddr:       cfg.JackalConfig.RPC,
		GRPCAddr:      cfg.JackalConfig.GRPC,
		GasPrice:      "0.02ujkl",
		GasAdjustment: 1.5,
	})
	if err != nil {
		return nil, err
	}

	q := uploader.NewQueue(w)

	app := App{
		w:   w,
		q:   q,
		cfg: cfg,
	}

	return &app, nil
}

func subscribeLogs(client *ethclient.Client, query ethereum.FilterQuery) (ethereum.Subscription, chan types.Log, error) {
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	return sub, logs, err
}

// waitForReceipt polls for the transaction receipt until it's available
func waitForReceipt(client *ethclient.Client, txHash common.Hash, finality int64, callBack func(receipt *types.Receipt)) error {
	var errCount int64
	for {
		if errCount > 10 {
			return errors.New("cannot get receipt")
		}

		time.Sleep(1 * time.Second)

		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			errCount++
			continue
		}

		latestBlock, err := client.BlockByNumber(context.Background(), nil)
		if err != nil {
			errCount++
			continue
		}

		latest := latestBlock.Number().Int64()
		txBlock := receipt.BlockNumber.Int64()

		blockDiff := latest - txBlock
		if blockDiff >= finality {
			callBack(receipt)
			return nil
		} else {
			log.Printf("Still waiting %d more blocks for %s...", finality-blockDiff, txHash.String())
		}
	}
}

func (a *App) Address() string {
	return a.w.AccAddress()
}
