package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"strconv"
)

type BiboxMessage struct {
	Event      string                 `json:"event"`
	Channel    string                 `json:"channel"`
	Parameters map[string]interface{} `json:"parameters"`
}

type BiboxExchange struct {
	Exchange
}

func (he BiboxExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he BiboxExchange) GetPrice(s string) {
}

func (he *BiboxExchange) Run(symbol string) {
	he.SetTradeFee(0.0005, 0.0005)

	//获取价格
	utils.StartTimer(time.Second, func() {
		u := "https://api.bibox.com/v1/mdata?cmd=marketAll"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("result").ForEach(func(key, value gjson.Result) bool {
				coin := value.Get("coin_symbol").String()
				base := value.Get("currency_symbol").String()
				last := value.Get("last").Float()

				symbol := coin + "-" + base
				symbol = strings.ToLower(symbol)
				he.SetPrice(symbol, last)
				return true
			})
		})
	})

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		u := "https://api.bibox.com/v1/mdata?cmd=pairList"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("result").ForEach(func(key, value gjson.Result) bool {
				symbol := value.Get("pair").String()
				m := utils.GetCoinByZb(symbol)
				coin := m["coin"]
				base := m["base"]
				newSymbol := coin + "-" + base
				newSymbol = strings.ToLower(newSymbol)
				he.SetSymbol(newSymbol, symbol)
				return true
			})
		})
	})

	//获取currency和转账费
	utils.StartTimer(time.Minute*30, func() {
		u := "https://bibox.zendesk.com/hc/zh-cn/articles/360002336133"
		utils.GetHtml("GET", u, nil, func(result *goquery.Document) {
			result.Find("#article-container > article > section.article-info" +
				" > div > div.article-body > div:nth-child(16) > table > tbody > tr").
				Each(func(i int, selection *goquery.Selection) {
				currency := selection.Find("td:nth-child(1)").Text()
				currency = strings.ToLower(currency)
				fee := selection.Find("td:nth-child(3)").Text()
				m := utils.GetFloatByBitfinex(fee)
				num := m["num"]
				if num != "" {
					info := NewTransferInfo()
					f, _ := strconv.ParseFloat(num, 64)
					info.WithdrawFee = f
					he.SetTransferFee(currency, info)
					he.SetCurrency(currency)
				}
			})
		})
	})
}

func (he BiboxExchange) FeesRun() {
}

func NewBiboxExchange() BigE {
	exchange := new(BiboxExchange)
	exchange.Exchange = Exchange{
		Name: "BiBox",
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
