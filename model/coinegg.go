package model

import (
	"BitCoin/cache"
	"github.com/tidwall/gjson"
	"BitCoin/utils"
	"time"
	"sync"
	"github.com/gorilla/websocket"
	"net/http"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"strconv"
)

var CoineggOnce sync.Once
var CoineggOnce2 sync.Once

type CoineggExchange struct {
	Exchange
}

func (he CoineggExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he CoineggExchange) FeesRun() {
}

func (he CoineggExchange) GetTransfer() {
	once.Do(func() {
		//获取转账手续费
		token := "p83T5HXezbSKLzNYEQhBy9GknGivAufxKWoRA2iSQUxFSTvNLPdIR0pR0ZZacbjW9LXgaSrAL_P47V9ZxHZRtg=="

		all := cache.GetInstance().HGetAll(he.Name + "-currency")
		result, _ := all.Result()

		transferFeeUrl := "https://exchange.fcoin.com/api/web/v1/accounts/withdraws/fee?currency="

		headers := http.Header{
			"token": []string{token},
		}

		for _, v := range result {
			utils.GetInfo("GET", transferFeeUrl+v, headers, func(result gjson.Result) {
				fee := result.Get("data").Float()
				cache.GetInstance().HSet(he.Name+"-transfer", v, fee)
			})
		}
	})
}

func (he CoineggExchange) GetPrice() {
	once2.Do(func() {
		//获取价格
		all := cache.GetInstance().HGetAll(he.Name + "-symbols")
		result, _ := all.Result()

		u := "wss://api.fcoin.com/v2/ws"
		utils.GetInfoWS2(u, nil,
			func(ws *websocket.Conn) {
				for _, v := range result {
					ws.WriteJSON(FcoinMessage{
						Cmd:  "sub",
						Args: []string{"ticker." + v},
					})
				}
			}, func(result gjson.Result) {
				last := result.Get("ticker.0").Float()
				symbol := result.Get("type").String()
				m := utils.GetSymbolByFcoin(symbol)
				if m != nil {
					if m["type"] == "ticker" {
						s := m["symbol"]
						cache.GetInstance().HSet(he.Name, s, last)
					}
				}
			})
	})
}

func (he CoineggExchange) Run(symbol string) {
	bases := []string{"btc", "usdt", "eth", "usc"}

	he.SetTradeFee(0.001, 0.001)

	//获取symbols和价格
	utils.StartTimer(time.Second, func() {
		url := "https://www.coinegg.com/coin/%s/allcoin"
		for _, v := range bases {
			baseUrl := fmt.Sprintf(url, v)

			utils.GetInfo("GET", baseUrl, nil, func(result gjson.Result) {
				result.ForEach(func(key, value gjson.Result) bool {
					s := key.String()
					last := value.Get("1").Float()
					symbol := s + "-" + v

					he.SetSymbol(symbol, symbol)
					he.SetPrice(symbol, last)
					return true
				})
			})
		}
	})

	utils.StartTimer(time.Minute*30, func() {
		u := "https://www.coinegg.com/fee.html"
		utils.GetHtml("GET", u, nil, func(result *goquery.Document) {
			//body > div.body > div.right.gonggao > div > ul:nth-child(3) > li.w170
			result.Find("body > div.body > div.right.gonggao > div > ul.noticeListHeadBody").Each(func(i int, selection *goquery.Selection) {
				text := selection.Find("li:nth-child(4)").Text()
				c := selection.Find("li:nth-child(2)").Text()
				m := utils.GetByZb(text)
				coin := m["coin"]
				info := NewTransferInfo()
				info.CanWithdraw = 1
				if coin != "" && strings.TrimSpace(coin) != "" {
					num := m["num"]
					f, _ := strconv.ParseFloat(num, 64)
					info.WithdrawFee = f
					he.SetTransferFee(c, info)
				} else {
					info.WithdrawFee = -1
					he.SetTransferFee(c, info)
				}
				he.SetCurrency(c)
			})
		})
	})
}

func NewCoineggExchange() BigE {
	exchange := new(CoineggExchange)

	exchange.Exchange = Exchange{
		Name: "CoinEgg",
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
