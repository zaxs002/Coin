package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"strings"
	"BitCoin/cache"
	"BitCoin/event"
	"net/url"
	"bytes"
	"net/http"
)

//Bittrex
type BittrexExchange struct {
	Exchange
}

func (be BittrexExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (be BittrexExchange) GetTransfer(s string) {
	all := cache.GetInstance().HGetAll(be.Name + "-currency")
	result, _ := all.Result()

	utils.StartTimer(time.Minute*30, func() {
		u := "https://www.bittrex.com/api/v2.0/pub/Currency/GetCurrencyInfo"

		var client = &http.Client{}
		if !IsServer {
			uProxy, _ := url.Parse("http://127.0.0.1:1080")

			client = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(uProxy),
				},
			}
		}

		client.Timeout = time.Second * 10

		for _, v := range result {
			resp, _ := client.PostForm(u, url.Values{"currencyName": {v},})

			buf := bytes.NewBuffer(make([]byte, 0, 512))
			buf.ReadFrom(resp.Body)
			result := gjson.ParseBytes(buf.Bytes())
			fee := result.Get("result.TxFee").Float()
			cache.GetInstance().HSet(be.Name+"-transfer", v, fee)
		}
	})
}

func (be *BittrexExchange) Run(symbol string) {
	cache.GetInstance().HSet(be.Name+"-tradeFee", "taker", 0.0025)
	cache.GetInstance().HSet(be.Name+"-tradeFee", "maker", 0.0025)

	event.Bus.Subscribe(be.Name+":gettransfer", be.GetTransfer)

	//获取currency
	utils.StartTimer(time.Minute*30, func() {
		u := "https://bittrex.com/api/v1.1/public/getcurrencies"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("result").ForEach(func(key, value gjson.Result) bool {
				currency := value.Get("Currency").String()
				currency = strings.ToLower(currency)
				cache.GetInstance().HSet(be.Name+"-currency", currency, currency)
				return true
			})
		})
		event.Bus.Publish(be.Name+":gettransfer", "")
		event.Bus.Unsubscribe(be.Name+":gettransfer", be.GetTransfer)
	})

	//获取价格
	utils.StartTimer(time.Second, func() {
		u := "https://bittrex.com/api/v1.1/public/getmarketsummaries"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("result").ForEach(func(key, value gjson.Result) bool {
				MarketName := value.Get("MarketName").String()
				Last := value.Get("Last").Float()
				MarketName = strings.ToLower(MarketName)
				m := utils.GetBaseCoinByBittrex(MarketName)
				base := m["base"]
				coin := m["coin"]
				cache.GetInstance().HSet(be.Name, coin+base, Last)
				cache.GetInstance().HSet(be.Name+"-symbols", coin+base, coin+base)
				return true
			})
		})
	})
}

func (be BittrexExchange) FeesRun() {
}

func NewBittrexExchange() BigE {
	exchange := new(BittrexExchange)
	exchange.Exchange = Exchange{
		Name: "Bittrex",
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
