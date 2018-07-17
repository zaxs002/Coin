package model

import (
	"BitCoin/utils"
	"strings"
	"BitCoin/cache"
	"github.com/tidwall/gjson"
	"time"
	"github.com/gorilla/websocket"
	"BitCoin/event"
)

type ExxMessage struct {
	Channel  string `json:"channel"`
	Event    string `json:"event"`
	Binary   string `json:"binary"`
	IsZip    string `json:"isZip"`
	LastTime string `json:"lastTime"`
}
type ExxMessage2 struct {
	Action   string `json:"action"`
	DataSize int64  `json:"dataSize"`
	DataType string `json:"dataType"`
}

//Exx
type ExxExchange struct {
	Exchange
}

func (ge ExxExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (ge ExxExchange) GetPrice(s string) {
	all := cache.GetInstance().HGetAll(ge.Name + "-symbols")
	result, _ := all.Result()

	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfoWS2("wss://kline.exx.com/websocket",
			nil,
			func(ws *websocket.Conn) {
				for k, v := range result {
					ws.WriteJSON(ExxMessage{
						Binary:   "false",
						Channel:  k + "_" + v + "_kline_1min",
						Event:    "addChannel",
						IsZip:    "false",
						LastTime: "1530708540000",
					})
				}
			},
			func(result gjson.Result) {
				channel := result.Get("channel").String()
				m := utils.GetCoinByExx(channel)
				symbol := m["coin"]
				last := result.Get("datas.data.0.4").Float()
				cache.GetInstance().HSet(ge.Name, symbol, last)
			})
	})
}

func (ge *ExxExchange) Run(symbol string) {
	event.Bus.Subscribe(ge.Name+"-getprice", ge.GetPrice)

	cache.GetInstance().HSet(ge.Name+"-tradeFee", "taker", 0.001)
	cache.GetInstance().HSet(ge.Name+"-tradeFee", "maker", 0.001)

	//获取currency和转账费
	utils.StartTimer(time.Minute*30, func() {
		u := "https://main.exx.com/api/web/V1_0/getCoinMaps"

		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("datas").ForEach(func(key, value gjson.Result) bool {
				currency := key.String()
				currency = strings.ToLower(currency)
				IsPayOut := value.Get("isPayOut").Bool()
				if IsPayOut {
					fee := value.Get("minFees").Float()
					cache.GetInstance().HSet(ge.Name+"-transfer", currency, fee)
				} else {
					cache.GetInstance().HSet(ge.Name+"-transfer", currency, -1)
				}
				cache.GetInstance().HSet(ge.Name+"-currency", currency, currency)
				return true
			})
		})
	})

	////获取symbols
	//utils.StartTimer(time.Minute*30, func() {
	//	u := "https://api.exx.com/data/v1/markets"
	//	utils.GetInfo("GET", u, nil, func(result gjson.Result) {
	//		result.ForEach(func(key, value gjson.Result) bool {
	//			fmt.Println(key)
	//			return true
	//		})
	//	})
	//})

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfoWS2("wss://ws.exx.com/websocket", nil,
			func(ws *websocket.Conn) {
				ws.WriteJSON(ExxMessage2{
					DataType: "EXX_MARKET_LIST_ALL",
					DataSize: 1,
					Action:   "ADD",
				})
			},
			func(result gjson.Result) {
				result.Get("market").ForEach(func(key, value gjson.Result) bool {
					base := value.Get("1").String()
					coin := value.Get("2").String()
					symbol := coin + base
					cache.GetInstance().HSet(ge.Name+"-symbols", symbol, base)
					return true
				})
				event.Bus.Publish(ge.Name+"-getprice", "")
				event.Bus.Unsubscribe(ge.Name+"-getprice", ge.GetPrice)
			})
	})
}

func (ge ExxExchange) FeesRun() {
}

func NewExxExchange() BigE {
	exchange := new(ExxExchange)
	exchange.Exchange = Exchange{
		Name: "Exx",
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
