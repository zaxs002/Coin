package main

import (
	"BitCoin/utils"
	"github.com/tidwall/gjson"
	"github.com/gorilla/websocket"
	"fmt"
	"net/http"
)

type UpbitMessage1 struct {
	Ticket string `json:"ticket"`
}

type UpbitMessage2 struct {
	Type  string   `json:"type"`
	Codes []string `json:"codes"`
}

type UpbitMessage3 struct {
	Type      string   `json:"type"`
	Codes     []string `json:"codes"`
	AccessKey string   `json:"accessKey"`
}

func main() {
	u := "wss://crix-websocket.upbit.com/sockjs/834/jtrh5nsc/websocket"

	utils.GetInfoWS2(u, http.Header{
		"Cookie": []string{"Cookie:__cfduid=d16b91e6ac840edb8331fc96a9f49aba91531445208; _ga=GA1.2.371518422.1531445218; _gid=GA1.2.1312130944.1531445218; __cfruid=873b5c92cd460cff02946c125c25a63e01a55c76-1531449729; amplitude_id_totalupbit.com=eyJkZXZpY2VJZCI6IjBjMDMxMzI1LWIwOGYtNGFiMC1iZjE5LTc4NWM5NzVjNjZhOVIiLCJ1c2VySWQiOiIzMDM5MDc4ZS1jMjFkLTRkNmEtYTBhNS0yNDUyYmIyMjJlN2QiLCJvcHRPdXQiOmZhbHNlLCJzZXNzaW9uSWQiOjE1MzE0NDk3MzI5MzUsImxhc3RFdmVudFRpbWUiOjE1MzE0NTExNzU0MzQsImV2ZW50SWQiOjI0LCJpZGVudGlmeUlkIjoyNiwic2VxdWVuY2VOdW1iZXIiOjUwfQ==; amplitude_id_sampleupbit.com=eyJkZXZpY2VJZCI6IjIyNjIzOTQ3LWM2OWQtNDljYy05ODg1LWI0NGU1ODc0NTZjZVIiLCJ1c2VySWQiOiIzMDM5MDc4ZS1jMjFkLTRkNmEtYTBhNS0yNDUyYmIyMjJlN2QiLCJvcHRPdXQiOmZhbHNlLCJzZXNzaW9uSWQiOjE1MzE0NDk3MzI5MzgsImxhc3RFdmVudFRpbWUiOjE1MzE0NTExNzU0MzYsImV2ZW50SWQiOjI0LCJpZGVudGlmeUlkIjoyNiwic2VxdWVuY2VOdW1iZXIiOjUwfQ==; JSESSIONID=dummy"},
	},
		func(ws *websocket.Conn) {
			ws.WriteJSON(UpbitMessage1{
				Ticket: "ram macbook",
			})
			ws.WriteJSON(UpbitMessage2{
				Type:  "recentCrix",
				Codes: []string{"CRIX.UPBIT.KRW-BTC"},
			})
			ws.WriteJSON(UpbitMessage3{
				Type:      "crixOrder",
				Codes:     []string{"ALL"},
				AccessKey: "EzxOA7AqVkP4E4MIOkNt9Ydp38ZxlcTkMufRU3Ob",
			})
		}, func(result gjson.Result) {
			fmt.Println(result)
		})
}
