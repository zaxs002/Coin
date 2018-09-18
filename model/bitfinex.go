package model

import (
	"fmt"
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"BitCoin/cache"
	"strings"
	"sync"
	"github.com/gorilla/websocket"
	"github.com/PuerkitoBio/goquery"
	"strconv"
)

//TODO
//Bitfinex
type BitfinexMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Pair    string `json:"pair"`
}
type BitfinexHeartBeat struct {
	Event string `json:"event"`
}
type BitfinexExchange struct {
	Exchange
}

var bitfinexGetPrice sync.Once
var bitfinexGetTransfer sync.Once

func (he BitfinexExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he BitfinexExchange) GetPrice() {
	bitfinexGetPrice.Do(func() {
		all := cache.GetInstance().HGetAll(he.Name + "-symbols")
		result, _ := all.Result()

		pairId := D{}

		u := "wss://api.bitfinex.com/ws/"

		utils.GetInfoWS3(u, nil,
			func(ws *websocket.Conn) {
				for _, v := range result {
					ws.WriteJSON(BitfinexMessage{
						Event:   "subscribe",
						Channel: "ticker",
						Pair:    v,
					})
				}
			},
			func(ws *websocket.Conn, result gjson.Result) {
				e := result.Get("event").String()

				if e == "subscribed" {
					chanid := result.Get("chanId").String()
					pair := result.Get("pair").String()
					pairId[chanid] = pair
				} else if e == "pong" {
				} else if e == "info" {
				} else {
					if result.Get("#").Int() == 2 {
						ws.WriteJSON(BitfinexHeartBeat{
							Event: "ping",
						})
					} else {
						chanid := result.Get("0").String()
						pair := pairId[chanid]
						price := result.Get("7").Float()
						if pair != nil {
							i := pair.(string)
							i = strings.ToLower(i)
							m := utils.GetSymbolByBitfinex(i)
							coin := m["coin"]
							base := m["base"]
							s := coin + "-" + base
							cache.GetInstance().HSet(he.Name, s, price)
						}
					}
				}
			})
	})
}

func (he BitfinexExchange) GetTransfer() {
	bitfinexGetTransfer.Do(func() {
		utils.StartTimer(time.Minute*30, func() {
			u := "https://www.bitfinex.com/fees"

			utils.GetHtml("GET", u, nil, func(result *goquery.Document) {
				result.Find("#fees-page > div > table:nth-child(16) > tbody >tr").Each(func(i int, selection *goquery.Selection) {
					key := selection.Find("span").Text()
					key = strings.Replace(key, " ", "", -1)
					key = strings.ToLower(key)
					short := cache.GetInstance().HGet("FullToShort", key).Val()
					if short == "" {
						short = cache.GetInstance().HGet("FullToShort", key).Val()
					}
					val := selection.Find("td.bfx-green-text.col-info").Text()
					val = strings.TrimSpace(val)
					bitfinex := utils.GetFloatByBitfinex(val)
					info := NewTransferInfo()
					if len(bitfinex) == 0 {
						info.WithdrawFee = 0
						info.CanWithdraw = 1
						if short != "" {
							he.SetTransferFee(short, info)
						}
					} else {
						coin := bitfinex["coin"]
						num := bitfinex["num"]
						n, _ := strconv.ParseFloat(num, 64)
						coin = strings.ToLower(coin)

						info.WithdrawFee = n
						info.CanWithdraw = 1
						he.SetTransferFee(short, info)
					}
				})
			})
		})
	})
}

func (he BitfinexExchange) Run(symbol string) {
	he.SetTradeFee(0.002, 0.001)

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		symbolsUrl := "https://api.bitfinex.com/v1/symbols"
		utils.GetInfo("GET", symbolsUrl, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				s := value.String()
				m := utils.GetSymbolByBitfinex(s)
				coin := m["coin"]
				base := m["base"]

				symbol := coin + "-" + base

				he.SetSymbol(symbol, s)
				return true
			})
		})
		he.GetPrice()
	})

	utils.StartTimer(time.Minute*30, func() {
		//获取currency
		currencyUrl := "https://www.bitfinex.com/account/_bootstrap/"

		utils.GetInfo("GET", currencyUrl, nil, func(result gjson.Result) {
			currencies := result.Get("all_currencies")
			currencies.ForEach(func(key, value gjson.Result) bool {
				s := value.String()

				he.SetCurrency(s)
				return true
			})

			result.Get("nice_ccy_names").ForEach(func(key, value gjson.Result) bool {
				val := value.String()
				k := key.String()
				val = strings.Replace(val, " ", "", -1)
				val = strings.ToLower(val)
				k = strings.ToLower(k)
				cache.GetInstance().HSet("FullToShort", val, k)
				return true
			})
		})

		he.GetTransfer()
	})

	//检测价格是否全部获取完成
	utils.StartTimerWithFlag(time.Second, he.Name, func() {
		he.check(he.Name)
	})
}

func (he BitfinexExchange) FeesRun() {
	fmt.Println("Old FeesRun")
}

func NewBitfinexExchange() BigE {
	exchange := new(BitfinexExchange)

	exchange.Exchange = Exchange{
		Name: "Bitfinex",
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
