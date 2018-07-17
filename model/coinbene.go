package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"BitCoin/cache"
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
	cache.GetInstance().HSet(he.Name+"-tradeFee", "taker", 0.001)
	cache.GetInstance().HSet(he.Name+"-tradeFee", "maker", 0.001)

	utils.StartTimer(time.Second, func() {
		u := "http://api.coinbene.com/v1/market/ticker?symbol=all"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("ticker").ForEach(func(key, value gjson.Result) bool {
				last := value.Get("last").Float()
				symbol := value.Get("symbol").String()
				lowSymbol := strings.ToLower(symbol)
				cache.GetInstance().HSet(he.Name+"-symbol", symbol, lowSymbol)
				cache.GetInstance().HSet(he.Name, lowSymbol, last)
				return true
			})
		})
	})

	u := "https://a.coinbene.com/account/account/list"
	token := "Bearer eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiIxOTc3Mzg1Iiwic2NvcGVzIjpbIlJPTEVfVVNFUiJdLCJzaXRlIjoiTUFJTiIsImxvZ2luSWQiOiJ3dWRpZ29kMTJAMTYzLmNvbSIsImVudiI6Ik1vemlsbGEvNS4wIChXaW5kb3dzIE5UIDEwLjA7IFdPVzY0KSBBcHBsZVdlYktpdC81MzcuMzYgKEtIVE1MLCBsaWtlIEdlY2tvKSBDaHJvbWUvNjMuMC4zMjM5LjEzMiBTYWZhcmkvNTM3LjM2IiwiYmFuayI6Ik1BSU4iLCJpc3MiOiJodHRwczovL3d3dy5jb2luYmVuZS5jb20iLCJpYXQiOjE1MzE3MTQzMzgsImV4cCI6MTUzMTcxNzkzOH0.Idd1eJWV3OMOWCepdmzAOROFHrGEU3sBTQLXQqpnsA2EX91EKKuyGPvcu7pAJhY_UymPuBWh2KhtxtSTfk1mqg"
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
			//minWithdraw := value.Get("minWithdraw").Float()
			cache.GetInstance().HSet(he.Name+"-currency", currency, currency)
			cache.GetInstance().HSet(he.Name+"-transfer", currency, fee)
			return true
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
