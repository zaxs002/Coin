package model

import (
	"BitCoin/cache"
	"github.com/tidwall/gjson"
	"BitCoin/utils"
	"time"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
	"net/url"
	"fmt"
	"bytes"
	"encoding/binary"
	"compress/gzip"
	"io/ioutil"
	"regexp"
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

var hadaxGetPrice sync.Once

func (he HadaxExchange) GetPrice() {
	hadaxGetPrice.Do(func() {
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
							//fmt.Println("订阅成功")
						}
					} else {
						if tick.String() != "" {
							symbol := gjson.GetBytes(datas, "ch")
							bi := utils.GetCoinByHuoBi(symbol.String())
							m := utils.GetBaseByHuobi(bi)
							coin := m["coin"]
							base := m["base"]
							ss := coin + "-" + base
							tick := tick.Get("close").Float()
							he.SetPrice(ss, tick)
						}
					}
				}
			}
		})
	})
}

func (he HadaxExchange) Run(symbol string) {
	he.SetTradeFee(0.002, 0.002)

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		symbolsUrl := "https://api.hadax.com/v1/hadax/settings/symbols"

		result := utils.FetchJson("GET", symbolsUrl, nil)
		datas := result.Get("data")
		datas.ForEach(func(key, value gjson.Result) bool {
			base := value.Get("base-currency").String()
			quote := value.Get("quote-currency").String()
			symbol := base + quote
			newSymbol := base + "-" + quote
			cache.GetInstance().HSet(he.Name+"-symbols", newSymbol, symbol)
			return true
		})
		he.GetPrice()
	})

	//获取转账手续费
	hb_pro_token := "O_TnMHUTc3b7upOIFHQQy2OWCyvTL2dnpmIPWKzDi8YY-uOP2m0-gvjE57ad1qDF"

	utils.StartTimer(time.Minute*30, func() {
		info := NewTransferInfo()

		var coins []string

		currenciesUrl := "https://api.hadax.com/v1/settings/currencys?language=zh-CN"

		utils.GetInfo("GET", currenciesUrl, nil, func(result gjson.Result) {
			coinsJson := result.Get("data")
			coinsJson.ForEach(func(key, value gjson.Result) bool {
				coinName := value.Get("name")
				minWithdraw := value.Get("withdraw-min-amount").Float()
				minDeposit := value.Get("deposit-min-amount").Float()
				withdrawEnabled := value.Get("withdraw-enabled").Bool()
				depositEnabled := value.Get("deposit-enabled").Bool()
				minConfirms := value.Get("fast-confirms").Float()
				coins = append(coins, coinName.String())
				s := coinName.String()

				he.SetCurrency(s)

				info.MinWithdraw = minWithdraw
				info.MinDeposit = minDeposit
				info.CanWithdraw = 0
				info.CanDeposit = 0

				if withdrawEnabled {
					info.CanWithdraw = 1
				}
				if depositEnabled {
					info.CanDeposit = 1
				}
				info.WithdrawMinConfirmations = minConfirms

				transferFeeUrl := "https://api.hadax.com/v1/dw/withdraw/get-prewithdraw?currency=" + s

				headers := http.Header{
					"hb-pro-token": []string{hb_pro_token},
				}
				utils.GetInfo("GET", transferFeeUrl, headers, func(result gjson.Result) {
					status := result.Get("status")

					if status.String() != "ok" {
						info.WithdrawFee = -1
					} else {
						datas := result.Get("data")

						fee := datas.Get("default-fee")
						maxFee := datas.Get("max-fee").Float()
						minFee := datas.Get("min-fee").Float()

						info.WithdrawFee = fee.Float()
						info.MaxWithdrawFee = maxFee
						info.MinWithdrawFee = minFee

						he.SetTransferFee(s, info)
					}
				})

				return true
			})
		})
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
