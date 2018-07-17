package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"BitCoin/cache"
	"strings"
	"strconv"
	"BitCoin/event"
)

type BinanceExchange struct {
	Exchange
}

func (be BinanceExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (be BinanceExchange) GetPrice(s string) {
	var u = "wss://stream.binance.com:9443/stream?streams=!miniTicker@arr@1000ms"
	//var u = "wss://stream.binance.com:9443/ws/%s@aggTrade"
	//var u = "wss://stream.binance.com:9443/ws/!ticker@arr"
	utils.StartTimer(time.Hour*24, func() {
		utils.GetInfoWS(u, nil, func(result gjson.Result) {
			coins := result.Get("data")
			coins.ForEach(func(key, value gjson.Result) bool {
				price := value.Get("c").Float()
				s := value.Get("s").String()
				s = strings.ToLower(s)
				cache.GetInstance().HSet(be.Name, s, price)
				return true
			})
		})
	})
}

func (be BinanceExchange) Run(symbol string) {
	event.Bus.Subscribe(be.Name+":getprice", be.GetPrice)

	//获取currency和transfer
	utils.StartTimer(time.Hour*1, func() {
		cache.GetInstance().HSet(be.Name+"-tradeFee", "maker", 0.0005)
		cache.GetInstance().HSet(be.Name+"-tradeFee", "taker", 0.0005)

		transferUrl := "https://www.binance.com/assetWithdraw/getAllAsset.html"

		utils.GetInfo("GET", transferUrl, nil, func(result gjson.Result) {
			r := result.Get("#.assetCode")
			array := r.Array()
			for k := range array {
				i := strconv.Itoa(k)
				name := result.Get(i + ".assetCode")
				name2Lower := strings.ToLower(name.String())
				fee := result.Get(i + ".transactionFee")
				be.TransferFees.Set(name2Lower, fee.Float())
				cache.GetInstance().HSet(be.Name+"-transfer", name2Lower, fee.Float())
				cache.GetInstance().HSet(be.Name+"-currency", name2Lower, name2Lower)
			}
		})
	})

	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfo("GET", "https://www.binance.com/exchange/public/product",
			nil, func(result gjson.Result) {
				symbols := result.Get("data.#.symbol")
				symbols.ForEach(func(key, value gjson.Result) bool {
					s := strings.ToLower(value.String())
					cache.GetInstance().HSet(be.Name+"-symbols", s, s)
					return true
				})
			})
		event.Bus.Publish(be.Name+":getprice", "")
		event.Bus.Unsubscribe(be.Name+":getprice", be.GetPrice)
	})

}

func (be BinanceExchange) FeesRun() {
}

func NewBinanceExchange() BigE {
	exchange := new(BinanceExchange)
	exchange.Exchange = Exchange{
		Name: "Binance",
		PriceQueue: LockMap{
			M: make(map[string]float64),
		},
		AmountDict: LockMap{
			M: make(map[string]float64),
		},
		TradeFees: LockMap{
			M: make(map[string]float64),
		},
		TransferFees: LockMap{
			M: make(map[string]float64),
		},
		Sub: exchange,
	}

	var duitai BigE = exchange
	return duitai
}
