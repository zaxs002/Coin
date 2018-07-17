package model

import (
	"fmt"
	"BitCoin/utils"
	"net/url"
	"net/http"
	"time"
	"github.com/tidwall/gjson"
	"bytes"
	"BitCoin/cache"
	"strings"
	"github.com/gorilla/websocket"
	"BitCoin/event"
	"encoding/json"
)

type OkexMessage struct {
	Event      string                 `json:"event"`
	Channel    string                 `json:"channel"`
	Parameters map[string]interface{} `json:"parameters"`
}

type OkexHeartBeat struct {
	Event string `json:"event"`
}
type OkexExchange struct {
	Exchange
}

func (he OkexExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he OkexExchange) GetPrice(s string) {
	utils.StartTimer(time.Hour*24, func() {
		//websocket
		all := cache.GetInstance().HGetAll(he.Name + "-symbols")
		result, _ := all.Result()

		apiKey := "c4eb13f2-b3d8-4446-9ddd-a48919e14a8e"
		secretKey := "A8E98839AAA88020FAD749DE33566A89"
		var m = map[string]interface{}{"api_key": apiKey}
		sign := utils.BuildSign(m, secretKey)

		var u = "wss://real.okex.com:10441/websocket"
		dialer := websocket.Dialer{}
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
		defer ws.Close()

		for _, v := range result {
			channel := "ok_sub_spot_" + v + "_ticker"
			ws.WriteJSON(OkexMessage{
				Event:   "addChannel",
				Channel: channel,
				Parameters: map[string]interface{}{
					"api_key": apiKey, "sign": sign,
				},
			})
		}
		utils.StartTimer(time.Second*30, func() {
			ws.WriteJSON(OkexHeartBeat{
				Event: "ping",
			})
		})
		for {
			_, m, err := ws.ReadMessage()

			if err != nil {
				for ; ; {
					var err error
					ws, _, err = dialer.Dial(u, nil)
					if err != nil {
						fmt.Println(err)
					} else {
						break
					}
				}
			}

			heartBeat := gjson.GetBytes(m, "event").String()

			if heartBeat == "" {
				channel := gjson.GetBytes(m, "0.channel")

				if channel.String() == "addChannel" {
					//c := gjson.GetBytes(m, "0.data.channel")
					//fmt.Printf("%s订阅成功\n", utils.GetCoinByOkex(c.String()))
				} else {
					bi := utils.GetCoinByOkex(channel.String())
					last := gjson.GetBytes(m, "0.data.last").Float()
					he.SetPrice(bi, last)
					cache.GetInstance().HSet(he.Name, bi, last)
				}
			} else {
			}
		}
		//websocket
	})
}

func (he OkexExchange) GetTransfer(token string) {
	utils.StartTimer(time.Minute*30, func() {
		all := cache.GetInstance().HGetAll(he.Name + "-currency")
		result, _ := all.Result()
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

		maxRequestCount := 5

		for k, v := range result {
			lower := strings.ToLower(v)
			fee := cache.GetInstance().HGet(he.Name+"-transfer", lower)
			f := fee.String()

			result := utils.GetJsonFromRedisString(f)
			o := result.Get("feeDefault")
			he.TransferFees.Set(lower, o.Float())

			url := "https://www.okex.com/v2/asset/withdraw?currencyId=" + k
			req, _ := http.NewRequest("GET", url, nil)
			req.Header = http.Header{
				"authorization": []string{token},
			}
			for i := maxRequestCount; i > 0; i-- {
				resp, err := client.Do(req)

				if err != nil {
					continue
				}

				buf := bytes.NewBuffer(make([]byte, 0, 512))

				buf.ReadFrom(resp.Body)
				resp.Body.Close()

				feeDefault := gjson.GetBytes(buf.Bytes(), "data.feeDefault")
				feeMax := gjson.GetBytes(buf.Bytes(), "data.feeMax")
				feeMin := gjson.GetBytes(buf.Bytes(), "data.feeMin")
				m := D{"feeMax": feeMax.Float(), "feeMin": feeMin.Float(), "feeDefault": feeDefault.Float()}
				marshal, _ := json.Marshal(m)
				cache.GetInstance().HSet(he.Name+"-transfer", lower, marshal)
				he.TransferFees.Set(lower, feeDefault.Float())
				break
			}
		}
	})
}

func (he OkexExchange) GetTrade(token string) {
	utils.StartTimer(time.Minute*30, func() {
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

		tradeFeeUrl := "https://www.okex.com/v2/spot/user-level"
		req, _ := http.NewRequest("GET", tradeFeeUrl, nil)
		req.Header = http.Header{
			"authorization": []string{token},
		}
		for {
			resp, e := client.Do(req)
			if e != nil {
				continue
			}
			buf := bytes.NewBuffer(make([]byte, 0, 512))
			buf.ReadFrom(resp.Body)

			takerFee := gjson.GetBytes(buf.Bytes(), "data.takerFeeRatio").Float()
			makerFee := gjson.GetBytes(buf.Bytes(), "data.makerFeeRatio").Float()

			takerFee = takerFee / 100
			makerFee = makerFee / 100

			cache.GetInstance().HSet(he.Name+"-tradeFee", "taker", takerFee)
			cache.GetInstance().HSet(he.Name+"-tradeFee", "maker", makerFee)
		}
	})
}

func (he OkexExchange) check(s string) {
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
func (he *OkexExchange) Run(symbol string) {
	coin := utils.GetCoinBySymbol(symbol)
	base := utils.GetBaseBySymbol(symbol)
	symbol = coin + "_" + base

	token := "eyJhbGciOiJIUzUxMiJ9.eyJqdGkiOiJhZjc2NzgzNS0yMGFlLTQ0M2UtODg2NC03MjAyYTc1MDRiYTRNRHpNIiwidWlkIjoieFpRVzYxdnN6MFI5MmJ0TEw0RDZDUT09Iiwic3ViIjoiMTg1KioqNzgyNyIsImVtbCI6Ind1ZGlnb2QxM0AxNjMuY29tIiwic3RhIjowLCJtaWQiOjAsImlhdCI6MTUyOTY0MjQ5NywiZXhwIjoxNTMwMjQ3Mjk3LCJpc3MiOiJva2NvaW4ifQ.sE35xEl0K99KHz8S4XLHQFzf2b8PALnq3NVpWen1NYlIJCThkbFKkhMDwjxCL2ONngq_1v76vp9Wu5DSWAnZ5g"

	event.Bus.Subscribe(he.Name+":getprice", he.GetPrice)
	event.Bus.Subscribe(he.Name+":gettransfer", he.GetTransfer)
	event.Bus.Subscribe(he.Name+":gettrade", he.GetTrade)
	event.Bus.Subscribe(he.Name+":check", he.check)

	//获取货币
	utils.StartTimer(time.Minute*1, func() {
		h := http.Header{
			"authorization": []string{token},
		}
		utils.GetInfo("GET", "https://www.okex.com/v2/asset/accounts/currencies", h, func(result gjson.Result) {
			currencies := result.Get("data.#.currency")
			currencyIds := result.Get("data.#.currencyId")
			currencyIdsArr := currencyIds.Array()

			count := 0
			currencies.ForEach(func(key, value gjson.Result) bool {
				id := currencyIdsArr[count]
				currency := value.String()
				count++
				cache.GetInstance().HSet(he.Name+"-currency", id.String(), currency)
				return true
			})
		})

		event.Bus.Publish(he.Name+":gettransfer", token)
		event.Bus.Publish(he.Name+":gettrade", token)
		event.Bus.Unsubscribe(he.Name+":gettransfer", he.GetTransfer)
		event.Bus.Unsubscribe(he.Name+":gettrade", he.GetTrade)
	})

	utils.StartTimer(time.Minute*30, func() {
		h := http.Header{
			"authorization": []string{token},
		}
		utils.GetInfo("GET", "https://www.okex.com/v2/spot/new-collect", h, func(result gjson.Result) {
			symbols := result.Get("data.#.symbol")
			symbols.ForEach(func(key, value gjson.Result) bool {
				s := value.String()
				cache.GetInstance().HSet(he.Name+"-symbols", s, s)
				return true
			})
		})
		event.Bus.Publish(he.Name+":getprice", "")
		event.Bus.Unsubscribe(he.Name+":getprice", he.GetPrice)
	})

	//检测价格是否全部获取完成
	utils.StartTimer(time.Second, func() {
		event.Bus.Publish(he.Name+":check", "")
	})
}

func (he OkexExchange) FeesRun() {
}

func NewOkexExchange() BigE {
	exchange := new(OkexExchange)
	exchange.Exchange = Exchange{
		Name: "Okex",
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
