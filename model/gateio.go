package model

import (
	"BitCoin/utils"
	"time"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"BitCoin/cache"
	"github.com/tidwall/gjson"
	"BitCoin/event"
	"github.com/gorilla/websocket"
)

type GateIOMessage struct {
	ID     int64    `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

//gateio
type GateIOExchange struct {
	Exchange
}

func (ge GateIOExchange) CheckCoinExist(symbol string) bool {
	return true
}
func (ge GateIOExchange) GetPrice(s string) {
	all := cache.GetInstance().HGetAll(ge.Name + "-symbols")
	result, _ := all.Result()
	var symbols []string
	for _, v := range result {
		symbols = append(symbols, v)
	}
	utils.GetInfoWS2("wss://ws.gate.io/v3/", nil,
		func(ws *websocket.Conn) {
			ws.WriteJSON(GateIOMessage{
				ID:     12312,
				Method: "ticker.subscribe",
				Params: symbols,
			})
		},
		func(result gjson.Result) {
			symbol := result.Get("params.0").String()
			last := result.Get("params.1.last").Float()
			symbol = strings.ToLower(symbol)
			m := utils.GetCoinByZb(symbol)
			coin := m["coin"]
			base := m["base"]
			cache.GetInstance().HSet(ge.Name, coin+base, last)
		})
}

func (ge *GateIOExchange) Run(symbol string) {

	cache.GetInstance().HSet(ge.Name+"-tradeFee", "taker", 0.002)
	cache.GetInstance().HSet(ge.Name+"-tradeFee", "maker", 0.002)

	event.Bus.Subscribe(ge.Name+"-getprice", ge.GetPrice)
	//获取currency和转账费
	utils.StartTimer(time.Minute*30, func() {
		utils.GetHtml("GET", "https://gateio.io/fee", nil, func(result *goquery.Document) {
			trs := result.Find("#feelist > tbody > tr")
			trs.Each(func(i int, selection *goquery.Selection) {
				currency := selection.Find("td:nth-child(2)").Text()
				transferFee := selection.Find(".fee-withdraw").Text()
				currency = strings.TrimSpace(currency)
				currency = strings.ToLower(currency)
				m := utils.GetCoinByGateIO(transferFee)
				if currency != "" {
					n := m["fee"]
					cache.GetInstance().HSet(ge.Name+"-currency", currency, currency)
					cache.GetInstance().HSet(ge.Name+"-transfer", currency, n)
				}
			})
		})
	})

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		u := "https://data.gateio.io/api2/1/pairs"

		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				s := strings.ToLower(value.String())
				m := utils.GetCoinByZb(s)
				base := m["base"]
				coin := m["coin"]
				cache.GetInstance().HSet(ge.Name+"-symbols", coin+base, value.String())
				return true
			})
		})
		event.Bus.Publish(ge.Name+"-getprice", "")
		event.Bus.Unsubscribe(ge.Name+"-getprice", ge.GetPrice)
	})

}

func (ge GateIOExchange) FeesRun() {
}

func NewGateIOExchange() BigE {
	exchange := new(GateIOExchange)
	exchange.Exchange = Exchange{
		Name: "GateIO",
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
