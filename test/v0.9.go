package main

import (
	"BitCoin/utils"
	"fmt"
	"time"
	"net/http"
	"net/url"
	"bytes"
	"github.com/tidwall/gjson"
	"strconv"
)

func main() {
	utils.StartTimer(time.Second*1, func() {
		transferUrl := "https://www.binance.com/assetWithdraw/getAllAsset.html"

		var client = &http.Client{}
		if true {
			uProxy, _ := url.Parse("http://127.0.0.1:1080")

			client = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(uProxy),
				},
			}
		}

		client.Timeout = time.Second * 10

		resp, _ := http.NewRequest("GET", transferUrl, nil)

		for {
			resp, err := client.Do(resp)

			if err != nil {
				fmt.Println(err)
				continue
			}

			buf := bytes.NewBuffer(make([]byte, 0, 512))

			buf.ReadFrom(resp.Body)

			resp.Body.Close()

			body := buf.Bytes()
			result := gjson.GetBytes(body, "0.assetCode")
			result = gjson.GetBytes(body, "#.assetCode")
			array := result.Array()
			for k := range array {
				i := strconv.Itoa(k)
				name := gjson.GetBytes(body, i+".assetCode")
				fee := gjson.GetBytes(body, i+".transactionFee")
			}
			break
		}
	})

	for {
		time.Sleep(time.Second)
	}
}
