package model

import (
	"strconv"
	"fmt"
)

var U1 = User{
	Name:       "Test1",
	AmountDict: make(map[string]float64),
}

var U2 = User{
	Name:       "Test2",
	AmountDict: make(map[string]float64),
}


func init() {
	for i := 0; i < 10; i++ {
		user := User{
			Name:       "Test" + strconv.Itoa(i),
			AmountDict: make(map[string]float64),
		}
		UserList = append(UserList, user)
	}
}

type User struct {
	Name       string
	AmountDict map[string]float64
}

func (u *User) Buy(symbol string, num float64, tradeType string) {
	//默认限价
	//profit := 1.2

	prices := 0.0
	minPrice := 10000000.0
	maxPrice := 0.0
	var minPriceExchange BigE
	var maxPriceExchange BigE
	for _, v := range ExchangeList {
		price := v.GetLastPrice(symbol)
		prices += price

		if minPrice >= price {
			minPrice = price
			minPriceExchange = v
		}

		if maxPrice <= price {
			maxPrice = price
			maxPriceExchange = v
		}
	}
	maxPriceExchange.GetName()
	avgPrice := prices / float64(len(ExchangeList))
	fmt.Println(symbol, "平均价: ", avgPrice, " 最高价: ", maxPrice, " 最低价: ", minPrice)

	fmt.Println(symbol, "显示价格: ", avgPrice)
	fmt.Println("从", minPriceExchange.GetName(), "买入")
	fmt.Println(symbol, "利润: ", avgPrice-minPrice)
}

func (u *User) WithDraw(symbol string, num float64) {
	maxAmountExchange := GetMaxAmountExchange(symbol)
	s := symbol[:3]
	maxAmountExchange.SetAmount(s, maxAmountExchange.GetAmount(s)-num)
	u.SetAmount(s, u.GetAmount(s)-num)
}

func (u *User) Deposit(symbol string, num float64) {
	minAmountExchange := GetMinAmountExchange(symbol)
	s := symbol[:3]
	minAmountExchange.SetAmount(s, minAmountExchange.GetAmount(s)+num)
	u.SetAmount(s, u.GetAmount(s)+num)
}

func (u *User) GetAmount(symbol string) float64 {
	return u.AmountDict[symbol]
}

func (u *User) SetAmount(symbol string, num float64) {
	u.AmountDict[symbol] = num
}
