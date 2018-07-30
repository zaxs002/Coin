package model

import (
	"BitCoin/cache"
	"bytes"
	"github.com/tidwall/gjson"
	"BitCoin/utils"
	"time"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"regexp"
	"github.com/gorilla/websocket"
	"fmt"
	"encoding/binary"
	"compress/gzip"
	"io/ioutil"
)

type HuoBiExchange struct {
	Exchange
}
type HuoBiMessage struct {
	Sub string `json:"sub"`
	Id  string `json:"id"`
}

type HuoBiPing struct {
	Ping int64 `json:"ping"`
}

type HuoBiPong struct {
	Pong int64 `json:"pong"`
}

func (he HuoBiExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he HuoBiExchange) FeesRun() {
}

var huobiGetPrice sync.Once

func (he HuoBiExchange) GetPrice() {
	huobiGetPrice.Do(func() {
		symbolss := cache.GetInstance().HGetAll(he.Name + "-symbols")
		var symbols, _ = symbolss.Result()
		utils.StartTimer(time.Hour*24, func() {
			var u = "wss://api.huobi.pro/ws"
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

			for k := range symbols {
				ws.WriteJSON(HuoBiMessage{
					Sub: "market." + k + ".kline.1min",
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
							m := utils.GetBaseByHuobi(bi)
							coin := m["coin"]
							base := m["base"]
							s := coin + "-" + base
							tick := tick.Get("close").Float()
							he.SetPrice(s, tick)
						}
					}
				}
			}
		})
	})
}

func (he HuoBiExchange) Run(symbol string) {
	he.SetTradeFee(0.002, 0.002)

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		symbolsUrl := "https://www.huobipro.com/-/x/pro/v1/settings/symbols?language=zh-cn"

		result := utils.FetchJson("GET", symbolsUrl, nil)
		datas := result.Get("data")
		datas.ForEach(func(key, value gjson.Result) bool {
			base := value.Get("base-currency").String()
			quote := value.Get("quote-currency").String()
			symbol := base + quote
			newSymbol := base + "-" + quote
			strings.ToUpper(symbol)

			he.SetSymbol(symbol, newSymbol)

			return true
		})
		he.GetPrice()
	})
	//获取转账手续费
	utils.StartTimer(time.Minute*30, func() {
		info := NewTransferInfo()
		currenciesUrl := "https://www.huobipro.com/-/x/pro/v1/settings/currencys?language=zh-CN"
		utils.GetInfo("GET", currenciesUrl, nil, func(result gjson.Result) {
			coinsJson := result.Get("data")
			coinsJson.ForEach(func(key, value gjson.Result) bool {
				coinName := value.Get("name")
				minWithdraw := value.Get("withdraw-min-amount").Float()
				minDeposit := value.Get("deposit-min-amount").Float()
				withdrawEnabled := value.Get("withdraw-enabled").Bool()
				depositEnabled := value.Get("deposit-enabled").Bool()
				minConfirms := value.Get("fast-confirms").Float()
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

				transferFeeUrl := "https://www.huobipro.com/-/x/pro/v1/dw/withdraw-virtual/fee-range?currency=" + s

				utils.GetInfo("GET", transferFeeUrl, nil, func(result gjson.Result) {
					status := result.Get("status")

					if status.String() != "ok" {
						info.WithdrawFee = -1
					} else {
						datas := result.Get("data")

						fee := datas.Get("default-amount")
						maxFee := datas.Get("max-amount").Float()
						minFee := datas.Get("min-amount").Float()

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

func NewHuoBiExchange() BigE {
	exchange := new(HuoBiExchange)

	exchange.Exchange = Exchange{
		Name: "HuoBi",
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
