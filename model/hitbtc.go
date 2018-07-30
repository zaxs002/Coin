package model

import (
	"BitCoin/utils"
	"strconv"
	"time"
	"github.com/tidwall/gjson"
	"github.com/gorilla/websocket"
	"strings"
	"BitCoin/cache"
	"BitCoin/event"
)

//TODO
type HitbtcMessage struct {
	Method string            `json:"method"`
	Params map[string]string `json:"params"`
	Id     string            `json:"id"`
}

type HitbtcHeartBeat struct {
	Event string `json:"event"`
}

//hitbtc
type HitbtcExchange struct {
	Exchange
}

func (he HitbtcExchange) CheckCoinExist(symbol string) bool {
	return true
}
func (he HitbtcExchange) GetPrice() {
	all := cache.GetInstance().HGetAll(he.Name + "-symbols")
	result, _ := all.Result()

	o2n := make(map[string]string)

	utils.GetInfoWS2("wss://api.hitbtc.com/api/2/ws", nil,
		func(ws *websocket.Conn) {
			for k, v := range result {
				ws.WriteJSON(HitbtcMessage{
					Id:     strconv.Itoa(int(time.Now().Unix())),
					Method: "subscribeTicker",
					Params: map[string]string{"symbol": v},
				})
				o2n[v] = k
			}
		},
		func(result gjson.Result) {
			method := result.Get("method").String()
			if method == "ticker" {
				symbol := result.Get("params.symbol").String()
				last := result.Get("params.last").Float()

				symbol = strings.ToLower(symbol)
				newSymbol := o2n[symbol]

				he.SetPrice(newSymbol, last)
			}
		})
}

func (he HitbtcExchange) GetTransfer(s string) {

}
func (he *HitbtcExchange) Run(symbol string) {
	he.SetTradeFee(0.001, 0.001)

	event.Bus.Subscribe(he.Name+"-getprice", he.GetPrice)
	utils.StartTimer(time.Minute*30, func() {
		//获取symbols
		u := "https://api.hitbtc.com/api/2/public/symbol"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				symbol := value.Get("id").String()
				baseCurrency := value.Get("baseCurrency").String()
				quoteCurrency := value.Get("quoteCurrency").String()
				symbol = strings.ToLower(symbol)

				s := baseCurrency + "-" + quoteCurrency
				s = strings.ToLower(s)

				he.SetSymbol(s, symbol)
				return true
			})
		})
		he.GetPrice()
	})

	utils.StartTimer(time.Minute*30, func() {
		info := NewTransferInfo()
		//获取currency
		u := "https://api.hitbtc.com/api/2/public/currency"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				currency := value.Get("id").String()
				fee := value.Get("payoutFee").Float()
				canWithdraw := value.Get("payoutEnabled").Bool()
				payinConfirmations := value.Get("payinConfirmations").Float()
				currency = strings.ToLower(currency)

				info.WithdrawFee = fee
				info.CanWithdraw = 0
				if canWithdraw {
					info.CanWithdraw = 1
				}
				info.WithdrawMinConfirmations = payinConfirmations

				he.SetCurrency(currency)
				he.SetTransferFee(currency, info)
				return true
			})
		})
	})
}

func (he HitbtcExchange) FeesRun() {
}

func NewHitbtcExchange() BigE {
	exchange := new(HitbtcExchange)
	exchange.Exchange = Exchange{
		Name: "Hitbtc",
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
