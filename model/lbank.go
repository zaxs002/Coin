package model

import (
	"BitCoin/utils"
	"time"
	"strings"
	"BitCoin/cache"
	"github.com/tidwall/gjson"
	"BitCoin/event"
	"github.com/PuerkitoBio/goquery"
	"fmt"
)

type LBankMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
}

//LBank
type LBankExchange struct {
	Exchange
}

func (le LBankExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (le LBankExchange) GetPrice(s string) {
	all := cache.GetInstance().HGetAll(le.Name + "-symbols")
	result, _ := all.Result()
	var symbols []string
	for _, v := range result {
		symbols = append(symbols, v)
	}
	for _, v := range symbols {
		go func() {
			utils.StartTimer(time.Second, func() {
				u := "https://www.lbank.io/request/history?symbol=%s&resolution=1&from=%d&to=%d"

				t := time.Now()
				m := utils.GetCoinByZb(v)
				symbol := m["coin"] + ":" + m["base"]
				timestamp := t.Unix()
				old := timestamp - 10

				s := fmt.Sprintf(u, symbol, old, timestamp)
				utils.GetInfo("GET", s, nil, func(result gjson.Result) {
					fmt.Println(result)
				})
			})
		}()
	}
}

func (le LBankExchange) Run(symbol string) {
	//cookie := "aliyungf_tc=AQAAAOh22n1cpw4AMEU2c4A6DXk0+5Fz; Hm_lvt_5776045435d31dc2fcb78afd31c2cdb0=1530516012,1530584000; _uuid=fbdb209b678eed65968a08f3db8a13cb775cd9ce312fc09ec84b42d28667502373fceb136254585e436ec793e38c63579684ce1c0cda5286b26dc2d44224edd28490e7221ab1b9a7921a1828f1250caf; _uname=wudigod12%40163.com; Hm_lpvt_5776045435d31dc2fcb78afd31c2cdb0=1530588044"

	cache.GetInstance().HSet(le.Name+"-tradeFee", "taker", 0.001)
	cache.GetInstance().HSet(le.Name+"-tradeFee", "maker", 0.001)

	event.Bus.Subscribe(le.Name+"-getprice", le.GetPrice)
	//获取currency和转账费
	utils.StartTimer(time.Minute*30, func() {
		utils.GetHtml("GET", "https://lbankinfo.zendesk.com/hc/zh-cn/articles/115002295114--%E8%B4%B9%E7%8E%87%E8%AF%B4%E6%98%8E",
			nil,
			func(result *goquery.Document) {
				trs := result.Find("#article-container > article > section.article-info > div > div.article-body > table:nth-child(6) > tbody > tr")
				trs.Each(func(i int, selection *goquery.Selection) {
					currency := selection.Find("td:nth-child(1)").Text()
					transferFee := selection.Find("td:nth-child(5)").Text()
					currency = strings.TrimSpace(currency)
					currency = strings.ToLower(currency)
					m := utils.GetByZb(transferFee)
					if currency != "币种" {
						n := m["num"]
						if n == "" {
							if currency == "btc" {
								cache.GetInstance().HSet(le.Name+"-transfer", currency, 0.0005)
							} else {
								cache.GetInstance().HSet(le.Name+"-transfer", currency, -1)
							}
						} else {
							cache.GetInstance().HSet(le.Name+"-transfer", currency, n)
						}
						cache.GetInstance().HSet(le.Name+"-currency", currency, currency)
					}
				})
			})
	})

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		u := "http://api.lbank.info/v1/currencyPairs.do"

		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				s := strings.ToLower(value.String())
				m := utils.GetCoinByZb(s)
				base := m["base"]
				coin := m["coin"]
				cache.GetInstance().HSet(le.Name+"-symbols", coin+base, value.String())
				return true
			})
		})
		event.Bus.Publish(le.Name+"-getprice", "")
		event.Bus.Unsubscribe(le.Name+"-getprice", le.GetPrice)
	})

}

func (le LBankExchange) FeesRun() {
}

func NewLBankExchange() BigE {
	exchange := new(LBankExchange)
	exchange.Exchange = Exchange{
		Name: "LBank",
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
