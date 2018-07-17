package model

import (
	"BitCoin/utils"
	"strconv"
	"time"
	"github.com/tidwall/gjson"
	"github.com/gorilla/websocket"
	"strings"
	"BitCoin/cache"
	"BitCoin/event"
)

type HitbtcMessage struct {
	Method string            `json:"method"`
	Params map[string]string `json:"params"`
	Id     string            `json:"id"`
}

type HitbtcHeartBeat struct {
	Event string `json:"event"`
}

//hitbtc
type HitbtcExchange struct {
	Exchange
}

func (he HitbtcExchange) CheckCoinExist(symbol string) bool {
	return true
}
func (he HitbtcExchange) GetPrice(s string) {
	all := cache.GetInstance().HGetAll(he.Name + "-symbols")
	result, _ := all.Result()
	utils.GetInfoWS2("wss://api.hitbtc.com/api/2/ws", nil,
		func(ws *websocket.Conn) {
			for _, v := range result {
				ws.WriteJSON(HitbtcMessage{
					Id:     strconv.Itoa(int(time.Now().Unix())),
					Method: "subscribeTicker",
					Params: map[string]string{"symbol": v},
				})
			}
		},
		func(result gjson.Result) {
			method := result.Get("method").String()
			if method == "ticker" {
				symbol := result.Get("params.symbol")
				last := result.Get("params.last")

				cache.GetInstance().HSet(he.Name, symbol.String(), last.Float())
			}
		})
}

func (he HitbtcExchange) GetTransfer(s string) {

}
func (he *HitbtcExchange) Run(symbol string) {
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
	//url := "https://api.hitbtc.com/api/2/public/ticker/" + symbol
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
	//	result := gjson.GetBytes(buf.Bytes(), "bid")
	//	ue.SetPrice(symbol, result.Float())
	//
	//	time.Sleep(1 * time.Second)
	//}

	cache.GetInstance().HSet(he.Name+"-tradeFee", "taker", 0.001)
	cache.GetInstance().HSet(he.Name+"-tradeFee", "maker", 0.001)

	event.Bus.Subscribe(he.Name+"-getprice", he.GetPrice)
	utils.StartTimer(time.Minute*30, func() {
		//获取symbols
		u := "https://api.hitbtc.com/api/2/public/symbol"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				symbol := value.Get("id").String()
				symbol = strings.ToLower(symbol)

				cache.GetInstance().HSet(he.Name+"-symbols", symbol, symbol)
				return true
			})
		})
		event.Bus.Publish(he.Name+"-getprice", "")
		event.Bus.Unsubscribe(he.Name+"-getprice", he.GetPrice)
	})

	utils.StartTimer(time.Minute*30, func() {
		//获取currency
		u := "https://api.hitbtc.com/api/2/public/currency"
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.ForEach(func(key, value gjson.Result) bool {
				currency := value.Get("id").String()
				fee := value.Get("payoutFee").Float()
				currency = strings.ToLower(currency)

				cache.GetInstance().HSet(he.Name+"-currency", currency, currency)
				cache.GetInstance().HSet(he.Name+"-transfer", currency, fee)
				return true
			})
		})
	})
}

func (he HitbtcExchange) FeesRun() {
}

func NewHitbtcExchange() BigE {
	exchange := new(HitbtcExchange)
	exchange.Exchange = Exchange{
		Name: "Hitbtc",
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
