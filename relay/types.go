package relay

import (
	"github.com/JackalLabs/mulberry/config"
	"github.com/JackalLabs/mulberry/jackal/uploader"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)

type App struct {
	w   *wallet.Wallet
	q   *uploader.Queue
	cfg config.Config
}

var ChainIDS = map[int64]string{
	1:     "Ethereum",
	8453:  "Base",
	137:   "Polygon",
	10:    "OP",
	42161: "Arbitrum",
}

const ConfirmationsNeeded = 12
