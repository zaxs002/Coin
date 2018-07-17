package model

import (
	"fmt"
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"BitCoin/cache"
	"BitCoin/event"
	"github.com/gorilla/websocket"
	"net/url"
	"net/http"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"strconv"
)

//Bitfinex
type BitfinexMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Pair    string `json:"pair"`
}
type BitfinexHeartBeat struct {
	Event string `json:"event"`
}
type BitfinexExchange struct {
	Exchange
}

func (he BitfinexExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he BitfinexExchange) GetPrice(s string) {
	all := cache.GetInstance().HGetAll(he.Name + "-symbols")
	result, _ := all.Result()

	pairId := D{}

	u := "wss://api.bitfinex.com/ws/"
	dialer := websocket.Dialer{
	}
	if !IsServer {
		uProxy, _ := url.Parse("http://127.0.0.1:1080")

		dialer = websocket.Dialer{
			Proxy: http.ProxyURL(uProxy),
		}
	}

	var ws *websocket.Conn
	for {
		var err error
		ws, _, err = dialer.Dial(u, nil)
		if err != nil {
			fmt.Println(err)
		} else {
			break
		}
	}

	for _, v := range result {
		ws.WriteJSON(BitfinexMessage{
			Event:   "subscribe",
			Channel: "ticker",
			Pair:    v,
		})
	}

	for {
		if ws == nil {
			for {
				var err error
				ws, _, err = dialer.Dial(u, nil)
				if err != nil {
					fmt.Println(err)
				} else {
					break
				}
			}
		}
		_, m, err := ws.ReadMessage()

		if err != nil {
			var err error
			ws, _, err = dialer.Dial(u, nil)
			if err != nil {
				fmt.Println(err)
			}
		}

		result := gjson.ParseBytes(m)
		event := result.Get("event").String()

		if event == "subscribed" {
			chanid := result.Get("chanId").String()
			pair := result.Get("pair").String()
			pairId[chanid] = pair
		} else if event == "pong" {
		} else if event == "info" {
		} else {
			if result.Get("#").Int() == 2 {
				ws.WriteJSON(BitfinexHeartBeat{
					Event: "ping",
				})
			} else {
				chanid := result.Get("0").String()
				pair := pairId[chanid]
				price := result.Get("7").Float()
				if pair != nil {
					i := pair.(string)
					i = strings.ToLower(i)
					cache.GetInstance().HSet(he.Name, i, price)
				}
			}
		}
	}
	ws.Close()

	//priceUrl := "https://api.bitfinex.com/v2/tickers?symbols=ALL"
	//utils.StartTimer(time.Second*3, func() {
	//	utils.GetInfo("GET", priceUrl, nil, func(result gjson.Result) {
	//		fmt.Println(result)
	//		result.ForEach(func(key, value gjson.Result) bool {
	//			oldName := value.Get("0").String()
	//			if !strings.Contains(oldName, "f") {
	//				name := strings.Replace(oldName, "t", "", -1)
	//				name = strings.ToLower(name)
	//				price := value.Get("7").Float()
	//				fmt.Println(name, "的价格:", price)
	//				cache.GetInstance().HSet(he.Name, name, price)
	//			}
	//			return true
	//		})
	//	})
	//})
}

func (he BitfinexExchange) GetTransfer(s string) {
	utils.StartTimer(time.Minute*30, func() {
		u := "https://www.bitfinex.com/fees"

		resp, _ := utils.Fetch("GET", u, nil)
		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		doc.Find("#fees-page > div > table:nth-child(16) > tbody >tr").Each(func(i int, selection *goquery.Selection) {
			key := selection.Find("span").Text()
			upperkey := strings.ToUpper(key)
			short := cache.GetInstance().HGet("FullToShort", key).Val()
			if short == "" {
				short = cache.GetInstance().HGet("FullToShort", upperkey).Val()
			}
			val := selection.Find("td.bfx-green-text.col-info").Text()
			val = strings.TrimSpace(val)
			bitfinex := utils.GetFloatByBitfinex(val)
			if len(bitfinex) == 0 {
				//fmt.Println(val, "免费")
			} else {
				coin := bitfinex["coin"]
				num := bitfinex["num"]
				n, _ := strconv.ParseFloat(num, 64)
				coin = strings.ToLower(coin)
				cache.GetInstance().HSet(he.Name+"-transfer", coin, n)
			}
		})
	})
}

func (he BitfinexExchange) Run(symbol string) {
	//oldSymbol := symbol
	//coin := utils.GetCoinBySymbol(symbol)
	//base := utils.GetBaseBySymbol(symbol)
	//symbol = base + "-" + coin
	//
	//utils.StartTimer(time.Millisecond*500, func() {
	//	var client = &http.Client{}
	//	if !IsServer {
	//		uProxy, _ := url.Parse("http://127.0.0.1:1080")
	//
	//		client = &http.Client{
	//			Transport: &http.Transport{
	//				Proxy: http.ProxyURL(uProxy),
	//			},
	//		}
	//	}
	//
	//	client.Timeout = time.Second * 10
	//
	//	//https://bittrex.com/api/v1.1/public/getticker?market=BTC-LTC
	//	url := "https://bittrex.com/api/v1.1/public/getticker"
	//	url += "?"
	//	url += "&market=" + symbol
	//
	//	resp, _ := http.NewRequest("GET", url, nil)
	//
	//	for {
	//		resp, err := client.Do(resp)
	//
	//		if err != nil {
	//			continue
	//		}
	//
	//		buf := bytes.NewBuffer(make([]byte, 0, 512))
	//
	//		buf.ReadFrom(resp.Body)
	//		resp.Body.Close()
	//
	//		result := gjson.GetBytes(buf.Bytes(), "result.Last")
	//		he.SetPrice(oldSymbol, result.Float())
	//		break
	//	}
	//})

	cache.GetInstance().HSet(he.Name+"-tradeFee", "taker", 0.002)
	cache.GetInstance().HSet(he.Name+"-tradeFee", "maker", 0.001)

	event.Bus.Subscribe(he.Name+":getprice", he.GetPrice)
	event.Bus.Subscribe(he.Name+":gettransfer", he.GetTransfer)

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		symbolsUrl := "https://api.bitfinex.com/v1/symbols"
		utils.GetInfo("GET", symbolsUrl, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				s := value.String()
				cache.GetInstance().HSet(he.Name+"-symbols", s, s)
				return true
			})
		})

		event.Bus.Publish(he.Name+":getprice", "")
		event.Bus.Unsubscribe(he.Name+":getprice", he.GetPrice)
	})

	utils.StartTimer(time.Minute*10, func() {
		//全名变缩写
		attrUrl := "https://api.feixiaohao.com/search/relatedallword"
		utils.GetInfo("GET", attrUrl, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				s := value.String()
				ss := strings.Split(s, "#")
				quanming := ss[2]
				suoxie := ss[4]

				cache.GetInstance().HSet("FullToShort", quanming, suoxie)
				return true
			})
		})
	})

	utils.StartTimer(time.Minute*30, func() {
		//获取currency
		currencyUrl := "https://www.bitfinex.com/account/_bootstrap/"

		utils.GetInfo("GET", currencyUrl, nil, func(result gjson.Result) {
			currencies := result.Get("all_currencies")
			currencies.ForEach(func(key, value gjson.Result) bool {
				s := value.String()
				s = strings.ToLower(s)
				cache.GetInstance().HSet(he.Name+"-currency", s, s)
				cache.GetInstance().HSet(he.Name+"-transfer", s, 0.0)
				return true
			})
		})
		event.Bus.Publish(he.Name+":gettransfer", "")
		event.Bus.Unsubscribe(he.Name+":gettransfer", he.GetTransfer)
	})
}

func (he BitfinexExchange) FeesRun() {
	fmt.Println("Old FeesRun")
}

func NewBitfinexExchange() BigE {
	exchange := new(BitfinexExchange)

	exchange.Exchange = Exchange{
		Name: "Bitfinex",
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
