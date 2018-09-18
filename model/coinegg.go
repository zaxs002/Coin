package model

import (
	"BitCoin/cache"
	"BitCoin/utils"
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/websocket"
	"github.com/robertkrimen/otto"
	"github.com/tidwall/gjson"
	"golang.org/x/net/proxy"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var CoineggOnce sync.Once
var CoineggOnce2 sync.Once

type CoineggExchange struct {
	Exchange
}

func (he CoineggExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he CoineggExchange) FeesRun() {
}

func (he CoineggExchange) GetTransfer() {
	once.Do(func() {
		//获取转账手续费
		token := "p83T5HXezbSKLzNYEQhBy9GknGivAufxKWoRA2iSQUxFSTvNLPdIR0pR0ZZacbjW9LXgaSrAL_P47V9ZxHZRtg=="

		all := cache.GetInstance().HGetAll(he.Name + "-currency")
		result, _ := all.Result()

		transferFeeUrl := "https://exchange.fcoin.com/api/web/v1/accounts/withdraws/fee?currency="

		headers := http.Header{
			"token": []string{token},
		}

		for _, v := range result {
			utils.GetInfo("GET", transferFeeUrl+v, headers, func(result gjson.Result) {
				fee := result.Get("data").Float()
				cache.GetInstance().HSet(he.Name+"-transfer", v, fee)
			})
		}
	})
}

func (he CoineggExchange) GetPrice() {
	once2.Do(func() {
		//获取价格
		all := cache.GetInstance().HGetAll(he.Name + "-symbols")
		result, _ := all.Result()

		u := "wss://api.fcoin.com/v2/ws"
		utils.GetInfoWS2(u, nil,
			func(ws *websocket.Conn) {
				for _, v := range result {
					ws.WriteJSON(FcoinMessage{
						Cmd:  "sub",
						Args: []string{"ticker." + v},
					})
				}
			}, func(result gjson.Result) {
				last := result.Get("ticker.0").Float()
				symbol := result.Get("type").String()
				m := utils.GetSymbolByFcoin(symbol)
				if m != nil {
					if m["type"] == "ticker" {
						s := m["symbol"]
						cache.GetInstance().HSet(he.Name, s, last)
					}
				}
			})
	})
}

func (he CoineggExchange) Run(symbol string) {
	bases := []string{"btc", "usdt", "eth", "usc"}

	he.SetTradeFee(0.001, 0.001)

	//获取symbols和价格
	h := http.Header{
		"user-agent":                []string{"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36"},
		"referer":                   []string{"https://www.coinegg.com/"},
		"accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8"},
		"accept-encoding":           []string{"gzip, deflate, br"},
		"accept-language":           []string{"zh-CN,zh;q=0.9,en;q=0.8,ru;q=0.7"},
		"upgrade-insecure-requests": []string{"1"},
	}

	he.HackCloudflare(h)

	he.GetCoinEggCookies(h)

	utils.StartTimer(time.Second, func() {
		url := "https://www.coinegg.com/coin/%s/allcoin"

		for _, v := range bases {
			baseUrl := fmt.Sprintf(url, v)

			utils.GetInfo("GET", baseUrl, h, func(result gjson.Result) {
				result.ForEach(func(key, value gjson.Result) bool {
					s := key.String()
					last := value.Get("1").Float()
					symbol := s + "-" + v

					he.SetSymbol(symbol, symbol)
					he.SetPrice(symbol, last)
					return true
				})
			})
		}
	})

	utils.StartTimer(time.Minute*30, func() {
		u := "https://www.coinegg.com/fee.html"
		utils.GetHtml("GET", u, nil, func(result *goquery.Document) {
			//body > div.body > div.right.gonggao > div > ul:nth-child(3) > li.w170
			result.Find("body > div.body > div.right.gonggao > div > ul.noticeListHeadBody").Each(func(i int, selection *goquery.Selection) {
				text := selection.Find("li:nth-child(4)").Text()
				c := selection.Find("li:nth-child(2)").Text()
				m := utils.GetByZb(text)
				coin := m["coin"]
				info := NewTransferInfo()
				info.CanWithdraw = 1
				if coin != "" && strings.TrimSpace(coin) != "" {
					num := m["num"]
					f, _ := strconv.ParseFloat(num, 64)
					info.WithdrawFee = f
					he.SetTransferFee(c, info)
				} else {
					info.WithdrawFee = -1
					he.SetTransferFee(c, info)
				}
				he.SetCurrency(c)
			})
		})
	})
}

func (he CoineggExchange) HackCloudflare(h http.Header) string {
	var client = &http.Client{}
	if !IsServer {
		dialer, e := proxy.SOCKS5("tcp", "127.0.0.1:1080", nil, proxy.Direct)
		if e != nil {
			fmt.Println("请确认代理服务器开启", e)
			os.Exit(1)
		}
		httpTransport := &http.Transport{}
		client = &http.Client{Transport: httpTransport}
		httpTransport.Dial = dialer.Dial
	}
	client.Timeout = time.Second * 20
	resp, _ := http.NewRequest("GET", "https://www.coinegg.com", nil)
	resp.Header = h
	response, _ := client.Do(resp)
	cookies := response.Cookies()
	c := cookies[0].Name + "=" + cookies[0].Value
	h.Add("cookie", c)

	buf := bytes.NewBuffer(make([]byte, 0, 512))
	buf.ReadFrom(response.Body)

	s := string(buf.Bytes())

	to2 := strings.Split(s, "setTimeout(function(){")
	to1 := strings.Split(to2[1], "}, 4000);")
	jsCode := to1[0]
	jsCode = strings.Replace(jsCode,
		"t.substr(r.length); t = t.substr(0,t.length-1);",
		"\"www.coinegg.com\"", -1)
	jsCode = strings.Replace(jsCode,
		"t = document.createElement('div');",
		"", -1)
	jsCode = strings.Replace(jsCode,
		`t.innerHTML="<a href='/'>x</a>";`,
		"", -1)
	jsCode = strings.Replace(jsCode,
		`t = t.firstChild.href;r = t.match(/https?:\/\//)[0];`,
		"", -1)
	jsCode = strings.Replace(jsCode,
		`t = t.substr(r.length); t = t.substr(0,t.length-1);`,
		"", -1)
	jsCode = strings.Replace(jsCode,
		`a = document.getElementById('jschl-answer');`,
		"", -1)
	jsCode = strings.Replace(jsCode,
		`f = document.getElementById('challenge-form');`,
		"", -1)
	jsCode = strings.Replace(jsCode,
		`f.action += location.hash;`,
		"", -1)
	jsCode = strings.Replace(jsCode,
		`f.submit();`,
		"", -1)
	jsCode = strings.Replace(jsCode,
		`a.value =`,
		"a =", -1)
	vm := otto.New()
	vm.Run(jsCode)
	value, _ := vm.Get("a")

	jschl_vc := strings.Split(s, `<input type="hidden" name="jschl_vc" value="`)[1]
	jschl_vc = strings.Split(jschl_vc, `"/>`)[0]

	pass := strings.Split(s, `<input type="hidden" name="pass" value="`)[1]
	pass = strings.Split(pass, `"/>`)[0]

	u := "https://www.coinegg.com/cdn-cgi/l/chk_jschl?"
	u += "jschl_vc=" + jschl_vc
	u += "&pass=" + pass
	u += "&jschl_answer=" + value.String()
	resp, _ = http.NewRequest("GET", u, nil)
	resp.Header = h

	println(u)
	fmt.Println(h)

	time.Sleep(4 * time.Second)
	response, _ = client.Do(resp)
	fmt.Println(response.Cookies())

	return ""
}
func (he CoineggExchange) GetCoinEggCookies(headers http.Header) {
	//u := "https://www.coinegg.com/cdn-cgi/l/chk_jschl"

}

func NewCoineggExchange() BigE {
	exchange := new(CoineggExchange)

	exchange.Exchange = Exchange{
		Name: "CoinEgg",
		PriceQueue: LockMap{
			M: make(map[string]float64),
		},
		AmountDict: LockMap{
			M: make(map[string]float64),
		},
		TradeFees: LockMap{
			M: make(map[string]float64),
		},
		TransferFees: LockMapString{
			M: make(map[string]string),
		},
		Sub: exchange,
	}

	var duitai BigE = exchange
	return duitai
}
