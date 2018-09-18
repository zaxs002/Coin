package model

import (
	"BitCoin/cache"
	"BitCoin/utils"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"strings"
	"sync"
	"time"
)

type OtcBtcMessage struct {
	Event string            `json:"event"`
	Data  map[string]string `json:"data"`
}

//OtcBtc
type OtcBtcExchange struct {
	Exchange
}

func (ge OtcBtcExchange) CheckCoinExist(symbol string) bool {
	return true
}

var otcbtcGetPrice sync.Once

func (ge OtcBtcExchange) GetPrice() {
	otcbtcGetPrice.Do(func() {
		all := cache.GetInstance().HGetAll(ge.Name + "-symbols")
		r, _ := all.Result()

		o2n := make(map[string]string)

		for k, v := range r {
			o2n[v] = k
		}

		utils.StartTimer(time.Minute*30, func() {
			utils.GetInfoWS2("wss://ws-ap1.pusher.com/app/a6f0d8b7baa8bdb7c392?protocol=7&version=4.2.1&flash=false",
				nil,
				func(ws *websocket.Conn) {
					ws.WriteJSON(OtcBtcMessage{
						Event: "pusher:subscribe",
						Data:  map[string]string{"channel": "market-global"},
					})
				},
				func(result gjson.Result) {
					if result.Get("event").String() == "tickers" {
						s := result.Get("data").String()
						v := gjson.Parse(s)
						v.ForEach(func(key, value gjson.Result) bool {
							k := key.String()
							symbol := o2n[k]
							last := value.Get("last").Float()
							cache.GetInstance().HSet(ge.Name, symbol, last)
							return true
						})
					}
				})
		})
	})
}

func (ge *OtcBtcExchange) Run(symbol string) {
	ge.SetTradeFee(0.0005, 0.0005)

	utils.StartTimer(time.Minute*30, func() {
		u := "https://bb.otcbtc.com/api/v2/markets"

		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				s := value.Get("id").String()
				tickerId := value.Get("ticker_id").String()
				m := utils.GetSymbolByOtcbtc(tickerId)
				coin := m["coin"]
				base := m["base"]
				symbol := coin + "-" + base
				ge.SetSymbol(symbol, s)
				return true
			})
		})
		ge.GetPrice()
	})

	utils.StartTimer(time.Minute*30, func() {
		u := "https://otcbtc.com/fee"

		utils.GetHtml("GET", u, nil, func(result *goquery.Document) {
			result.Find("#fee-style > table > tbody > tr").Each(func(i int, selection *goquery.Selection) {
				currency := selection.Find("td.text-left.currency-info").Text()
				fee := selection.Find("td.text-center.withdrawals-fee").Text()
				currency = strings.TrimSpace(currency)
				currency = strings.ToLower(currency)
				fee = strings.TrimSpace(fee)
				f := utils.GetFloatByBitfinex(fee)

				ge.SetCurrency(currency)
				info := NewTransferInfo()
				num := f["num"]
				if num == "" {
					info.CanWithdraw = 0
				} else {
					info.WithdrawFee = num
					info.CanWithdraw = 1
				}
				ge.SetTransferFee(currency, info)
			})
		})
	})
}

func (ge OtcBtcExchange) FeesRun() {
}

func NewOtcBtcExchange() BigE {
	exchange := new(OtcBtcExchange)
	exchange.Exchange = Exchange{
		Name: "OtcBtc",
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
