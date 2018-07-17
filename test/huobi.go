package main

import (
	"net/url"
	"github.com/gorilla/websocket"
	"net/http"
	"fmt"
	"github.com/tidwall/gjson"
	"compress/gzip"
	"bytes"
	"io/ioutil"
	"encoding/binary"
)

type Message struct {
	Req string `json:"req"`
	Id  string `json:"id"`
}

type Ping struct {
	Ping int64 `json:"ping"`
}

type Pong struct {
	Pong int64 `json:"pong"`
}

func main() {
	var u = "wss://api.huobipro.com/ws"
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

	ws.WriteJSON(Message{
		Req: "market.ethbtc.kline.1min",
		Id:  "id1",
	})
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

		binary.Write(b, binary.LittleEndian, m)
		r, _ := gzip.NewReader(b)
		datas, _ := ioutil.ReadAll(r)
		r.Close()

		result := gjson.GetBytes(datas, "ping")
		result2 := gjson.GetBytes(datas, "data")

		if result.Int() > 0 {
			ws.WriteJSON(Pong{
				Pong: result.Int(),
			})
			ws.WriteJSON(Message{
				Req: "market.ethbtc.kline.1min",
				Id:  "id10",
			})
		} else {
			arr := result2.Array()
			fmt.Println(arr[len(arr)-1])
		}
	}
}
