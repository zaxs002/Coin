package main

import (
	"net/http"
	"fmt"
	"github.com/tidwall/gjson"
	"bytes"
	"BitCoin/model"
)

func main() {
	result := getSymbols()

	fmt.Println(result)
	var ExchangeList []model.BigE

	huobi := model.NewHuoBiExchange()
	binance := model.NewBinanceExchange()
	kucoin := model.NewKuCoinExchange()
	bitfinex := model.NewBitfinexExchange()
	poloniex := model.NewPoloniexExchange()
	okex := model.NewOkexExchange()
	upbit := model.NewUpbitExchange()
	hitbtc := model.NewHitbtcExchange()
	fatbtc := model.NewFatbtcExchange()
	exx := model.NewExxExchange()
	zb := model.NewZbExchange()
	allcoin := model.NewAllcoinExchange()
	quoine := model.NewQuoineExchange()

	ExchangeList = append(ExchangeList, huobi)
	ExchangeList = append(ExchangeList, binance)
	ExchangeList = append(ExchangeList, kucoin)
	ExchangeList = append(ExchangeList, bitfinex)
	ExchangeList = append(ExchangeList, poloniex)
	ExchangeList = append(ExchangeList, okex)
	ExchangeList = append(ExchangeList, upbit)
	ExchangeList = append(ExchangeList, hitbtc)
	ExchangeList = append(ExchangeList, fatbtc)
	ExchangeList = append(ExchangeList, exx)
	ExchangeList = append(ExchangeList, zb)
	ExchangeList = append(ExchangeList, allcoin)
	ExchangeList = append(ExchangeList, quoine)

	result.ForEach(func(key, value gjson.Result) bool {
		for _, e := range ExchangeList {
			e.CreateRun(value.String())
		}
		return true
	})
}

func getSymbols() gjson.Result {
	//uProxy, _ := url.Parse("http://127.0.0.1:1080")
	client := &http.Client{
		//Transport: &http.Transport{
		//	Proxy: http.ProxyURL(uProxy),
		//},
	}

	url :="https://api.hitbtc.com/api/2/public/symbol"

	resp, err := client.Get(url)

	if err != nil {
		fmt.Println(err)
	}
	buf := bytes.NewBuffer(make([]byte, 0, 512))

	buf.ReadFrom(resp.Body)
	results := gjson.GetBytes(buf.Bytes(), "#[quoteCurrency==\"BTC\"]#.id")

	return results
}
