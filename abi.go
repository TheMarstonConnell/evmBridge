package main

import (
	"context"
	_ "embed"
	"encoding/hex"
	"evmBridge/jackal/uploader"
	"fmt"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	storageTypes "github.com/jackalLabs/canine-chain/v3/x/storage/types"
	"log"
	"strings"
)

var ABI = `[
    {
      "type": "function",
      "name": "postFile",
      "inputs": [
        {
          "name": "merkle",
          "type": "string",
          "internalType": "string"
        },
        {
          "name": "filesize",
          "type": "uint64",
          "internalType": "uint64"
        }
      ],
      "outputs": [],
      "stateMutability": "payable"
    },
    {
      "type": "event",
      "name": "PostedFile",
      "inputs": [
        {
          "name": "sender",
          "type": "address",
          "indexed": false,
          "internalType": "address"
        },
        {
          "name": "merkle",
          "type": "string",
          "indexed": false,
          "internalType": "string"
        },
        {
          "name": "size",
          "type": "uint64",
          "indexed": false,
          "internalType": "uint64"
        }
      ],
      "anonymous": false
    }
  ]`

var eventABI abi.ABI

func init() {
	var err error
	eventABI, err = abi.JSON(strings.NewReader(ABI))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}
}

func handleLog(vLog *types.Log, w *wallet.Wallet, q *uploader.Queue) {

	event := struct {
		Sender common.Address
		Merkle string
		Size   uint64
	}{}

	e, err := eventABI.Unpack("PostedFile", vLog.Data)
	if err != nil {
		log.Fatalf("Failed to unpack log data: %v", err)
		return
	}

	fmt.Println(len(e))

	err = eventABI.UnpackIntoInterface(&event, "PostedFile", vLog.Data)
	if err != nil {
		log.Fatalf("Failed to unpack log data: %v", err)
		return
	}

	log.Printf("Event details: %+v", event)

	merkleRoot, err := hex.DecodeString(event.Merkle)
	if err != nil {
		log.Fatalf("Failed to decode merkle: %v", err)
		return
	}

	abci, err := w.Client.RPCClient.ABCIInfo(context.Background())
	if err != nil {
		log.Fatalf("Failed to query ABCI: %v", err)
		return
	}

	log.Printf("Relaying for %s\n", event.Sender.String())

	msg := &storageTypes.MsgPostFile{
		Creator:       w.AccAddress(),
		Merkle:        merkleRoot,
		FileSize:      int64(event.Size),
		ProofInterval: 40,
		ProofType:     0,
		MaxProofs:     3,
		Note:          fmt.Sprintf("{\"memo\":\"Relayed from EVM for %s\"}", event.Sender.String()),
	}

	msg.Expires = abci.Response.LastBlockHeight + ((100 * 365 * 24 * 60 * 60) / 6)
	if err := msg.ValidateBasic(); err != nil {
		log.Fatalf("Failed to validate message: %v", err)
		return
	}

	res, err := q.Post(msg)
	if err != nil {
		log.Fatalf("Failed to post message: %v", err)
		return
	}

	if res == nil {
		log.Fatalf("something went wrong, response is empty")
		return
	}

	log.Println(res.RawLog)
	log.Println(res.TxHash)

}
