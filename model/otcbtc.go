package model

import (
	"BitCoin/utils"
	"strings"
	"BitCoin/cache"
	"github.com/tidwall/gjson"
	"github.com/gorilla/websocket"
	"time"
	"github.com/PuerkitoBio/goquery"
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
func (ge OtcBtcExchange) GetPrice(s string) {
}

func (ge *OtcBtcExchange) Run(symbol string) {
	cache.GetInstance().HSet(ge.Name+"-tradeFee", "taker", 0.0005)
	cache.GetInstance().HSet(ge.Name+"-tradeFee", "maker", 0.0005)

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
						last := value.Get("last").Float()
						cache.GetInstance().HSet(ge.Name, key.String(), last)
						return true
					})
				}
			})
	})

	utils.StartTimer(time.Minute*30, func() {
		u := "https://bb.otcbtc.com/api/v2/markets"

		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				s := value.Get("id").String()
				cache.GetInstance().HSet(ge.Name+"-symbols", s, s)
				return true
			})
		})
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
				cache.GetInstance().HSet(ge.Name+"-currency", currency, currency)
				num := f["num"]
				if num == "" {
					cache.GetInstance().HSet(ge.Name+"-transfer", currency, -1)
				} else {
					cache.GetInstance().HSet(ge.Name+"-transfer", currency, num)
				}
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
