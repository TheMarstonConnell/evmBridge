package uploader

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	walletTypes "github.com/desmos-labs/cosmos-go-wallet/types"
	"github.com/desmos-labs/cosmos-go-wallet/wallet"
)

type MsgHolder struct {
	m   sdk.Msg
	r   *sdk.TxResponse
	wg  *sync.WaitGroup
	err error
}

type Queue struct {
	messages []*MsgHolder
	w        *wallet.Wallet
	stopped  bool
	jklPrice float64
}

func NewQueue(w *wallet.Wallet) *Queue {
	q := Queue{
		messages: make([]*MsgHolder, 0),
		w:        w,
	}
	return &q
}

func (q *Queue) Stop() {
	q.stopped = true
}

func (q *Queue) Listen() {
	go func() {
		for !q.stopped {
			time.Sleep(time.Millisecond * 1000)
			q.popAndPost(10)
		}
	}()

	go func() {
		for !q.stopped {
			time.Sleep(time.Minute * 10)
			_ = q.UpdateGecko() // updating price oracle every 5 minutes, we don't care about errors tbh
		}
	}()
}

func (q *Queue) popAndPost(count int) {
	if len(q.messages) == 0 {
		return
	}

	if count > len(q.messages) {
		count = len(q.messages)
	}

	msgs := make([]*MsgHolder, count)

	for i := 0; i < count; i++ {
		m := q.messages[0]
		q.messages = q.messages[1:]
		msgs[i] = m
	}

	ms := make([]sdk.Msg, count)
	for i, msg := range msgs {
		ms[i] = msg.m
	}

	data := walletTypes.NewTransactionData(
		ms...,
	).WithGasAuto().WithFeeAuto()

	res, err := q.w.BroadcastTxCommit(data)
	if err != nil {
		fmt.Println(err)
	}
	if res == nil {
		fmt.Println("response is for sure empty")
	}
	for _, msg := range msgs {
		msg.r = res
		msg.err = err
		msg.wg.Done()
	}
}

func (q *Queue) Post(msg sdk.Msg) (*sdk.TxResponse, error) {
	fmt.Println("posting message...")

	var wg sync.WaitGroup
	m := MsgHolder{
		m:  msg,
		wg: &wg,
	}
	wg.Add(1)

	q.messages = append(q.messages, &m)

	fmt.Println("waiting...")

	wg.Wait()

	return m.r, nil
}

type GeckoRes struct {
	JackalPrice Price `json:"jackal-protocol"`
}

type Price struct {
	USDPrice float64 `json:"usd"`
}

func (q *Queue) UpdateGecko() error {
	const u = "https://api.coingecko.com/api/v3/simple/price?ids=jackal-protocol&vs_currencies=usd"
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var priceResp GeckoRes
	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return err
	}

	q.jklPrice = priceResp.JackalPrice.USDPrice

	return nil
}

func (q *Queue) GetCost(kbs int64, hours int64) int64 {
	pricePerTBPerMonth := 15.0 * float64(kbs) * float64(hours)

	quantifiedPricePerTBPerMonth := pricePerTBPerMonth / 3.0
	pricePerGbPerMonth := quantifiedPricePerTBPerMonth / 1000.0
	pricePerMbPerMonth := pricePerGbPerMonth / 1000.0
	pricePerKbPerMonth := pricePerMbPerMonth / 1000.0
	pricePerHour := pricePerKbPerMonth / 720.0

	ujklUnit := 1000000.0
	jklCost := pricePerHour / q.jklPrice

	ujklCost := jklCost * ujklUnit

	return int64(ujklCost)
}
