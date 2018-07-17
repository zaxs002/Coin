package main

import (
	"regexp"
	"BitCoin/utils"
	"fmt"
)

func main() {
	s := "market.swftcbtc.kline.1min"

	var myExp = utils.MyRegexp{regexp.MustCompile(`^market\.(?P<coin>(\w+)*)\.`)}
	m := myExp.FindStringSubmatchMap(s)
	if s, ok := m["coin"]; ok {
		coin := utils.GetCoinBySymbol(s)
		base := utils.GetBaseBySymbol(s)
		fmt.Println(coin+base)
	}

}