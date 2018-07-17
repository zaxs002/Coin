package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"BitCoin/cache"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

type BitzExchange struct {
	Exchange
}

func (he BitzExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he *BitzExchange) Run(symbol string) {
	cache.GetInstance().HSet(he.Name+"-tradeFee", "taker", 0.001)
	cache.GetInstance().HSet(he.Name+"-tradeFee", "maker", 0.001)

	//获取transferFee
	utils.StartTimer(time.Minute*30, func() {
		u := "https://www.bit-z.pro/about/fee"
		utils.GetHtml("GET", u, nil, func(result *goquery.Document) {
			result.Find("body > div.wrap > div.fee_main.clearfix > div " +
				"> table:nth-child(4) > tbody > tr").Each(func(i int, selection *goquery.Selection) {
				//currency := selection.Find("td:nth-child(1)").Text()
				fee := selection.Find("td:nth-child(2)").Text()
				m := utils.GetByZb(fee)
				coin := m["coin"]
				num := m["num"]
				coin = strings.ToLower(coin)
				cache.GetInstance().HSet(he.Name+"-transfer", coin, num)
			})
		})
	})

	utils.StartTimer(time.Second, func() {
		u := "https://www.bcex.top/Api_Market/getPriceList"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				value.ForEach(func(key, value gjson.Result) bool {
					base := value.Get("coin_to").String()
					coin := value.Get("coin_from").String()
					last := value.Get("current").Float()
					cache.GetInstance().HSet(he.Name, coin+base, last)
					cache.GetInstance().HSet(he.Name+"-symbol", coin+base, coin+base)
					return true
				})
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

func (he BitzExchange) FeesRun() {
}

func NewBitzExchange() BigE {
	exchange := new(BitzExchange)
	exchange.Exchange = Exchange{
		Name: "Bitz",
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
