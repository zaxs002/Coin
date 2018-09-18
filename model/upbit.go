package model

import (
	"time"
	"fmt"
	"github.com/tidwall/gjson"
	"BitCoin/utils"
	"strings"
	"sync"
	"BitCoin/cache"
)

//upbit
type UpbitExchange struct {
	Exchange
}

func (ue UpbitExchange) CheckCoinExist(symbol string) bool {
	return true
}

var upbitGetPrice sync.Once

func (ue UpbitExchange) GetPrice() {
	upbitGetPrice.Do(func() {
		u := "https://api.upbit.com/v1/ticker?markets="
		all := cache.GetInstance().HGetAll(ue.Name + "-symbols")
		r, _ := all.Result()
		o2n := make(map[string]string)
		for k := range r {
			arr := strings.Split(k, "-")
			s := arr[1] + "-" + arr[0] + ","
			o2n[s] = k
			u += s
		}
		utils.StartTimer(time.Second, func() {
			utils.GetInfo("GET", u, nil, func(result gjson.Result) {
				result.ForEach(func(key, value gjson.Result) bool {
					s := value.Get("market").String()
					last := value.Get("trade_price").Float()

					symbol := o2n[s]

					ue.SetPrice(symbol, last)
					return true
				})
			})
		})
	})
}

func (ue *UpbitExchange) Run(symbol string) {
	ue.SetTradeFee(0.0025, 0.0025)
	//获取currency
	currencyUrl := "https://api-manager.upbit.com/api/v1/kv/UPBIT_PC_COIN_DEPOSIT_AND_WITHDRAW_GUIDE"
	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfo("GET", currencyUrl, nil, func(result gjson.Result) {
			s := result.Get("data").String()
			p := gjson.Parse(s)
			p.ForEach(func(key, value gjson.Result) bool {
				info := NewTransferInfo()
				s := value.Get("currency").String()
				withdrawFee := value.Get("withdrawFee").Float()
				info.WithdrawFee = withdrawFee
				info.CanWithdraw = 1
				ue.SetTransferFee(s, info)
				ue.SetCurrency(s)
				return true
			})
		})
	})

	//获取symbols
	symbolsUrl := "https://s3.ap-northeast-2.amazonaws.com/crix-production/crix_master"
	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfo("GET", symbolsUrl, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				baseCurrencyCode := value.Get("baseCurrencyCode").String()
				quoteCurrencyCode := value.Get("quoteCurrencyCode").String()
				symbol := baseCurrencyCode + "-" + quoteCurrencyCode
				symbol = strings.ToLower(symbol)

				ue.SetSymbol(symbol, symbol)
				return true
			})
		})
		ue.GetPrice()
	})

}

func (he UpbitExchange) FeesRun() {
	fmt.Println("Old FeesRun")
}

func NewUpbitExchange() BigE {
	exchange := new(UpbitExchange)
	exchange.Exchange = Exchange{
		Name: "Upbit",
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
