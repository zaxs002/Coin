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

type BithumbExchange struct {
	Exchange
}
type BithumbMessage struct {
	Sub string `json:"sub"`
	Id  string `json:"id"`
}


func (he BithumbExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he BithumbExchange) FeesRun() {
}

func (he BithumbExchange) GetPrice(s string) {
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

func (he BithumbExchange) Run(symbol string) {
	event.Bus.Subscribe(he.Name+":getprice", he.GetPrice)

	//获取symbols
	utils.StartTimer(time.Minute*30, func() {
		symbolsUrl := "https://www.huobipro.com/-/x/pro/v1/settings/symbols?language=zh-cn"

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
		tradeFeeUrl := "https://www.huobipro.com/-/x/pro/v1/settings/fee?r=x91neu0q2cn&symbols=btcusdt,bchusdt,ethusdt,etcusdt,ltcusdt,eosusdt,xrpusdt,omgusdt,dashusdt,zecusdt,iotausdt,adausdt,steemusdt,socusdt,ctxcusdt,actusdt,btmusdt,btsusdt,ontusdt,iostusdt,htusdt,trxusdt,dtausdt,neousdt,qtumusdt,smtusdt,elausdt,venusdt,thetausdt,sntusdt,zilusdt,xemusdt,nasusdt,ruffusdt,hsrusdt,letusdt,mdsusdt,storjusdt,elfusdt,itcusdt,cvcusdt,gntusdt,bchbtc,ethbtc,ltcbtc,etcbtc,eosbtc,omgbtc,xrpbtc,dashbtc,zecbtc,iotabtc,adabtc,steembtc,polybtc,edubtc,kanbtc,lbabtc,wanbtc,bftbtc,btmbtc,ontbtc,iostbtc,htbtc,trxbtc,smtbtc,elabtc,wiccbtc,ocnbtc,zlabtc,abtbtc,mtxbtc,nasbtc,venbtc,dtabtc,neobtc,waxbtc,btsbtc,zilbtc,thetabtc,ctxcbtc,srnbtc,xembtc,icxbtc,dgdbtc,chatbtc,wprbtc,lunbtc,swftcbtc,sntbtc,meetbtc,yeebtc,elfbtc,letbtc,qtumbtc,lskbtc,itcbtc,socbtc,qashbtc,mdsbtc,ekobtc,topcbtc,mtnbtc,actbtc,hsrbtc,stkbtc,storjbtc,gnxbtc,dbcbtc,sncbtc,cmtbtc,tnbbtc,ruffbtc,qunbtc,zrxbtc,kncbtc,blzbtc,propybtc,rpxbtc,appcbtc,aidocbtc,powrbtc,cvcbtc,paybtc,qspbtc,datbtc,rdnbtc,mcobtc,rcnbtc,manabtc,utkbtc,tntbtc,gasbtc,batbtc,ostbtc,linkbtc,gntbtc,mtlbtc,evxbtc,reqbtc,adxbtc,astbtc,engbtc,saltbtc,bifibtc,bcxbtc,bcdbtc,sbtcbtc,btgbtc,eoseth,omgeth,iotaeth,adaeth,steemeth,polyeth,edueth,kaneth,lbaeth,waneth,bfteth,zrxeth,asteth,knceth,onteth,hteth,btmeth,iosteth,smteth,elaeth,trxeth,abteth,naseth,ocneth,wicceth,zileth,ctxceth,zlaeth,wpreth,dtaeth,mtxeth,thetaeth,srneth,veneth,btseth,waxeth,hsreth,icxeth,mtneth,acteth,blzeth,qasheth,ruffeth,cmteth,elfeth,meeteth,soceth,qtumeth,itceth,swftceth,yeeeth,lsketh,luneth,leteth,gnxeth,chateth,ekoeth,topceth,dgdeth,stketh,mdseth,dbceth,snceth,payeth,quneth,aidoceth,tnbeth,appceth,rdneth,utketh,powreth,bateth,propyeth,manaeth,reqeth,cvceth,qspeth,evxeth,dateth,mcoeth,gnteth,gaseth,osteth,linketh,rcneth,tnteth,engeth,salteth,adxeth"

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

		currenciesUrl := "https://www.huobipro.com/-/x/pro/v1/settings/currencys?language=zh-CN"

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

		for _, s := range coins {
			fee := cache.GetInstance().HGet(he.Name+"-transfer", s)
			f, _ := fee.Float64()
			he.TransferFees.Set(s, f)

			transferFeeUrl := "https://www.huobipro.com/-/x/pro/v1/dw/withdraw-virtual/fee-range?currency=" + s

			transferRequest, _ := http.NewRequest("GET", transferFeeUrl, nil)
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

				fee := datas.Get("default-amount")

				cache.GetInstance().HSet(he.Name+"-transfer", s, fee.Float())

				he.TransferFees.Set(s, fee.Float())
				break
			}
		}
	})
}

func NewBithumbExchange() BigE {
	exchange := new(BithumbExchange)

	exchange.Exchange = Exchange{
		Name: "Bithumb",
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
