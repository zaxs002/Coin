package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"strings"
	"net/http"
)

type CoinbeneExchange struct {
	Exchange
}

func (he CoinbeneExchange) CheckCoinExist(symbol string) bool {
	return true
}

//需要token 登陆
func (he *CoinbeneExchange) Run(symbol string) {
	he.SetTradeFee(0.001, 0.001)

	utils.StartTimer(time.Second, func() {
		u := "http://api.coinbene.com/v1/market/ticker?symbol=all"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("ticker").ForEach(func(key, value gjson.Result) bool {
				last := value.Get("last").Float()
				symbol := value.Get("symbol").String()
				lowSymbol := strings.ToLower(symbol)

				m := utils.GetBaseBySymbol2(lowSymbol)
				coin := m["coin"]
				base := m["base"]

				symbol = coin + "-" + base

				he.SetSymbol(symbol, lowSymbol)
				he.SetPrice(symbol, last)
				return true
			})
		})
	})

	u := "https://a.coinbene.com/account/account/list"
	token := "Bearer eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiIxOTc3Mzg1Iiwic2NvcGVzIjpbIlJPTEVfVVNFUiJdLCJzaXRlIjoiTUFJTiIsImxvZ2luSWQiOiJ3dWRpZ29kMTJAMTYzLmNvbSIsImVudiI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdPVzY0KSBBcHBsZVdlYktpdC81MzcuMzYgKEtIVE1MLCBsaWtlIEdlY2tvKSBDaHJvbWUvNjMuMC4zMjM5LjEzMiBTYWZhcmkvNTM3LjM2IiwiYmFuayI6Ik1BSU4iLCJpc3MiOiJodHRwczovL3d3dy5jb2luYmVuZS5jb20iLCJpYXQiOjE1MzI1NzA2MDUsImV4cCI6MTUzMjU3NDIwNX0.ypSQmONVyGGUZ-JX1I1TTWXWuY0UZkXyMyog9S6a6Uo-wwvNPbi6DpvuvmGUOqOrSXXTBGuU719IwUcPoTR1Fw"
	site := "MAIN"
	headers := http.Header{
		"site":          []string{site},
		"Authorization": []string{token},
		"lang":          []string{"zh_CN"},
	}
	utils.StartTimer(time.Minute*30, func() {
		utils.GetInfo("GET", u, headers, func(result gjson.Result) {
			result.Get("data.list").ForEach(func(key, value gjson.Result) bool {
				currency := value.Get("asset").String()
				fee := value.Get("withdrawFee").Float()
				minWithdraw := value.Get("minWithdraw").Float()
				withdrawMinConfirmations := value.Get("withdrawMinConfirmations").Float()
				withdraw := value.Get("withdraw").Bool()

				info := NewTransferInfo()
				info.WithdrawFee = fee
				info.MinWithdraw = minWithdraw
				info.WithdrawMinConfirmations = withdrawMinConfirmations
				if withdraw {
					info.CanWithdraw = 1
				} else {
					info.CanWithdraw = 0
				}

				he.SetTransferFee(currency, info)
				he.SetCurrency(currency)
				return true
			})
		})
	})
}

func (he CoinbeneExchange) FeesRun() {
}

func NewCoinbeneExchange() BigE {
	exchange := new(CoinbeneExchange)
	exchange.Exchange = Exchange{
		Name: "Coinbene",
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
