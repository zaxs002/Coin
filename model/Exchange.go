package model

import (
	"net/http"
	"fmt"
	"bytes"
	"github.com/tidwall/gjson"
	"time"
	"sync"
	"strings"
	"strconv"
	"reflect"
	"net/url"
	"BitCoin/config"
	"encoding/json"
	"BitCoin/cache"
)

var IsServer = config.IsServer

type LockMap struct {
	sync.RWMutex
	M map[string]float64
}
type LockMapString struct {
	sync.RWMutex
	M map[string]string
}

func (lm *LockMap) Get(k string) interface{} {
	lm.RLock()
	if f, ok := lm.M[k]; ok {
		lm.RUnlock()
		return f
	} else {
		lm.RUnlock()
		return -1.0
	}
}

func (lm *LockMap) Set(k string, v float64) {
	lm.Lock()
	lm.M[k] = v
	lm.Unlock()
}
func (lm *LockMapString) Get(k string) interface{} {
	lm.RLock()
	if f, ok := lm.M[k]; ok {
		lm.RUnlock()
		return f
	} else {
		lm.RUnlock()
		return ""
	}
}

func (lm *LockMapString) Set(k string, v string) {
	lm.Lock()
	lm.M[k] = v
	lm.Unlock()
}

type BigE interface {
	CreateRun(symbol string)
	Run(symbol string)
	FeesRun()
	SetPrice(symbol string, num float64)
	GetLastPrice(symbol string) float64
	GetAmount(symbol string) float64
	SetAmount(symbol string, num float64)
	GetName() string
	CheckCoinExist(symbol string) bool
	GetTransferFee(coin string) float64
	GetTradeFee(symbol string, flag string) float64
	GetPrices() map[string]float64
}

type Exchange struct {
	PriceQueue   LockMap
	AmountDict   LockMap
	ValidSymbols []string
	Name         string
	TradeFees    LockMap
	TransferFees LockMapString
	Sub          interface{}

	TSDoOnce sync.Once
}

type TransferFee struct {
	WithdrawFee              interface{} //默认转账费
	MaxWithdrawFee           interface{} //最大转账费
	MinWithdrawFee           interface{} //最小转账费
	MinWithdraw              interface{} //单笔最小转账数量
	MaxWithdraw              interface{} //单笔最大转账数量
	MaxDayWithdraw           interface{} //单日最大转账数量
	CanWithdraw              interface{} //1可以 0不可以 -1未设置
	WithdrawMinConfirmations interface{} //最小确认
	CanDeposit               interface{} //1可以 0不可以 -1未设置
	MinDeposit               interface{} //最小存款数量
}

func NewTransferInfo() TransferFee {
	return TransferFee{
		-1,
		-1,
		-1,
		-1,
		-1,
		-1,
		-1,
		-1,
		-1,
		-1,
	}
}

func (e Exchange) CallMethod(method string, params []interface{}) reflect.Value {
	f := reflect.ValueOf(e.Sub).MethodByName(method)
	if f.IsValid() {
		args := make([]reflect.Value, len(params))
		for k, param := range params {
			args[k] = reflect.ValueOf(param)
		}
		ret := f.Call(args)
		if len(ret) > 0 {
			return ret[0]
		}
		return reflect.Value{}
	} else {
		fmt.Println("can't call " + method)
		return reflect.Value{}
	}
}

func (e Exchange) CreateRun(symbol string) {
	b := e.CallMethod("CheckCoinExist", []interface{}{symbol}).Bool()
	if b {
		//go e.Run(symbol)
		e.CallMethod("Run", []interface{}{symbol})
		//go e.CallMethod("FeesRun", []interface{}{})
	} else {
		e.SetPrice(symbol, -1)
	}
}

func (e Exchange) Run(symbol string) {
	fmt.Println("Old Run")
}

func (e Exchange) FeesRun() {
	fmt.Println("Old FeesRun")
}

func (e Exchange) check(flag string) {
	symbolLen := cache.GetInstance().HLen(e.Name + "-symbols").Val()
	priceLen := cache.GetInstance().HLen(e.Name).Val()
	currencyLen := cache.GetInstance().HLen(e.Name + "-currency").Val()
	transferLen := cache.GetInstance().HLen(e.Name + "-transfer").Val()

	if symbolLen == priceLen && symbolLen != 0 {
	}

	if currencyLen == transferLen && transferLen != 0 {
	}

	if symbolLen == priceLen && currencyLen == transferLen && transferLen != 0 {
		cache.GetInstance().HSet("Flag", e.Name, 1)
		println(e.Name + "全部获取完成")
	}
}

func (e Exchange) GetLastPrice(symbol string) float64 {
	r, _ := cache.GetInstance().HGet(e.Name, symbol).Float64()
	return r
}

func (e Exchange) GetPrices() map[string]float64 {
	return e.PriceQueue.M
}

func (e Exchange) SetPrice(symbol string, num float64) {
	cache.GetInstance().HSet(e.Name, symbol, num)
	//e.PriceQueue.Set(symbol, num)
}

func (e Exchange) SetSymbol(symbol string, symbol2 string) {
	cache.GetInstance().HSet(e.Name+"-symbols", symbol, symbol2)
}

func (e Exchange) SetCurrency(currency string) {
	currency = strings.ToLower(currency)
	cache.GetInstance().HSet(e.Name+"-currency", currency, currency)
}

func (e Exchange) SetCurrency2(currency string, currency2 string) {
	currency = strings.ToLower(currency)
	currency2 = strings.ToLower(currency2)
	cache.GetInstance().HSet(e.Name+"-currency", currency, currency2)
}

func (e Exchange) SetTransferFee(currency string, transfer TransferFee) {
	currency = strings.ToLower(currency)
	m := make(map[string]interface{})
	s := reflect.ValueOf(&transfer).Elem()
	typ := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		name := typ.Field(i).Name
		m[name] = f.Interface()
	}
	j, _ := json.Marshal(m)
	cache.GetInstance().HSet(e.Name+"-transfer", currency, j)
}

func (e Exchange) SetTradeFee(taker float64, maker float64) {
	cache.GetInstance().HSet(e.Name+"-tradeFee", "taker", taker)
	cache.GetInstance().HSet(e.Name+"-tradeFee", "maker", maker)
}

func (e Exchange) GetAmount(symbol string) float64 {
	f := e.AmountDict.Get(symbol).(float64)
	if f < 0 {
		return 0
	}
	return f
}

func (e Exchange) SetAmount(symbol string, num float64) {
	e.AmountDict.Set(symbol, num)
}

func (e Exchange) GetName() string {
	return e.Name
}

func (e Exchange) CheckCoinExist(symbol string) bool {
	fmt.Println("Old CheckCoinExist")
	return false
}

func (e Exchange) GetTransferFee(coin string) float64 {
	return e.TransferFees.Get(coin).(float64)
}

func (e Exchange) GetTradeFee(symbol string, flag string) float64 {
	return e.TradeFees.Get(symbol + "-" + flag).(float64)
}

func GetMinAmountExchange(symbol string) BigE {
	minAmount := 10000000.0
	var minAmountExchange BigE
	for _, v := range ExchangeList {
		amount := v.GetAmount(symbol[:3])
		if amount < minAmount {
			minAmount = amount
			minAmountExchange = v
		}
	}
	return minAmountExchange
}

func GetMaxAmountExchange(symbol string) BigE {
	maxAmount := -1.0
	var maxAmountExchange BigE
	for _, v := range ExchangeList {
		amount := v.GetAmount(symbol[:3])
		if amount > maxAmount {
			maxAmount = amount
			maxAmountExchange = v
		}
	}
	return maxAmountExchange
}

//fatbtc
type FatbtcExchange struct {
	Exchange
}

func (ue FatbtcExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (ue *FatbtcExchange) Run(symbol string) {
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

	oldSymbol := symbol
	containBtc := strings.Contains(symbol, "btc")
	coin := ""
	if containBtc {
		coins := strings.Split(symbol, "btc")
		if len(coins) < 2 {
			return
		}
		coin = coins[0]
	}
	symbol = coin + "btc"

	timestamp := time.Now().Unix()
	url := "https://www.fatbtc.com/m/allticker/" + strconv.Itoa(int(timestamp)) + "000"

	resp, _ := http.NewRequest("GET", url, nil)

	for {
		resp, err := client.Do(resp)

		if err != nil {
			fmt.Println(err)
			continue
		}

		buf := bytes.NewBuffer(make([]byte, 0, 512))

		buf.ReadFrom(resp.Body)
		resp.Body.Close()

		jsonNode := "data." + symbol + "_ticker" + ".close"
		result := gjson.GetBytes(buf.Bytes(), jsonNode)
		ue.SetPrice(oldSymbol, result.Float())

		time.Sleep(1 * time.Second)
	}
}

func (ue FatbtcExchange) FeesRun() {
	fmt.Println("Old FeesRun")
}

func NewFatbtcExchange() BigE {
	exchange := new(FatbtcExchange)
	exchange.Exchange = Exchange{
		Name: "Fatbtc",
		PriceQueue: LockMap{
			M: make(map[string]float64),
		},
		AmountDict: LockMap{
			M: make(map[string]float64),
		},
		Sub: exchange,
	}
	var duitai BigE = exchange
	return duitai
}

//allcoin
type AllcoinExchange struct {
	Exchange
}

func (ue AllcoinExchange) CheckCoinExist(symbol string) bool {
	return true
}

func (ue *AllcoinExchange) Run(symbol string) {
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

	oldSymbol := symbol
	containBtc := strings.Contains(symbol, "btc")
	coin := ""
	if containBtc {
		coins := strings.Split(symbol, "btc")
		if len(coins) < 2 {
			return
		}
		coin = coins[0]
	}
	symbol = coin + "_btc"

	url := "https://api.allcoin.com/api/v1/ticker" +
		"?symbol=" + symbol

	resp, _ := http.NewRequest("GET", url, nil)

	for {
		resp, err := client.Do(resp)

		if err != nil {
			fmt.Println(err)
			continue
		}

		buf := bytes.NewBuffer(make([]byte, 0, 512))

		buf.ReadFrom(resp.Body)
		resp.Body.Close()

		jsonNode := "ticker.last"
		result := gjson.GetBytes(buf.Bytes(), jsonNode)
		ue.SetPrice(oldSymbol, result.Float())

		time.Sleep(1 * time.Second)
	}
}

func (ue AllcoinExchange) FeesRun() {
	fmt.Println("Old FeesRun")
}

func NewAllcoinExchange() BigE {
	exchange := new(AllcoinExchange)
	exchange.Exchange = Exchange{
		Name: "Allcoin",
		PriceQueue: LockMap{
			M: make(map[string]float64),
		},
		AmountDict: LockMap{
			M: make(map[string]float64),
		},
		Sub: exchange,
	}
	var duitai BigE = exchange
	return duitai
}

//quoine
type QuoineExchange struct {
	Exchange
}

func (ue *QuoineExchange) CreateRun(symbol string) {
	go ue.Run(symbol)
}

func (ue *QuoineExchange) Run(symbol string) {
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

	buf := bytes.NewBuffer(make([]byte, 0, 512))
	preUrl := "https://api.quoine.com/products"

	preResp, _ := http.NewRequest("GET", preUrl, nil)
	for {
		resp, err := client.Do(preResp)

		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		break
	}

	id := gjson.GetBytes(buf.Bytes(), "#[currency_pair_code==\""+strings.ToUpper(symbol)+"\"].id")

	url := "https://api.quoine.com/products" + "/" + id.String()

	respTWo, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}

	for {
		resp, err := client.Do(respTWo)

		if err != nil {
			fmt.Println(err)
			continue
		}

		buf := bytes.NewBuffer(make([]byte, 0, 512))

		buf.ReadFrom(resp.Body)
		resp.Body.Close()

		jsonNode := "market_bid"
		result := gjson.GetBytes(buf.Bytes(), jsonNode)
		ue.SetPrice(symbol, result.Float())

		time.Sleep(1 * time.Second)
	}
}

func (ue QuoineExchange) FeesRun() {
	fmt.Println("Old FeesRun")
}

func NewQuoineExchange() BigE {
	exchange := QuoineExchange{
		Exchange{
			Name: "Quoine",
			PriceQueue: LockMap{
				M: make(map[string]float64),
			},
			AmountDict: LockMap{
				M: make(map[string]float64),
			},
		},
	}
	var duitai BigE = &exchange
	return duitai
}
