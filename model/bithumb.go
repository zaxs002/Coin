package model

import (
	"BitCoin/utils"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BithumbExchange struct {
	Exchange
}
type BithumbMessage struct {
	Currency string `json:"currency"`
}

func (he BithumbExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he BithumbExchange) FeesRun() {
}

var bithumbOnce sync.Once

func (he BithumbExchange) GetPrice() {
	bithumbOnce.Do(func() {
		utils.StartTimer(time.Hour*24, func() {
			var u = "wss://wss.bithumb.com/public"
			utils.GetInfoWS2(u, nil,
				func(ws *websocket.Conn) {
					ws.WriteJSON(BithumbMessage{
						Currency: "BTC",
					})
				}, func(result gjson.Result) {
					if result.Get("header").Get("service").String() == "ticker" {
						result.Get("data").ForEach(func(key, value gjson.Result) bool {
							fmt.Println(value)
							return true
						})
					}
				})
		})
	})
}

func (he BithumbExchange) Run(symbol string) {
	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		symbolsUrl := "https://api.bithumb.com/public/ticker/ALL"

		utils.GetInfo("GET", symbolsUrl, nil, func(result gjson.Result) {
			if result.Get("status").String() == "0000" {
				btcBuy := 0.0
				btcSell := 0.0
				ethBuy := 0.0
				ethSell := 0.0
				result.Get("data").ForEach(func(key, value gjson.Result) bool {
					currency := key.String()
					currency = strings.ToLower(currency)
					buyPrice := value.Get("buy_price").Float()
					sellPrice := value.Get("sell_price").Float()
					he.SetCurrency(currency)
					if currency == "btc" {
						btcBuy = buyPrice
						btcSell = sellPrice
					} else if currency == "eth" {
						ethBuy = buyPrice
						ethSell = sellPrice
						he.SetPrice("eth-btc", ethBuy/btcBuy)
					} else if currency == "date" {
						return false
					} else {
						btcSymbol := currency + "-btc"
						ethSymbol := currency + "-eth"
						btcPrice := buyPrice / btcBuy
						ethPrice := buyPrice / ethBuy
						he.SetSymbol(btcSymbol, btcSymbol)
						he.SetSymbol(ethSymbol, ethSymbol)
						he.SetPrice(btcSymbol, btcPrice)
						he.SetPrice(ethSymbol, ethPrice)
					}
					return true
				})
			}
		})

	})

	//获取交易手续费
	utils.StartTimer(time.Hour*1, func() {
		tradeFeeUrl := "https://www.bithumb.com/u1/US138"
		flag := true

		for ; flag; {
			utils.GetHtml("GET", tradeFeeUrl, nil, func(result *goquery.Document) {
				feeHtml1 := result.Find("#contents_f > table.g_table_list.fee > tbody > tr:nth-child(1) > td:nth-child(2)")
				if feeHtml1 == nil {
				} else {
					flag = false
					feeHtml := feeHtml1.Text()
					f := utils.GetFeeByBithumb(feeHtml)
					fs := f["num"]
					fee, _ := strconv.ParseFloat(fs, 64)
					he.SetTradeFee(fee/100, fee/100)

					result.Find("#contents_f > table.g_table_list.fee_in_out > tbody >tr").Each(func(i int, tr *goquery.Selection) {
						_, exists := tr.Attr("data-coin")
						if exists {
							td := tr.Find("td.money_type")
							divs := tr.Find("div.right.out_fee")
							coinText := td.Text()
							feeText := divs.Text()
							if feeText != "" {
								c := utils.GetCoinByBithumb(coinText)
								info := NewTransferInfo()
								info.CanDeposit = 1
								info.CanWithdraw = 1
								info.WithdrawFee = feeText
								he.SetTransferFee(c["num"], info)
							}
						}
					})
				}
			})
			time.Sleep(1 * time.Second)
		}
	})
}

func NewBithumbExchange() BigE {
	exchange := new(BithumbExchange)

	exchange.Exchange = Exchange{
		Name: "Bithumb",
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
