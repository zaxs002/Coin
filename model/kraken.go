package model

import (
	"BitCoin/cache"
	"BitCoin/utils"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"strings"
	"sync"
	"time"
)

type KrakenMessage struct {
	Event      string                 `json:"event"`
	Channel    string                 `json:"channel"`
	Parameters map[string]interface{} `json:"parameters"`
}

type KrakenHeartBeat struct {
	Event string `json:"event"`
}
type KrakenExchange struct {
	Exchange
}

func (he KrakenExchange) CheckCoinExist(symbol string) bool {
	return true
}

var krakenGetPrice sync.Once
var krakenGetTransfer sync.Once
var krakenGetTrade sync.Once
var krakenCheck sync.Once

func (he KrakenExchange) GetPrice() {
	okexGetPrice.Do(func() {
		utils.StartTimer(time.Hour*24, func() {
			//websocket
			all := cache.GetInstance().HGetAll(he.Name + "-symbols")
			result, _ := all.Result()

			apiKey := "c4eb13f2-b3d8-4446-9ddd-a48919e14a8e"
			secretKey := "A8E98839AAA88020FAD749DE33566A89"
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

func (he KrakenExchange) GetTransfer() {
	okexGetTransfer.Do(func() {
		utils.StartTimer(time.Minute*30, func() {
			all := cache.GetInstance().HGetAll(he.Name + "-currency")
			aa, _ := all.Result()

			for k, v := range aa {
				info := NewTransferInfo()

				url := "https://www.okex.com/v2/asset/withdraw?currencyId=" + v

				utils.GetInfo("GET", url, nil, func(result gjson.Result) {
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

func (he KrakenExchange) GetTrade() {
	okexGetTrade.Do(func() {
		utils.StartTimer(time.Minute*30, func() {
			tradeFeeUrl := "https://www.okex.com/v2/spot/user-level"

			utils.GetInfo("GET", tradeFeeUrl, nil, func(result gjson.Result) {
				takerFee := result.Get("data.takerFeeRatio").Float()
				makerFee := result.Get("data.makerFeeRatio").Float()

				takerFee = takerFee / 100
				makerFee = makerFee / 100

				he.SetTradeFee(takerFee, makerFee)
			})
		})
	})
}

func (he KrakenExchange) Run(symbol string) {
	//获取货币
	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfo("GET", "https://api.kraken.com/0/public/AssetPairs", nil, func(result gjson.Result) {
			pairs := result.Get("result")

			pairs.ForEach(func(key, value gjson.Result) bool {
				base := value.Get("base").String()
				quote := value.Get("quote").String()
				symbol := quote + "-" + base
				he.SetSymbol(symbol, symbol)

				taker := value.Get("fees.1.1").Float()
				maker := value.Get("fees_maker.1.1").Float()
				he.SetTradeFee(taker, maker)
				return true
			})
		})
		he.GetPrice()
	})

	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfo("GET", "https://api.kraken.com/0/public/Assets", nil, func(result gjson.Result) {
			result.Get("result").ForEach(func(key, value gjson.Result) bool {
				alt := value.Get("altname").String()
				he.SetCurrency(alt)
				return true
			})
		})
	})

	//检测价格是否全部获取完成
	utils.StartTimerWithFlag(time.Second, he.Name, func() {
		he.check(he.Name)
	})
}

func (he KrakenExchange) FeesRun() {
}

func NewKrakenExchange() BigE {
	exchange := new(KrakenExchange)
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
