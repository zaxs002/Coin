package main

import (
	"net/url"
	"net/http"
	"time"
	"fmt"
	"bytes"
	"github.com/tidwall/gjson"
)

var IsServer = false

func GetInfo(u string, header http.Header, callback func(result gjson.Result)) {
	maxRequestCount := 5
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

	req, _ := http.NewRequest("GET", u, nil)

	if header != nil {
		req.Header = header
	}

	for i := maxRequestCount; i > 0; i-- {
		resp, err := client.Do(req)

		if err != nil {
			fmt.Println(err)
			continue
		}

		buf := bytes.NewBuffer(make([]byte, 0, 512))

		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		result := gjson.ParseBytes(buf.Bytes())
		callback(result)

		break
	}
}

func main() {
	GetInfo("https://www.huobipro.com/-/x/pro/v1/settings/symbols?language=zh-cn", nil, func(result gjson.Result) {
		result.Get("data").ForEach(func(key, value gjson.Result) bool {
			base := value.Get("base-currency").String()
			quote := value.Get("quote-currency").String()
			symbol := base + quote
			fmt.Println(symbol)
			return true
		})
	})
}
