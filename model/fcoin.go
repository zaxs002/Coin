package model

import (
	"BitCoin/cache"
	"github.com/tidwall/gjson"
	"BitCoin/utils"
	"time"
	"sync"
	"github.com/gorilla/websocket"
	"net/http"
	"fmt"
)

var once sync.Once
var once2 sync.Once

type FcoinExchange struct {
	Exchange
}

type FcoinMessage struct {
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
}

func (he FcoinExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he FcoinExchange) FeesRun() {
}

func (he FcoinExchange) GetTransfer() {
	once.Do(func() {
		//获取转账手续费
		token := "CsdpoJ4ujZdwfnFJ10znw-5cnUBuVIrpwEk52IYWW3rnAWz3BKMdY8oDr0C9I52iQfqctB9XyjGUtxCglnOJ7g=="
		agent := "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"

		transferFeeUrl := "https://exchange.fcoin.com/api/web/v1/accounts/withdraw/config"

		headers := http.Header{
			"token":      []string{token},
			"user-agent": []string{agent},
		}

		info := NewTransferInfo()
		utils.GetInfo("GET", transferFeeUrl, headers, func(result gjson.Result) {
			fmt.Println(result)

			result.Get("data").ForEach(func(key, value gjson.Result) bool {
				currency := value.Get("currency").String()
				fee := value.Get("fees").Float()
				minAmount := value.Get("single_min_amount").Float()
				maxAmount := value.Get("single_max_amount").Float()

				info.WithdrawFee = fee
				info.MinWithdraw = minAmount
				info.MaxWithdraw = maxAmount
				info.CanWithdraw = 1
				if minAmount <= 0 {
					info.CanWithdraw = 0
				}
				he.SetTransferFee(currency, info)
				return true
			})
		})
	})
}

func (he FcoinExchange) GetPrice() {
	once2.Do(func() {
		//获取价格
		all := cache.GetInstance().HGetAll(he.Name + "-symbols")
		result, _ := all.Result()

		o2n := make(map[string]string)

		u := "wss://api.fcoin.com/v2/ws"
		utils.GetInfoWS2(u, nil,
			func(ws *websocket.Conn) {
				for k, v := range result {
					ws.WriteJSON(FcoinMessage{
						Cmd:  "sub",
						Args: []string{"ticker." + v},
					})
					o2n[v] = k
				}
			}, func(result gjson.Result) {
				last := result.Get("ticker.0").Float()
				symbol := result.Get("type").String()
				m := utils.GetSymbolByFcoin(symbol)
				if m != nil {
					if m["type"] == "ticker" {
						s := m["symbol"]
						n := o2n[s]
						he.SetPrice(n, last)
					}
				}
			})
	})
}

func (he FcoinExchange) Run(symbol string) {
	he.SetTradeFee(0.001, 0.001)

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		symbolsUrl := "https://api.fcoin.com/v2/public/symbols"

		utils.GetInfo("GET", symbolsUrl, nil, func(result gjson.Result) {
			result.Get("data").ForEach(func(key, value gjson.Result) bool {
				symbol := value.Get("name").String()
				base := value.Get("quote_currency").String()
				coin := value.Get("base_currency").String()
				s := coin + "-" + base
				he.SetSymbol(s, symbol)
				return true
			})
		})
		he.GetPrice()
	})

	//获取currency
	utils.StartTimer(time.Minute*30, func() {
		currencyUrl := "https://api.fcoin.com/v2/public/currencies"
		utils.GetInfo("GET", currencyUrl, nil, func(result gjson.Result) {
			result.Get("data").ForEach(func(key, value gjson.Result) bool {
				currency := value.String()
				he.SetCurrency(currency)
				return true
			})
		})
		he.GetTransfer()
	})
}

func NewFcoinExchange() BigE {
	exchange := new(FcoinExchange)

	exchange.Exchange = Exchange{
		Name: "Fcoin",
		PriceQueue: LockMap{
			M: make(map[string]float64),
		},
		AmountDict: LockMap{
			M: make(map[string]float64),
		},
		TradeFees: LockMap{
			M: make(map[string]float64),
		},
		TransferFees: LockMapString{
			M: make(map[string]string),
		},
		Sub: exchange,
	}

	var duitai BigE = exchange
	return duitai
}
