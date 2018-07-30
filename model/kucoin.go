package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"BitCoin/cache"
	"BitCoin/event"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

type KuCoinExchange struct {
	Exchange
}

func (he KuCoinExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he KuCoinExchange) check(s string) {
	println("---------")
	symbolLen := cache.GetInstance().HLen(he.Name + "-symbols").Val()
	priceLen := cache.GetInstance().HLen(he.Name).Val()
	currencyLen := cache.GetInstance().HLen(he.Name + "-currency").Val()
	transferLen := cache.GetInstance().HLen(he.Name + "-transfer").Val()

	if symbolLen == priceLen && symbolLen != 0 {
		println(he.Name + " 价格全部获取完成")
	}

	if currencyLen == transferLen && transferLen != 0 {
		println(he.Name + " 转账手续费获取完成")
	}
	if symbolLen == priceLen && currencyLen == transferLen && transferLen != 0 {
		event.Bus.Unsubscribe(he.Name+":check", he.check)
	}
}

func (he *KuCoinExchange) Run(symbol string) {
	he.SetTradeFee(0.001, 0.001)

	//获取价格
	u := "https://api.kucoin.com/v1/open/tick"
	utils.StartTimer(time.Second, func() {
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("data").ForEach(func(key, value gjson.Result) bool {
				coin := value.Get("coinType").String()
				base := value.Get("coinTypePair").String()
				last := value.Get("buy").Float()
				symbol := coin + "-" + base
				s := coin + base

				symbol = strings.ToLower(symbol)
				s = strings.ToLower(s)

				he.SetPrice(symbol, last)
				he.SetSymbol(symbol, s)
				return true
			})
		})
	})

	//获取转账手续费
	transferUrl := "https://news.kucoin.com/fee/"
	info := NewTransferInfo()
	utils.StartTimer(time.Minute*30, func() {
		utils.GetHtml("GET", transferUrl, nil, func(result *goquery.Document) {
			result.Find("#loop-container > div > article > div.post-content > table > tbody > tr").Each(func(i int, selection *goquery.Selection) {
				currency := selection.Find("td:nth-child(1)").Text()
				fee := selection.Find("td:nth-child(2)").Text()
				//最小提现额
				min := selection.Find("td:nth-child(3)").Text()
				if currency != "虚拟货币" {
					info.MinWithdraw = min
					if fee == "Free" {
						info.WithdrawFee = 0
					} else {
						info.WithdrawFee = fee
					}
					he.SetTransferFee(currency, info)
					he.SetCurrency(currency)
				}
			})
		})
	})
}

func (he KuCoinExchange) FeesRun() {
}

func NewKuCoinExchange() BigE {
	exchange := new(KuCoinExchange)
	exchange.Exchange = Exchange{
		Name: "KuCoin",
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
