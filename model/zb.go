package model

import (
	"BitCoin/utils"
	"time"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"BitCoin/cache"
	"github.com/tidwall/gjson"
	"github.com/gorilla/websocket"
	"BitCoin/event"
)

type ZbMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
}

//zb
type ZbExchange struct {
	Exchange
}

func (ue ZbExchange) CheckCoinExist(symbol string) bool {
	return true
}
func (ue ZbExchange) GetPrice(s string) {
	all := cache.GetInstance().HGetAll(ue.Name + "-symbols")
	result, _ := all.Result()
	utils.GetInfoWS2("wss://api.zb.com:9999/websocket", nil,
		func(ws *websocket.Conn) {
			for _, v := range result {
				ws.WriteJSON(ZbMessage{
					Event:   "addChannel",
					Channel: v + "_ticker",
				})
			}
		},
		func(result gjson.Result) {
			success := result.Get("success").Bool()
			if success {
				result.Get("data").ForEach(func(key, value gjson.Result) bool {
					m := utils.GetCoinByZb(key.String())
					base := m["base"]
					coin := m["coin"]
					symbol := base + coin
					cache.GetInstance().HSet(ue.Name+"-symbols", symbol, symbol)
					return true
				})
			} else {
				channel := result.Get("channel").String()
				last := result.Get("ticker.last").Float()
				symbol := utils.GetCoinByZb2(channel)["symbol"]
				cache.GetInstance().HSet(ue.Name, symbol, last)
			}
		})
}

func (ue *ZbExchange) Run(symbol string) {
	//var client = &http.Client{}
	//if !IsServer {
	//	uProxy, _ := url.Parse("http://127.0.0.1:1080")
	//
	//	client = &http.Client{
	//		Transport: &http.Transport{
	//			Proxy: http.ProxyURL(uProxy),
	//		},
	//	}
	//}
	//
	//client.Timeout = time.Second * 10
	//
	//oldSymbol := symbol
	//containBtc := strings.Contains(symbol, "btc")
	//coin := ""
	//if containBtc {
	//	coins := strings.Split(symbol, "btc")
	//	if len(coins) < 2 {
	//		return
	//	}
	//	coin = coins[0]
	//}
	//symbol = coin + "_btc"
	//
	//url := "http://api.zb.com/data/v1/ticker" +
	//	"?market=" + symbol
	//
	//resp, _ := http.NewRequest("GET", url, nil)
	//
	//for {
	//	resp, err := client.Do(resp)
	//
	//	if err != nil {
	//		fmt.Println(err)
	//		continue
	//	}
	//
	//	buf := bytes.NewBuffer(make([]byte, 0, 512))
	//
	//	buf.ReadFrom(resp.Body)
	//	resp.Body.Close()
	//
	//	jsonNode := "ticker.last"
	//	result := gjson.GetBytes(buf.Bytes(), jsonNode)
	//	ue.SetPrice(oldSymbol, result.Float())
	//
	//	time.Sleep(1 * time.Second)
	//}

	cache.GetInstance().HSet(ue.Name+"-tradeFee", "taker", 0.001)
	cache.GetInstance().HSet(ue.Name+"-tradeFee", "maker", 0.001)

	event.Bus.Subscribe(ue.Name+"-getprice", ue.GetPrice)
	//获取currency和转账费
	utils.StartTimer(time.Minute*30, func() {
		utils.GetHtml("GET", "https://www.bitkk.com/i/rate", nil, func(result *goquery.Document) {
			trs := result.Find("body > div.ch-body > div.envor-content > section.envor-section > div > div > div > article > table > tbody > tr")
			trs.Each(func(i int, selection *goquery.Selection) {
				currency := selection.Find("td:nth-child(1)").Text()
				transferFee := selection.Find("td:nth-child(7)").Text()
				currency = strings.TrimSpace(currency)
				currency = strings.ToLower(currency)
				m := utils.GetByZb(transferFee)
				if currency != "" {
					n := m["num"]
					if n == "" {
						n = "-1"
					}
					cache.GetInstance().HSet(ue.Name+"-currency", currency, currency)
					cache.GetInstance().HSet(ue.Name+"-transfer", currency, n)
				}
			})
		})
	})

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		u := "http://api.zb.com/data/v1/markets"

		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				m := utils.GetCoinByZb(key.String())
				base := m["base"]
				coin := m["coin"]
				cache.GetInstance().HSet(ue.Name+"-symbols", coin+base, coin+base)
				return true
			})
		})
		event.Bus.Publish(ue.Name+"-getprice", "")
		event.Bus.Unsubscribe(ue.Name+"-getprice", ue.GetPrice)
	})

}

func (ue ZbExchange) FeesRun() {
}

func NewZbExchange() BigE {
	exchange := new(ZbExchange)
	exchange.Exchange = Exchange{
		Name: "Zb",
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
