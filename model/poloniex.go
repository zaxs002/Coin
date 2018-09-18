package model

import (
	"fmt"
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"strings"
	"sync"
	"github.com/gorilla/websocket"
	"BitCoin/cache"
)

type PoloniexMessage struct {
	Command string `json:"command"`
	Channel string `json:"channel"`
}

//poloniex
type PoloniexExchange struct {
	Exchange
}

func (he PoloniexExchange) CheckCoinExist(symbol string) bool {
	return true
}

var poloniexGetPrice sync.Once

func (he PoloniexExchange) GetPrice() {
	poloniexGetPrice.Do(func() {
		u := "wss://api2.poloniex.com/"
		utils.StartTimer(time.Hour*12, func() {
			all := cache.GetInstance().HGetAll(he.Name + "-symbols")
			r, _ := all.Result()

			id2s := make(map[string]string)

			utils.GetInfoWS3(u, nil,
				func(ws *websocket.Conn) {
					for k, v := range r {
						v = strings.ToUpper(v)
						ws.WriteJSON(PoloniexMessage{
							Channel: v,
							Command: "subscribe",
						})
						id2s[v] = k
					}
				},
				func(ws *websocket.Conn, result gjson.Result) {
					id := result.Get("0").String()
					s := id2s[id]

					o := result.Get("2.0.0").String()
					if o == "o" {
						last := result.Get("2.0.2").Float()

						he.SetPrice(s, last)
					}
				})
		})
	})
}

func (he *PoloniexExchange) Run(symbol string) {
	he.SetTradeFee(0.002, 0.001)
	//获取symbols
	symbolsUrl := "https://www.poloniex.com/public?command=returnTicker"
	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfo("GET", symbolsUrl, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				s := key.String()
				s = strings.ToLower(s)
				id := value.Get("id").String()
				last := value.Get("last").Float()
				he.SetSymbol(s, id)
				he.SetPrice(s, last)
				return true
			})
		})
		he.GetPrice()
	})

	//获取currency
	currencyUrl := "https://poloniex.com/public?command=returnCurrencies"
	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfo("GET", currencyUrl, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				currency := key.String()
				currency = strings.ToLower(currency)

				fee := value.Get("txFee").Float()
				minConf := value.Get("minConf").Float()
				disabled := value.Get("disabled").Bool()
				info := NewTransferInfo()
				info.WithdrawFee = fee
				info.MinWithdraw = minConf
				info.CanWithdraw = 1
				if disabled {
					info.CanWithdraw = 0
				}
				he.SetCurrency(currency)
				he.SetTransferFee(currency, info)
				return true
			})
		})
	})
}

func (he *PoloniexExchange) FeesRun() {
	fmt.Println("Old FeesRun")
}

func NewPoloniexExchange() BigE {
	exchange := new(PoloniexExchange)
	exchange.Exchange = Exchange{
		Name: "Poloniex",
		PriceQueue: LockMap{
			M: make(map[string]float64),
		},
		AmountDict: LockMap{
			M: make(map[string]float64),
		},
		Sub: exchange,
	}
	var duitai BigE = exchange
	return duitai
}
