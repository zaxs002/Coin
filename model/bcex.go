package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"BitCoin/cache"
	"net/http"
)

type BcexExchange struct {
	Exchange
}

func (he BcexExchange) CheckCoinExist(symbol string) bool {
	return true
}

//需要token 登陆
func (he *BcexExchange) Run(symbol string) {
	cache.GetInstance().HSet(he.Name+"-tradeFee", "taker", 0.002)
	cache.GetInstance().HSet(he.Name+"-tradeFee", "maker", 0)

	he.SetTradeFee(0.002, 0)

	utils.StartTimer(time.Second, func() {
		u := "https://www.bcex.top/Api_Market/getPriceList"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				value.ForEach(func(key, value gjson.Result) bool {
					base := value.Get("coin_to").String()
					coin := value.Get("coin_from").String()
					last := value.Get("current").Float()
					s := coin + "-" + base
					he.SetPrice(s, last)
					he.SetSymbol(s, s)
					return true
				})
				return true
			})
		})
	})

	u := "https://a.coinbene.com/account/account/list"
	token := "Bearer eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiIxOTc3Mzg1Iiwic2NvcGVzIjpbIlJPTEVfVVNFUiJdLCJzaXRlIjoiTUFJTiIsImxvZ2luSWQiOiJ3dWRpZ29kMTJAMTYzLmNvbSIsImVudiI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdPVzY0KSBBcHBsZVdlYktpdC81MzcuMzYgKEtIVE1MLCBsaWtlIEdlY2tvKSBDaHJvbWUvNjMuMC4zMjM5LjEzMiBTYWZhcmkvNTM3LjM2IiwiYmFuayI6Ik1BSU4iLCJpc3MiOiJodHRwczovL3d3dy5jb2luYmVuZS5jb20iLCJpYXQiOjE1MzI0OTM1ODEsImV4cCI6MTUzMjQ5NzE4MX0.tWP9EiAe4u-bJL9tdCuOeJK9vApPzKpHJFEk9IIhSQIcUcFpgcDvDnd66u13GGoi_CWKsbyPWoVow_70UVFU1A"
	site := "MAIN"
	headers := http.Header{
		"site":          []string{site},
		"Authorization": []string{token},
		"lang":          []string{"zh_CN"},
	}
	utils.GetInfo("GET", u, headers, func(result gjson.Result) {
		result.Get("data.list").ForEach(func(key, value gjson.Result) bool {
			currency := value.Get("asset").String()
			fee := value.Get("withdrawFee").Float()
			minWithdraw := value.Get("minWithdraw").Float()

			info := NewTransferInfo()
			info.MinWithdraw = minWithdraw
			info.WithdrawFee = fee
			info.CanWithdraw = 1
			info.WithdrawMinConfirmations = 1

			he.SetTransferFee(currency, info)
			he.SetCurrency(currency)
			return true
		})
	})
}

func (he BcexExchange) FeesRun() {
}

func NewBcexExchange() BigE {
	exchange := new(BcexExchange)
	exchange.Exchange = Exchange{
		Name: "Bcex",
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
		Sub: exchange,
	}
	var duitai BigE = exchange
	return duitai
}
