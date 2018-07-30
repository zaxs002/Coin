package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"strconv"
)

type BitzExchange struct {
	Exchange
}

func (he BitzExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he *BitzExchange) Run(symbol string) {
	he.SetTradeFee(0.001, 0.011)

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

				f, _ := strconv.ParseFloat(num, 64)
				info := NewTransferInfo()
				info.WithdrawFee = f
				he.SetTransferFee(coin, info)
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
					s := coin + "-" + base

					he.SetSymbol(s, s)
					he.SetPrice(s, last)
					return true
				})
				return true
			})
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
