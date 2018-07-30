package model

import (
	"BitCoin/utils"
	"time"
	"strings"
	"github.com/tidwall/gjson"
	"github.com/PuerkitoBio/goquery"
	"sync"
)

type LBankMessage struct {
	Channel string `json:"channel"`
}

//LBank
type LBankExchange struct {
	Exchange
}

func (le LBankExchange) CheckCoinExist(symbol string) bool {
	return true
}

var lbankGetPrice sync.Once

func (le LBankExchange) GetPrice() {
}

func (le LBankExchange) Run(symbol string) {
	le.SetTradeFee(0.001, 0.001)

	//获取currency和转账费
	utils.StartTimer(time.Minute*30, func() {
		utils.GetHtml("GET", "https://lbankinfo.zendesk.com/hc/zh-cn/articles/115002295114--%E8%B4%B9%E7%8E%87%E8%AF%B4%E6%98%8E",
			nil,
			func(result *goquery.Document) {
				trs := result.Find("#article-container > article > section.article-info > div > div.article-body > table:nth-child(6) > tbody > tr")
				trs.Each(func(i int, selection *goquery.Selection) {
					currency := selection.Find("td:nth-child(1)").Text()
					transferFee := selection.Find("td:nth-child(5)").Text()
					minWithdraw := selection.Find("td:nth-child(2)").Text()
					maxWithdraw := selection.Find("td:nth-child(3)").Text()
					currency = strings.TrimSpace(currency)
					currency = strings.ToLower(currency)

					minWithdraw = strings.TrimSpace(minWithdraw)
					maxWithdraw = strings.TrimSpace(maxWithdraw)
					minWithdraw = strings.ToLower(minWithdraw)
					maxWithdraw = strings.ToLower(maxWithdraw)
					minWithdraw = strings.Replace(minWithdraw, ",", "", -1)
					maxWithdraw = strings.Replace(maxWithdraw, ",", "", -1)

					info := NewTransferInfo()

					if minWithdraw == "暂未开放提币" {
						info.CanWithdraw = 0
						le.SetTransferFee(currency, info)
					} else {
						info.CanWithdraw = 1

						m := utils.GetByZb(transferFee)
						m2 := utils.GetByZb(minWithdraw)
						m3 := utils.GetByZb(maxWithdraw)
						if currency != "币种" {
							n := m["num"]
							n2 := m2["num"]
							n3 := m3["num"]
							info.MinWithdraw = n2
							info.MaxWithdraw = n3
							if n == "" {
								if currency == "btc" {
									info.WithdrawFee = 0.0005
								} else {
									info.WithdrawFee = -1
								}
							} else {
								info.WithdrawFee = n
							}
							le.SetTransferFee(currency, info)
							le.SetCurrency(currency)
						}
					}

				})
			})
	})

	//获取symbols
	utils.StartTimer(time.Second, func() {
		u := "https://www.lbank.io/request/tick"

		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("data").ForEach(func(key, value gjson.Result) bool {
				s := value.Get("symbol").String()
				last := value.Get("l.price").Float()
				m := utils.GetCoinByLbank(s)
				base := m["base"]
				coin := m["coin"]
				symbol := coin + "-" + base
				s = coin + base
				le.SetSymbol(symbol, s)
				le.SetPrice(symbol, last)
				return true
			})
		})
		le.GetPrice()
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
