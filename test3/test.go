package main

import (
	"github.com/gorilla/websocket"
	"fmt"
	"github.com/tidwall/gjson"
	"BitCoin/utils"
	"net/url"
	"net/http"
	"sort"
	"crypto/md5"
	"strings"
	"time"
)

type OkexMessage struct {
	Event      string                 `json:"event"`
	Channel    string                 `json:"channel"`
	Parameters map[string]interface{} `json:"parameters"`
}

type OkexHeartBeat struct {
	Event string `json:"event"`
}

func buildSign(params map[string]interface{}, secretKey string) string {
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
	fmt.Println(sign)
	data := []byte(sign + "secret_key=" + secretKey)
	sum := md5.Sum(data)
	md5str := fmt.Sprintf("%x", sum)
	md5str = strings.ToUpper(md5str)
	return md5str
}

func main() {
	ValidSymbols := []string{"eth_btc", "etc_btc", "eos_btc", "xrp_btc"}
	apiKey := "c4eb13f2-b3d8-4446-9ddd-a48919e14a8e"
	secretKey := "A8E98839AAA88020FAD749DE33566A89"
	var m = map[string]interface{}{"api_key": apiKey}
	sign := buildSign(m, secretKey)

	var u = "wss://real.okex.com:10441/websocket"
	uProxy, _ := url.Parse("http://127.0.0.1:1080")
	dialer := websocket.Dialer{
		Proxy: http.ProxyURL(uProxy),
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

	for _, v := range ValidSymbols {
		channel := "ok_sub_spot_" + v + "_ticker"
		ws.WriteJSON(OkexMessage{
			Event:   "addChannel",
			Channel: channel,
			Parameters: map[string]interface{}{
				"api_key": apiKey, "sign": sign,
			},
		})
		fmt.Println("发送查询:", v)
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
				c := gjson.GetBytes(m, "0.data.channel")
				fmt.Printf("%s订阅成功\n", utils.GetCoinByOkex(c.String()))
			} else {
				bi := utils.GetCoinByOkex(channel.String())
				last := gjson.GetBytes(m, "0.data.last").Float()
				fmt.Printf("%s的价格:%f\n", bi, last)
			}
		}else{
			fmt.Println(string(m))
		}
	}
}
