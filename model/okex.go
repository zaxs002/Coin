package model

import (
	"BitCoin/cache"
	"BitCoin/utils"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"net/http"
	"strings"
	"sync"
	"time"
)

type OkexMessage struct {
	Event      string                 `json:"event"`
	Channel    string                 `json:"channel"`
	Parameters map[string]interface{} `json:"parameters"`
}

type OkexHeartBeat struct {
	Event string `json:"event"`
}
type OkexExchange struct {
	Exchange
}

func (he OkexExchange) CheckCoinExist(symbol string) bool {
	return true
}

var okexGetPrice sync.Once
var okexGetTransfer sync.Once
var okexGetTrade sync.Once
var okexCheck sync.Once

func (he OkexExchange) GetPrice() {
	okexGetPrice.Do(func() {
		utils.StartTimer(time.Hour*24, func() {
			//websocket
			all := cache.GetInstance().HGetAll(he.Name + "-symbols")
			result, _ := all.Result()

			apiKey := "606ff446-0ba1-4753-9dcd-c56d0953fbe3"
			secretKey := "A91BC888BBB94E7F57D5C1CE6F0D88DC"
			var m = map[string]interface{}{"api_key": apiKey}
			sign := utils.BuildSign(m, secretKey)

			var u = "wss://real.okex.com:10441/websocket"

			utils.GetInfoWS3(u, nil,
				func(ws *websocket.Conn) {
					for _, v := range result {
						channel := "ok_sub_spot_" + v + "_ticker"
						ws.WriteJSON(OkexMessage{
							Event:   "addChannel",
							Channel: channel,
							Parameters: map[string]interface{}{
								"api_key": apiKey, "sign": sign,
							},
						})
					}
					utils.StartTimer(time.Second*30, func() {
						ws.WriteJSON(OkexHeartBeat{
							Event: "ping",
						})
					})
				},
				func(ws *websocket.Conn, result gjson.Result) {
					heartBeat := result.Get("event").String()

					if heartBeat == "" {
						channel := result.Get("0.channel")

						if channel.String() == "addChannel" {
							//c := result.Get("0.data.channel")
							//fmt.Printf("%s订阅成功\n", utils.GetCoinByOkex(c.String()))
						} else {
							bi := utils.GetCoinByOkex(channel.String())
							last := result.Get("0.data.last").Float()
							he.SetPrice(bi, last)
						}
					} else {
					}
				})
			//websocket
		})
	})
}

func (he OkexExchange) GetTransfer(token string) {
	okexGetTransfer.Do(func() {
		utils.StartTimer(time.Minute*30, func() {
			all := cache.GetInstance().HGetAll(he.Name + "-currency")
			aa, _ := all.Result()

			headers := http.Header{
				"authorization": []string{token},
			}
			for k, v := range aa {
				info := NewTransferInfo()

				url := "https://www.okex.com/v2/asset/withdraw?currencyId=" + v

				utils.GetInfo("GET", url, headers, func(result gjson.Result) {
					lower := strings.ToLower(k)

					feeDefault := result.Get("data.feeDefault").Float()
					feeMax := result.Get("data.feeMax").Float()
					feeMin := result.Get("data.feeMin").Float()
					confirmNum := result.Get("data.confirmNum").Float()
					singleMin := result.Get("data.singleMin").Float()
					singleMax := result.Get("data.singleMax").Float()
					canWithdrawAddress := result.Get("data.canWithdrawAddress").Bool()

					info.MinWithdrawFee = feeMin
					info.MaxWithdrawFee = feeMax
					info.WithdrawFee = feeDefault
					info.WithdrawMinConfirmations = confirmNum
					info.MinWithdraw = singleMin
					info.MaxWithdraw = singleMax
					info.CanWithdraw = 0
					if canWithdrawAddress {
						info.CanWithdraw = 1
					}

					he.SetTransferFee(lower, info)
				})
			}

		})
	})
}

func (he OkexExchange) GetTrade(token string) {
	okexGetTrade.Do(func() {
		utils.StartTimer(time.Minute*30, func() {
			headers := http.Header{
				"authorization": []string{token},
			}
			tradeFeeUrl := "https://www.okex.com/v2/spot/user-level"

			utils.GetInfo("GET", tradeFeeUrl, headers, func(result gjson.Result) {
				takerFee := result.Get("data.takerFeeRatio").Float()
				makerFee := result.Get("data.makerFeeRatio").Float()

				takerFee = takerFee / 100
				makerFee = makerFee / 100

				he.SetTradeFee(takerFee, makerFee)
			})
		})
	})
}

func (he *OkexExchange) Run(symbol string) {
	cache.GetInstance().HDel("Flag", he.Name)

	coin := utils.GetCoinBySymbol(symbol)
	base := utils.GetBaseBySymbol(symbol)
	symbol = coin + "_" + base

	token := "eyJhbGciOiJIUzUxMiJ9.eyJqdGkiOiJleDExMDE1MzY5ODQxODYyMDczREY1REQzNTMzQUY2RDc4eEtsYyIsInVpZCI6ImlTbU1YbzEwUXpiQW5kckZLNWtGaVE9PSIsInN0YSI6MCwibWlkIjowLCJpYXQiOjE1MzcyMzk3NzgsImV4cCI6MTUzNzg0NDU3OCwiYmlkIjowLCJkb20iOiJ3d3cub2tleC5jb20iLCJpc3MiOiJva2NvaW4ifQ.cAL9dlKgk9t7xxwYTCUre32C_9YY_y4hPXC0QVs1NRf7riUIT3eav4qAH9BFZQk9DpuNrL4rtxIRVfg8uWjHeQ"

	//获取货币
	utils.StartTimer(time.Minute*1, func() {
		h := http.Header{
			"authorization": []string{token},
		}
		utils.GetInfo("GET", "https://www.okex.com/v2/asset/accounts/currencies", h, func(result gjson.Result) {
			currencies := result.Get("data.#.currency")
			currencyIds := result.Get("data.#.currencyId")
			currencyIdsArr := currencyIds.Array()

			count := 0
			currencies.ForEach(func(key, value gjson.Result) bool {
				id := currencyIdsArr[count]
				currency := value.String()
				count++
				he.SetCurrency2(currency, id.String())
				return true
			})
		})

		he.GetTransfer(token)
		he.GetTrade(token)
	})

	utils.StartTimer(time.Minute*30, func() {
		h := http.Header{
			"authorization": []string{token},
		}
		utils.GetInfo("GET", "https://www.okex.com/v2/spot/new-collect", h, func(result gjson.Result) {
			symbols := result.Get("data.#.symbol")
			symbols.ForEach(func(key, value gjson.Result) bool {
				s := value.String()
				newSymbol := strings.Replace(s, "_", "-", -1)

				he.SetSymbol(newSymbol, s)

				return true
			})
		})
		he.GetPrice()
	})

	//检测价格是否全部获取完成
	utils.StartTimerWithFlag(time.Second, he.Name, func() {
		he.check(he.Name)
	})
}

func (he OkexExchange) FeesRun() {
}

func NewOkexExchange() BigE {
	exchange := new(OkexExchange)
	exchange.Exchange = Exchange{
		Name: "Okex",
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
		Sub:      exchange,
		TSDoOnce: sync.Once{},
	}
	var duitai BigE = exchange
	return duitai
}
