package model

import (
	"BitCoin/cache"
	"github.com/gorilla/websocket"
	"fmt"
	"bytes"
	"encoding/binary"
	"compress/gzip"
	"io/ioutil"
	"github.com/tidwall/gjson"
	"BitCoin/utils"
	"regexp"
	"time"
	"net/http"
	"net/url"
	"BitCoin/event"
	"strings"
)

type HadaxExchange struct {
	Exchange
}
type HadaxMessage struct {
	Sub string `json:"sub"`
	Id  string `json:"id"`
}

type HadaxPing struct {
	Ping int64 `json:"ping"`
}

type HadaxPong struct {
	Pong int64 `json:"pong"`
}

func (he HadaxExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he HadaxExchange) FeesRun() {
}

func (he HadaxExchange) GetPrice(s string) {
	symbolss := cache.GetInstance().HGetAll(he.Name + "-symbols")
	var symbols, _ = symbolss.Result()
	utils.StartTimer(time.Hour*24, func() {
		var u = "wss://api.hadax.com/ws"
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

		for _, v := range symbols {
			ws.WriteJSON(HuoBiMessage{
				Sub: "market." + v + ".kline.1min",
				Id:  "id1",
			})
		}
		b := new(bytes.Buffer)

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

			if string(m) == "" {
				continue
			}

			binary.Write(b, binary.LittleEndian, m)
			r, _ := gzip.NewReader(b)
			datas, _ := ioutil.ReadAll(r)
			r.Close()

			result := gjson.GetBytes(datas, "ping")
			tick := gjson.GetBytes(datas, "tick")

			if result.Int() > 0 {
				ws.WriteJSON(HuoBiPong{
					Pong: result.Int(),
				})
			} else {
				status := gjson.GetBytes(datas, "status")
				if status.String() == "ok" {
					subbed := gjson.GetBytes(datas, "subbed")
					var myExp = utils.MyRegexp{regexp.MustCompile(`^market.(?P<coin>(\w+)*)`)}
					m := myExp.FindStringSubmatchMap(subbed.String())
					if _, ok := m["coin"]; ok {
						//fmt.Println("订阅成功", s)
					}
				} else {
					if tick.String() != "" {
						symbol := gjson.GetBytes(datas, "ch")
						bi := utils.GetCoinByHuoBi(symbol.String())
						tick := tick.Get("close").Float()
						he.SetPrice(bi, tick)
					}
				}
			}
		}
	})
}

func (he HadaxExchange) Run(symbol string) {
	event.Bus.Subscribe(he.Name+":getprice", he.GetPrice)

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		symbolsUrl := "https://api.hadax.com/v1/hadax/settings/symbols"

		result := utils.FetchJson("GET", symbolsUrl, nil)
		datas := result.Get("data")
		datas.ForEach(func(key, value gjson.Result) bool {
			base := value.Get("base-currency").String()
			quote := value.Get("quote-currency").String()
			symbol := base + quote
			strings.ToUpper(symbol)
			cache.GetInstance().HSet(he.Name+"-symbols", symbol, symbol)
			return true
		})
		event.Bus.Publish(he.Name+":getprice", "")
		event.Bus.Unsubscribe(he.Name+":getprice", he.GetPrice)
	})
	//获取交易手续费
	utils.StartTimer(time.Hour*1, func() {
		tradeFeeUrl := "https://api.hadax.com/v1/settings/fee?r=ed6xcq7q61&symbols=muskbtc,tosbtc,bcvbtc,dacbtc,idtbtc,pntbtc,zjltbtc,lymbtc,sspbtc,fairbtc,yccbtc,xmxbtc,ektbtc,ftibtc,seelebtc,gvebtc,bkbtbtc,aebtc,renbtc,pcbtc,getbtc,manbtc,hotbtc,gtcbtc,portalbtc,datxbtc,18tbtc,butbtc,lxtbtc,cdcbtc,uuubtc,aacbtc,cnnbtc,uipbtc,ucbtc,gscbtc,iicbtc,mexbtc,egccbtc,shebtc,musketh,toseth,bcveth,daceth,idteth,pnteth,zjlteth,lymeth,sspeth,faireth,ycceth,xmxeth,ekteth,ftieth,seeleeth,gveeth,bkbteth,aeeth,reneth,pceth,geteth,maneth,hoteth,gtceth,portaleth,datxeth,18teth,buteth,lxteth,cdceth,uuueth,aaceth,cnneth,uipeth,uceth,gsceth,iiceth,mexeth,egcceth,sheeth"

		utils.GetInfo("GET", tradeFeeUrl, nil, func(result gjson.Result) {
			symbols := result.Get("data")

			makerFee := 0.0
			takerFee := 0.0

			symbols.ForEach(func(key, value gjson.Result) bool {
				makerFee = value.Get("maker-fee").Float()
				takerFee = value.Get("taker-fee").Float()
				return false
			})
			cache.GetInstance().HSet(he.Name+"-tradeFee", "maker", makerFee)
			cache.GetInstance().HSet(he.Name+"-tradeFee", "taker", takerFee)
		})
	})

	//获取转账手续费
	utils.StartTimer(time.Minute*15, func() {
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

		var coins []string

		currenciesUrl := "https://api.hadax.com/v1/settings/currencys?language=zh-CN"

		currenciesRequest, _ := http.NewRequest("GET", currenciesUrl, nil)
		maxRequestCount := 5

		for i := maxRequestCount; i > 0; i-- {
			resp, err := client.Do(currenciesRequest)
			if err != nil {
				continue
			}

			buf := bytes.NewBuffer(make([]byte, 0, 512))

			buf.ReadFrom(resp.Body)

			resp.Body.Close()

			status := gjson.GetBytes(buf.Bytes(), "status")

			if status.String() != "ok" {
				continue
			}

			coinsJson := gjson.GetBytes(buf.Bytes(), "data")
			coinsJson.ForEach(func(key, value gjson.Result) bool {
				coinName := value.Get("name")
				coins = append(coins, coinName.String())
				s := coinName.String()
				cache.GetInstance().HSet(he.Name+"-currency", s, s)
				return true
			})
		}

		hb_pro_token := "mu0ArFv5dLa0-SGvF5fHQxFsnDPwgx2Q4WlwnzXElLwY-uOP2m0-gvjE57ad1qDF"

		for _, s := range coins {
			fee := cache.GetInstance().HGet(he.Name+"-transfer", s)
			f, _ := fee.Float64()
			he.TransferFees.Set(s, f)

			transferFeeUrl := "https://api.hadax.com/v1/dw/withdraw/get-prewithdraw?currency=" + s

			transferRequest, _ := http.NewRequest("GET", transferFeeUrl, nil)
			transferRequest.Header = http.Header{
				"hb-pro-token": []string{hb_pro_token},
			}
			for i := maxRequestCount; i > 0; i-- {
				resp, err := client.Do(transferRequest)

				if err != nil {
					continue
				}

				buf := bytes.NewBuffer(make([]byte, 0, 512))

				buf.ReadFrom(resp.Body)

				resp.Body.Close()

				status := gjson.GetBytes(buf.Bytes(), "status")

				if status.String() != "ok" {
					he.TransferFees.Set(s, -1.0)
					cache.GetInstance().HSet(he.Name+"-transfer", s, -1.0)
					break
				}

				datas := gjson.GetBytes(buf.Bytes(), "data")

				fee := datas.Get("default-fee")

				cache.GetInstance().HSet(he.Name+"-transfer", s, fee.Float())

				he.TransferFees.Set(s, fee.Float())
				break
			}
		}
	})
}

func NewHadaxExchange() BigE {
	exchange := new(HadaxExchange)

	exchange.Exchange = Exchange{
		Name: "Hadax",
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
