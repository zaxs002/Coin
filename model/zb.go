package model

import (
	"BitCoin/utils"
	"time"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"github.com/tidwall/gjson"
	"sync"
	"BitCoin/cache"
	"github.com/gorilla/websocket"
)

type ZbMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
}

//zb
type ZbExchange struct {
	Exchange
}

func (ue ZbExchange) CheckCoinExist(symbol string) bool {
	return true
}

var zbGetPrice sync.Once

func (ue ZbExchange) GetPrice() {
	zbGetPrice.Do(func() {
		utils.StartTimer(time.Minute*30, func() {
			all := cache.GetInstance().HGetAll(ue.Name + "-symbols")
			aa, _ := all.Result()

			o2n := make(map[string]string)

			utils.GetInfoWS2("wss://api.zb.com:9999/websocket", nil,
				func(ws *websocket.Conn) {
					for k, v := range aa {
						ws.WriteJSON(ZbMessage{
							Event:   "addChannel",
							Channel: v + "_ticker",
						})
						o2n[v] = k
					}
				},
				func(result gjson.Result) {
					success := result.Get("success").Bool()
					if success {
						result.Get("data").ForEach(func(key, value gjson.Result) bool {
							m := utils.GetCoinByZb(key.String())
							base := m["base"]
							coin := m["coin"]
							s := base + coin
							symbol := base + "-" + coin
							ue.SetSymbol(symbol, s)
							return true
						})
					} else {
						channel := result.Get("channel").String()
						last := result.Get("ticker.last").Float()
						symbol := utils.GetCoinByZb2(channel)["symbol"]
						s := o2n[symbol]

						ue.SetPrice(s, last)
					}
				})
		})
	})
}

func (ue *ZbExchange) Run(symbol string) {
	ue.SetTradeFee(0.001, 0.001)

	//获取currency和转账费
	utils.StartTimer(time.Minute*30, func() {
		utils.GetHtml("GET", "https://www.zb.cn/i/rate", nil, func(result *goquery.Document) {
			trs := result.Find("body > div.ch-body > div.envor-content > section.envor-section > div > div > div > article > table > tbody > tr")
			trs.Each(func(i int, selection *goquery.Selection) {
				currency := selection.Find("td:nth-child(1)").Text()
				transferFee := selection.Find("td:nth-child(7)").Text()
				singleNum := selection.Find("td:nth-child(8)").Text()
				dayNum := selection.Find("td:nth-child(9)").Text()
				currency = strings.TrimSpace(currency)
				currency = strings.ToLower(currency)

				singleNum = strings.TrimSpace(singleNum)
				singleNum = strings.ToLower(singleNum)
				singleNum = strings.Replace(singleNum, ",", "", -1)

				dayNum = strings.TrimSpace(dayNum)
				dayNum = strings.ToLower(dayNum)
				dayNum = strings.Replace(dayNum, ",", "", -1)

				m := utils.GetByZb(transferFee)
				m2 := utils.GetByZb(singleNum)
				m3 := utils.GetByZb(dayNum)
				info := NewTransferInfo()
				info.CanWithdraw = 0
				if currency != "" {
					n := m["num"]
					n2 := m2["num"]
					n3 := m3["num"]
					if n == "" {
						n = "-1"
					}
					if n2 == "" {
						n2 = "-1"
					}
					if n3 == "" {
						n3 = "-1"
					}
					info.CanWithdraw = 1
					info.WithdrawFee = n
					info.MaxWithdraw = n2
					info.MaxDayWithdraw = n3
					ue.SetCurrency(currency)
					ue.SetTransferFee(currency, info)
				}
			})
		})
	})

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		u := "http://api.zb.com/data/v1/markets"

		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				m := utils.GetCoinByZb(key.String())
				base := m["base"]
				coin := m["coin"]
				s := coin + base
				newSymbol := coin + "-" + base
				ue.SetSymbol(newSymbol, s)
				return true
			})
		})
		ue.GetPrice()
	})

}

func (ue ZbExchange) FeesRun() {
}

func NewZbExchange() BigE {
	exchange := new(ZbExchange)
	exchange.Exchange = Exchange{
		Name: "Zb",
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
