package utils

import (
	"time"
	"regexp"
	"github.com/tidwall/gjson"
	"BitCoin/config"
	"net/http"
	"net/url"
	"bytes"
	"strings"
	"sort"
	"fmt"
	"crypto/md5"
	"github.com/gorilla/websocket"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/proxy"
	"os"
)

func StartTimer(duration time.Duration, f func()) {
	go func() {
		for {
			f()
			now := time.Now()
			next := now.Add(duration)
			//next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			t := time.NewTimer(next.Sub(now))
			<-t.C
		}
	}()
}

type MyRegexp struct {
	*regexp.Regexp
}

//add a new method to our new regular expression type
func (r *MyRegexp) FindStringSubmatchMap(s string) map[string]string {
	captures := make(map[string]string)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return captures
	}

	for i, name := range r.SubexpNames() {
		//Ignore the whole regexp match and unnamed groups
		if i == 0 || name == "" {
			continue
		}

		captures[name] = match[i]

	}
	return captures
}
func GetCoinByHuoBi(s string) string {
	var myExp = MyRegexp{regexp.MustCompile(`^market\.(?P<coin>(\w+)*)\.`)}
	m := myExp.FindStringSubmatchMap(s)
	if s, ok := m["coin"]; ok {
		return s
	}
	return ""
}
func BuildSign(params map[string]interface{}, secretKey string) string {
	sign := ""
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if len(params) == 1 {
			sign += k + "=" + params[k].(string)
		} else {
			sign += k + "=" + params[k].(string) + "&"
		}
	}
	data := []byte(sign + "secret_key=" + secretKey)
	sum := md5.Sum(data)
	md5str := fmt.Sprintf("%x", sum)
	md5str = strings.ToUpper(md5str)
	return md5str
}
func GetCoinByOkex(s string) string {
	var myExp = MyRegexp{regexp.MustCompile(`^ok_sub_spot_(?P<coin>(\w+)*)_ticker`)}
	m := myExp.FindStringSubmatchMap(s)
	if s, ok := m["coin"]; ok {
		return strings.Replace(s, "_", "-", -1)
	}
	return ""
}

func GetCoinBySymbol(symbol string) string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)(?P<base>(btc|eth|usdt))$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	if s, ok := m["coin"]; ok {
		return s
	}
	return ""
}

func GetCoinByZb(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)_(?P<base>(\w+)*)$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetCoinByLbank(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)/(?P<base>(\w+)*)$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}
func GetCoinByExx(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)_\w+_kline_1min$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetCoinByGateIO(s string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^0%\+(?P<fee>((\d+)*)(\d+(\.\d+)?))`)}
	m := myExp.FindStringSubmatchMap(s)
	return m
}

func GetCoinByZb2(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<symbol>(\w+)*)_`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetFloatByBitfinex(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`(?P<num>(\d+(\.\d+)?)) (?P<coin>(\w+)*)`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}
func GetByZb(s string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`(?P<num>(\d+(\.\d+)?)) ?(?P<coin>[a-zA-Z]+)`)}
	m := myExp.FindStringSubmatchMap(s)

	if m["num"] == "000" {
		var myExp = MyRegexp{regexp.MustCompile(`(?P<num>(\d+.?\d+)) ?(?P<coin>[a-zA-Z]+)`)}
		d := myExp.FindStringSubmatchMap(s)
		i := d["num"]
		dn := strings.Replace(i, ",", "", -1)
		d["num"] = dn
		return d
	}
	return m
}

func GetBaseBySymbol(symbol string) string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)(?P<base>(btc|eth|usdt))$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	if s, ok := m["base"]; ok {
		return s
	}
	return ""
}

func GetBaseBySymbol2(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)(?P<base>(btc|eth|usdt))$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetBaseByHuobi(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)(?P<base>(btc|eth|usdt|ht))$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetSymbolByBitfinex(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)(?P<base>(btc|eth|usd|eos|jpy|eur))$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetSymbolByBigOne(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)-(?P<base>(\w+)*)$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetSymbolByCryptopia(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<coin>(\w+)*)/(?P<base>(\w+)*)$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetBaseCoinByBittrex(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<base>(btc|eth|usdt|usd))-(?P<coin>(\w+)*)$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetSymbolByFcoin(symbol string) map[string]string {
	var myExp = MyRegexp{regexp.MustCompile(`^(?P<type>(ticker)).(?P<symbol>(\w+)*)$`)}
	m := myExp.FindStringSubmatchMap(symbol)
	return m
}

func GetJsonFromRedisString(redisStr string) gjson.Result {
	var myExp = MyRegexp{regexp.MustCompile(`{(?P<json>(.+)*)}$`)}
	m := myExp.FindStringSubmatchMap(redisStr)
	j := ""
	if s, ok := m["json"]; ok {
		j = "{" + s + "}"
	}
	g := gjson.Parse(j)
	return g
}

var IsServer = config.IsServer

func Fetch(method string, u string, header http.Header) (*http.Response, error) {
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

	resp, _ := http.NewRequest(method, u, nil)

	if header != nil {
		resp.Header = header
	}

	return client.Do(resp)
}
func FetchWithBody(method string, u string, header http.Header, body url.Values) (*http.Response, error) {
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
	var resp *http.Request
	if method == "POST" {
		return client.PostForm(u, body)
	} else {
		resp, _ = http.NewRequest(method, u, nil)
		if header != nil {
			resp.Header = header
		}

		return client.Do(resp)
	}
}

func FetchJsonWithTry(method string, u string, maxTry int64, header http.Header) {
	maxRequestCount := maxTry
	for i := maxRequestCount; i > 0; i-- {
		resp, e := Fetch(method, u, header)

		if e != nil {
			continue
		}

		buf := bytes.NewBuffer(make([]byte, 0, 512))
		buf.ReadFrom(resp.Body)
		gjson.GetBytes(buf.Bytes(), "data")
	}
}

func FetchJson(method string, u string, header http.Header) gjson.Result {
	maxRequestCount := 5
	for i := maxRequestCount; i > 0; i-- {
		resp, e := Fetch(method, u, header)

		if e != nil {
			continue
		}

		buf := bytes.NewBuffer(make([]byte, 0, 512))
		buf.ReadFrom(resp.Body)
		return gjson.ParseBytes(buf.Bytes())
	}
	return gjson.Result{}
}

func GetInfo(method string, u string, header http.Header, callback func(result gjson.Result)) {
	maxRequestCount := 5
	for i := maxRequestCount; i > 0; i-- {
		resp, e := Fetch(method, u, header)

		if e != nil {
			continue
		}

		buf := bytes.NewBuffer(make([]byte, 0, 512))
		buf.ReadFrom(resp.Body)
		result := gjson.ParseBytes(buf.Bytes())
		callback(result)
		break
	}
}
func GetInfoWithBody(method string, u string, header http.Header, body url.Values, callback func(result gjson.Result)) {
	maxRequestCount := 5
	for i := maxRequestCount; i > 0; i-- {
		resp, e := FetchWithBody(method, u, header, body)

		if e != nil {
			continue
		}

		buf := bytes.NewBuffer(make([]byte, 0, 512))
		buf.ReadFrom(resp.Body)
		result := gjson.ParseBytes(buf.Bytes())
		callback(result)
		break
	}
}
func GetHtml(method string, u string, header http.Header, callback func(result *goquery.Document)) {
	maxRequestCount := 5
	for i := maxRequestCount; i > 0; i-- {
		resp, e := Fetch(method, u, header)

		if e != nil {
			continue
		}

		doc, _ := goquery.NewDocumentFromReader(resp.Body)
		callback(doc)
		break
	}
}

func GetInfoWS(u string, header http.Header, callback func(result gjson.Result)) {
	dialer := websocket.Dialer{
	}
	if !IsServer {
		d, e := proxy.SOCKS5("tcp", "127.0.0.1:1080", nil, proxy.Direct)
		if e != nil {
			fmt.Println("请确认代理服务器开启", e)
			os.Exit(1)
		}

		dialer = websocket.Dialer{
			NetDial: d.Dial,
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

		callback(gjson.ParseBytes(m))
	}
	ws.Close()
}

func GetInfoWS2(u string, header http.Header, callback func(*websocket.Conn),
	callback2 func(result gjson.Result)) {
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
		ws, _, err = dialer.Dial(u, header)
		if err != nil {
			fmt.Println(err)
		} else {
			break
		}
	}

	callback(ws)

	for {
		if ws == nil {
			for {
				var err error
				ws, _, err = dialer.Dial(u, header)
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
			ws, _, err = dialer.Dial(u, header)
			if err != nil {
				fmt.Println(err)
			}
		}

		callback2(gjson.ParseBytes(m))
	}
	ws.Close()
}
func GetInfoWS3(u string, header http.Header, callback func(*websocket.Conn),
	callback2 func(ws *websocket.Conn, result gjson.Result)) {
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
		ws, _, err = dialer.Dial(u, header)
		if err != nil {
			fmt.Println(err)
		} else {
			break
		}
	}

	callback(ws)

	for {
		if ws == nil {
			for {
				var err error
				ws, _, err = dialer.Dial(u, header)
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
			ws, _, err = dialer.Dial(u, header)
			if err != nil {
				fmt.Println(err)
			}
		}

		callback2(ws, gjson.ParseBytes(m))
	}
	ws.Close()
}
