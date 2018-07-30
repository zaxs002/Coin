package model

import (
	"BitCoin/utils"
	"time"
	"github.com/tidwall/gjson"
	"BitCoin/cache"
	"BitCoin/event"
	"strings"
)

type CryptopiaExchange struct {
	Exchange
}

func (he CryptopiaExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (he CryptopiaExchange) check(s string) {
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

func (he *CryptopiaExchange) Run(symbol string) {
	he.SetTradeFee(0.002, 0.002)

	//获取价格和symbols
	u := "https://www.cryptopia.co.nz/api/GetMarkets"
	utils.StartTimer(time.Second, func() {
		utils.GetInfo("GET", u, nil, func(result gjson.Result) {
			result.Get("Data").ForEach(func(key, value gjson.Result) bool {
				symbol := value.Get("Label").String()
				symbol = strings.ToLower(symbol)
				m := utils.GetSymbolByCryptopia(symbol)
				coin := m["coin"]
				base := m["base"]
				symbol = coin + "-" + base
				last := value.Get("LastPrice").Float()

				he.SetPrice(symbol, last)
				he.SetSymbol(symbol, symbol)
				return true
			})
		})
	})

	//获取currency
	currencyUrl := "https://www.cryptopia.co.nz/api/GetCurrencies"
	utils.StartTimer(time.Hour*30, func() {
		utils.GetInfo("GET", currencyUrl, nil, func(result gjson.Result) {
			result.Get("Data").ForEach(func(key, value gjson.Result) bool {
				currency := value.Get("Symbol").String()
				fee := value.Get("WithdrawFee").Float()
				MinWithdraw := value.Get("MinWithdraw").Float()
				MaxWithdraw := value.Get("MaxWithdraw").Float()

				info := NewTransferInfo()
				info.WithdrawFee = fee
				info.MinWithdraw = MinWithdraw
				info.MaxWithdraw = MaxWithdraw
				he.SetTransferFee(currency, info)
				return true
			})
		})
	})

}

func (he CryptopiaExchange) FeesRun() {
}

func NewCryptopiaExchange() BigE {
	exchange := new(CryptopiaExchange)
	exchange.Exchange = Exchange{
		Name: "Cryptopia",
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
