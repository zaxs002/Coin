package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"strings"
	"strconv"
	"sync"
	"BitCoin/cache"
)

type BinanceExchange struct {
	Exchange
}

func (be BinanceExchange) CheckCoinExist(symbol string) bool {
	return true
}

var binanceGetPrice sync.Once

func (be BinanceExchange) GetPrice() {
	binanceGetPrice.Do(func() {
		all := cache.GetInstance().HGetAll(be.Name + "-symbols")
		r, _ := all.Result()

		o2n := make(map[string]string)

		for k, v := range r {
			o2n[v] = k
		}

		var u = "wss://stream.binance.com:9443/stream?streams=!miniTicker@arr@1000ms"
		utils.StartTimer(time.Hour*24, func() {
			utils.GetInfoWS(u, nil, func(result gjson.Result) {
				coins := result.Get("data")
				coins.ForEach(func(key, value gjson.Result) bool {
					price := value.Get("c").Float()
					s := value.Get("s").String()
					s = strings.ToLower(s)

					k := o2n[s]

					be.SetPrice(k, price)
					return true
				})
			})
		})
	})
}

func (be BinanceExchange) Run(symbol string) {
	be.SetTradeFee(0.0005, 0.0005)

	//获取currency和transfer
	utils.StartTimer(time.Hour*1, func() {
		transferUrl := "https://www.binance.com/assetWithdraw/getAllAsset.html"

		utils.GetInfo("GET", transferUrl, nil, func(result gjson.Result) {
			r := result.Get("#.assetCode")
			array := r.Array()
			for k := range array {
				i := strconv.Itoa(k)
				name := result.Get(i + ".assetCode")
				enableWithdraw := result.Get(i + ".enableWithdraw").Bool()
				name2Lower := strings.ToLower(name.String())
				fee := result.Get(i + ".transactionFee")

				minProductWithdraw := result.Get(i + ".minProductWithdraw").Float()
				confirmTimes := result.Get(i + ".confirmTimes").Float()

				info := NewTransferInfo()
				canWithdraw := 0.0
				if enableWithdraw {
					canWithdraw = 1
				}
				info.CanWithdraw = canWithdraw
				info.WithdrawFee = fee.Float()
				info.MinWithdraw = minProductWithdraw
				info.WithdrawMinConfirmations = confirmTimes

				be.SetTransferFee(name2Lower, info)
				be.SetCurrency(name2Lower)
			}
		})
	})

	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfo("GET", "https://www.binance.com/exchange/public/product",
			nil, func(result gjson.Result) {
				result.Get("data").ForEach(func(key, value gjson.Result) bool {
					coin := value.Get("baseAsset").String()
					base := value.Get("quoteAsset").String()
					symbol := coin + "-" + base
					s := coin + base

					symbol = strings.ToLower(symbol)
					s = strings.ToLower(s)

					be.SetSymbol(symbol, s)
					return true
				})
			})
		be.GetPrice()
	})

	//检测价格是否全部获取完成
	utils.StartTimerWithFlag(time.Second, be.Name, func() {
		be.check(be.Name)
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
		TransferFees: LockMapString{
			M: make(map[string]string),
		},
		Sub:      exchange,
		TSDoOnce: sync.Once{},
	}

	var duitai BigE = exchange
	return duitai
}
