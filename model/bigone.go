package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"BitCoin/cache"
	"BitCoin/event"
	"net/url"
	"strings"
)

type BigOneExchange struct {
	Exchange
}

func (he BigOneExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he BigOneExchange) check(s string) {
	println("---------")
	symbolLen := cache.GetInstance().HLen(he.Name + "-symbols").Val()
	priceLen := cache.GetInstance().HLen(he.Name).Val()
	currencyLen := cache.GetInstance().HLen(he.Name + "-currency").Val()
	transferLen := cache.GetInstance().HLen(he.Name + "-transfer").Val()

	if symbolLen == priceLen && symbolLen != 0 {
		println(he.Name + " 价格全部获取完成")
	}

	if currencyLen == transferLen && transferLen != 0 {
		println(he.Name + " 转账手续费获取完成")
	}
	if symbolLen == priceLen && currencyLen == transferLen && transferLen != 0 {
		event.Bus.Unsubscribe(he.Name+":check", he.check)
	}
}

func (he *BigOneExchange) Run(symbol string) {
	he.SetTradeFee(0.001, 0.001)

	//获取价格
	u := "https://big.one/api/v2/tickers"
	utils.StartTimer(time.Second, func() {
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("data").ForEach(func(key, value gjson.Result) bool {
				symbol := value.Get("market_id").String()
				symbol = strings.ToLower(symbol)
				m := utils.GetSymbolByBigOne(symbol)
				coin := m["coin"]
				base := m["base"]
				symbol = coin + "-" + base
				last := value.Get("close").Float()

				he.SetPrice(symbol, last)
				he.SetSymbol(symbol, symbol)
				return true
			})
		})
	})

	//获取转账手续费
	transferUrl := "https://big.one/api/graphql"
	utils.StartTimer(time.Minute*30, func() {
		body := url.Values{
			"operationName": []string{},
			"variables":     []string{},
			"query": []string{
				`{
  markets {
    ...Market
  }
}

fragment Market on Market {
  baseAsset {
    ...Asset
  }
}

fragment Asset on Asset {
  symbol
  withdrawalFee
  isWithdrawalEnabled
}`,
			},
		}

		utils.GetInfoWithBody("POST", transferUrl, nil, body, func(result gjson.Result) {
			result.Get("data.markets").ForEach(func(key, value gjson.Result) bool {
				fee := value.Get("baseAsset.withdrawalFee").Float()
				currency := value.Get("baseAsset.symbol").String()
				isWithdrawalEnabled := value.Get("baseAsset.isWithdrawalEnabled").Bool()
				currency = strings.ToLower(currency)

				info := NewTransferInfo()
				canWithdraw := 0.0
				if isWithdrawalEnabled {
					canWithdraw = 1
				} else {
					canWithdraw = 0
				}
				info.CanWithdraw = canWithdraw
				info.WithdrawFee = fee
				he.SetTransferFee(currency, info)
				he.SetCurrency(currency)
				return true
			})
		})
	})
}

func (he BigOneExchange) FeesRun() {
}

func NewBigOneExchange() BigE {
	exchange := new(BigOneExchange)
	exchange.Exchange = Exchange{
		Name: "BigOne",
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
